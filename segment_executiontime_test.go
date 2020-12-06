package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutionTimeWriterDefaultThresholdEnabled(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("executionTime", nil).Return(1337)
	executionTime := &executiontime{
		env: env,
	}
	assert.True(t, executionTime.enabled())
}

func TestExecutionTimeWriterDefaultThresholdDisabled(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("executionTime", nil).Return(1)
	executionTime := &executiontime{
		env: env,
	}
	assert.False(t, executionTime.enabled())
}

func TestExecutionTimeWriterCustomThresholdEnabled(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("executionTime", nil).Return(99)
	props := &properties{
		values: map[Property]interface{}{
			ThresholdProperty: float64(10),
		},
	}
	executionTime := &executiontime{
		env:   env,
		props: props,
	}
	assert.True(t, executionTime.enabled())
}

func TestExecutionTimeWriterCustomThresholdDisabled(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("executionTime", nil).Return(99)
	props := &properties{
		values: map[Property]interface{}{
			ThresholdProperty: float64(100),
		},
	}
	executionTime := &executiontime{
		env:   env,
		props: props,
	}
	assert.False(t, executionTime.enabled())
}

func TestExecutionTimeWriterDuration(t *testing.T) {
	input := 1337
	expected := "1.337s"
	env := new(MockedEnvironment)
	env.On("executionTime", nil).Return(input)
	executionTime := &executiontime{
		env: env,
	}
	executionTime.enabled()
	assert.Equal(t, expected, executionTime.output)
}

func TestExecutionTimeWriterDuration2(t *testing.T) {
	input := 13371337
	expected := "3h42m51.337s"
	env := new(MockedEnvironment)
	env.On("executionTime", nil).Return(input)
	executionTime := &executiontime{
		env: env,
	}
	executionTime.enabled()
	assert.Equal(t, expected, executionTime.output)
}
