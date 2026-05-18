package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

// CopilotCLI segment displays GitHub Copilot CLI session information
type CopilotCLI struct {
	Base
	markedChar   string
	unmarkedChar string
	CopilotCLIData
}

// CopilotCLIData represents the parsed Copilot CLI JSON data
type CopilotCLIData struct {
	Model          AIModel                 `json:"model"`
	Workspace      CopilotCLIWorkspace     `json:"workspace"`
	TranscriptPath string                  `json:"transcript_path"`
	CWD            string                  `json:"cwd"`
	Version        string                  `json:"version"`
	SessionID      string                  `json:"session_id"`
	SessionName    string                  `json:"session_name"`
	Username       string                  `json:"username"`
	ContextWindow  CopilotCLIContextWindow `json:"context_window"`
	Cost           CopilotCLICost          `json:"cost"`
	Remote         CopilotCLIRemote        `json:"remote"`
}

// CopilotCLIWorkspace represents workspace directory information
type CopilotCLIWorkspace struct {
	CurrentDir string `json:"current_dir"`
}

// CopilotCLIRemote represents remote connection state
type CopilotCLIRemote struct {
	Connected bool `json:"connected"`
}

// CopilotCLICost represents cost and duration information
type CopilotCLICost struct {
	TotalDurationMS      DurationMS `json:"total_duration_ms"`
	TotalAPIDurationMS   DurationMS `json:"total_api_duration_ms"`
	TotalLinesAdded      int        `json:"total_lines_added"`
	TotalLinesRemoved    int        `json:"total_lines_removed"`
	TotalPremiumRequests int        `json:"total_premium_requests"`
}

// CopilotCLIContextWindow represents token usage information
type CopilotCLIContextWindow struct {
	ContextWindowSize     *int     `json:"context_window_size"`
	UsedPercentage        *float64 `json:"used_percentage"`
	RemainingPercentage   *float64 `json:"remaining_percentage"`
	RemainingTokens       *int     `json:"remaining_tokens"`
	TotalInputTokens      int      `json:"total_input_tokens"`
	TotalOutputTokens     int      `json:"total_output_tokens"`
	TotalCacheReadTokens  int      `json:"total_cache_read_tokens"`
	TotalCacheWriteTokens int      `json:"total_cache_write_tokens"`
	TotalReasoningTokens  int      `json:"total_reasoning_tokens"`
	TotalTokens           int      `json:"total_tokens"`
	LastCallInputTokens   int      `json:"last_call_input_tokens"`
	LastCallOutputTokens  int      `json:"last_call_output_tokens"`
	CurrentContextTokens  int      `json:"current_context_tokens"`
}

func (c *CopilotCLI) Template() string {
	return " \uec1e {{ .Model.DisplayName }} \uf2d0 {{ .TokenGauge }} "
}

func (c *CopilotCLI) Enabled() bool {
	log.Debug("copilot_cli segment: checking if enabled")

	data, found := cache.Get[CopilotCLIData](cache.Session, cache.COPILOTCLICACHE)
	if !found {
		log.Debug("copilot_cli segment: no data found in session cache")
		return false
	}

	log.Debug("copilot_cli segment: found data in session cache")
	log.Debugf("copilot_cli segment: model=%s, session=%s", data.Model.DisplayName, data.SessionID)

	c.CopilotCLIData = data

	c.markedChar = c.options.String(gaugeMarkedChar, "▰")
	c.unmarkedChar = c.options.String(gaugeUnmarkedChar, "▱")

	return true
}

// TokenUsagePercent returns the percentage of context window used.
// Uses pre-calculated UsedPercentage when available; falls back to computing
// from CurrentContextTokens / ContextWindowSize; returns 0 when unavailable.
func (c *CopilotCLI) TokenUsagePercent() text.Percentage {
	if c.ContextWindow.UsedPercentage != nil {
		v := *c.ContextWindow.UsedPercentage
		if v > 100 {
			return 100
		}

		if v < 0 {
			return 0
		}

		return text.Percentage(int(v + 0.5))
	}

	if c.ContextWindow.ContextWindowSize == nil || *c.ContextWindow.ContextWindowSize <= 0 {
		return 0
	}

	tokens := c.ContextWindow.CurrentContextTokens
	if tokens <= 0 {
		tokens = c.ContextWindow.TotalTokens
	}

	if tokens <= 0 {
		return 0
	}

	percent := (float64(tokens) * 100.0) / float64(*c.ContextWindow.ContextWindowSize)

	rounded := int(percent + 0.5)
	if rounded > 100 {
		return 100
	}

	return text.Percentage(rounded)
}

// TokenGauge returns a 5-block gauge showing remaining context window capacity using the configured characters.
func (c *CopilotCLI) TokenGauge() string {
	return c.TokenUsagePercent().GaugeWith(c.markedChar, c.unmarkedChar)
}

// TokenGaugeUsed returns a 5-block gauge showing used context window capacity using the configured characters.
func (c *CopilotCLI) TokenGaugeUsed() string {
	return c.TokenUsagePercent().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

// FormattedTokens returns a human-readable string of current context tokens.
func (c *CopilotCLI) FormattedTokens() string {
	tokens := c.ContextWindow.CurrentContextTokens
	if tokens <= 0 {
		tokens = c.ContextWindow.TotalTokens
	}

	return formatTokenCount(tokens)
}

// FormattedDuration returns total session duration as "Xm Ys".
func (c *CopilotCLI) FormattedDuration() string {
	return c.Cost.TotalDurationMS.String()
}

// FormattedAPIDuration returns API wait time as "Xm Ys".
func (c *CopilotCLI) FormattedAPIDuration() string {
	return c.Cost.TotalAPIDurationMS.String()
}

// RemainingPercent returns the percentage of context window remaining (0-100).
func (c *CopilotCLI) RemainingPercent() text.Percentage {
	if c.ContextWindow.RemainingPercentage != nil {
		v := *c.ContextWindow.RemainingPercentage
		if v > 100 {
			return 100
		}

		if v < 0 {
			return 0
		}

		return text.Percentage(int(v + 0.5))
	}

	used := int(c.TokenUsagePercent())
	remaining := 100 - used
	if remaining < 0 {
		return 0
	}

	return text.Percentage(remaining)
}

// RemainingTokensCount returns the number of remaining context window tokens.
func (c *CopilotCLI) RemainingTokensCount() int {
	if c.ContextWindow.RemainingTokens != nil {
		return *c.ContextWindow.RemainingTokens
	}

	if c.ContextWindow.ContextWindowSize == nil || *c.ContextWindow.ContextWindowSize <= 0 {
		return 0
	}

	tokens := c.ContextWindow.CurrentContextTokens
	if tokens <= 0 {
		tokens = c.ContextWindow.TotalTokens
	}

	remaining := *c.ContextWindow.ContextWindowSize - tokens
	if remaining < 0 {
		return 0
	}

	return remaining
}
