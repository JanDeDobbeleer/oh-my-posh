package segments

import (
	"encoding/json"
	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

// CopilotUsage represents usage statistics for a specific quota type.
type CopilotUsage struct {
	Used      int             `json:"used"`
	Limit     int             `json:"limit"`
	Percent   text.Percentage `json:"percent"`
	Remaining text.Percentage `json:"remaining"`
	Unlimited bool            `json:"unlimited"`
}

// Copilot displays GitHub Copilot usage statistics.
type Copilot struct {
	Base
	BillingCycleEnd string       `json:"billing_cycle_end"`
	Premium         CopilotUsage `json:"premium"`
	Inline          CopilotUsage `json:"inline"`
	Chat            CopilotUsage `json:"chat"`
}

const (
	copilotAPIURL = "https://api.github.com/copilot_internal/user"
)

// copilotQuotaSnapshot represents a single quota type.
type copilotQuotaSnapshot struct {
	Entitlement int  `json:"entitlement"`
	Remaining   int  `json:"remaining"`
	Unlimited   bool `json:"unlimited"`
}

// copilotQuotaSnapshots represents the quota snapshots structure.
type copilotQuotaSnapshots struct {
	PremiumInteractions copilotQuotaSnapshot `json:"premium_interactions"`
	Completions         copilotQuotaSnapshot `json:"completions"`
	Chat                copilotQuotaSnapshot `json:"chat"`
}

// copilotAPIResponse represents the API response structure.
type copilotAPIResponse struct {
	QuotaSnapshots    *copilotQuotaSnapshots `json:"quota_snapshots"`
	QuotaResetDate    string                 `json:"quota_reset_date"`
	QuotaResetDateUTC string                 `json:"quota_reset_date_utc"`
}

func (c *Copilot) Template() string {
	return " \uec1e {{ .Premium.Percent.Gauge }} "
}

func (c *Copilot) Enabled() bool {
	err := c.setStatus()
	if err != nil {
		log.Error(err)
		return false
	}

	return true
}

func (c *Copilot) getAccessToken() string {
	// Check cache from `oh-my-posh auth copilot`
	if cachedToken, OK := cache.Get[string](cache.Device, auth.CopilotTokenKey); OK && len(cachedToken) != 0 {
		return cachedToken
	}

	return ""
}

func (c *Copilot) getResult() (*copilotAPIResponse, error) {
	accessToken := c.getAccessToken()
	if len(accessToken) == 0 {
		return nil, &noAccessTokenError{}
	}

	log.Debug("found access token")

	httpTimeout := c.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	addAuthHeader := func(request *http.Request) {
		request.Header.Set("Authorization", "Bearer "+accessToken)
		request.Header.Set("User-Agent", "GitHub-Copilot-Usage-Tray")
		request.Header.Set("Accept", "application/json")
		request.Header.Set("Content-Type", "application/json")
	}

	body, err := c.env.HTTPRequest(copilotAPIURL, nil, httpTimeout, addAuthHeader)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Debug("executed HTTP request successfully")

	response := new(copilotAPIResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Copilot) setStatus() error {
	response, err := c.getResult()
	if err != nil {
		return err
	}

	// Extract quota data from response - try different paths
	quotaSnapshots := c.extractQuotaSnapshots(response)
	if quotaSnapshots == nil {
		return &noQuotaDataError{}
	}

	// Calculate premium usage
	c.Premium = c.calculateUsage(quotaSnapshots.PremiumInteractions)

	// Calculate inline usage (completions)
	c.Inline = c.calculateUsage(quotaSnapshots.Completions)

	// Calculate chat usage
	c.Chat = c.calculateUsage(quotaSnapshots.Chat)

	// Set billing cycle end date
	c.BillingCycleEnd = response.QuotaResetDate
	if c.BillingCycleEnd == "" {
		c.BillingCycleEnd = response.QuotaResetDateUTC
	}

	return nil
}

func (c *Copilot) extractQuotaSnapshots(response *copilotAPIResponse) *copilotQuotaSnapshots {
	if response == nil {
		return nil
	}

	// Use root-level quota_snapshots
	if response.QuotaSnapshots != nil {
		return response.QuotaSnapshots
	}

	return nil
}

func (c *Copilot) calculateUsage(snapshot copilotQuotaSnapshot) CopilotUsage {
	if snapshot.Unlimited {
		return CopilotUsage{
			Used:      0,
			Limit:     0,
			Percent:   text.Percentage(0),
			Remaining: text.Percentage(100),
			Unlimited: true,
		}
	}

	used := max(snapshot.Entitlement-snapshot.Remaining, 0)
	percent := c.calculatePercent(used, snapshot.Entitlement)
	remainingPercent := 100 - percent

	return CopilotUsage{
		Used:      used,
		Limit:     snapshot.Entitlement,
		Percent:   text.Percentage(percent),
		Remaining: text.Percentage(remainingPercent),
		Unlimited: false,
	}
}

func (c *Copilot) calculatePercent(used, limit int) int {
	if limit <= 0 {
		return 0
	}

	percent := (used * 100) / limit
	if percent > 100 {
		return 100
	}

	return percent
}

// Custom error types for better error handling

type noQuotaDataError struct{}

func (e *noQuotaDataError) Error() string {
	return "no quota data in response"
}

type noAccessTokenError struct{}

func (e *noAccessTokenError) Error() string {
	return "no access token available, use 'oh-my-posh auth copilot' to authenticate"
}
