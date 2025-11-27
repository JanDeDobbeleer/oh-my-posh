package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

const (
	copilotTestURL = "https://api.github.com/copilot_internal/user"
)

func TestCopilotSegment(t *testing.T) {
	cases := []struct {
		Case                 string
		JSONResponse         string
		ExpectedString       string
		Template             string
		HasToken             bool
		ExpectedEnabled      bool
		ExpectedAuthenticate bool
		HasError             bool
	}{
		{
			Case: "Valid response with usage data",
			JSONResponse: `{
				"userInfo": {
					"quota_snapshots": {
						"premium_interactions": {
							"entitlement": 50,
							"remaining": 35
						},
						"completions": {
							"entitlement": 2000,
							"remaining": 1500
						}
					},
					"billing_cycle_start": "2025-01-01T00:00:00Z",
					"quota_reset_date": "2025-02-01T00:00:00Z"
				}
			}`,
			ExpectedString:  "\ue272 15/50 | 500/2000",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case: "Full premium usage",
			JSONResponse: `{
				"userInfo": {
					"quota_snapshots": {
						"premium_interactions": {
							"entitlement": 50,
							"remaining": 0
						},
						"completions": {
							"entitlement": 2000,
							"remaining": 0
						}
					},
					"quota_reset_date": "2025-02-01T00:00:00Z"
				}
			}`,
			ExpectedString:  "\ue272 50/50 | 2000/2000",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case: "No usage",
			JSONResponse: `{
				"userInfo": {
					"quota_snapshots": {
						"premium_interactions": {
							"entitlement": 50,
							"remaining": 50
						},
						"completions": {
							"entitlement": 2000,
							"remaining": 2000
						}
					},
					"quota_reset_date": "2025-02-01T00:00:00Z"
				}
			}`,
			ExpectedString:  "\ue272 0/50 | 0/2000",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case: "Custom template with percentages",
			JSONResponse: `{
				"userInfo": {
					"quota_snapshots": {
						"premium_interactions": {
							"entitlement": 100,
							"remaining": 50
						},
						"completions": {
							"entitlement": 1000,
							"remaining": 750
						}
					},
					"quota_reset_date": "2025-02-01T00:00:00Z"
				}
			}`,
			Template:        " {{ .PremiumPercent }}% | {{ .StandardPercent }}% ",
			ExpectedString:  "50% | 25%",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case:                 "No access token",
			ExpectedEnabled:      true,
			ExpectedAuthenticate: true,
			ExpectedString:       "run 'oh-my-posh auth copilot' to authenticate",
			Template:             "{{ .Error }}",
			HasError:             false,
		},
		{
			Case:            "API error",
			HasError:        true,
			ExpectedEnabled: false,
			HasToken:        true,
		},
		{
			Case:            "Invalid JSON response",
			JSONResponse:    "invalid json",
			ExpectedEnabled: false,
			HasToken:        true,
		},
		{
			Case: "Empty quota data",
			JSONResponse: `{
				"userInfo": {}
			}`,
			ExpectedEnabled: false,
			HasToken:        true,
		},
		{
			Case: "Null userInfo",
			JSONResponse: `{
				"userInfo": null
			}`,
			ExpectedEnabled: false,
			HasToken:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := &mock.Environment{}
			props := properties.Map{}

			// Setup cached token mock
			if tc.HasToken {
				cache.Set(cache.Device, CopilotTokenKey, "ghp_test_token", cache.INFINITE)
			} else {
				cache.Delete(cache.Device, CopilotTokenKey)
			}

			// Setup HTTP request mock
			var httpErr error
			if tc.HasError {
				httpErr = errors.New("request failed")
			}

			if tc.HasToken {
				env.On("HTTPRequest", copilotTestURL).Return([]byte(tc.JSONResponse), httpErr)
			}

			c := &Copilot{}
			c.Init(props, env)

			enabled := c.Enabled()
			assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
			assert.Equal(t, tc.ExpectedAuthenticate, c.Authenticate, tc.Case)

			if !enabled {
				return
			}

			template := tc.Template
			if template == "" {
				template = c.Template()
			}

			assert.Equal(t, tc.ExpectedString, renderTemplate(env, template, c), tc.Case)
		})
	}
}

func TestCopilotPercentageCalculation(t *testing.T) {
	cases := []struct {
		Case            string
		Used            int
		Limit           int
		ExpectedPercent int
	}{
		{
			Case:            "50 percent",
			Used:            50,
			Limit:           100,
			ExpectedPercent: 50,
		},
		{
			Case:            "Zero limit",
			Used:            10,
			Limit:           0,
			ExpectedPercent: 0,
		},
		{
			Case:            "Negative limit",
			Used:            10,
			Limit:           -5,
			ExpectedPercent: 0,
		},
		{
			Case:            "Over 100 percent caps at 100",
			Used:            150,
			Limit:           100,
			ExpectedPercent: 100,
		},
		{
			Case:            "Zero used",
			Used:            0,
			Limit:           100,
			ExpectedPercent: 0,
		},
		{
			Case:            "Full usage",
			Used:            100,
			Limit:           100,
			ExpectedPercent: 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			c := &Copilot{}
			result := c.calculatePercent(tc.Used, tc.Limit)
			assert.Equal(t, tc.ExpectedPercent, result, tc.Case)
		})
	}
}

func TestCopilotBillingCycleEnd(t *testing.T) {
	env := &mock.Environment{}
	props := properties.Map{}

	jsonResponse := `{
		"userInfo": {
			"quota_snapshots": {
				"premium_interactions": {
					"entitlement": 50,
					"remaining": 35
				},
				"completions": {
					"entitlement": 2000,
					"remaining": 1500
				}
			},
			"quota_reset_date": "2025-02-15T00:00:00Z"
		}
	}`

	cache.Set(cache.Device, CopilotTokenKey, "ghp_test_token", cache.INFINITE)
	env.On("HTTPRequest", copilotTestURL).Return([]byte(jsonResponse), nil)

	c := &Copilot{}
	c.Init(props, env)

	enabled := c.Enabled()
	assert.True(t, enabled)
	assert.Equal(t, "2025-02-15T00:00:00Z", c.BillingCycleEnd)
}
