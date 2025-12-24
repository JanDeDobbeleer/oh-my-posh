package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"

	"github.com/stretchr/testify/assert"
)

func TestClaudeSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ClaudeStatus    string
		ExpectedModel   string
		ExpectedSession string
		ExpectedEnabled bool
	}{
		{
			Case:            "Empty environment variable",
			ClaudeStatus:    "",
			ExpectedEnabled: false,
		},
		{
			Case:            "Invalid JSON",
			ClaudeStatus:    "invalid json",
			ExpectedEnabled: false,
		},
		{
			Case: "Valid JSON with all fields",
			ClaudeStatus: `{
				"session_id": "abc123",
				"model": {
					"id": "claude-opus-4-1",
					"display_name": "Opus"
				},
				"workspace": {
					"current_dir": "/repo/project",
					"project_dir": "/repo"
				},
				"cost": {
					"total_cost_usd": 0.01,
					"total_duration_ms": 45000
				},
				"context_window": {
					"total_input_tokens": 15234,
					"total_output_tokens": 4521,
					"context_window_size": 200000,
					"current_usage": {
						"input_tokens": 8500,
						"output_tokens": 1200
					}
				}
			}`,
			ExpectedEnabled: true,
			ExpectedModel:   "Opus",
			ExpectedSession: "abc123",
		},
		{
			Case: "Valid JSON with partial fields",
			ClaudeStatus: `{
				"session_id": "xyz789",
				"model": {
					"id": "claude-sonnet-3-5",
					"display_name": "Sonnet 3.5"
				},
				"workspace": {},
				"cost": {},
				"context_window": {
					"total_input_tokens": 1000,
					"total_output_tokens": 500,
					"context_window_size": 100000
				}
			}`,
			ExpectedEnabled: true,
			ExpectedModel:   "Sonnet 3.5",
			ExpectedSession: "xyz789",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", poshClaudeStatusEnv).Return(tc.ClaudeStatus)

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
