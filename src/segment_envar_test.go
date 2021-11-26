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
	e := &envvar{
		env: env,
		props: map[Property]interface{}{
			VarName: name,
		},
	}
	assert.True(t, e.enabled())
	assert.Equal(t, expected, e.string())
}

func TestEnvvarNotAvailable(t *testing.T) {
	name := "HERP"
	expected := ""
	env := new(MockedEnvironment)
	env.On("getenv", name).Return(expected)
	e := &envvar{
		env: env,
		props: map[Property]interface{}{
			VarName: name,
		},
	}
	assert.False(t, e.enabled())
}
