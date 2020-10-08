package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOsInfo(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getRuntimeGOOS", nil).Return("windows")
	props := &properties{
		values:     map[Property]interface{}{Windows: "win"},
		foreground: "#fff",
		background: "#000",
	}
	osInfo := &osInfo{
		env: env,
		props: props,
	}
	want := "win"
	got := osInfo.string()
	assert.Equal(t, want, got)
}
