package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(MockedEnvironment)
	env.On("getShellName", nil).Return(expected, nil)
	props := &properties{}
	s := &shell{
		env:   env,
		props: props,
	}
	assert.Equal(t, expected, s.string())
}
