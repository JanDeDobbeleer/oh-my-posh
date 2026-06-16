package segments

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/text"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodexSegment(t *testing.T) {
	cases := []struct {
		Data            *CodexData
		Case            string
		ExpectedModel   string
		ExpectedThread  string
		ExpectedEnabled bool
	}{
		{
			Case:            "No cache data",
			Data:            nil,
			ExpectedEnabled: false,
		},
		{
			Case: "Valid data",
			Data: &CodexData{
				ThreadID: "019e9eac-83ec-7393-ae18-cc2e566394d5",
				Model: CodexModel{
					ID:          "gpt-5.5",
					DisplayName: "GPT-5.5",
				},
			},
			ExpectedEnabled: true,
			ExpectedModel:   "GPT-5.5",
			ExpectedThread:  "019e9eac-83ec-7393-ae18-cc2e566394d5",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			if tc.Data != nil {
				cache.Set(cache.Session, cache.CODEXCACHE, *tc.Data, cache.INFINITE)
			} else {
				cache.Delete(cache.Session, cache.CODEXCACHE)
			}

			env := new(mock.Environment)
			if tc.Data == nil {
				env.On("Getenv", defaultCodexStatusFileEnv).Return("")
				env.On("Getenv", defaultCodexStatusJSONEnv).Return("")
			}

			segment := &Codex{
				Base: Base{
					env: env,
					options: options.Map{
						discoverLocalStatus: false,
					},
				},
			}

			enabled := segment.Enabled()
			assert.Equal(t, tc.ExpectedEnabled, enabled)

			if tc.ExpectedEnabled {
				assert.Equal(t, tc.ExpectedModel, segment.Model.DisplayName)
				assert.Equal(t, tc.ExpectedThread, segment.ThreadID)
			}

			cache.Delete(cache.Session, cache.CODEXCACHE)
		})
	}
}

func TestCodexDataUnmarshalWrappedTokenCountEvent(t *testing.T) {
	input := []byte(`{
		"type": "event_msg",
		"payload": {
			"type": "token_count",
			"model": "gpt-5.5",
			"thread_id": "thread-1",
			"info": {
				"total_token_usage": {
					"input_tokens": 1000,
					"cached_input_tokens": 200,
					"output_tokens": 300,
					"reasoning_output_tokens": 50,
					"total_tokens": 1300
				},
				"last_token_usage": {
					"input_tokens": 100,
					"output_tokens": 30,
					"total_tokens": 130
				},
				"model_context_window": 260000
			},
			"rate_limits": {
				"plan_type": "pro",
				"primary": {
					"used_percent": 12.5,
					"window_minutes": 300,
					"resets_at": 1780794869
				},
				"secondary": {
					"used_percent": 26,
					"window_minutes": 10080,
					"resets_at": 1781137565
				}
			}
		}
	}`)

	var data CodexData
	require.NoError(t, json.Unmarshal(input, &data))
	assert.Equal(t, "token_count", data.Type)
	assert.Equal(t, "gpt-5.5", data.Model.DisplayName)
	assert.Equal(t, 1300, data.Info.TotalTokenUsage.TotalTokens)
	assert.Equal(t, 26.0, *data.RateLimits.Secondary.UsedPercent)
}

func TestCodexDataUnmarshalUnsupportedEvent(t *testing.T) {
	input := []byte(`{
		"type": "event_msg",
		"payload": {
			"type": "agent_message",
			"message": "not token data"
		}
	}`)

	var data CodexData
	require.Error(t, json.Unmarshal(input, &data))
}

func TestCodexTemplateUsesExplicitFields(t *testing.T) {
	weekly := 26.0
	segment := &Codex{
		CodexData: CodexData{
			Model: CodexModel{
				ID: "gpt-5.5",
			},
			Info: CodexTokenInfo{
				LastTokenUsage: CodexTokenUsage{
					TotalTokens: 4000,
				},
				ModelContextWindow: 10000,
			},
			RateLimits: &CodexRateLimits{
				Secondary: &CodexRateLimitWindow{
					UsedPercent: &weekly,
				},
			},
		},
	}

	assert.Equal(t, "\ue7cf gpt-5.5 \uf2d0 ▰▰▰▱▱ \uf017 7d 74%", renderTemplate(new(mock.Environment), segment.Template(), segment))
}

func TestCodexSegmentEmptyPayloadDisabled(t *testing.T) {
	cache.Delete(cache.Session, cache.CODEXCACHE)

	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("")
	env.On("Getenv", defaultCodexStatusJSONEnv).Return(`{"type":"token_count"}`)

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				discoverLocalStatus: false,
			},
		},
	}

	assert.False(t, segment.Enabled())
}

func TestCodexSegmentStatusFile(t *testing.T) {
	cache.Delete(cache.Session, cache.CODEXCACHE)

	status := `{
		"type": "token_count",
		"thread_id": "thread-from-file",
		"model": {
			"id": "gpt-5.5",
			"display_name": "GPT-5.5"
		}
	}`

	env := new(mock.Environment)
	env.On("FileContent", "/tmp/codex-status.json").Return(status)

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				statusFile: "/tmp/codex-status.json",
			},
		},
	}

	require.True(t, segment.Enabled())
	assert.Equal(t, "thread-from-file", segment.ThreadID)
	assert.Equal(t, "GPT-5.5", segment.Model.DisplayName)
}

func TestCodexSegmentStatusFileExpandsHome(t *testing.T) {
	cache.Delete(cache.Session, cache.CODEXCACHE)

	status := `{
		"type": "token_count",
		"thread_id": "thread-from-home-file",
		"model": "gpt-5.5"
	}`

	env := new(mock.Environment)
	env.On("Home").Return("/home/test")
	env.On("FileContent", filepath.Join("/home/test", ".codex", "codex-status.json")).Return(status)

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				statusFile: "~/.codex/codex-status.json",
			},
		},
	}

	require.True(t, segment.Enabled())
	assert.Equal(t, "thread-from-home-file", segment.ThreadID)
}

func TestCodexSegmentStatusFileEnv(t *testing.T) {
	cache.Delete(cache.Session, cache.CODEXCACHE)

	status := `{
		"type": "token_count",
		"thread_id": "thread-from-file-env",
		"model": "gpt-5.5"
	}`

	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("/tmp/codex-status.json")
	env.On("FileContent", "/tmp/codex-status.json").Return(status)

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				discoverLocalStatus: false,
			},
		},
	}

	require.True(t, segment.Enabled())
	assert.Equal(t, "thread-from-file-env", segment.ThreadID)
	assert.Equal(t, "gpt-5.5", segment.Model.DisplayName)
}

func TestCodexSegmentStatusEnv(t *testing.T) {
	cache.Delete(cache.Session, cache.CODEXCACHE)

	status := `{
		"type": "token_count",
		"thread_id": "thread-from-env",
		"model": "gpt-5.5"
	}`

	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("")
	env.On("Getenv", defaultCodexStatusJSONEnv).Return(status)

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				discoverLocalStatus: false,
			},
		},
	}

	require.True(t, segment.Enabled())
	assert.Equal(t, "thread-from-env", segment.ThreadID)
	assert.Equal(t, "gpt-5.5", segment.Model.DisplayName)
}

func TestCodexSegmentInvalidCacheFallsBackToStatusFile(t *testing.T) {
	cache.Set(cache.Session, cache.CODEXCACHE, CodexData{}, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.CODEXCACHE)

	status := `{
		"type": "token_count",
		"thread_id": "thread-from-file-after-cache-miss",
		"model": "gpt-5.5"
	}`

	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("/tmp/codex-status.json")
	env.On("FileContent", "/tmp/codex-status.json").Return(status)

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				discoverLocalStatus: false,
			},
		},
	}

	require.True(t, segment.Enabled())
	assert.Equal(t, "thread-from-file-after-cache-miss", segment.ThreadID)
	assert.Equal(t, "gpt-5.5", segment.Model.DisplayName)

	_, found := cache.Get[CodexData](cache.Session, cache.CODEXCACHE)
	assert.False(t, found)
}

func TestCodexSegmentInvalidStatusFileFallsBackToStatusEnv(t *testing.T) {
	cache.Delete(cache.Session, cache.CODEXCACHE)

	status := `{
		"type": "token_count",
		"thread_id": "thread-from-env-after-file-miss",
		"model": "gpt-5.5"
	}`

	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("/tmp/codex-status.json")
	env.On("FileContent", "/tmp/codex-status.json").Return(`{"type":"token_count"}`)
	env.On("Getenv", defaultCodexStatusJSONEnv).Return(status)

	segment := &Codex{
		Base: Base{
			env:     env,
			options: options.Map{},
		},
	}

	require.True(t, segment.Enabled())
	assert.Equal(t, "thread-from-env-after-file-miss", segment.ThreadID)
	assert.Equal(t, "gpt-5.5", segment.Model.DisplayName)
}

func TestCodexSegmentEmptyRateLimitsDisabled(t *testing.T) {
	cache.Set(cache.Session, cache.CODEXCACHE, CodexData{
		RateLimits: &CodexRateLimits{},
	}, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.CODEXCACHE)

	env := new(mock.Environment)
	env.On("Getenv", defaultCodexStatusFileEnv).Return("")
	env.On("Getenv", defaultCodexStatusJSONEnv).Return("")

	segment := &Codex{
		Base: Base{
			env: env,
			options: options.Map{
				discoverLocalStatus: false,
			},
		},
	}

	assert.False(t, segment.Enabled())
}

func TestCodexFormattedTextUsesOptionsWithoutEnabled(t *testing.T) {
	primary := 12.5
	secondary := 26.0
	segment := &Codex{
		Base: Base{
			options: options.Map{
				displayFiveHourLimit: true,
				displayReasoning:     true,
				displayTokens:        true,
				gaugeMarkedChar:      "#",
				gaugeUnmarkedChar:    "-",
			},
		},
		CodexData: CodexData{
			ReasoningEffort: "high",
			Model: CodexModel{
				ID:          "gpt-5.5",
				DisplayName: "GPT-5.5",
			},
			Info: CodexTokenInfo{
				TotalTokenUsage: CodexTokenUsage{
					TotalTokens: 1200,
				},
				ModelContextWindow: 10000,
			},
			RateLimits: &CodexRateLimits{
				Primary: &CodexRateLimitWindow{
					UsedPercent: &primary,
				},
				Secondary: &CodexRateLimitWindow{
					UsedPercent: &secondary,
				},
			},
		},
	}

	assert.Equal(t, "\ue7cf GPT-5.5 (high) \uf2d0 ####- tok 1.2K \uf017 5h 87% 7d 74%", segment.FormattedText())
}

func TestCodexFormattedTextOmitsUnknownContext(t *testing.T) {
	segment := &Codex{
		CodexData: CodexData{
			Model: CodexModel{
				DisplayName: "GPT-5.5",
			},
			Info: CodexTokenInfo{
				TotalTokenUsage: CodexTokenUsage{
					TotalTokens: 1200,
				},
			},
		},
	}

	assert.Empty(t, segment.TokenGauge())
	assert.Empty(t, segment.TokenGaugeUsed())
	assert.Equal(t, "\ue7cf GPT-5.5", segment.FormattedText())
}

func TestCodexTokenUsagePercent(t *testing.T) {
	cases := []struct {
		Case            string
		Info            CodexTokenInfo
		ExpectedPercent text.Percentage
	}{
		{
			Case:            "No data",
			Info:            CodexTokenInfo{},
			ExpectedPercent: 0,
		},
		{
			Case: "Rounded usage",
			Info: CodexTokenInfo{
				TotalTokenUsage: CodexTokenUsage{
					TotalTokens: 1300,
				},
				ModelContextWindow: 10000,
			},
			ExpectedPercent: 13,
		},
		{
			Case: "Usage capped",
			Info: CodexTokenInfo{
				TotalTokenUsage: CodexTokenUsage{
					TotalTokens: 12000,
				},
				ModelContextWindow: 10000,
			},
			ExpectedPercent: 100,
		},
		{
			Case: "Last usage preferred over cumulative total",
			Info: CodexTokenInfo{
				TotalTokenUsage: CodexTokenUsage{
					TotalTokens: 12000,
				},
				LastTokenUsage: CodexTokenUsage{
					TotalTokens: 2500,
				},
				ModelContextWindow: 10000,
			},
			ExpectedPercent: 25,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			segment := &Codex{
				CodexData: CodexData{
					Info: tc.Info,
				},
			}
			assert.Equal(t, tc.ExpectedPercent, segment.TokenUsagePercent())
		})
	}
}

func TestCodexFormattedTokens(t *testing.T) {
	segment := &Codex{
		CodexData: CodexData{
			Info: CodexTokenInfo{
				TotalTokenUsage: CodexTokenUsage{
					InputTokens:           1200,
					CachedInputTokens:     400,
					OutputTokens:          500000,
					ReasoningOutputTokens: 42,
					TotalTokens:           1500000,
				},
				LastTokenUsage: CodexTokenUsage{
					TotalTokens: 987,
				},
			},
		},
	}

	assert.Equal(t, "1.5M", segment.FormattedTokens())
	assert.Equal(t, "1.2K", segment.FormattedInputTokens())
	assert.Equal(t, "500.0K", segment.FormattedOutputTokens())
	assert.Equal(t, "42", segment.FormattedReasoningTokens())
	assert.Equal(t, "400", segment.FormattedCachedInputTokens())
	assert.Equal(t, "987", segment.FormattedLastTokens())
}

func TestCodexRemainingContext(t *testing.T) {
	segment := &Codex{
		CodexData: CodexData{
			Info: CodexTokenInfo{
				TotalTokenUsage: CodexTokenUsage{
					TotalTokens: 250000,
				},
				LastTokenUsage: CodexTokenUsage{
					TotalTokens: 25000,
				},
				ModelContextWindow: 100000,
			},
		},
	}

	assert.Equal(t, text.Percentage(75), segment.RemainingPercent())
	assert.Equal(t, 75000, segment.RemainingTokensCount())
}

func TestCodexRateLimitUsage(t *testing.T) {
	primary := 12.5
	secondary := 26.0
	reset := int64(1781137565)
	segment := &Codex{
		Base: Base{
			options: options.Map{},
		},
		CodexData: CodexData{
			RateLimits: &CodexRateLimits{
				Primary: &CodexRateLimitWindow{
					UsedPercent:   &primary,
					WindowMinutes: 300,
				},
				Secondary: &CodexRateLimitWindow{
					UsedPercent:   &secondary,
					WindowMinutes: 10080,
					ResetsAt:      &reset,
				},
			},
		},
	}

	assert.Equal(t, text.Percentage(13), segment.FiveHourUsage())
	assert.Equal(t, text.Percentage(26), segment.WeeklyUsage())
	assert.Equal(t, text.Percentage(87), segment.FiveHourRemaining())
	assert.Equal(t, text.Percentage(74), segment.WeeklyRemaining())
	assert.Equal(t, "74%", segment.FormattedWeeklyRemaining())
	assert.Equal(t, "5h", segment.FiveHourLimitLabel())
	assert.Equal(t, "7d", segment.WeeklyLimitLabel())
	assert.Equal(t, "7d 74%", segment.FormattedLimits())
	assert.Equal(t, time.Unix(reset, 0), segment.WeeklyResetsAt())
}

func TestCodexRateLimitLabelsUseWindowMinutes(t *testing.T) {
	primary := 12.5
	secondary := 26.0
	segment := &Codex{
		Base: Base{
			options: options.Map{
				displayFiveHourLimit: true,
			},
		},
		CodexData: CodexData{
			RateLimits: &CodexRateLimits{
				Primary: &CodexRateLimitWindow{
					UsedPercent:   &primary,
					WindowMinutes: 90,
				},
				Secondary: &CodexRateLimitWindow{
					UsedPercent:   &secondary,
					WindowMinutes: 4320,
				},
			},
		},
	}

	assert.Equal(t, "90m", segment.FiveHourLimitLabel())
	assert.Equal(t, "3d", segment.WeeklyLimitLabel())
	assert.Equal(t, "90m 87% 3d 74%", segment.FormattedLimits())
}

func TestCodexFormattedLimitsOmitsUnknownWindows(t *testing.T) {
	primary := 12.5
	segment := &Codex{
		Base: Base{
			options: options.Map{
				displayFiveHourLimit: true,
			},
		},
		CodexData: CodexData{
			RateLimits: &CodexRateLimits{
				Primary: &CodexRateLimitWindow{
					UsedPercent: &primary,
				},
				Secondary: &CodexRateLimitWindow{},
			},
		},
	}

	assert.Equal(t, "5h 87%", segment.FormattedLimits())
	assert.Equal(t, "87%", segment.FormattedFiveHourRemaining())
	assert.Empty(t, segment.FormattedWeeklyRemaining())
	assert.NotEmpty(t, segment.FiveHourGauge())
	assert.NotEmpty(t, segment.FiveHourRemainingGauge())
	assert.Empty(t, segment.WeeklyGauge())
	assert.Empty(t, segment.WeeklyRemainingGauge())
}

func TestCodexFormattedLimitsEmptyWhenUsageUnknown(t *testing.T) {
	segment := &Codex{
		Base: Base{
			options: options.Map{
				displayFiveHourLimit: true,
			},
		},
		CodexData: CodexData{
			RateLimits: &CodexRateLimits{},
		},
	}

	assert.Empty(t, segment.FormattedLimits())
	assert.Empty(t, segment.FormattedFiveHourRemaining())
	assert.Empty(t, segment.FormattedWeeklyRemaining())
	assert.Empty(t, segment.FiveHourGauge())
	assert.Empty(t, segment.WeeklyGauge())
	assert.Empty(t, segment.FiveHourRemainingGauge())
	assert.Empty(t, segment.WeeklyRemainingGauge())
}

func TestCodexTokenGaugeCustomChars(t *testing.T) {
	data := CodexData{
		Info: CodexTokenInfo{
			TotalTokenUsage: CodexTokenUsage{
				TotalTokens: 60000,
			},
			ModelContextWindow: 100000,
		},
	}
	cache.Set(cache.Session, cache.CODEXCACHE, data, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.CODEXCACHE)

	segment := &Codex{
		Base: Base{
			env: new(mock.Environment),
			options: options.Map{
				gaugeMarkedChar:   "█",
				gaugeUnmarkedChar: "░",
			},
		},
	}

	require.True(t, segment.Enabled())
	assert.Contains(t, segment.TokenGaugeUsed(), "█")
	assert.Contains(t, segment.TokenGauge(), "░")
}

func TestCodexDisplayOptions(t *testing.T) {
	primary := 12.5
	secondary := 26.0
	data := CodexData{
		ThreadID:        "thread-1",
		ReasoningEffort: "high",
		Model: CodexModel{
			ID:          "gpt-5.5",
			DisplayName: "GPT-5.5",
		},
		Info: CodexTokenInfo{
			TotalTokenUsage: CodexTokenUsage{
				TotalTokens: 1200,
			},
			ModelContextWindow: 10000,
		},
		RateLimits: &CodexRateLimits{
			Primary: &CodexRateLimitWindow{
				UsedPercent: &primary,
			},
			Secondary: &CodexRateLimitWindow{
				UsedPercent: &secondary,
			},
		},
	}
	cache.Set(cache.Session, cache.CODEXCACHE, data, cache.INFINITE)
	defer cache.Delete(cache.Session, cache.CODEXCACHE)

	cases := []struct {
		Case              string
		Options           options.Map
		ExpectedModel     bool
		ExpectedReason    bool
		ExpectedContext   bool
		ExpectedTokens    bool
		ExpectedFiveHour  bool
		ExpectedWeekly    bool
		ExpectedModelText string
		ExpectedLimits    string
	}{
		{
			Case:              "Defaults",
			Options:           options.Map{},
			ExpectedModel:     true,
			ExpectedReason:    false,
			ExpectedContext:   true,
			ExpectedTokens:    false,
			ExpectedFiveHour:  false,
			ExpectedWeekly:    true,
			ExpectedModelText: "GPT-5.5",
			ExpectedLimits:    "7d 74%",
		},
		{
			Case: "Everything enabled",
			Options: options.Map{
				displayFiveHourLimit: true,
				displayReasoning:     true,
				displayTokens:        true,
			},
			ExpectedModel:     true,
			ExpectedReason:    true,
			ExpectedContext:   true,
			ExpectedTokens:    true,
			ExpectedFiveHour:  true,
			ExpectedWeekly:    true,
			ExpectedModelText: "GPT-5.5 (high)",
			ExpectedLimits:    "5h 87% 7d 74%",
		},
		{
			Case: "Minimal limits only",
			Options: options.Map{
				displayModel:         false,
				displayContext:       false,
				displayWeeklyLimit:   false,
				displayFiveHourLimit: true,
			},
			ExpectedModel:     false,
			ExpectedReason:    false,
			ExpectedContext:   false,
			ExpectedTokens:    false,
			ExpectedFiveHour:  true,
			ExpectedWeekly:    false,
			ExpectedModelText: "",
			ExpectedLimits:    "5h 87%",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			segment := &Codex{
				Base: Base{
					env:     new(mock.Environment),
					options: tc.Options,
				},
			}

			require.True(t, segment.Enabled())
			assert.Equal(t, tc.ExpectedModel, segment.DisplayModel())
			assert.Equal(t, tc.ExpectedReason, segment.DisplayReasoning())
			assert.Equal(t, tc.ExpectedContext, segment.DisplayContext())
			assert.Equal(t, tc.ExpectedTokens, segment.DisplayTokens())
			assert.Equal(t, tc.ExpectedFiveHour, segment.DisplayFiveHourLimit())
			assert.Equal(t, tc.ExpectedWeekly, segment.DisplayWeeklyLimit())
			assert.Equal(t, tc.ExpectedModelText, segment.FormattedModel())
			assert.Equal(t, tc.ExpectedLimits, segment.FormattedLimits())
		})
	}
}
