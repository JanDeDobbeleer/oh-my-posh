package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"

	"github.com/stretchr/testify/assert"
)

//nolint:dupl
func TestAntigravitySegment(t *testing.T) {
	cases := []struct {
		Data            *AntigravityData
		Case            string
		ExpectedModel   string
		ExpectedSession string
		ExpectedEnabled bool
	}{
		{
			Case:            "No cache data",
			Data:            nil,
			ExpectedEnabled: false,
		},
		{
			Case: "Valid data with model and session",
			Data: &AntigravityData{
				SessionID: "eb2d84b0-1091",
				Model: AIModel{
					ID:          "gemini-3-pro",
					DisplayName: "Gemini 3 Pro",
				},
				Version: "1.0.0",
			},
			ExpectedEnabled: true,
			ExpectedModel:   "Gemini 3 Pro",
			ExpectedSession: "eb2d84b0-1091",
		},
		{
			Case: "Valid data with empty model",
			Data: &AntigravityData{
				SessionID: "test-123",
				Model:     AIModel{},
			},
			ExpectedEnabled: true,
			ExpectedModel:   "",
			ExpectedSession: "test-123",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			if tc.Data != nil {
				cache.Set(cache.Session, cache.ANTIGRAVITYCACHE, *tc.Data, cache.INFINITE)
			} else {
				cache.Delete(cache.Session, cache.ANTIGRAVITYCACHE)
			}

			env := new(mock.Environment)
			segment := &Antigravity{
				Base: Base{
					env:     env,
					options: options.Map{},
				},
			}

			enabled := segment.Enabled()
			assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

			if tc.ExpectedEnabled {
				assert.Equal(t, tc.ExpectedModel, segment.Model.DisplayName, tc.Case)
				assert.Equal(t, tc.ExpectedSession, segment.SessionID, tc.Case)
			}

			cache.Delete(cache.Session, cache.ANTIGRAVITYCACHE)
		})
	}
}

func TestAntigravityTokenUsagePercent(t *testing.T) {
	t.Cleanup(func() {
		cache.Delete(cache.Session, cache.ANTIGRAVITYCACHE)
	})

	cases := []struct {
		Case            string
		ContextWindow   AntigravityContextWindow
		ExpectedPercent text.Percentage
	}{
		{
			Case:            "No data available",
			ContextWindow:   AntigravityContextWindow{},
			ExpectedPercent: 0,
		},
		{
			Case: "UsedPercentage provided",
			ContextWindow: AntigravityContextWindow{
				UsedPercentage: new(61.7),
			},
			ExpectedPercent: 62,
		},
		{
			Case: "UsedPercentage over 100 capped",
			ContextWindow: AntigravityContextWindow{
				UsedPercentage: new(105.0),
			},
			ExpectedPercent: 100,
		},
		{
			Case: "Computed from CurrentUsage",
			ContextWindow: AntigravityContextWindow{
				CurrentUsage:      &AntigravityCurrentUsage{InputTokens: 50000},
				ContextWindowSize: 200000,
			},
			ExpectedPercent: 25,
		},
		{
			Case: "Computed from total tokens fallback",
			ContextWindow: AntigravityContextWindow{
				TotalInputTokens:  50000,
				TotalOutputTokens: 50000,
				ContextWindowSize: 200000,
			},
			ExpectedPercent: 50,
		},
		{
			Case: "ContextWindowSize zero, no UsedPercentage",
			ContextWindow: AntigravityContextWindow{
				CurrentUsage: &AntigravityCurrentUsage{InputTokens: 50000},
			},
			ExpectedPercent: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			data := AntigravityData{ContextWindow: tc.ContextWindow}
			cache.Set(cache.Session, cache.ANTIGRAVITYCACHE, data, cache.INFINITE)

			env := new(mock.Environment)
			segment := &Antigravity{
				Base: Base{
					env:     env,
					options: options.Map{},
				},
			}

			enabled := segment.Enabled()
			assert.True(t, enabled, tc.Case)
			assert.Equal(t, tc.ExpectedPercent, segment.TokenUsagePercent(), tc.Case)

			cache.Delete(cache.Session, cache.ANTIGRAVITYCACHE)
		})
	}
}

//nolint:dupl
func TestAntigravityRemainingPercent(t *testing.T) {
	cases := []struct {
		Case          string
		ContextWindow AntigravityContextWindow
		Expected      text.Percentage
	}{
		{
			Case: "RemainingPercentage provided",
			ContextWindow: AntigravityContextWindow{
				RemainingPercentage: new(40.0),
			},
			Expected: 40,
		},
		{
			Case: "Computed from used percentage",
			ContextWindow: AntigravityContextWindow{
				UsedPercentage: new(75.0),
			},
			Expected: 25,
		},
		{
			Case:          "No data",
			ContextWindow: AntigravityContextWindow{},
			Expected:      100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			data := AntigravityData{ContextWindow: tc.ContextWindow}
			cache.Set(cache.Session, cache.ANTIGRAVITYCACHE, data, cache.INFINITE)

			env := new(mock.Environment)
			segment := &Antigravity{
				Base: Base{
					env:     env,
					options: options.Map{},
				},
			}

			segment.Enabled()
			assert.Equal(t, tc.Expected, segment.RemainingPercent(), tc.Case)

			cache.Delete(cache.Session, cache.ANTIGRAVITYCACHE)
		})
	}
}

func TestAntigravityFormattedTokens(t *testing.T) {
	cases := []struct {
		Case          string
		Expected      string
		ContextWindow AntigravityContextWindow
	}{
		{
			Case:          "Zero tokens",
			ContextWindow: AntigravityContextWindow{},
			Expected:      "0",
		},
		{
			Case: "Current usage preferred",
			ContextWindow: AntigravityContextWindow{
				CurrentUsage:      &AntigravityCurrentUsage{InputTokens: 1234},
				TotalInputTokens:  9999,
				TotalOutputTokens: 9999,
			},
			Expected: "1.2K",
		},
		{
			Case: "Falls back to total tokens",
			ContextWindow: AntigravityContextWindow{
				TotalInputTokens:  250000,
				TotalOutputTokens: 250000,
			},
			Expected: "500.0K",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			data := AntigravityData{ContextWindow: tc.ContextWindow}
			cache.Set(cache.Session, cache.ANTIGRAVITYCACHE, data, cache.INFINITE)

			env := new(mock.Environment)
			segment := &Antigravity{
				Base: Base{
					env:     env,
					options: options.Map{},
				},
			}

			segment.Enabled()
			assert.Equal(t, tc.Expected, segment.FormattedTokens(), tc.Case)

			cache.Delete(cache.Session, cache.ANTIGRAVITYCACHE)
		})
	}
}

func TestAntigravityTokenGaugeCustomChars(t *testing.T) {
	usedPct := 60.0
	data := AntigravityData{
		ContextWindow: AntigravityContextWindow{
			UsedPercentage: &usedPct,
		},
	}
	cache.Set(cache.Session, cache.ANTIGRAVITYCACHE, data, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.ANTIGRAVITYCACHE)

	env := new(mock.Environment)
	segment := &Antigravity{
		Base: Base{
			env: env,
			options: options.Map{
				gaugeMarkedChar:   "█",
				gaugeUnmarkedChar: "░",
			},
		},
	}

	segment.Enabled()
	gauge := segment.TokenGaugeUsed()
	assert.NotEmpty(t, gauge)
	assert.Contains(t, gauge, "█")
}
