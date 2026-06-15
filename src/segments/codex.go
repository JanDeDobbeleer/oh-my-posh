package segments

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

// Codex segment displays OpenAI Codex session information.
type Codex struct {
	Base
	CodexData
}

// CodexData represents Codex statusline data.
type CodexData struct {
	Type            string           `json:"type"`
	ThreadID        string           `json:"thread_id"`
	CWD             string           `json:"cwd"`
	Version         string           `json:"version"`
	ReasoningEffort string           `json:"reasoning_effort"`
	ApprovalMode    string           `json:"approval_mode"`
	SandboxPolicy   string           `json:"sandbox_policy"`
	Model           CodexModel       `json:"model"`
	Info            CodexTokenInfo   `json:"info"`
	RateLimits      *CodexRateLimits `json:"rate_limits"`
}

// CodexModel represents OpenAI Codex model information.
type CodexModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// CodexTokenInfo represents token count information emitted by Codex.
type CodexTokenInfo struct {
	TotalTokenUsage    CodexTokenUsage `json:"total_token_usage"`
	LastTokenUsage     CodexTokenUsage `json:"last_token_usage"`
	ModelContextWindow int             `json:"model_context_window"`
}

// CodexTokenUsage represents a token usage snapshot.
type CodexTokenUsage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
	TotalTokens           int `json:"total_tokens"`
}

// CodexRateLimits represents Codex rate limit information.
type CodexRateLimits struct {
	LimitID              string                `json:"limit_id"`
	LimitName            string                `json:"limit_name"`
	Primary              *CodexRateLimitWindow `json:"primary"`
	Secondary            *CodexRateLimitWindow `json:"secondary"`
	PlanType             string                `json:"plan_type"`
	RateLimitReachedType string                `json:"rate_limit_reached_type"`
}

// CodexRateLimitWindow represents one Codex rate limit window.
type CodexRateLimitWindow struct {
	UsedPercent   *float64 `json:"used_percent"`
	WindowMinutes int      `json:"window_minutes"`
	ResetsAt      *int64   `json:"resets_at"`
}

const (
	displayModel         options.Option = "display_model"
	displayReasoning     options.Option = "display_reasoning"
	displayContext       options.Option = "display_context"
	displayTokens        options.Option = "display_tokens"
	displayFiveHourLimit options.Option = "display_five_hour_limit"
	displayWeeklyLimit   options.Option = "display_weekly_limit"
	discoverLocalStatus  options.Option = "discover_local_status"
	codexHome            options.Option = "codex_home"
	sessionRoot          options.Option = "session_root"
	codexSessionID       options.Option = "session_id"
	statusFile           options.Option = "status_file"
	statusFileEnv        options.Option = "status_file_env"
	statusJSONEnv        options.Option = "status_json_env"

	defaultCodexStatusFileEnv = "POSH_CODEX_STATUS_FILE"
	defaultCodexStatusJSONEnv = "POSH_CODEX_STATUS"
)

func (m *CodexModel) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err == nil {
		m.ID = id
		m.DisplayName = id
		return nil
	}

	type model CodexModel
	var parsed model
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	*m = CodexModel(parsed)
	if m.DisplayName == "" {
		m.DisplayName = m.ID
	}

	return nil
}

func (c *CodexData) UnmarshalJSON(data []byte) error {
	type codexData CodexData

	var wrapped struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Type == "event_msg" && len(wrapped.Payload) > 0 {
		var payload struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(wrapped.Payload, &payload); err != nil {
			return err
		}

		if payload.Type != "token_count" {
			return fmt.Errorf("unsupported codex event payload type: %s", payload.Type)
		}

		return json.Unmarshal(wrapped.Payload, (*codexData)(c))
	}

	var parsed codexData
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	*c = CodexData(parsed)
	return nil
}

func (c *Codex) Template() string {
	return " {{ .FormattedText }} "
}

func (c *Codex) Enabled() bool {
	log.Debug("codex segment: checking if enabled")

	data, found := cache.Get[CodexData](cache.Session, cache.CODEXCACHE)
	if found {
		if !data.hasStatus() {
			log.Debug("codex segment: cached data has no status information, ignoring cache entry")
			cache.Delete(cache.Session, cache.CODEXCACHE)
		} else {
			log.Debug("codex segment: found data in session cache")
			log.Debugf("codex segment: model=%s, thread=%s", data.Model.DisplayName, data.ThreadID)

			c.CodexData = data
			return true
		}
	} else {
		log.Debug("codex segment: no data found in session cache")
	}

	if data, found = c.statusFileData(); found {
		if !data.hasStatus() {
			log.Debug("codex segment: status file has no status information, checking environment variable")
		} else {
			log.Debug("codex segment: found data in status file")
			log.Debugf("codex segment: model=%s, thread=%s", data.Model.DisplayName, data.ThreadID)

			c.CodexData = data
			return true
		}
	}

	if data, found = c.statusEnvData(); found {
		if !data.hasStatus() {
			log.Debug("codex segment: status environment variable has no status information")
		} else {
			log.Debug("codex segment: found data in status environment variable")
			log.Debugf("codex segment: model=%s, thread=%s", data.Model.DisplayName, data.ThreadID)

			c.CodexData = data
			return true
		}
	}

	if data, found = c.localStatusData(); found {
		log.Debug("codex segment: found data in local session transcripts")
		log.Debugf("codex segment: model=%s, thread=%s", data.Model.DisplayName, data.ThreadID)

		c.CodexData = data
		return true
	}

	log.Debug("codex segment: no data found")
	return false
}

func (c *CodexData) hasStatus() bool {
	return c.Model.DisplayName != "" ||
		c.Model.ID != "" ||
		c.Info.TotalTokenUsage.TotalTokens > 0 ||
		c.Info.LastTokenUsage.TotalTokens > 0 ||
		(c.RateLimits != nil && c.RateLimits.hasUsage())
}

func (c *CodexRateLimits) hasUsage() bool {
	if c == nil {
		return false
	}

	return codexRateLimitWindowHasUsage(c.Primary) || codexRateLimitWindowHasUsage(c.Secondary)
}

func codexRateLimitWindowHasUsage(window *CodexRateLimitWindow) bool {
	return window != nil && window.UsedPercent != nil
}

func (c *Codex) statusFileData() (CodexData, bool) {
	if c.env == nil {
		return CodexData{}, false
	}

	file := c.optionString(statusFile, "")
	if file == "" {
		fileEnv := c.optionString(statusFileEnv, defaultCodexStatusFileEnv)
		if fileEnv == "" {
			return CodexData{}, false
		}

		file = c.env.Getenv(fileEnv)
	}

	if file == "" {
		return CodexData{}, false
	}

	return parseCodexStatus(c.env.FileContent(file))
}

func (c *Codex) optionBool(option options.Option, defaultValue bool) bool {
	if c.options == nil {
		return defaultValue
	}

	return c.options.Bool(option, defaultValue)
}

func (c *Codex) optionString(option options.Option, defaultValue string) string {
	if c.options == nil {
		return defaultValue
	}

	return c.options.String(option, defaultValue)
}

func (c *Codex) statusEnvData() (CodexData, bool) {
	if c.env == nil {
		return CodexData{}, false
	}

	envName := c.optionString(statusJSONEnv, defaultCodexStatusJSONEnv)
	if envName == "" {
		return CodexData{}, false
	}

	return parseCodexStatus(c.env.Getenv(envName))
}

func (c *Codex) localStatusData() (CodexData, bool) {
	if c.env == nil || !c.optionBool(discoverLocalStatus, true) {
		return CodexData{}, false
	}

	localOptions := ResolveCodexLocalStatusOptions(CodexLocalStatusOptions{
		CodexHome:   c.optionString(codexHome, ""),
		SessionRoot: c.optionString(sessionRoot, ""),
		SessionID:   c.optionString(codexSessionID, ""),
	}, c.env.Getenv("CODEX_HOME"), c.env.Home())

	data, err := CodexStatusFromLocalSessions(localOptions)
	if err != nil {
		log.Debugf("codex segment: failed to discover local status data: %v", err)
		return CodexData{}, false
	}

	if !data.hasStatus() {
		log.Debug("codex segment: local session data has no status information")
		return CodexData{}, false
	}

	return data, true
}

func (c *Codex) CacheKey() (string, bool) {
	parts := []string{"codex", "discover", fmt.Sprintf("%t", c.optionBool(discoverLocalStatus, true))}

	if value := c.optionString(statusFile, ""); value != "" {
		parts = append(parts, "file", codexCacheKeyHash(value))
	}

	if value := c.optionString(statusFileEnv, defaultCodexStatusFileEnv); value != "" {
		parts = append(parts, "file-env", value)
		if c.env != nil {
			if file := c.env.Getenv(value); file != "" {
				parts = append(parts, "file", codexCacheKeyHash(file))
			}
		}
	}

	if value := c.optionString(statusJSONEnv, defaultCodexStatusJSONEnv); value != "" {
		parts = append(parts, "json-env", value)
	}

	if value := c.optionString(codexSessionID, ""); value != "" {
		parts = append(parts, "session", value)
	}

	if value := c.optionString(sessionRoot, ""); value != "" {
		parts = append(parts, "root", codexCacheKeyHash(value))
	} else if value := c.optionString(codexHome, ""); value != "" {
		parts = append(parts, "root", codexCacheKeyHash(filepath.Join(value, "sessions")))
	} else if c.env != nil {
		localOptions := ResolveCodexLocalStatusOptions(CodexLocalStatusOptions{}, c.env.Getenv("CODEX_HOME"), c.env.Home())
		if localOptions.SessionRoot != "" {
			parts = append(parts, "root", codexCacheKeyHash(localOptions.SessionRoot))
		}
	}

	return strings.Join(parts, "|"), true
}

func codexCacheKeyHash(value string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}

func parseCodexStatus(status string) (CodexData, bool) {
	status = strings.TrimSpace(status)
	if status == "" {
		return CodexData{}, false
	}

	var data CodexData
	if err := json.Unmarshal([]byte(status), &data); err != nil {
		log.Debugf("codex segment: failed to parse status data: %v", err)
		return CodexData{}, false
	}

	return data, true
}

// DisplayModel returns whether the default template should display the model name.
func (c *Codex) DisplayModel() bool {
	return c.optionBool(displayModel, true)
}

// DisplayReasoning returns whether the default template should display the reasoning level.
func (c *Codex) DisplayReasoning() bool {
	return c.optionBool(displayReasoning, false)
}

// DisplayContext returns whether the default template should display context usage.
func (c *Codex) DisplayContext() bool {
	return c.optionBool(displayContext, true)
}

// DisplayTokens returns whether the default template should display total tokens.
func (c *Codex) DisplayTokens() bool {
	return c.optionBool(displayTokens, false)
}

// DisplayFiveHourLimit returns whether the default template should display the 5-hour rolling limit.
func (c *Codex) DisplayFiveHourLimit() bool {
	return c.optionBool(displayFiveHourLimit, false)
}

// DisplayWeeklyLimit returns whether the default template should display the weekly rolling limit.
func (c *Codex) DisplayWeeklyLimit() bool {
	return c.optionBool(displayWeeklyLimit, true)
}

// FormattedText returns the default Codex segment text based on display options.
func (c *Codex) FormattedText() string {
	parts := []string{}

	if model := c.FormattedModel(); model != "" {
		parts = append(parts, "\ue7cf "+model)
	}

	if c.DisplayContext() {
		if gauge := c.TokenGauge(); gauge != "" {
			parts = append(parts, "\uf2d0 "+gauge)
		}
	}

	if c.DisplayTokens() {
		parts = append(parts, "tok "+c.FormattedTokens())
	}

	if limits := c.FormattedLimits(); limits != "" {
		parts = append(parts, "\uf017 "+limits)
	}

	return strings.Join(parts, " ")
}

// FormattedModel returns model and reasoning text based on display options.
func (c *Codex) FormattedModel() string {
	model := c.Model.DisplayName
	if model == "" {
		model = c.Model.ID
	}

	if !c.DisplayModel() {
		model = ""
	}

	if !c.DisplayReasoning() || c.ReasoningEffort == "" {
		return model
	}

	if model == "" {
		return c.ReasoningEffort
	}

	reasoning := "(" + c.ReasoningEffort + ")"
	if strings.Contains(strings.ToLower(model), strings.ToLower(reasoning)) {
		return model
	}

	return fmt.Sprintf("%s %s", model, reasoning)
}

// TokenUsagePercent returns the percentage of the Codex model context window used.
func (c *Codex) TokenUsagePercent() text.Percentage {
	tokens := c.contextTokens()

	if c.Info.ModelContextWindow <= 0 || tokens <= 0 {
		return 0
	}

	percent := (float64(tokens) * 100.0) / float64(c.Info.ModelContextWindow)
	rounded := int(percent + 0.5)
	if rounded > 100 {
		return 100
	}

	return text.Percentage(rounded)
}

func (c *Codex) hasContextUsage() bool {
	return c.Info.ModelContextWindow > 0 && c.contextTokens() > 0
}

func (c *Codex) contextTokens() int {
	tokens := c.Info.LastTokenUsage.TotalTokens
	if tokens <= 0 {
		tokens = c.Info.TotalTokenUsage.TotalTokens
	}

	return tokens
}

// TokenGauge returns a 5-block gauge showing remaining context capacity using the configured characters.
func (c *Codex) TokenGauge() string {
	if !c.hasContextUsage() {
		return ""
	}

	return c.TokenUsagePercent().GaugeWith(c.optionString(gaugeMarkedChar, "▰"), c.optionString(gaugeUnmarkedChar, "▱"))
}

// TokenGaugeUsed returns a 5-block gauge showing used context capacity using the configured characters.
func (c *Codex) TokenGaugeUsed() string {
	if !c.hasContextUsage() {
		return ""
	}

	return c.TokenUsagePercent().GaugeUsedWith(c.optionString(gaugeMarkedChar, "▰"), c.optionString(gaugeUnmarkedChar, "▱"))
}

// FormattedTokens returns total session tokens as a human-readable string.
func (c *Codex) FormattedTokens() string {
	return formatTokenCount(c.Info.TotalTokenUsage.TotalTokens)
}

// FormattedInputTokens returns total input tokens as a human-readable string.
func (c *Codex) FormattedInputTokens() string {
	return formatTokenCount(c.Info.TotalTokenUsage.InputTokens)
}

// FormattedOutputTokens returns total output tokens as a human-readable string.
func (c *Codex) FormattedOutputTokens() string {
	return formatTokenCount(c.Info.TotalTokenUsage.OutputTokens)
}

// FormattedReasoningTokens returns total reasoning output tokens as a human-readable string.
func (c *Codex) FormattedReasoningTokens() string {
	return formatTokenCount(c.Info.TotalTokenUsage.ReasoningOutputTokens)
}

// FormattedCachedInputTokens returns cached input tokens as a human-readable string.
func (c *Codex) FormattedCachedInputTokens() string {
	return formatTokenCount(c.Info.TotalTokenUsage.CachedInputTokens)
}

// FormattedLastTokens returns the latest response token count as a human-readable string.
func (c *Codex) FormattedLastTokens() string {
	return formatTokenCount(c.Info.LastTokenUsage.TotalTokens)
}

// RemainingPercent returns the percentage of context window remaining (0-100).
func (c *Codex) RemainingPercent() text.Percentage {
	return remainingPercentage(c.TokenUsagePercent())
}

// RemainingTokensCount returns the number of remaining context window tokens.
func (c *Codex) RemainingTokensCount() int {
	if c.Info.ModelContextWindow <= 0 {
		return 0
	}

	tokens := c.contextTokens()
	remaining := c.Info.ModelContextWindow - tokens
	if remaining < 0 {
		return 0
	}

	return remaining
}

// FiveHourUsage returns the primary rolling window usage percentage.
func (c *Codex) FiveHourUsage() text.Percentage {
	return codexRateLimitUsage(c.RateLimits, func(l *CodexRateLimits) *CodexRateLimitWindow {
		return l.Primary
	})
}

// WeeklyUsage returns the secondary rolling window usage percentage.
func (c *Codex) WeeklyUsage() text.Percentage {
	return codexRateLimitUsage(c.RateLimits, func(l *CodexRateLimits) *CodexRateLimitWindow {
		return l.Secondary
	})
}

// FiveHourRemaining returns the primary rolling window remaining percentage.
func (c *Codex) FiveHourRemaining() text.Percentage {
	return remainingPercentage(c.FiveHourUsage())
}

// WeeklyRemaining returns the secondary rolling window remaining percentage.
func (c *Codex) WeeklyRemaining() text.Percentage {
	return remainingPercentage(c.WeeklyUsage())
}

// FiveHourGauge returns a 5-block gauge showing primary window usage.
func (c *Codex) FiveHourGauge() string {
	if !codexRateLimitHasUsage(c.RateLimits, codexPrimaryRateLimitWindow) {
		return ""
	}

	return c.FiveHourUsage().GaugeUsedWith(c.optionString(gaugeMarkedChar, "▰"), c.optionString(gaugeUnmarkedChar, "▱"))
}

// WeeklyGauge returns a 5-block gauge showing weekly window usage.
func (c *Codex) WeeklyGauge() string {
	if !codexRateLimitHasUsage(c.RateLimits, codexSecondaryRateLimitWindow) {
		return ""
	}

	return c.WeeklyUsage().GaugeUsedWith(c.optionString(gaugeMarkedChar, "▰"), c.optionString(gaugeUnmarkedChar, "▱"))
}

// FiveHourRemainingGauge returns a 5-block gauge showing primary quota remaining.
func (c *Codex) FiveHourRemainingGauge() string {
	if !codexRateLimitHasUsage(c.RateLimits, codexPrimaryRateLimitWindow) {
		return ""
	}

	return c.FiveHourUsage().GaugeWith(c.optionString(gaugeMarkedChar, "▰"), c.optionString(gaugeUnmarkedChar, "▱"))
}

// WeeklyRemainingGauge returns a 5-block gauge showing weekly quota remaining.
func (c *Codex) WeeklyRemainingGauge() string {
	if !codexRateLimitHasUsage(c.RateLimits, codexSecondaryRateLimitWindow) {
		return ""
	}

	return c.WeeklyUsage().GaugeWith(c.optionString(gaugeMarkedChar, "▰"), c.optionString(gaugeUnmarkedChar, "▱"))
}

// FormattedFiveHourRemaining returns the primary rolling window remaining percentage with a percent sign.
func (c *Codex) FormattedFiveHourRemaining() string {
	if !codexRateLimitHasUsage(c.RateLimits, codexPrimaryRateLimitWindow) {
		return ""
	}

	return fmt.Sprintf("%d%%", c.FiveHourRemaining())
}

// FormattedWeeklyRemaining returns the weekly rolling window remaining percentage with a percent sign.
func (c *Codex) FormattedWeeklyRemaining() string {
	if !codexRateLimitHasUsage(c.RateLimits, codexSecondaryRateLimitWindow) {
		return ""
	}

	return fmt.Sprintf("%d%%", c.WeeklyRemaining())
}

// FormattedLimits returns rate limit text for the default template.
func (c *Codex) FormattedLimits() string {
	limits := []string{}

	if c.DisplayFiveHourLimit() {
		if remaining := c.FormattedFiveHourRemaining(); remaining != "" {
			limits = append(limits, "5h "+remaining)
		}
	}

	if c.DisplayWeeklyLimit() {
		if remaining := c.FormattedWeeklyRemaining(); remaining != "" {
			limits = append(limits, "7d "+remaining)
		}
	}

	return strings.Join(limits, " ")
}

// FiveHourResetsAt returns the reset time for the primary rolling window.
func (c *Codex) FiveHourResetsAt() time.Time {
	return codexRateLimitReset(c.RateLimits, func(l *CodexRateLimits) *CodexRateLimitWindow {
		return l.Primary
	})
}

// WeeklyResetsAt returns the reset time for the secondary rolling window.
func (c *Codex) WeeklyResetsAt() time.Time {
	return codexRateLimitReset(c.RateLimits, func(l *CodexRateLimits) *CodexRateLimitWindow {
		return l.Secondary
	})
}

// FiveHourResetsIn returns the signed duration until the primary rolling window resets.
// Returns 0 when unavailable, negative when the window already reset.
func (c *Codex) FiveHourResetsIn() time.Duration {
	return codexRateLimitResetsIn(c.RateLimits, func(l *CodexRateLimits) *CodexRateLimitWindow {
		return l.Primary
	})
}

// WeeklyResetsIn returns the signed duration until the secondary rolling window resets.
// Returns 0 when unavailable, negative when the window already reset.
func (c *Codex) WeeklyResetsIn() time.Duration {
	return codexRateLimitResetsIn(c.RateLimits, func(l *CodexRateLimits) *CodexRateLimitWindow {
		return l.Secondary
	})
}

func remainingPercentage(used text.Percentage) text.Percentage {
	remaining := 100 - int(used)
	if remaining < 0 {
		return 0
	}

	return text.Percentage(remaining)
}

func codexRateLimitUsage(limits *CodexRateLimits, window func(*CodexRateLimits) *CodexRateLimitWindow) text.Percentage {
	if !codexRateLimitHasUsage(limits, window) {
		return 0
	}

	w := window(limits)
	if w == nil || w.UsedPercent == nil {
		return 0
	}

	percent := int(*w.UsedPercent + 0.5)
	if percent > 100 {
		return 100
	}

	if percent < 0 {
		return 0
	}

	return text.Percentage(percent)
}

func codexRateLimitHasUsage(limits *CodexRateLimits, window func(*CodexRateLimits) *CodexRateLimitWindow) bool {
	if limits == nil || window == nil {
		return false
	}

	return codexRateLimitWindowHasUsage(window(limits))
}

func codexPrimaryRateLimitWindow(limits *CodexRateLimits) *CodexRateLimitWindow {
	return limits.Primary
}

func codexSecondaryRateLimitWindow(limits *CodexRateLimits) *CodexRateLimitWindow {
	return limits.Secondary
}

func codexRateLimitReset(limits *CodexRateLimits, window func(*CodexRateLimits) *CodexRateLimitWindow) time.Time {
	if limits == nil || window == nil {
		return time.Time{}
	}

	w := window(limits)
	if w == nil || w.ResetsAt == nil {
		return time.Time{}
	}

	return time.Unix(*w.ResetsAt, 0)
}

func codexRateLimitResetsIn(limits *CodexRateLimits, window func(*CodexRateLimits) *CodexRateLimitWindow) time.Duration {
	t := codexRateLimitReset(limits, window)
	if t.IsZero() {
		return 0
	}

	return time.Until(t)
}
