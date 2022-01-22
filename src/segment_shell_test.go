package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(MockedEnvironment)
	env.On("getShellName").Return(expected, nil)
	env.onTemplate()
	s := &shell{
		env:   env,
		props: properties{},
	}
	_ = s.enabled()
	assert.Equal(t, expected, s.string())
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
		env.On("getShellName").Return(tc.Expected, nil)
		env.onTemplate()
		s := &shell{
			env: env,
			props: properties{
				MappedShellNames: map[string]string{"pwsh": "PS"},
			},
		}
		_ = s.enabled()
		got := s.string()
		assert.Equal(t, tc.Expected, got)
	}
}
