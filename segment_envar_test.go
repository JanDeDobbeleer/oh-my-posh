package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvvarAvailable(t *testing.T) {
	name := "HERP"
	expected := "derp"
	env := new(MockedEnvironment)
	env.On("getenv", name).Return(expected)
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
	env := new(MockedEnvironment)
	env.On("getenv", name).Return(expected)
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
