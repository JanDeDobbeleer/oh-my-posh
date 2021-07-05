package main

import (
	"testing"

	"oh-my-posh/runtime"

	"github.com/stretchr/testify/assert"
)

type dotnetArgs struct {
	enabled         bool
	version         string
	exitCode        int
	unsupportedIcon string
	displayVersion  bool
}

func bootStrapDotnetTest(args *dotnetArgs) *dotnet {
	env := new(runtime.MockedEnvironment)
	env.On("HasCommand", "dotnet").Return(args.enabled)
	if args.exitCode != 0 {
		err := &runtime.CommandError{ExitCode: args.exitCode}
		env.On("RunCommand", "dotnet", []string{"--version"}).Return("", err)
	} else {
		env.On("RunCommand", "dotnet", []string{"--version"}).Return(args.version, nil)
	}

	env.On("HasFiles", "*.cs").Return(true)
	env.On("GetPathSeperator", nil).Return("")
	env.On("Getcwd", nil).Return("/usr/home/project")
	env.On("HomeDir", nil).Return("/usr/home")
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
	args := &dotnetArgs{
		enabled:         true,
		displayVersion:  true,
		exitCode:        dotnetExitCode,
		unsupportedIcon: expected,
	}
	dotnet := bootStrapDotnetTest(args)
	assert.True(t, dotnet.enabled())
	assert.Equal(t, expected, dotnet.string())
}
