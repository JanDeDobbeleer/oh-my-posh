package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	env := &environment{}
	props := &properties{
		values: map[Property]interface{}{
			Command: "echo hello",
		},
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, c.string(), "hello")
}

func TestExecuteMultipleCommandsOrFirst(t *testing.T) {
	env := &environment{}
	props := &properties{
		values: map[Property]interface{}{
			Command: "exit 1 || echo hello",
		},
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, c.string(), "hello")
}

func TestExecuteMultipleCommandsOrSecond(t *testing.T) {
	env := &environment{}
	props := &properties{
		values: map[Property]interface{}{
			Command: "echo hello || echo world",
		},
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, c.string(), "hello")
}

func TestExecuteMultipleCommandsAnd(t *testing.T) {
	env := &environment{}
	props := &properties{
		values: map[Property]interface{}{
			Command: "echo hello && echo world",
		},
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, c.string(), "helloworld")
}

func TestExecuteSingleCommandEmpty(t *testing.T) {
	env := &environment{}
	props := &properties{
		values: map[Property]interface{}{
			Command: "",
		},
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.False(t, enabled)
}

func TestExecuteSingleCommandNoCommandProperty(t *testing.T) {
	env := &environment{}
	props := &properties{}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.True(t, enabled)
	assert.Equal(t, "no command specified", c.value)
}

func TestExecuteMultipleCommandsAndDisabled(t *testing.T) {
	env := &environment{}
	props := &properties{
		values: map[Property]interface{}{
			Command: "echo && echo",
		},
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.False(t, enabled)
}

func TestExecuteMultipleCommandsOrDisabled(t *testing.T) {
	env := &environment{}
	props := &properties{
		values: map[Property]interface{}{
			Command: "echo|| echo",
		},
	}
	c := &command{
		props: props,
		env:   env,
	}
	enabled := c.enabled()
	assert.False(t, enabled)
}
