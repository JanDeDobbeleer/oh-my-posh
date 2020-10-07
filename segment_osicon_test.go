package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOsIcon(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getRuntimeGOOS", nil).Return("windows")
	osicon := &osicon{env: env}
	want := "\uF17A"
	got := osicon.string()
	assert.Equal(t, want, got)
}
