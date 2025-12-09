package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

const (
	copilotTestURL = "https://api.github.com/copilot_internal/user"
)

func TestCopilotSegment(t *testing.T) {
	cases := []struct {
		Case            string
		JSONResponse    string
		ExpectedString  string
		Template        string
		HasToken        bool
		ExpectedEnabled bool
		HasError        bool
	}{
		{
			Case: "Valid response with usage data",
			JSONResponse: `{
				"quota_snapshots": {
					"premium_interactions": {
						"entitlement": 50,
						"remaining": 35,
						"unlimited": false
					},
					"completions": {
						"entitlement": 2000,
						"remaining": 1500,
						"unlimited": false
					},
					"chat": {
						"entitlement": 100,
						"remaining": 80,
						"unlimited": false
					}
				},
				"quota_reset_date": "2025-02-01T00:00:00Z"
			}`,
			Template:        " \uec1e {{ .Premium.Used }}/{{ .Premium.Limit }} | {{ .Inline.Used }}/{{ .Inline.Limit }} | {{ .Chat.Used }}/{{ .Chat.Limit }} ",
			ExpectedString:  "\uec1e 15/50 | 500/2000 | 20/100",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case: "Full premium usage",
			JSONResponse: `{
				"quota_snapshots": {
					"premium_interactions": {
						"entitlement": 50,
						"remaining": 0,
						"unlimited": false
					},
					"completions": {
						"entitlement": 2000,
						"remaining": 0,
						"unlimited": false
					},
					"chat": {
						"entitlement": 100,
						"remaining": 0,
						"unlimited": false
					}
				},
				"quota_reset_date": "2025-02-01T00:00:00Z"
			}`,
			Template:        " \uec1e {{ .Premium.Used }}/{{ .Premium.Limit }} | {{ .Inline.Used }}/{{ .Inline.Limit }} | {{ .Chat.Used }}/{{ .Chat.Limit }} ",
			ExpectedString:  "\uec1e 50/50 | 2000/2000 | 100/100",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case: "No usage",
			JSONResponse: `{
				"quota_snapshots": {
					"premium_interactions": {
						"entitlement": 50,
						"remaining": 50,
						"unlimited": false
					},
					"completions": {
						"entitlement": 2000,
						"remaining": 2000,
						"unlimited": false
					},
					"chat": {
						"entitlement": 0,
						"remaining": 0,
						"unlimited": true
					}
				},
				"quota_reset_date": "2025-02-01T00:00:00Z"
			}`,
			Template:        " \uec1e {{ .Premium.Used }}/{{ .Premium.Limit }} | {{ .Inline.Used }}/{{ .Inline.Limit }} | {{ .Chat.Used }}/{{ .Chat.Limit }} ",
			ExpectedString:  "\uec1e 0/50 | 0/2000 | 0/0",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case: "Custom template with percentages",
			JSONResponse: `{
				"quota_snapshots": {
					"premium_interactions": {
						"entitlement": 100,
						"remaining": 50,
						"unlimited": false
					},
					"completions": {
						"entitlement": 1000,
						"remaining": 750,
						"unlimited": false
					},
					"chat": {
						"entitlement": 200,
						"remaining": 100,
						"unlimited": false
					}
				},
				"quota_reset_date": "2025-02-01T00:00:00Z"
			}`,
			Template:        " {{ .Premium.Percent }}% | {{ .Inline.Percent }}% | {{ .Chat.Percent }}% ",
			ExpectedString:  "50% | 25% | 50%",
			ExpectedEnabled: true,
			HasToken:        true,
		},
		{
			Case:            "No access token",
			ExpectedEnabled: false,
			HasError:        false,
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
			Case:            "Empty quota data",
			JSONResponse:    `{}`,
			ExpectedEnabled: false,
			HasToken:        true,
		},
		{
			Case:            "Null quota_snapshots",
			JSONResponse:    `{"quota_snapshots": null}`,
			ExpectedEnabled: false,
			HasToken:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := &mock.Environment{}
			props := options.Map{}

			// Setup cached token mock
			if tc.HasToken {
				cache.Set(cache.Device, auth.CopilotTokenKey, "ghp_test_token", cache.INFINITE)
			} else {
				cache.Delete(cache.Device, auth.CopilotTokenKey)
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

func TestCopilotPercentageGauge(t *testing.T) {
	cases := []struct {
		Case          string
		ExpectedGauge string
		Percent       CopilotPercentage
	}{
		{
			Case:          "0 percent used (100% remaining)",
			Percent:       CopilotPercentage(0),
			ExpectedGauge: "▰▰▰▰▰",
		},
		{
			Case:          "20 percent used (80% remaining - 4 blocks)",
			Percent:       CopilotPercentage(20),
			ExpectedGauge: "▰▰▰▰▱",
		},
		{
			Case:          "40 percent used (60% remaining - 3 blocks)",
			Percent:       CopilotPercentage(40),
			ExpectedGauge: "▰▰▰▱▱",
		},
		{
			Case:          "60 percent used (40% remaining - 2 blocks)",
			Percent:       CopilotPercentage(60),
			ExpectedGauge: "▰▰▱▱▱",
		},
		{
			Case:          "80 percent used (20% remaining - 1 block)",
			Percent:       CopilotPercentage(80),
			ExpectedGauge: "▰▱▱▱▱",
		},
		{
			Case:          "100 percent used (0% remaining - 0 blocks)",
			Percent:       CopilotPercentage(100),
			ExpectedGauge: "▱▱▱▱▱",
		},
		{
			Case:          "50 percent used (50% remaining - 2.5 rounds to 2 blocks)",
			Percent:       CopilotPercentage(50),
			ExpectedGauge: "▰▰▱▱▱",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			result := tc.Percent.Gauge()
			assert.Equal(t, tc.ExpectedGauge, result, tc.Case)
		})
	}
}

func TestCopilotRemainingPercentage(t *testing.T) {
	env := &mock.Environment{}
	props := options.Map{}

	jsonResponse := `{
		"quota_snapshots": {
			"premium_interactions": {
				"entitlement": 100,
				"remaining": 75,
				"unlimited": false
			},
			"completions": {
				"entitlement": 1000,
				"remaining": 600,
				"unlimited": false
			},
			"chat": {
				"entitlement": 200,
				"remaining": 0,
				"unlimited": false
			}
		},
		"quota_reset_date": "2025-02-15T00:00:00Z"
	}`

	cache.Set(cache.Device, auth.CopilotTokenKey, "ghp_test_token", cache.INFINITE)
	env.On("HTTPRequest", copilotTestURL).Return([]byte(jsonResponse), nil)

	c := &Copilot{}
	c.Init(props, env)

	enabled := c.Enabled()
	assert.True(t, enabled)

	// Test Premium: 100 entitlement - 75 remaining = 25 used (25% used, 75% remaining)
	assert.Equal(t, CopilotPercentage(25), c.Premium.Percent)
	assert.Equal(t, CopilotPercentage(75), c.Premium.Remaining)

	// Test Inline: 1000 entitlement - 600 remaining = 400 used (40% used, 60% remaining)
	assert.Equal(t, CopilotPercentage(40), c.Inline.Percent)
	assert.Equal(t, CopilotPercentage(60), c.Inline.Remaining)

	// Test Chat: 200 entitlement - 0 remaining = 200 used (100% used, 0% remaining)
	assert.Equal(t, CopilotPercentage(100), c.Chat.Percent)
	assert.Equal(t, CopilotPercentage(0), c.Chat.Remaining)
}

func TestCopilotBillingCycleEnd(t *testing.T) {
	env := &mock.Environment{}
	props := options.Map{}

	jsonResponse := `{
		"quota_snapshots": {
			"premium_interactions": {
				"entitlement": 50,
				"remaining": 35,
				"unlimited": false
			},
			"completions": {
				"entitlement": 2000,
				"remaining": 1500,
				"unlimited": false
			},
			"chat": {
				"entitlement": 100,
				"remaining": 80,
				"unlimited": false
			}
		},
		"quota_reset_date": "2025-02-15T00:00:00Z"
	}`

	cache.Set(cache.Device, auth.CopilotTokenKey, "ghp_test_token", cache.INFINITE)
	env.On("HTTPRequest", copilotTestURL).Return([]byte(jsonResponse), nil)

	c := &Copilot{}
	c.Init(props, env)

	enabled := c.Enabled()
	assert.True(t, enabled)
	assert.Equal(t, "2025-02-15T00:00:00Z", c.BillingCycleEnd)
}
