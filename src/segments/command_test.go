package segments

import (
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	props := properties.Map{
		Command: "echo hello",
	}
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", renderTemplate(env, c.Template(), c))
}

func TestExecuteMultipleCommandsOrFirst(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "exit 1").Return("")
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	env.On("RunShellCommand", "bash", "exit 1 || echo hello").Return("hello")
	props := properties.Map{
		Command: "exit 1 || echo hello",
	}
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", renderTemplate(env, c.Template(), c))
}

func TestExecuteMultipleCommandsOrSecond(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	env.On("RunShellCommand", "bash", "echo world").Return("world")
	props := properties.Map{
		Command: "echo hello || echo world",
	}
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.True(t, enabled)
	assert.Equal(t, "hello", renderTemplate(env, c.Template(), c))
}

func TestExecuteMultipleCommandsAnd(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo hello").Return("hello")
	env.On("RunShellCommand", "bash", "echo world").Return("world")
	props := properties.Map{
		Command: "echo hello && echo world",
	}
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.True(t, enabled)
	assert.Equal(t, "helloworld", renderTemplate(env, c.Template(), c))
}

func TestExecuteSingleCommandEmpty(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "").Return("")
	props := properties.Map{
		Command: "",
	}
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.False(t, enabled)
}

func TestExecuteSingleCommandNoCommandProperty(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo no command specified").Return("no command specified")
	var props properties.Map
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.True(t, enabled)
	assert.Equal(t, "no command specified", c.Output)
}

func TestExecuteMultipleCommandsAndDisabled(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo").Return("")
	props := properties.Map{
		Command: "echo && echo",
	}
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.False(t, enabled)
}

func TestExecuteMultipleCommandsOrDisabled(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "bash").Return(true)
	env.On("RunShellCommand", "bash", "echo").Return("")
	env.On("RunShellCommand", "bash", "echo|| echo").Return("")
	props := properties.Map{
		Command: "echo|| echo",
	}
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.False(t, enabled)
}
