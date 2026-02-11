package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"

	"github.com/stretchr/testify/assert"
)

func TestClaudeSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ClaudeData      *ClaudeData
		ExpectedModel   string
		ExpectedSession string
		ExpectedEnabled bool
	}{
		{
			Case:            "No cache data",
			ClaudeData:      nil,
			ExpectedEnabled: false,
		},
		{
			Case: "Valid cache data with all fields",
			ClaudeData: &ClaudeData{
				SessionID: "abc123",
				Model: ClaudeModel{
					ID:          "claude-opus-4-1",
					DisplayName: "Opus",
				},
				Workspace: ClaudeWorkspace{
					CurrentDir: "/repo/project",
					ProjectDir: "/repo",
				},
				Cost: ClaudeCost{
					TotalCostUSD:    0.01,
					TotalDurationMS: 45000,
				},
				ContextWindow: ClaudeContextWindow{
					TotalInputTokens:  15234,
					TotalOutputTokens: 4521,
					ContextWindowSize: 200000,
					CurrentUsage: &ClaudeCurrentUsage{
						InputTokens:  8500,
						OutputTokens: 1200,
					},
				},
			},
			ExpectedEnabled: true,
			ExpectedModel:   "Opus",
			ExpectedSession: "abc123",
		},
		{
			Case: "Valid cache data with partial fields",
			ClaudeData: &ClaudeData{
				SessionID: "xyz789",
				Model: ClaudeModel{
					ID:          "claude-sonnet-3-5",
					DisplayName: "Sonnet 3.5",
				},
				ContextWindow: ClaudeContextWindow{
					TotalInputTokens:  1000,
					TotalOutputTokens: 500,
					ContextWindowSize: 100000,
				},
			},
			ExpectedEnabled: true,
			ExpectedModel:   "Sonnet 3.5",
			ExpectedSession: "xyz789",
		},
	}

	for _, tc := range cases {
		// Setup cache for test
		if tc.ClaudeData != nil {
			cache.Set(cache.Session, cache.CLAUDECACHE, *tc.ClaudeData, cache.INFINITE)
		} else {
			cache.Delete(cache.Session, cache.CLAUDECACHE)
		}

		env := new(mock.Environment)
		claude := &Claude{
			Base: Base{
				env:     env,
				options: options.Map{},
			},
		}

		enabled := claude.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedModel, claude.Model.DisplayName, tc.Case)
			assert.Equal(t, tc.ExpectedSession, claude.SessionID, tc.Case)
		}
	}
}

func TestClaudeTokenUsagePercent(t *testing.T) {
	cases := []struct {
		UsedPercentage           *int
		Case                     string
		InputTokens              int
		OutputTokens             int
		CurrentInput             int
		CacheCreationInputTokens int
		CacheReadInputTokens     int
		ContextWindow            int
		ExpectedPercent          text.Percentage
		HasCurrentUsage          bool
	}{
		{
			Case:            "Uses UsedPercentage when available",
			UsedPercentage:  new(42),
			ContextWindow:   200000,
			ExpectedPercent: 42,
		},
		{
			Case:            "UsedPercentage capped at 100",
			UsedPercentage:  new(150),
			ContextWindow:   200000,
			ExpectedPercent: 100,
		},
		{
			Case:            "UsedPercentage zero is valid",
			UsedPercentage:  new(0),
			ContextWindow:   200000,
			ExpectedPercent: 0,
		},
		{
			Case:            "Context reset - both UsedPercentage and CurrentUsage nil",
			UsedPercentage:  nil,
			HasCurrentUsage: false,
			InputTokens:     50000, // High cumulative total - should be ignored
			OutputTokens:    50000,
			ContextWindow:   200000,
			ExpectedPercent: 0, // Should return 0 after reset, not fallback to total
		},
		{
			Case:            "Zero context window (no UsedPercentage)",
			HasCurrentUsage: true,
			InputTokens:     1000,
			OutputTokens:    500,
			ContextWindow:   0,
			ExpectedPercent: 0,
		},
		{
			Case:            "10% usage (fallback to total)",
			HasCurrentUsage: true,
			InputTokens:     8000,
			OutputTokens:    2000,
			ContextWindow:   100000,
			ExpectedPercent: 10,
		},
		{
			Case:            "50% usage (fallback to total)",
			HasCurrentUsage: true,
			InputTokens:     50000,
			OutputTokens:    50000,
			ContextWindow:   200000,
			ExpectedPercent: 50,
		},
		{
			Case:            "Over 100% usage (capped)",
			HasCurrentUsage: true,
			InputTokens:     100000,
			OutputTokens:    50000,
			ContextWindow:   100000,
			ExpectedPercent: 100,
		},
		{
			Case:            "Uses CurrentUsage input tokens",
			HasCurrentUsage: true,
			InputTokens:     100000, // High cumulative total
			OutputTokens:    50000,
			CurrentInput:    20000, // Current context input
			ContextWindow:   200000,
			ExpectedPercent: 10, // Should use current input (20000/200000 = 10%)
		},
		{
			Case:                     "Uses CurrentUsage with cache tokens",
			HasCurrentUsage:          true,
			InputTokens:              100000, // High cumulative total
			OutputTokens:             50000,
			CurrentInput:             10000,
			CacheCreationInputTokens: 5000,
			CacheReadInputTokens:     5000,
			ContextWindow:            200000,
			ExpectedPercent:          10, // (10000+5000+5000)/200000 = 10%
		},
		{
			Case:            "Uses CurrentUsage after compact (low current, high total)",
			HasCurrentUsage: true,
			InputTokens:     100000, // High cumulative total
			OutputTokens:    50000,
			CurrentInput:    6000, // Low current context (after compact)
			ContextWindow:   200000,
			ExpectedPercent: 3, // Should use current (6000/200000 = 3%)
		},
		{
			Case:            "Fallback to total when CurrentUsage is zero",
			HasCurrentUsage: true,
			InputTokens:     20000,
			OutputTokens:    10000,
			CurrentInput:    0,
			ContextWindow:   100000,
			ExpectedPercent: 30, // Should fallback to total (30000/100000 = 30%)
		},
	}

	for _, tc := range cases {
		claude := &Claude{}
		claude.ContextWindow.TotalInputTokens = tc.InputTokens
		claude.ContextWindow.TotalOutputTokens = tc.OutputTokens
		if tc.HasCurrentUsage {
			claude.ContextWindow.CurrentUsage = &ClaudeCurrentUsage{
				InputTokens:              tc.CurrentInput,
				CacheCreationInputTokens: tc.CacheCreationInputTokens,
				CacheReadInputTokens:     tc.CacheReadInputTokens,
			}
		}
		claude.ContextWindow.UsedPercentage = tc.UsedPercentage
		claude.ContextWindow.ContextWindowSize = tc.ContextWindow

		percent := claude.TokenUsagePercent()
		assert.Equal(t, tc.ExpectedPercent, percent, tc.Case)
	}
}

// intPtr is a helper to create a pointer to an int value
//
//go:fix inline
func intPtr(i int) *int {
	return new(i)
}

func TestClaudeFormattedCost(t *testing.T) {
	cases := []struct {
		Case         string
		ExpectedCost string
		CostUSD      float64
	}{
		{
			Case:         "Very small cost",
			CostUSD:      0.0012,
			ExpectedCost: "$0.0012",
		},
		{
			Case:         "Small cost",
			CostUSD:      0.0099,
			ExpectedCost: "$0.0099",
		},
		{
			Case:         "Regular cost",
			CostUSD:      0.15,
			ExpectedCost: "$0.15",
		},
		{
			Case:         "Large cost",
			CostUSD:      12.34,
			ExpectedCost: "$12.34",
		},
	}

	for _, tc := range cases {
		claude := &Claude{}
		claude.Cost.TotalCostUSD = tc.CostUSD

		formatted := claude.FormattedCost()
		assert.Equal(t, tc.ExpectedCost, formatted, tc.Case)
	}
}

func TestClaudeFormattedTokens(t *testing.T) {
	cases := []struct {
		Case                     string
		ExpectedFormat           string
		InputTokens              int
		OutputTokens             int
		CurrentInput             int
		CacheCreationInputTokens int
		CacheReadInputTokens     int
		HasCurrentUsage          bool
	}{
		{
			Case:            "Small token count (fallback to total)",
			HasCurrentUsage: true,
			InputTokens:     300,
			OutputTokens:    200,
			ExpectedFormat:  "500",
		},
		{
			Case:            "Thousands (fallback to total)",
			HasCurrentUsage: true,
			InputTokens:     8500,
			OutputTokens:    1500,
			ExpectedFormat:  "10.0K",
		},
		{
			Case:            "Tens of thousands (fallback to total)",
			HasCurrentUsage: true,
			InputTokens:     50000,
			OutputTokens:    25000,
			ExpectedFormat:  "75.0K",
		},
		{
			Case:            "Millions (fallback to total)",
			HasCurrentUsage: true,
			InputTokens:     1500000,
			OutputTokens:    500000,
			ExpectedFormat:  "2.0M",
		},
		{
			Case:            "Uses CurrentUsage input tokens",
			HasCurrentUsage: true,
			InputTokens:     100000, // High cumulative total
			OutputTokens:    50000,
			CurrentInput:    10000, // Current context input
			ExpectedFormat:  "10.0K",
		},
		{
			Case:                     "Uses CurrentUsage with cache tokens",
			HasCurrentUsage:          true,
			InputTokens:              100000, // High cumulative total
			OutputTokens:             50000,
			CurrentInput:             5000,
			CacheCreationInputTokens: 2500,
			CacheReadInputTokens:     2500,
			ExpectedFormat:           "10.0K", // 5000+2500+2500 = 10000
		},
		{
			Case:            "Uses CurrentUsage after compact (low current)",
			HasCurrentUsage: true,
			InputTokens:     500000, // High cumulative total
			OutputTokens:    200000,
			CurrentInput:    500, // Low current context (after compact)
			ExpectedFormat:  "500",
		},
		{
			Case:            "Fallback to total when CurrentUsage is zero",
			HasCurrentUsage: true,
			InputTokens:     50000,
			OutputTokens:    25000,
			CurrentInput:    0,
			ExpectedFormat:  "75.0K", // Should fallback to total
		},
		{
			Case:            "Nil CurrentUsage falls back to total",
			HasCurrentUsage: false,
			InputTokens:     50000,
			OutputTokens:    25000,
			ExpectedFormat:  "75.0K", // Should fallback to total
		},
	}

	for _, tc := range cases {
		claude := &Claude{}
		claude.ContextWindow.TotalInputTokens = tc.InputTokens
		claude.ContextWindow.TotalOutputTokens = tc.OutputTokens
		if tc.HasCurrentUsage {
			claude.ContextWindow.CurrentUsage = &ClaudeCurrentUsage{
				InputTokens:              tc.CurrentInput,
				CacheCreationInputTokens: tc.CacheCreationInputTokens,
				CacheReadInputTokens:     tc.CacheReadInputTokens,
			}
		}

		formatted := claude.FormattedTokens()
		assert.Equal(t, tc.ExpectedFormat, formatted, tc.Case)
	}
}
