package segments

import (
	"errors"
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
		DueOutput       string
		DueErr          error
		ScheduledOutput string
		ScheduledErr    error
		WaitingOutput   string
		WaitingErr      error
		ContextOutput   string
		ContextErr      error
		ExpectedEnabled bool
		ExpectedString  string
	}{
		{
			Case:            "happy path",
			HasCommand:      true,
			DueOutput:       "3",
			ScheduledOutput: "1",
			WaitingOutput:   "2",
			ContextOutput:   "work",
			ExpectedEnabled: true,
			ExpectedString:  "\uf4a0 3",
		},
		{
			Case:            "no command",
			HasCommand:      false,
			ExpectedEnabled: false,
			ExpectedString:  "",
		},
		{
			Case:            "all zeros",
			HasCommand:      true,
			DueOutput:       "0",
			ScheduledOutput: "0",
			WaitingOutput:   "0",
			ContextOutput:   "",
			ExpectedEnabled: true,
			ExpectedString:  "\uf4a0 0",
		},
		{
			Case:            "custom template",
			Template:        " D:{{.Due}} S:{{.Scheduled}} W:{{.Waiting}} [{{.Context}}] ",
			HasCommand:      true,
			DueOutput:       "5",
			ScheduledOutput: "2",
			WaitingOutput:   "1",
			ContextOutput:   "home",
			ExpectedEnabled: true,
			ExpectedString:  "D:5 S:2 W:1 [home]",
		},
		{
			Case:            "due error returns zero",
			HasCommand:      true,
			DueOutput:       "",
			DueErr:          errors.New("command failed"),
			ScheduledOutput: "1",
			WaitingOutput:   "0",
			ContextOutput:   "",
			ExpectedEnabled: true,
			ExpectedString:  "\uf4a0 0",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		env.On("HasCommand", "task").Return(tc.HasCommand)
		env.On("RunCommand", "task", []string{"+PENDING", "due.before:tomorrow", "count"}).Return(tc.DueOutput, tc.DueErr)
		env.On("RunCommand", "task", []string{"+PENDING", "scheduled.before:tomorrow", "count"}).Return(tc.ScheduledOutput, tc.ScheduledErr)
		env.On("RunCommand", "task", []string{"+WAITING", "count"}).Return(tc.WaitingOutput, tc.WaitingErr)
		env.On("RunCommand", "task", []string{"_get", "rc.context"}).Return(tc.ContextOutput, tc.ContextErr)

		props := options.Map{}

		tw := &Taskwarrior{}
		tw.Init(props, env)

		tmpl := tc.Template
		if tmpl == "" {
			tmpl = tw.Template()
		}

		assert.Equal(t, tc.ExpectedEnabled, tw.Enabled(), tc.Case)

		if !tc.ExpectedEnabled {
			continue
		}

		got := renderTemplate(env, tmpl, tw)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
