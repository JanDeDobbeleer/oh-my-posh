package main

import (
	"testing"

	"oh-my-posh/runtime"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(runtime.MockedEnvironment)
	env.On("GetShellName", nil).Return(expected, nil)
	props := &properties{}
	s := &shell{
		env:   env,
		props: props,
	}
	assert.Equal(t, expected, s.string())
}
