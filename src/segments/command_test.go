package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/mock"
	"github.com/jandedobbeleer/oh-my-posh/properties"

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
	env.On("RunShellCommand", "bash", "").Return("")
	var props properties.Map
	c := &Cmd{
		props: props,
		env:   env,
	}
	enabled := c.Enabled()
	assert.False(t, enabled)
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

func TestExecuteScript(t *testing.T) {
	cases := []struct {
		Case            string
		Output          string
		HasScript       bool
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "Output",
			Output:          "Hello World",
			ExpectedString:  "Hello World",
			ExpectedEnabled: true,
		},
		{
			Case:            "No output",
			ExpectedEnabled: false,
		},
	}
	for _, tc := range cases {
		script := "../test/script.sh"
		env := new(mock.MockedEnvironment)
		env.On("HasCommand", "bash").Return(true)
		env.On("RunShellCommand", "bash", script).Return(tc.Output)
		props := properties.Map{
			Script: script,
		}
		c := &Cmd{
			props: props,
			env:   env,
		}
		enabled := c.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, c.Template(), c))
		}
	}
}
