package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	props := properties{
		Command: "echo hello",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", renderTemplate(env, c.template(), c))
}

func TestExecuteMultipleCommandsOrFirst(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "exit 1").Return("")
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	env.On("RunShellCommand", "bash", "exit 1 || echo hello").Return("hello")
	props := properties{
		Command: "exit 1 || echo hello",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", renderTemplate(env, c.template(), c))
}

func TestExecuteMultipleCommandsOrSecond(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	env.On("RunShellCommand", "bash", "echo world").Return("world")
	props := properties{
		Command: "echo hello || echo world",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", renderTemplate(env, c.template(), c))
}

func TestExecuteMultipleCommandsAnd(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	env.On("RunShellCommand", "bash", "echo world").Return("world")
	props := properties{
		Command: "echo hello && echo world",
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "helloworld", renderTemplate(env, c.template(), c))
}

func TestExecuteSingleCommandEmpty(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "").Return("")
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
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo no command specified").Return("no command specified")
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
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo").Return("")
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
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo").Return("")
	env.On("RunShellCommand", "bash", "echo|| echo").Return("")
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
