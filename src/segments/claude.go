package segments

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

// Claude segment displays Claude Code session information
type Claude struct {
	Base
	ClaudeData
}

// ClaudeData represents the parsed Claude JSON data
type ClaudeData struct {
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
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// ClaudeCost represents cost and duration information
type ClaudeCost struct {
	TotalCostUSD    float64 `json:"total_cost_usd"`
	TotalDurationMS int64   `json:"total_duration_ms"`
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
)

func (c *Claude) Template() string {
	return " \U000f0bc9 {{ .Model.DisplayName }} \uf2d0 {{ .TokenUsagePercent.Gauge }} "
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

	return true
}

// TokenUsagePercent returns the percentage of context window used.
// Uses pre-calculated UsedPercentage when available (resets on compact/clear),
// falls back to calculating from CurrentUsage.
func (c *Claude) TokenUsagePercent() text.Percentage {
	// Prefer pre-calculated UsedPercentage - most accurate and resets on compact/clear
	if c.ContextWindow.UsedPercentage != nil {
		if *c.ContextWindow.UsedPercentage > 100 {
			return 100
		}
		return text.Percentage(*c.ContextWindow.UsedPercentage)
	}

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

// FormattedCost returns the cost formatted as a currency string
func (c *Claude) FormattedCost() string {
	if c.Cost.TotalCostUSD < 0.01 {
		return fmt.Sprintf("$%.4f", c.Cost.TotalCostUSD)
	}

	return fmt.Sprintf("$%.2f", c.Cost.TotalCostUSD)
}

// FormattedTokens returns a human-readable string of current context tokens.
// Uses CurrentUsage (which represents actual context and resets on compact/clear).
func (c *Claude) FormattedTokens() string {
	var currentTokens int

	if c.ContextWindow.CurrentUsage != nil {
		currentTokens = c.ContextWindow.CurrentUsage.InputTokens +
			c.ContextWindow.CurrentUsage.CacheCreationInputTokens +
			c.ContextWindow.CurrentUsage.CacheReadInputTokens
	}

	if currentTokens < int(thousand) {
		return fmt.Sprintf("%d", currentTokens)
	}

	if currentTokens < int(million) {
		return fmt.Sprintf("%.1fK", float64(currentTokens)/thousand)
	}

	return fmt.Sprintf("%.1fM", float64(currentTokens)/million)
}
