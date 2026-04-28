package segments

import (
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

// Claude segment displays Claude Code session information
type Claude struct {
	Base
	markedChar   string
	unmarkedChar string
	ClaudeData
}

// ClaudeData represents the parsed Claude JSON data
type ClaudeData struct {
	RateLimits    *ClaudeRateLimits   `json:"rate_limits"`
	Model         ClaudeModel         `json:"model"`
	Workspace     ClaudeWorkspace     `json:"workspace"`
	SessionID     string              `json:"session_id"`
	ContextWindow ClaudeContextWindow `json:"context_window"`
	Cost          ClaudeCost          `json:"cost"`
}

// ClaudeModel represents the AI model information
type ClaudeModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ClaudeWorkspace represents workspace directory information
type ClaudeWorkspace struct {
	CurrentDir  string `json:"current_dir"`
	ProjectDir  string `json:"project_dir"`
	GitWorktree string `json:"git_worktree"`
}

// DurationMS is a duration in milliseconds that formats as "Xm Ys".
type DurationMS int64

func (d DurationMS) String() string {
	totalSeconds := int64(d) / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// ClaudeCost represents cost and duration information
type ClaudeCost struct {
	TotalCostUSD       float64    `json:"total_cost_usd"`
	TotalDurationMS    DurationMS `json:"total_duration_ms"`
	TotalAPIDurationMS DurationMS `json:"total_api_duration_ms"`
	TotalLinesAdded    int        `json:"total_lines_added"`
	TotalLinesRemoved  int        `json:"total_lines_removed"`
}

// ClaudeRateLimitWindow represents a single rate limit time window.
type ClaudeRateLimitWindow struct {
	UsedPercentage *float64 `json:"used_percentage"`
	ResetsAt       *int64   `json:"resets_at"`
}

// ClaudeRateLimits represents rate limit information across time windows.
type ClaudeRateLimits struct {
	FiveHour *ClaudeRateLimitWindow `json:"five_hour"`
	SevenDay *ClaudeRateLimitWindow `json:"seven_day"`
}

// ClaudeContextWindow represents token usage information
type ClaudeContextWindow struct {
	UsedPercentage      *int                `json:"used_percentage"`
	RemainingPercentage *int                `json:"remaining_percentage"`
	CurrentUsage        *ClaudeCurrentUsage `json:"current_usage"`
	TotalInputTokens    int                 `json:"total_input_tokens"`
	TotalOutputTokens   int                 `json:"total_output_tokens"`
	ContextWindowSize   int                 `json:"context_window_size"`
}

// ClaudeCurrentUsage represents current context window usage from the last API call
type ClaudeCurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

const (
	thousand = 1000.0
	million  = 1000000.0

	gaugeMarkedChar   options.Option = "gauge_marked_char"
	gaugeUnmarkedChar options.Option = "gauge_unmarked_char"
)

func (c *Claude) Template() string {
	return " \U000f0bc9 {{ .Model.DisplayName }} \uf2d0 {{ .TokenGauge }} "
}

func (c *Claude) Enabled() bool {
	log.Debug("claude segment: checking if enabled")

	// Try to get Claude data from session cache
	claudeData, found := cache.Get[ClaudeData](cache.Session, cache.CLAUDECACHE)
	if !found {
		log.Debug("claude segment: no Claude data found in session cache")
		return false
	}

	log.Debug("claude segment: found Claude data in session cache")
	log.Debugf("claude segment: model=%s, session=%s", claudeData.Model.DisplayName, claudeData.SessionID)

	// Copy the data to our embedded struct
	c.ClaudeData = claudeData

	c.markedChar = c.options.String(gaugeMarkedChar, "▰")
	c.unmarkedChar = c.options.String(gaugeUnmarkedChar, "▱")

	return true
}

// TokenUsagePercent returns the percentage of context window used.
// Uses pre-calculated UsedPercentage when available (resets on compact/clear),
// falls back to calculating from CurrentUsage, then to total tokens for backwards compatibility.
func (c *Claude) TokenUsagePercent() text.Percentage {
	// Prefer pre-calculated UsedPercentage - most accurate and resets on compact/clear
	// When UsedPercentage is nil (null in JSON), context was reset - return 0
	if c.ContextWindow.UsedPercentage != nil {
		if *c.ContextWindow.UsedPercentage > 100 {
			return 100
		}
		return text.Percentage(*c.ContextWindow.UsedPercentage)
	}

	// UsedPercentage is nil - check if CurrentUsage is also nil (indicates reset/clear)
	if c.ContextWindow.CurrentUsage == nil {
		return 0
	}

	if c.ContextWindow.ContextWindowSize <= 0 {
		return 0
	}

	// Calculate from CurrentUsage (includes cache tokens for accurate context measurement)
	currentTokens := c.ContextWindow.CurrentUsage.InputTokens +
		c.ContextWindow.CurrentUsage.CacheCreationInputTokens +
		c.ContextWindow.CurrentUsage.CacheReadInputTokens

	// Fallback to total tokens if CurrentUsage is not provided (backwards compatibility)
	if currentTokens <= 0 {
		currentTokens = c.ContextWindow.TotalInputTokens + c.ContextWindow.TotalOutputTokens
	}

	if currentTokens <= 0 {
		return 0
	}

	// Use floating-point arithmetic for accurate percentage calculation
	percent := (float64(currentTokens) * 100.0) / float64(c.ContextWindow.ContextWindowSize)

	// Round to nearest integer and cap at 100
	roundedPercent := int(percent + 0.5)
	if roundedPercent > 100 {
		return 100
	}

	return text.Percentage(roundedPercent)
}

// TokenGauge returns a 5-block gauge showing remaining context window capacity using the configured characters.
func (c *Claude) TokenGauge() string {
	return c.TokenUsagePercent().GaugeWith(c.markedChar, c.unmarkedChar)
}

// TokenGaugeUsed returns a 5-block gauge showing used context window capacity using the configured characters.
func (c *Claude) TokenGaugeUsed() string {
	return c.TokenUsagePercent().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

// FiveHourGauge returns a 5-block gauge showing 5-hour rate limit usage using the configured characters.
func (c *Claude) FiveHourGauge() string {
	return c.FiveHourUsage().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

// SevenDayGauge returns a 5-block gauge showing 7-day rate limit usage using the configured characters.
func (c *Claude) SevenDayGauge() string {
	return c.SevenDayUsage().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

// FormattedCost returns the cost formatted as a currency string
func (c *Claude) FormattedCost() string {
	if c.Cost.TotalCostUSD < 0.01 {
		return fmt.Sprintf("$%.4f", c.Cost.TotalCostUSD)
	}

	return fmt.Sprintf("$%.2f", c.Cost.TotalCostUSD)
}

// FormattedDuration returns total session duration as "Xm Ys".
func (c *Claude) FormattedDuration() string {
	return c.Cost.TotalDurationMS.String()
}

// FormattedAPIDuration returns API wait time as "Xm Ys".
func (c *Claude) FormattedAPIDuration() string {
	return c.Cost.TotalAPIDurationMS.String()
}

// rateLimitPercentage extracts a percentage from a rate limit window with nil-safety.
func rateLimitPercentage(limits *ClaudeRateLimits, window func(*ClaudeRateLimits) *ClaudeRateLimitWindow) text.Percentage {
	if limits == nil {
		return 0
	}

	w := window(limits)
	if w == nil || w.UsedPercentage == nil {
		return 0
	}

	percent := int(*w.UsedPercentage + 0.5)
	if percent > 100 {
		return 100
	}

	return text.Percentage(percent)
}

// FiveHourUsage returns the 5-hour rolling window rate limit usage as a Percentage.
func (c *Claude) FiveHourUsage() text.Percentage {
	return rateLimitPercentage(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.FiveHour
	})
}

// SevenDayUsage returns the 7-day window rate limit usage as a Percentage.
func (c *Claude) SevenDayUsage() text.Percentage {
	return rateLimitPercentage(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.SevenDay
	})
}

// rateLimitResetsAt extracts the reset time from a rate limit window with nil-safety.
func rateLimitResetsAt(limits *ClaudeRateLimits, window func(*ClaudeRateLimits) *ClaudeRateLimitWindow) time.Time {
	if limits == nil {
		return time.Time{}
	}

	w := window(limits)
	if w == nil || w.ResetsAt == nil {
		return time.Time{}
	}

	return time.Unix(*w.ResetsAt, 0)
}

// FiveHourResetsAt returns the reset time for the 5-hour rolling window, or zero if unavailable.
func (c *Claude) FiveHourResetsAt() time.Time {
	return rateLimitResetsAt(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.FiveHour
	})
}

// SevenDayResetsAt returns the reset time for the 7-day rolling window, or zero if unavailable.
func (c *Claude) SevenDayResetsAt() time.Time {
	return rateLimitResetsAt(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.SevenDay
	})
}

// rateLimitResetsIn returns the duration until a rate limit window resets.
func rateLimitResetsIn(limits *ClaudeRateLimits, window func(*ClaudeRateLimits) *ClaudeRateLimitWindow) time.Duration {
	t := rateLimitResetsAt(limits, window)
	if t.IsZero() {
		return 0
	}

	return time.Until(t)
}

// FiveHourResetsIn returns the duration until the 5-hour rolling window resets, or 0 if unavailable.
func (c *Claude) FiveHourResetsIn() time.Duration {
	return rateLimitResetsIn(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.FiveHour
	})
}

// SevenDayResetsIn returns the duration until the 7-day rolling window resets, or 0 if unavailable.
func (c *Claude) SevenDayResetsIn() time.Duration {
	return rateLimitResetsIn(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.SevenDay
	})
}

// FormattedTokens returns a human-readable string of current context tokens.
// Uses CurrentUsage (which represents actual context and resets on compact/clear)
// with fallback to total tokens for backwards compatibility.
func (c *Claude) FormattedTokens() string {
	var currentTokens int

	// Use CurrentUsage for display - includes cache tokens for accurate context measurement
	// When CurrentUsage is nil (context reset), fall back to total tokens
	if c.ContextWindow.CurrentUsage != nil {
		currentTokens = c.ContextWindow.CurrentUsage.InputTokens +
			c.ContextWindow.CurrentUsage.CacheCreationInputTokens +
			c.ContextWindow.CurrentUsage.CacheReadInputTokens
	}

	// Fallback to total tokens if CurrentUsage is not provided (backwards compatibility)
	if currentTokens <= 0 {
		currentTokens = c.ContextWindow.TotalInputTokens + c.ContextWindow.TotalOutputTokens
	}

	if currentTokens < int(thousand) {
		return fmt.Sprintf("%d", currentTokens)
	}

	if currentTokens < int(million) {
		return fmt.Sprintf("%.1fK", float64(currentTokens)/thousand)
	}

	return fmt.Sprintf("%.1fM", float64(currentTokens)/million)
}
