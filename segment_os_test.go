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
		env:   env,
		props: props,
	}
	want := "win"
	got := osInfo.string()
	assert.Equal(t, want, got)
}

func TestWSL(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("getRuntimeGOOS", nil).Return("linux")
	env.On("getenv", "WSL_DISTRO_NAME").Return("debian")
	env.On("getPlatform", nil).Return("debian")
	props := &properties{
		values: map[Property]interface{}{
			WSL:          "WSL TEST",
			WSLSeparator: " @ ",
		},
	}
	osInfo := &osInfo{
		env:   env,
		props: props,
	}
	want := "WSL TEST @ \uF306"
	got := osInfo.string()
	assert.Equal(t, want, got)
}
