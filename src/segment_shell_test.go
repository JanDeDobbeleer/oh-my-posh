package main

import (
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(mock.MockedEnvironment)
	env.On("Shell").Return(expected, nil)
	s := &Shell{
		env:   env,
		props: properties.Map{},
	}
	_ = s.Enabled()
	assert.Equal(t, expected, renderTemplate(env, s.Template(), s))
}

func TestUseMappedShellNames(t *testing.T) {
	cases := []struct {
		Shell    string
		Expected string
	}{
		{Shell: "zsh", Expected: "zsh"},
		{Shell: "pwsh", Expected: "PS"},
		{Shell: "PWSH", Expected: "PS"},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Shell").Return(tc.Expected, nil)
		s := &Shell{
			env: env,
			props: properties.Map{
				MappedShellNames: map[string]string{"pwsh": "PS"},
			},
		}
		_ = s.Enabled()
		got := renderTemplate(env, s.Template(), s)
		assert.Equal(t, tc.Expected, got)
	}
}
