package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type process struct {
	mock.Mock
}

// Pid is the process ID for this process.
func (p *process) Pid() int {
	args := p.Called(nil)
	return args.Int(0)
}

// PPid is the parent process ID for this process.
func (p *process) PPid() int {
	args := p.Called(nil)
	return args.Int(0)
}

// Executable name running this process. This is not a path to the
// executable.
func (p *process) Executable() string {
	args := p.Called(nil)
	return args.String(0)
}

func TestWriteCurrentShell(t *testing.T) {
	expected := "zsh"
	env := new(MockedEnvironment)
	process := new(process)
	process.On("Executable", nil).Return(expected)
	env.On("getParentProcess", nil).Return(process, nil)
	props := &properties{}
	s := &shell{
		env:   env,
		props: props,
	}
	assert.Equal(t, "zsh", s.string())
}

func TestWriteCurrentShellError(t *testing.T) {
	err := errors.New("Oh no, what shell is this?")
	env := new(MockedEnvironment)
	process := new(process)
	env.On("getParentProcess", nil).Return(process, err)
	props := &properties{}
	s := &shell{
		env:   env,
		props: props,
	}
	assert.Equal(t, "unknown", s.string())
}
