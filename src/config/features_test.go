package config

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/stretchr/testify/assert"
)

func TestFeatures(t *testing.T) {
	cases := []struct {
		Case     string
		Shell    string
		Daemon   bool
		Async    bool
		Expected shell.Features
	}{
		{
			Case:     "Daemon enabled, supported shell",
			Shell:    shell.ZSH,
			Daemon:   true,
			Expected: shell.Daemon,
		},
		{
			Case:     "Daemon enabled, unsupported shell",
			Shell:    shell.CMD,
			Daemon:   true,
			Expected: 0,
		},
		{
			Case:     "Daemon disabled",
			Shell:    shell.ZSH,
			Daemon:   false,
			Expected: 0,
		},
		{
			Case:     "Async enabled",
			Shell:    shell.ZSH,
			Async:    true,
			Expected: shell.Async,
		},
		{
			Case:     "Async enabled, unsupported shell",
			Shell:    shell.CMD,
			Async:    true,
			Expected: 0,
		},
		{
			Case:     "Async and Daemon enabled",
			Shell:    shell.ZSH,
			Async:    true,
			Daemon:   true,
			Expected: shell.Async | shell.Daemon,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		env.On("Shell").Return(tc.Shell)

		cfg := &Config{
			Async:   tc.Async,
			Upgrade: &upgrade.Config{},
		}

		got := cfg.Features(env, tc.Daemon)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
