package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(MockedEnvironment)
	env.On("getShellName", nil).Return(expected, nil)
	s := &shell{
		env:   env,
		props: properties{},
	}
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
		env.On("getShellName", nil).Return(tc.Expected, nil)
		s := &shell{
			env: env,
			props: properties{
				MappedShellNames: map[string]string{"pwsh": "PS"},
			},
		}
		got := s.string()
		assert.Equal(t, tc.Expected, got)
	}
}
