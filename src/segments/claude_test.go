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
					CurrentUsage: ClaudeCurrentUsage{
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
		Case            string
		InputTokens     int
		OutputTokens    int
		ContextWindow   int
		ExpectedPercent text.Percentage
	}{
		{
			Case:            "Zero context window",
			InputTokens:     1000,
			OutputTokens:    500,
			ContextWindow:   0,
			ExpectedPercent: 0,
		},
		{
			Case:            "10% usage",
			InputTokens:     8000,
			OutputTokens:    2000,
			ContextWindow:   100000,
			ExpectedPercent: 10,
		},
		{
			Case:            "50% usage",
			InputTokens:     50000,
			OutputTokens:    50000,
			ContextWindow:   200000,
			ExpectedPercent: 50,
		},
		{
			Case:            "Over 100% usage (capped)",
			InputTokens:     100000,
			OutputTokens:    50000,
			ContextWindow:   100000,
			ExpectedPercent: 100,
		},
	}

	for _, tc := range cases {
		claude := &Claude{}
		claude.ContextWindow.TotalInputTokens = tc.InputTokens
		claude.ContextWindow.TotalOutputTokens = tc.OutputTokens
		claude.ContextWindow.ContextWindowSize = tc.ContextWindow

		percent := claude.TokenUsagePercent()
		assert.Equal(t, tc.ExpectedPercent, percent, tc.Case)
	}
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
		Case           string
		ExpectedFormat string
		InputTokens    int
		OutputTokens   int
	}{
		{
			Case:           "Small token count",
			InputTokens:    300,
			OutputTokens:   200,
			ExpectedFormat: "500",
		},
		{
			Case:           "Thousands",
			InputTokens:    8500,
			OutputTokens:   1500,
			ExpectedFormat: "10.0K",
		},
		{
			Case:           "Tens of thousands",
			InputTokens:    50000,
			OutputTokens:   25000,
			ExpectedFormat: "75.0K",
		},
		{
			Case:           "Millions",
			InputTokens:    1500000,
			OutputTokens:   500000,
			ExpectedFormat: "2.0M",
		},
	}

	for _, tc := range cases {
		claude := &Claude{}
		claude.ContextWindow.TotalInputTokens = tc.InputTokens
		claude.ContextWindow.TotalOutputTokens = tc.OutputTokens

		formatted := claude.FormattedTokens()
		assert.Equal(t, tc.ExpectedFormat, formatted, tc.Case)
	}
}
