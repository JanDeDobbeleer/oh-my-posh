package template

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	cases := []struct {
		Case        string
		Command     string
		ReturnValue string
		ReturnError error
		Template    string
		Expected    string
		Args        []string
		ShouldError bool
	}{
		{
			Case:        "simple command",
			Command:     "echo",
			Args:        []string{},
			ReturnValue: "hello",
			Template:    `{{ cmd "echo" }}`,
			Expected:    "hello",
		},
		{
			Case:        "command with args",
			Command:     "git",
			Args:        []string{"log", "--oneline", "-1"},
			ReturnValue: "abc1234 initial commit",
			Template:    `{{ cmd "git" "log" "--oneline" "-1" }}`,
			Expected:    "abc1234 initial commit",
		},
		{
			Case:        "output is trimmed",
			Command:     "echo",
			Args:        []string{},
			ReturnValue: "  trimmed output  \n",
			Template:    `{{ cmd "echo" }}`,
			Expected:    "trimmed output",
		},
		{
			Case:        "command error returns error",
			Command:     "badcmd",
			Args:        []string{},
			ReturnValue: "",
			ReturnError: errors.New("command not found"),
			Template:    `{{ cmd "badcmd" }}`,
			ShouldError: true,
		},
	}

	for _, tc := range cases {
		e := &mock.Environment{}
		e.On("Shell").Return("foo")
		e.On("RunCommand", tc.Command, tc.Args).Return(tc.ReturnValue, tc.ReturnError)

		Cache = new(cache.Template)
		Init(e, nil, nil)

		text, err := Render(tc.Template, nil)
		if tc.ShouldError {
			assert.Error(t, err, tc.Case)
			continue
		}

		assert.NoError(t, err, tc.Case)
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
