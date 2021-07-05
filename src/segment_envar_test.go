package main

import (
	"testing"

	"oh-my-posh/runtime"

	"github.com/stretchr/testify/assert"
)

func TestEnvvarAvailable(t *testing.T) {
	name := "HERP"
	expected := "derp"
	env := new(runtime.MockedEnvironment)
	env.On("Getenv", name).Return(expected)
	props := &properties{
		values: map[Property]interface{}{
			VarName: name,
		},
	}
	e := &envvar{
		env:   env,
		props: props,
	}
	assert.True(t, e.enabled())
	assert.Equal(t, expected, e.string())
}

func TestEnvvarNotAvailable(t *testing.T) {
	name := "HERP"
	expected := ""
	env := new(runtime.MockedEnvironment)
	env.On("Getenv", name).Return(expected)
	props := &properties{
		values: map[Property]interface{}{
			VarName: name,
		},
	}
	e := &envvar{
		env:   env,
		props: props,
	}
	assert.False(t, e.enabled())
}
