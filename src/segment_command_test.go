package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "echo hello").Return("hello")
	env.onTemplate()
	props := properties{
		Command: "echo hello",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", c.string())
}

func TestExecuteMultipleCommandsOrFirst(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "exit 1").Return("")
	env.On("runShellCommand", "bash", "echo hello").Return("hello")
	env.On("runShellCommand", "bash", "exit 1 || echo hello").Return("hello")
	env.onTemplate()
	props := properties{
		Command: "exit 1 || echo hello",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", c.string())
}

func TestExecuteMultipleCommandsOrSecond(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "echo hello").Return("hello")
	env.On("runShellCommand", "bash", "echo world").Return("world")
	env.onTemplate()
	props := properties{
		Command: "echo hello || echo world",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", c.string())
}

func TestExecuteMultipleCommandsAnd(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "echo hello").Return("hello")
	env.On("runShellCommand", "bash", "echo world").Return("world")
	env.onTemplate()
	props := properties{
		Command: "echo hello && echo world",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "helloworld", c.string())
}

func TestExecuteSingleCommandEmpty(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "").Return("")
	env.onTemplate()
	props := properties{
		Command: "",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.False(t, enabled)
}

func TestExecuteSingleCommandNoCommandProperty(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "echo no command specified").Return("no command specified")
	env.onTemplate()
	var props properties
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "no command specified", c.Output)
}

func TestExecuteMultipleCommandsAndDisabled(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "echo").Return("")
	env.onTemplate()
	props := properties{
		Command: "echo && echo",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.False(t, enabled)
}

func TestExecuteMultipleCommandsOrDisabled(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "bash").Return(true)
	env.On("runShellCommand", "bash", "echo").Return("")
	env.On("runShellCommand", "bash", "echo|| echo").Return("")
	env.onTemplate()
	props := properties{
		Command: "echo|| echo",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.False(t, enabled)
}
