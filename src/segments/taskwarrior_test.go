package segments

import (
	"errors"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestTaskwarrior(t *testing.T) {
	cases := []struct {
		Case            string
		Template        string
		HasCommand      bool
		ConfiguredTags  map[string]string
		TagOutputs      map[string]string
		TagErrors       map[string]error
		ContextCmd      string
		ContextOutput   string
		ContextErr      error
		ExpectedEnabled bool
		ExpectedTags    map[string]int
		ExpectedContext string
	}{
		{
			Case:       "happy path default tags",
			HasCommand: true,
			TagOutputs: map[string]string{
				"+PENDING due.before:tomorrow count":       "3",
				"+PENDING scheduled.before:tomorrow count": "1",
				"+WAITING count":                           "2",
			},
			ExpectedEnabled: true,
			ExpectedTags: map[string]int{
				"Due":       3,
				"Scheduled": 1,
				"Waiting":   2,
			},
		},
		{
			Case:            "no command",
			HasCommand:      false,
			ExpectedEnabled: false,
		},
		{
			Case:       "all zeros",
			HasCommand: true,
			TagOutputs: map[string]string{
				"+PENDING due.before:tomorrow count":       "0",
				"+PENDING scheduled.before:tomorrow count": "0",
				"+WAITING count":                           "0",
			},
			ExpectedEnabled: true,
			ExpectedTags: map[string]int{
				"Due":       0,
				"Scheduled": 0,
				"Waiting":   0,
			},
		},
		{
			Case:       "custom tags only",
			HasCommand: true,
			ConfiguredTags: map[string]string{
				"urgent": "+PENDING +OVERDUE",
			},
			TagOutputs: map[string]string{
				"+PENDING +OVERDUE count": "5",
			},
			ExpectedEnabled: true,
			ExpectedTags: map[string]int{
				"Urgent": 5,
			},
		},
		{
			Case:       "with context",
			HasCommand: true,
			ConfiguredTags: map[string]string{
				"due": "+PENDING due.before:tomorrow",
			},
			TagOutputs: map[string]string{
				"+PENDING due.before:tomorrow count": "3",
			},
			ContextCmd:      "_get rc.context",
			ContextOutput:   "work",
			ExpectedEnabled: true,
			ExpectedTags: map[string]int{
				"Due": 3,
			},
			ExpectedContext: "work",
		},
		{
			Case:       "command error returns zero for that tag",
			HasCommand: true,
			ConfiguredTags: map[string]string{
				"due": "+PENDING due.before:tomorrow",
			},
			TagErrors: map[string]error{
				"+PENDING due.before:tomorrow count": errors.New("command failed"),
			},
			ExpectedEnabled: true,
			ExpectedTags: map[string]int{
				"Due": 0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("HasCommand", "task").Return(tc.HasCommand)

			// Determine which tags to configure
			configuredTags := tc.ConfiguredTags
			if configuredTags == nil && tc.HasCommand {
				// default tags
				configuredTags = map[string]string{
					"due":       "+PENDING due.before:tomorrow",
					"scheduled": "+PENDING scheduled.before:tomorrow",
					"waiting":   "+WAITING",
				}
			}

			for _, filter := range configuredTags {
				args := splitAndAppend(filter, "count")
				var output string
				var err error
				if tc.TagOutputs != nil {
					output = tc.TagOutputs[filter+" count"]
				}
				if tc.TagErrors != nil {
					err = tc.TagErrors[filter+" count"]
				}
				env.On("RunCommand", "task", args).Return(output, err)
			}

			if tc.ContextCmd != "" {
				contextArgs := splitArgs(tc.ContextCmd)
				env.On("RunCommand", "task", contextArgs).Return(tc.ContextOutput, tc.ContextErr)
			}

			props := options.Map{}
			if tc.ConfiguredTags != nil {
				props[TaskwarriorTags] = tc.ConfiguredTags
			}
			if tc.ContextCmd != "" {
				props[TaskwarriorCtxCmd] = tc.ContextCmd
			}

			tw := &Taskwarrior{}
			tw.Init(props, env)

			assert.Equal(t, tc.ExpectedEnabled, tw.Enabled(), tc.Case)

			if !tc.ExpectedEnabled {
				return
			}

			assert.Equal(t, tc.ExpectedTags, tw.Tags, tc.Case)
			assert.Equal(t, tc.ExpectedContext, tw.Context, tc.Case)
		})
	}
}

// splitAndAppend splits a filter string into fields and appends extra args.
func splitAndAppend(filter string, extra ...string) []string {
	parts := splitArgs(filter)
	return append(parts, extra...)
}

// splitArgs splits a space-separated argument string into a slice.
func splitArgs(s string) []string {
	var result []string
	for _, f := range strings.Fields(s) {
		result = append(result, f)
	}
	return result
}
