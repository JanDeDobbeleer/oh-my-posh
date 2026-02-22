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
		Case               string
		HasCommand         bool
		ConfiguredCommands map[string]string
		CommandOutputs     map[string]string
		CommandErrors      map[string]error
		ExpectedEnabled    bool
		ExpectedCommands   map[string]string
	}{
		{
			Case:       "happy path default commands",
			HasCommand: true,
			CommandOutputs: map[string]string{
				"+PENDING due.before:tomorrow count":       "3",
				"+PENDING scheduled.before:tomorrow count": "1",
				"+WAITING count":                           "2",
			},
			ExpectedEnabled: true,
			ExpectedCommands: map[string]string{
				"Due":       "3",
				"Scheduled": "1",
				"Waiting":   "2",
				"Context":   "",
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
			CommandOutputs: map[string]string{
				"+PENDING due.before:tomorrow count":       "0",
				"+PENDING scheduled.before:tomorrow count": "0",
				"+WAITING count":                           "0",
			},
			ExpectedEnabled: true,
			ExpectedCommands: map[string]string{
				"Due":       "0",
				"Scheduled": "0",
				"Waiting":   "0",
				"Context":   "",
			},
		},
		{
			Case:       "custom commands only",
			HasCommand: true,
			ConfiguredCommands: map[string]string{
				"urgent": "+PENDING +OVERDUE count",
			},
			CommandOutputs: map[string]string{
				"+PENDING +OVERDUE count": "5",
			},
			ExpectedEnabled: true,
			ExpectedCommands: map[string]string{
				"Urgent": "5",
			},
		},
		{
			Case:       "context command via commands map",
			HasCommand: true,
			ConfiguredCommands: map[string]string{
				"due":     "+PENDING due.before:tomorrow count",
				"context": "_get rc.context",
			},
			CommandOutputs: map[string]string{
				"+PENDING due.before:tomorrow count": "3",
				"_get rc.context":                    "work",
			},
			ExpectedEnabled: true,
			ExpectedCommands: map[string]string{
				"Due":     "3",
				"Context": "work",
			},
		},
		{
			Case:       "command error returns empty string for that command",
			HasCommand: true,
			ConfiguredCommands: map[string]string{
				"due": "+PENDING due.before:tomorrow count",
			},
			CommandErrors: map[string]error{
				"+PENDING due.before:tomorrow count": errors.New("command failed"),
			},
			ExpectedEnabled: true,
			ExpectedCommands: map[string]string{
				"Due": "",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("HasCommand", "task").Return(tc.HasCommand)

			configuredCommands := tc.ConfiguredCommands
			if configuredCommands == nil && tc.HasCommand {
				// default commands
				configuredCommands = map[string]string{
					"due":       "+PENDING due.before:tomorrow count",
					"scheduled": "+PENDING scheduled.before:tomorrow count",
					"waiting":   "+WAITING count",
					"context":   "_get rc.context",
				}
			}

			for _, args := range configuredCommands {
				splitArgs := splitTaskArgs(args)
				var output string
				var err error
				if tc.CommandOutputs != nil {
					output = tc.CommandOutputs[args]
				}
				if tc.CommandErrors != nil {
					err = tc.CommandErrors[args]
				}
				env.On("RunCommand", "task", splitArgs).Return(output, err)
			}

			props := options.Map{}
			if tc.ConfiguredCommands != nil {
				props[TaskwarriorCommands] = tc.ConfiguredCommands
			}

			tw := &Taskwarrior{}
			tw.Init(props, env)

			assert.Equal(t, tc.ExpectedEnabled, tw.Enabled(), tc.Case)

			if !tc.ExpectedEnabled {
				return
			}

			assert.Equal(t, tc.ExpectedCommands, tw.Commands, tc.Case)
		})
	}
}

// splitTaskArgs splits a space-separated argument string into a slice.
func splitTaskArgs(s string) []string {
	return append([]string{}, strings.Fields(s)...)
}
