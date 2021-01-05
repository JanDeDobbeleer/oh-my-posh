package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type dotnetArgs struct {
	enabled         bool
	version         string
	unsupported     bool
	unsupportedIcon string
	displayVersion  bool
}

func bootStrapDotnetTest(args *dotnetArgs) *dotnet {
	env := new(MockedEnvironment)
	env.On("hasCommand", "dotnet").Return(args.enabled)
	if args.unsupported {
		err := &commandError{exitCode: 145}
		env.On("runCommand", "dotnet", []string{"--version"}).Return("", err)
	} else {
		env.On("runCommand", "dotnet", []string{"--version"}).Return(args.version, nil)
	}

	env.On("hasFiles", "*.cs").Return(true)
	env.On("getPathSeperator", nil).Return("")
	props := &properties{
		values: map[Property]interface{}{
			DisplayVersion:               args.displayVersion,
			UnsupportedDotnetVersionIcon: args.unsupportedIcon,
		},
	}
	dotnet := &dotnet{}
	dotnet.init(props, env)
	return dotnet
}

func TestEnabledDotnetNotFound(t *testing.T) {
	args := &dotnetArgs{
		enabled: false,
	}
	dotnet := bootStrapDotnetTest(args)
	assert.True(t, dotnet.enabled())
}

func TestDotnetVersionNotDisplayed(t *testing.T) {
	args := &dotnetArgs{
		enabled:        true,
		displayVersion: false,
		version:        "3.1.402",
	}
	dotnet := bootStrapDotnetTest(args)
	assert.True(t, dotnet.enabled())
	assert.Equal(t, "", dotnet.string())
}

func TestDotnetVersionDisplayed(t *testing.T) {
	expected := "3.1.402"
	args := &dotnetArgs{
		enabled:        true,
		displayVersion: true,
		version:        expected,
	}
	dotnet := bootStrapDotnetTest(args)
	assert.True(t, dotnet.enabled())
	assert.Equal(t, expected, dotnet.string())
}

func TestDotnetVersionUnsupported(t *testing.T) {
	expected := "x"
	args := &dotnetArgs{
		enabled:         true,
		displayVersion:  true,
		unsupported:     true,
		unsupportedIcon: expected,
	}
	dotnet := bootStrapDotnetTest(args)
	assert.True(t, dotnet.enabled())
	assert.Equal(t, expected, dotnet.string())
}
