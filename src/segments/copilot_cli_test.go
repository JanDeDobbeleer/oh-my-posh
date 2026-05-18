package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"

	"github.com/stretchr/testify/assert"
)

func TestCopilotCLISegment(t *testing.T) {
	cases := []struct {
		Data            *CopilotCLIData
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
			Data: &CopilotCLIData{
				SessionID: "eb2d84b0-1091",
				Model: AIModel{
					ID:          "gpt-4o",
					DisplayName: "GPT-4o",
				},
				Version: "1.0.48",
				Cost: CopilotCLICost{
					TotalDurationMS: 668,
					TotalLinesAdded: 10,
				},
			},
			ExpectedEnabled: true,
			ExpectedModel:   "GPT-4o",
			ExpectedSession: "eb2d84b0-1091",
		},
		{
			Case: "Valid data with empty model",
			Data: &CopilotCLIData{
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
				cache.Set(cache.Session, cache.COPILOTCLICACHE, *tc.Data, cache.INFINITE)
			} else {
				cache.Delete(cache.Session, cache.COPILOTCLICACHE)
			}

			env := new(mock.Environment)
			segment := &CopilotCLI{
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

			cache.Delete(cache.Session, cache.COPILOTCLICACHE)
		})
	}
}

func TestCopilotCLITokenUsagePercent(t *testing.T) {
	t.Cleanup(func() {
		cache.Delete(cache.Session, cache.COPILOTCLICACHE)
	})

	cases := []struct {
		Case            string
		ContextWindow   CopilotCLIContextWindow
		ExpectedPercent text.Percentage
	}{
		{
			Case:            "No data available",
			ContextWindow:   CopilotCLIContextWindow{},
			ExpectedPercent: 0,
		},
		{
			Case: "UsedPercentage provided",
			ContextWindow: CopilotCLIContextWindow{
				UsedPercentage: new(61.7),
			},
			ExpectedPercent: 62,
		},
		{
			Case: "UsedPercentage zero",
			ContextWindow: CopilotCLIContextWindow{
				UsedPercentage: new(0.0),
			},
			ExpectedPercent: 0,
		},
		{
			Case: "UsedPercentage over 100 capped",
			ContextWindow: CopilotCLIContextWindow{
				UsedPercentage: new(105.0),
			},
			ExpectedPercent: 100,
		},
		{
			Case: "Computed from CurrentContextTokens",
			ContextWindow: CopilotCLIContextWindow{
				CurrentContextTokens: 50000,
				ContextWindowSize:    new(200000),
			},
			ExpectedPercent: 25,
		},
		{
			Case: "Computed from TotalTokens fallback",
			ContextWindow: CopilotCLIContextWindow{
				TotalTokens:       100000,
				ContextWindowSize: new(200000),
			},
			ExpectedPercent: 50,
		},
		{
			Case: "ContextWindowSize nil, no UsedPercentage",
			ContextWindow: CopilotCLIContextWindow{
				CurrentContextTokens: 50000,
				ContextWindowSize:    nil,
			},
			ExpectedPercent: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			data := CopilotCLIData{ContextWindow: tc.ContextWindow}
			cache.Set(cache.Session, cache.COPILOTCLICACHE, data, cache.INFINITE)

			env := new(mock.Environment)
			segment := &CopilotCLI{
				Base: Base{
					env:     env,
					options: options.Map{},
				},
			}

			enabled := segment.Enabled()
			assert.True(t, enabled, tc.Case)
			assert.Equal(t, tc.ExpectedPercent, segment.TokenUsagePercent(), tc.Case)

			cache.Delete(cache.Session, cache.COPILOTCLICACHE)
		})
	}
}

func TestCopilotCLIFormattedTokens(t *testing.T) {
	cases := []struct {
		Case          string
		Expected      string
		ContextWindow CopilotCLIContextWindow
	}{
		{
			Case:          "Zero tokens",
			ContextWindow: CopilotCLIContextWindow{},
			Expected:      "0",
		},
		{
			Case: "Current context tokens preferred",
			ContextWindow: CopilotCLIContextWindow{
				CurrentContextTokens: 1234,
				TotalTokens:          9999,
			},
			Expected: "1.2K",
		},
		{
			Case: "Falls back to total tokens",
			ContextWindow: CopilotCLIContextWindow{
				TotalTokens: 500000,
			},
			Expected: "500.0K",
		},
		{
			Case: "Millions",
			ContextWindow: CopilotCLIContextWindow{
				CurrentContextTokens: 1500000,
			},
			Expected: "1.5M",
		},
		{
			Case: "Small count",
			ContextWindow: CopilotCLIContextWindow{
				CurrentContextTokens: 42,
			},
			Expected: "42",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			data := CopilotCLIData{ContextWindow: tc.ContextWindow}
			cache.Set(cache.Session, cache.COPILOTCLICACHE, data, cache.INFINITE)

			env := new(mock.Environment)
			segment := &CopilotCLI{
				Base: Base{
					env:     env,
					options: options.Map{},
				},
			}

			segment.Enabled()
			assert.Equal(t, tc.Expected, segment.FormattedTokens(), tc.Case)

			cache.Delete(cache.Session, cache.COPILOTCLICACHE)
		})
	}
}

func TestCopilotCLIFormattedDuration(t *testing.T) {
	data := CopilotCLIData{
		Cost: CopilotCLICost{
			TotalDurationMS:    90000, // 1m 30s
			TotalAPIDurationMS: 30000, // 0m 30s
		},
	}
	cache.Set(cache.Session, cache.COPILOTCLICACHE, data, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.COPILOTCLICACHE)

	env := new(mock.Environment)
	segment := &CopilotCLI{
		Base: Base{
			env:     env,
			options: options.Map{},
		},
	}

	segment.Enabled()
	assert.Equal(t, "1m 30s", segment.FormattedDuration())
	assert.Equal(t, "0m 30s", segment.FormattedAPIDuration())
}

func TestCopilotCLITokenGaugeCustomChars(t *testing.T) {
	usedPct := 60.0
	data := CopilotCLIData{
		ContextWindow: CopilotCLIContextWindow{
			UsedPercentage: &usedPct,
		},
	}
	cache.Set(cache.Session, cache.COPILOTCLICACHE, data, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.COPILOTCLICACHE)

	env := new(mock.Environment)
	segment := &CopilotCLI{
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

func TestCopilotCLIRemainingPercent(t *testing.T) {
	cases := []struct {
		Case          string
		ContextWindow CopilotCLIContextWindow
		Expected      text.Percentage
	}{
		{
			Case: "RemainingPercentage provided",
			ContextWindow: CopilotCLIContextWindow{
				RemainingPercentage: new(40.0),
			},
			Expected: 40,
		},
		{
			Case: "Computed from used percentage",
			ContextWindow: CopilotCLIContextWindow{
				UsedPercentage: new(75.0),
			},
			Expected: 25,
		},
		{
			Case:          "No data",
			ContextWindow: CopilotCLIContextWindow{},
			Expected:      100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			data := CopilotCLIData{ContextWindow: tc.ContextWindow}
			cache.Set(cache.Session, cache.COPILOTCLICACHE, data, cache.INFINITE)

			env := new(mock.Environment)
			segment := &CopilotCLI{
				Base: Base{
					env:     env,
					options: options.Map{},
				},
			}

			segment.Enabled()
			assert.Equal(t, tc.Expected, segment.RemainingPercent(), tc.Case)

			cache.Delete(cache.Session, cache.COPILOTCLICACHE)
		})
	}
}
