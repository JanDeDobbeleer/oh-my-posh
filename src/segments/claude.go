package segments

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

type Claude struct {
	Base
	markedChar   string
	unmarkedChar string
	ClaudeData
}

type ClaudeData struct {
	Effort            *ClaudeEffort       `json:"effort"`
	Worktree          *ClaudeWorktree     `json:"worktree"`
	OutputStyle       *ClaudeOutputStyle  `json:"output_style"`
	Vim               *ClaudeVim          `json:"vim"`
	RateLimits        *ClaudeRateLimits   `json:"rate_limits"`
	Thinking          *ClaudeThinking     `json:"thinking"`
	PR                *ClaudePR           `json:"pr"`
	Agent             *ClaudeAgent        `json:"agent"`
	PromptCache       *ClaudePromptCache  `json:"prompt_cache"`
	Model             AIModel             `json:"model"`
	TranscriptPath    string              `json:"transcript_path"`
	PromptID          string              `json:"prompt_id"`
	SessionName       string              `json:"session_name"`
	SessionID         string              `json:"session_id"`
	Version           string              `json:"version"`
	CWD               string              `json:"cwd"`
	Workspace         ClaudeWorkspace     `json:"workspace"`
	ContextWindow     ClaudeContextWindow `json:"context_window"`
	Cost              ClaudeCost          `json:"cost"`
	Exceeds200KTokens bool                `json:"exceeds_200k_tokens"`
	FastMode          bool                `json:"fast_mode"`
}

// Shared across AI CLI segments (e.g. copilot_cli).
type AIModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type ClaudeWorkspace struct {
	Repo        *ClaudeRepo `json:"repo"`
	CurrentDir  string      `json:"current_dir"`
	ProjectDir  string      `json:"project_dir"`
	GitWorktree string      `json:"git_worktree"`
	AddedDirs   []string    `json:"added_dirs"`
}

// Parsed from the origin remote.
type ClaudeRepo struct {
	Host  string `json:"host"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

// Nil when the statusline payload does not report an output style.
type ClaudeOutputStyle struct {
	Name string `json:"name"`
}

// Nil when the active model does not support reasoning effort.
type ClaudeEffort struct {
	Level string `json:"level"`
}

// Nil when the statusline payload does not report thinking state.
type ClaudeThinking struct {
	Enabled bool `json:"enabled"`
}

// Nil when vim mode is disabled.
type ClaudeVim struct {
	Mode string `json:"mode"`
}

// Nil when no agent is active.
type ClaudeAgent struct {
	Name string `json:"name"`
}

type ClaudePR struct {
	Number      json.Number `json:"number"`
	URL         string      `json:"url"`
	ReviewState string      `json:"review_state"`
}

// Nil when the session is not running inside a Claude Code worktree.
type ClaudeWorktree struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	OriginalCWD    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"`
}

// Formats as "Xm Ys" (see String()).
type DurationMS int64

func (d DurationMS) String() string {
	totalSeconds := int64(d) / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

type ClaudeCost struct {
	TotalCostUSD       float64    `json:"total_cost_usd"`
	TotalDurationMS    DurationMS `json:"total_duration_ms"`
	TotalAPIDurationMS DurationMS `json:"total_api_duration_ms"`
	TotalLinesAdded    int        `json:"total_lines_added"`
	TotalLinesRemoved  int        `json:"total_lines_removed"`
}

type ClaudeRateLimitWindow struct {
	UsedPercentage *float64 `json:"used_percentage"`
	ResetsAt       *int64   `json:"resets_at"`
}

type ClaudeRateLimits struct {
	FiveHour *ClaudeRateLimitWindow `json:"five_hour"`
	SevenDay *ClaudeRateLimitWindow `json:"seven_day"`
}

// Nil until the first API response of the session (Claude Code v2.1.251+).
type ClaudePromptCache struct {
	ExpiresAt           *int64               `json:"expires_at"`
	HitRatio            *float64             `json:"hit_ratio"`
	LastMissAt          *int64               `json:"last_miss_at"`
	RecacheTokensIfCold *int                 `json:"recache_tokens_if_cold"`
	LastMissCause       *ClaudeLastMissCause `json:"last_miss_cause"`
	MissCauses          map[string]int       `json:"miss_causes"`
	TTL                 string               `json:"ttl"`
	Requests            int                  `json:"requests"`
	Misses              int                  `json:"misses"`
	ExpectedRebuilds    int                  `json:"expected_rebuilds"`
	CacheWriteTokens    int                  `json:"cache_write_tokens"`
	MissRecacheTokens   int                  `json:"miss_recache_tokens"`
	Warm                bool                 `json:"warm"`
	CachingObserved     bool                 `json:"caching_observed"`
}

// Nil until the first miss or when no cause was identified (Claude Code v2.1.260+).
type ClaudeLastMissCause struct {
	Causes          []string `json:"causes"`
	ToolsAdded      int      `json:"tools_added"`
	ToolsRemoved    int      `json:"tools_removed"`
	SystemCharDelta int      `json:"system_char_delta"`
}

type ClaudeContextWindow struct {
	UsedPercentage      *int                `json:"used_percentage"`
	RemainingPercentage *int                `json:"remaining_percentage"`
	CurrentUsage        *ClaudeCurrentUsage `json:"current_usage"`
	TotalInputTokens    int                 `json:"total_input_tokens"`
	TotalOutputTokens   int                 `json:"total_output_tokens"`
	ContextWindowSize   int                 `json:"context_window_size"`
}

// Reflects the last API call, not a cumulative total.
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

func formatTokenCount(n int) string {
	if n < int(thousand) {
		return fmt.Sprintf("%d", n)
	}

	if n < int(million) {
		return fmt.Sprintf("%.1fK", float64(n)/thousand)
	}

	return fmt.Sprintf("%.1fM", float64(n)/million)
}

func (c *Claude) Template() string {
	return " \U000f0bc9 {{ .Model.DisplayName }} \uf2d0 {{ .TokenGauge }} "
}

func (c *Claude) Enabled() bool {
	log.Debug("claude segment: checking if enabled")

	// Try to get Claude data from session cache
	claudeData, found := cache.Session.Get[ClaudeData](cache.CLAUDECACHE)
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

// Shows remaining capacity; see TokenGaugeUsed for the used view.
func (c *Claude) TokenGauge() string {
	return c.TokenUsagePercent().GaugeWith(c.markedChar, c.unmarkedChar)
}

// Shows used capacity, unlike TokenGauge which shows remaining.
func (c *Claude) TokenGaugeUsed() string {
	return c.TokenUsagePercent().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

func (c *Claude) FiveHourGauge() string {
	return c.FiveHourUsage().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

func (c *Claude) SevenDayGauge() string {
	return c.SevenDayUsage().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

func (c *Claude) FormattedCost() string {
	if c.Cost.TotalCostUSD < 0.01 {
		return fmt.Sprintf("$%.4f", c.Cost.TotalCostUSD)
	}

	return fmt.Sprintf("$%.2f", c.Cost.TotalCostUSD)
}

func (c *Claude) FormattedDuration() string {
	return c.Cost.TotalDurationMS.String()
}

func (c *Claude) FormattedAPIDuration() string {
	return c.Cost.TotalAPIDurationMS.String()
}

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

func (c *Claude) FiveHourUsage() text.Percentage {
	return rateLimitPercentage(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.FiveHour
	})
}

func (c *Claude) SevenDayUsage() text.Percentage {
	return rateLimitPercentage(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.SevenDay
	})
}

func rateLimitResetsAt(limits *ClaudeRateLimits, window func(*ClaudeRateLimits) *ClaudeRateLimitWindow) time.Time {
	if limits == nil || window == nil {
		return time.Time{}
	}

	w := window(limits)
	if w == nil || w.ResetsAt == nil {
		return time.Time{}
	}

	return time.Unix(*w.ResetsAt, 0)
}

func (c *Claude) FiveHourResetsAt() time.Time {
	return rateLimitResetsAt(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.FiveHour
	})
}

func (c *Claude) SevenDayResetsAt() time.Time {
	return rateLimitResetsAt(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.SevenDay
	})
}

// Returns 0 when data is unavailable, negative when the window already reset, positive otherwise.
func rateLimitResetsIn(limits *ClaudeRateLimits, window func(*ClaudeRateLimits) *ClaudeRateLimitWindow) time.Duration {
	t := rateLimitResetsAt(limits, window)
	if t.IsZero() {
		return 0
	}

	return time.Until(t)
}

func (c *Claude) FiveHourResetsIn() time.Duration {
	return rateLimitResetsIn(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.FiveHour
	})
}

func (c *Claude) SevenDayResetsIn() time.Duration {
	return rateLimitResetsIn(c.RateLimits, func(r *ClaudeRateLimits) *ClaudeRateLimitWindow {
		return r.SevenDay
	})
}

// Uses CurrentUsage (actual context, resets on compact/clear), falling back to total tokens for backwards compatibility.
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

	return formatTokenCount(currentTokens)
}

// Returns 0 when the hit ratio is unavailable.
func (c *Claude) PromptCacheHitRatio() text.Percentage {
	if c.PromptCache == nil || c.PromptCache.HitRatio == nil {
		return 0
	}

	percent := int(*c.PromptCache.HitRatio*100 + 0.5)
	if percent < 0 {
		return 0
	}

	if percent > 100 {
		return 100
	}

	return text.Percentage(percent)
}

func (c *Claude) PromptCacheGauge() string {
	return c.PromptCacheHitRatio().GaugeUsedWith(c.markedChar, c.unmarkedChar)
}

// Returns the zero time when the cache has no cached prefix.
func (c *Claude) PromptCacheExpiresAt() time.Time {
	if c.PromptCache == nil || c.PromptCache.ExpiresAt == nil {
		return time.Time{}
	}

	return time.Unix(*c.PromptCache.ExpiresAt, 0)
}

// Returns 0 when unavailable, negative when the cache already expired.
func (c *Claude) PromptCacheExpiresIn() time.Duration {
	t := c.PromptCacheExpiresAt()
	if t.IsZero() {
		return 0
	}

	return time.Until(t)
}

func (c *Claude) FormattedCacheWriteTokens() string {
	if c.PromptCache == nil {
		return formatTokenCount(0)
	}

	return formatTokenCount(c.PromptCache.CacheWriteTokens)
}

func (c *Claude) FormattedMissRecacheTokens() string {
	if c.PromptCache == nil {
		return formatTokenCount(0)
	}

	return formatTokenCount(c.PromptCache.MissRecacheTokens)
}
