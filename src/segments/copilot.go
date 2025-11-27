package segments

import (
	"encoding/json"
	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

// CopilotTokenKey is the cache key for the Copilot OAuth token.
const CopilotTokenKey = "copilot_token"

// Copilot displays GitHub Copilot usage statistics.
type Copilot struct {
	Base
	BillingCycleEnd string
	Error           string
	PremiumUsed     int
	PremiumLimit    int
	PremiumPercent  int
	StandardUsed    int
	StandardLimit   int
	StandardPercent int
	Authenticate    bool
}

const (
	copilotAPIURL = "https://api.github.com/copilot_internal/user"
)

// copilotQuotaSnapshot represents a single quota type.
type copilotQuotaSnapshot struct {
	Entitlement int `json:"entitlement"`
	Remaining   int `json:"remaining"`
}

// copilotQuotaSnapshots represents the quota snapshots structure.
type copilotQuotaSnapshots struct {
	PremiumInteractions copilotQuotaSnapshot `json:"premium_interactions"`
	Completions         copilotQuotaSnapshot `json:"completions"`
}

// copilotUserInfo represents user information with quota data.
type copilotUserInfo struct {
	QuotaSnapshots    *copilotQuotaSnapshots `json:"quota_snapshots"`
	BillingCycleStart string                 `json:"billing_cycle_start"`
	QuotaResetDate    string                 `json:"quota_reset_date"`
}

// copilotAPIResponse represents the API response structure.
type copilotAPIResponse struct {
	UserInfo *copilotUserInfo `json:"userInfo"`
}

func (c *Copilot) Template() string {
	return " \ue272 {{ .PremiumUsed }}/{{ .PremiumLimit }} | {{ .StandardUsed }}/{{ .StandardLimit }} "
}

func (c *Copilot) Enabled() bool {
	err := c.setStatus()
	if err == nil {
		return true
	}

	// Check if this is an authentication error (no token or 403)
	if _, isNoToken := err.(*noTokenError); isNoToken {
		c.Authenticate = true
		c.Error = "run 'oh-my-posh auth copilot' to authenticate"
		return true
	}

	log.Error(err)
	return false
}

func (c *Copilot) getAccessToken() string {
	// Check cache from `oh-my-posh auth copilot`
	if cachedToken, OK := cache.Get[string](cache.Device, CopilotTokenKey); OK && len(cachedToken) != 0 {
		return cachedToken
	}

	return ""
}

func (c *Copilot) getResult() (*copilotAPIResponse, error) {
	accessToken := c.getAccessToken()
	if len(accessToken) == 0 {
		c.Error = "no access token"
		return nil, &noTokenError{}
	}

	httpTimeout := c.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	addAuthHeader := func(request *http.Request) {
		request.Header.Add("Authorization", "Bearer "+accessToken)
		request.Header.Add("User-Agent", "GitHub-Copilot-Usage-Tray")
		request.Header.Add("X-GitHub-Api-Version", "2025-05-01")
		request.Header.Add("Accept", "application/json")
	}

	body, err := c.env.HTTPRequest(copilotAPIURL, nil, httpTimeout, addAuthHeader)
	if err != nil {
		c.Error = "request failed"
		return nil, err
	}

	response := new(copilotAPIResponse)
	err = json.Unmarshal(body, &response)
	if err != nil {
		c.Error = "parse error"
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
		c.Error = "no quota data"
		return &noQuotaDataError{}
	}

	// Calculate premium usage
	c.PremiumLimit = quotaSnapshots.PremiumInteractions.Entitlement
	c.PremiumUsed = max(c.PremiumLimit-quotaSnapshots.PremiumInteractions.Remaining, 0)
	c.PremiumPercent = c.calculatePercent(c.PremiumUsed, c.PremiumLimit)

	// Calculate standard usage (completions)
	c.StandardLimit = quotaSnapshots.Completions.Entitlement
	c.StandardUsed = max(c.StandardLimit-quotaSnapshots.Completions.Remaining, 0)
	c.StandardPercent = c.calculatePercent(c.StandardUsed, c.StandardLimit)

	// Set billing cycle end date
	if response.UserInfo != nil {
		c.BillingCycleEnd = response.UserInfo.QuotaResetDate
	}

	return nil
}

func (c *Copilot) extractQuotaSnapshots(response *copilotAPIResponse) *copilotQuotaSnapshots {
	if response == nil {
		return nil
	}

	// Try userInfo.quota_snapshots path
	if response.UserInfo != nil && response.UserInfo.QuotaSnapshots != nil {
		return response.UserInfo.QuotaSnapshots
	}

	return nil
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

type noTokenError struct{}

func (e *noTokenError) Error() string {
	return "no access token available"
}

type noQuotaDataError struct{}

func (e *noQuotaDataError) Error() string {
	return "no quota data in response"
}
