package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(MockedEnvironment)
	env.On("Shell").Return(expected, nil)
	s := &shell{
		env:   env,
		props: properties{},
	}
	_ = s.enabled()
	assert.Equal(t, expected, renderTemplate(env, s.template(), s))
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
		env := new(MockedEnvironment)
		env.On("Shell").Return(tc.Expected, nil)
		s := &shell{
			env: env,
			props: properties{
				MappedShellNames: map[string]string{"pwsh": "PS"},
			},
		}
		_ = s.enabled()
		got := renderTemplate(env, s.template(), s)
		assert.Equal(t, tc.Expected, got)
	}
}
