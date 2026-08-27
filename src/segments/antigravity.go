package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

type Antigravity struct {
	Base
	markedChar   string
	unmarkedChar string
	AntigravityData
}

type AntigravityData struct {
	Quota                   map[string]AntigravityQuota `json:"quota"`
	VCS                     *AntigravityVCS             `json:"vcs"`
	Sandbox                 *AntigravitySandbox         `json:"sandbox"`
	Vim                     *ClaudeVim                  `json:"vim"`
	Model                   AIModel                     `json:"model"`
	Workspace               AntigravityWorkspace        `json:"workspace"`
	TranscriptPath          string                      `json:"transcript_path"`
	CWD                     string                      `json:"cwd"`
	SessionID               string                      `json:"session_id"`
	ConversationID          string                      `json:"conversation_id"`
	Version                 string                      `json:"version"`
	Product                 string                      `json:"product"`
	AgentState              string                      `json:"agent_state"`
	PlanTier                string                      `json:"plan_tier"`
	Email                   string                      `json:"email"`
	ExecutionMode           string                      `json:"execution_mode"`
	ContextWindow           AntigravityContextWindow    `json:"context_window"`
	ArtifactCount           int                         `json:"artifact_count"`
	PendingInputCount       int                         `json:"pending_input_count"`
	TaskCount               int                         `json:"task_count"`
	TerminalWidth           int                         `json:"terminal_width"`
	Exceeds200KTokens       bool                        `json:"exceeds_200k_tokens"`
	ToolConfirmationPending bool                        `json:"tool_confirmation_pending"`
}

type AntigravityWorkspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type AntigravityContextWindow struct {
	CurrentUsage        *AntigravityCurrentUsage `json:"current_usage"`
	UsedPercentage      *float64                 `json:"used_percentage"`
	RemainingPercentage *float64                 `json:"remaining_percentage"`
	TotalInputTokens    int                      `json:"total_input_tokens"`
	TotalOutputTokens   int                      `json:"total_output_tokens"`
	ContextWindowSize   int                      `json:"context_window_size"`
}

// Reflects the last API call, not a cumulative total.
type AntigravityCurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// Keyed by a model or quota bucket id (e.g. "gemini-weekly"), not necessarily a model id.
type AntigravityQuota struct {
	RemainingFraction *float64 `json:"remaining_fraction"`
	ResetInSeconds    *int     `json:"reset_in_seconds"`
	ResetTime         string   `json:"reset_time"`
}

// Nil when the session is not backed by a VCS checkout.
type AntigravityVCS struct {
	Type   string `json:"type"`
	Branch string `json:"branch"`
	Dirty  bool   `json:"dirty"`
}

// Nil when sandboxing is not in use.
type AntigravitySandbox struct {
	AllowNetwork *bool `json:"allow_network"`
	Enabled      bool  `json:"enabled"`
}

func (a *Antigravity) Template() string {
	return " \uf135 {{ .Model.DisplayName }} \uf2d0 {{ .TokenGauge }} "
}

func (a *Antigravity) Enabled() bool {
	log.Debug("antigravity segment: checking if enabled")

	data, found := cache.Get[AntigravityData](cache.Session, cache.ANTIGRAVITYCACHE)
	if !found {
		log.Debug("antigravity segment: no data found in session cache")
		return false
	}

	log.Debug("antigravity segment: found data in session cache")
	log.Debugf("antigravity segment: model=%s, session=%s", data.Model.DisplayName, data.SessionID)

	a.AntigravityData = data

	a.markedChar = a.options.String(gaugeMarkedChar, "▰")
	a.unmarkedChar = a.options.String(gaugeUnmarkedChar, "▱")

	return true
}

// Uses pre-calculated UsedPercentage when available, falls back to calculating from
// CurrentUsage, then to total tokens for backwards compatibility.
func (a *Antigravity) TokenUsagePercent() text.Percentage {
	if a.ContextWindow.UsedPercentage != nil {
		v := *a.ContextWindow.UsedPercentage
		if v > 100 {
			return 100
		}

		if v < 0 {
			return 0
		}

		return text.Percentage(int(v + 0.5))
	}

	if a.ContextWindow.ContextWindowSize <= 0 {
		return 0
	}

	var currentTokens int
	if a.ContextWindow.CurrentUsage != nil {
		currentTokens = a.ContextWindow.CurrentUsage.InputTokens +
			a.ContextWindow.CurrentUsage.CacheCreationInputTokens +
			a.ContextWindow.CurrentUsage.CacheReadInputTokens
	}

	if currentTokens <= 0 {
		currentTokens = a.ContextWindow.TotalInputTokens + a.ContextWindow.TotalOutputTokens
	}

	if currentTokens <= 0 {
		return 0
	}

	percent := (float64(currentTokens) * 100.0) / float64(a.ContextWindow.ContextWindowSize)

	rounded := int(percent + 0.5)
	if rounded > 100 {
		return 100
	}

	return text.Percentage(rounded)
}

// Shows remaining capacity; see TokenGaugeUsed for the used view.
func (a *Antigravity) TokenGauge() string {
	return a.TokenUsagePercent().GaugeWith(a.markedChar, a.unmarkedChar)
}

// Shows used capacity, unlike TokenGauge which shows remaining.
func (a *Antigravity) TokenGaugeUsed() string {
	return a.TokenUsagePercent().GaugeUsedWith(a.markedChar, a.unmarkedChar)
}

// Uses CurrentUsage (actual context, resets on compact/clear), falling back to total tokens.
func (a *Antigravity) FormattedTokens() string {
	var currentTokens int

	if a.ContextWindow.CurrentUsage != nil {
		currentTokens = a.ContextWindow.CurrentUsage.InputTokens +
			a.ContextWindow.CurrentUsage.CacheCreationInputTokens +
			a.ContextWindow.CurrentUsage.CacheReadInputTokens
	}

	if currentTokens <= 0 {
		currentTokens = a.ContextWindow.TotalInputTokens + a.ContextWindow.TotalOutputTokens
	}

	return formatTokenCount(currentTokens)
}

func (a *Antigravity) RemainingPercent() text.Percentage {
	if a.ContextWindow.RemainingPercentage != nil {
		v := *a.ContextWindow.RemainingPercentage
		if v > 100 {
			return 100
		}

		if v < 0 {
			return 0
		}

		return text.Percentage(int(v + 0.5))
	}

	remaining := 100 - int(a.TokenUsagePercent())
	if remaining < 0 {
		return 0
	}

	return text.Percentage(remaining)
}
