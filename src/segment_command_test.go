// +build !windows

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	env := &environment{}
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
	assert.Equal(t, "hello", c.string())
}

func TestExecuteMultipleCommandsOrFirst(t *testing.T) {
	env := &environment{}
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
	assert.Equal(t, "hello", c.string())
}

func TestExecuteMultipleCommandsOrSecond(t *testing.T) {
	env := &environment{}
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
	assert.Equal(t, "hello", c.string())
}

func TestExecuteMultipleCommandsAnd(t *testing.T) {
	env := &environment{}
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
	assert.Equal(t, "helloworld", c.string())
}

func TestExecuteSingleCommandEmpty(t *testing.T) {
	env := &environment{}
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
	debug := false
	env.init(&args{
		Debug: &debug,
	})
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
