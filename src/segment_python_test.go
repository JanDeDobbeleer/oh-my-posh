package main

import (
	"testing"

	"oh-my-posh/runtime"

	"github.com/alecthomas/assert"
)

func TestPythonVirtualEnv(t *testing.T) {
	cases := []struct {
		Expected            string
		VirtualEnvName      string
		CondaEnvName        string
		CondaDefaultEnvName string
		PyEnvName           string
		DisplayVersion      bool
		DisplayDefault      bool
	}{
		{Expected: "VENV", VirtualEnvName: "VENV"},
		{Expected: "CONDA", CondaEnvName: "CONDA"},
		{Expected: "CONDA", CondaDefaultEnvName: "CONDA"},
		{Expected: "", CondaDefaultEnvName: "base"},
		{Expected: "base", CondaDefaultEnvName: "base", DisplayDefault: true},
		{Expected: "PYENV", PyEnvName: "PYENV"},
		{Expected: "PYENV 3.8.4", PyEnvName: "PYENV", DisplayVersion: true},
	}

	for _, tc := range cases {
		env := new(runtime.MockedEnvironment)
		env.On("HasCommand", "python").Return(true)
		env.On("RunCommand", "python", []string{"--version"}).Return("Python 3.8.4", nil)
		env.On("HasFiles", "*.py").Return(true)
		env.On("Getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_ENV_PATH").Return(tc.CondaEnvName)
		env.On("Getenv", "CONDA_DEFAULT_ENV").Return(tc.CondaDefaultEnvName)
		env.On("Getenv", "PYENV_VERSION").Return(tc.PyEnvName)
		env.On("GetPathSeperator", nil).Return("")
		env.On("Getcwd", nil).Return("/usr/home/project")
		env.On("HomeDir", nil).Return("/usr/home")
		props := &properties{
			values: map[Property]interface{}{
				DisplayVersion:    tc.DisplayVersion,
				DisplayVirtualEnv: true,
				DisplayDefault:    tc.DisplayDefault,
			},
		}
		python := &python{}
		python.init(props, env)
		assert.True(t, python.enabled())
		assert.Equal(t, tc.Expected, python.string())
	}
}

func TestPythonPythonInContext(t *testing.T) {
	cases := []struct {
		Expected       bool
		VirtualEnvName string
	}{
		{Expected: true, VirtualEnvName: "VENV"},
		{Expected: false, VirtualEnvName: ""},
	}

	for _, tc := range cases {
		env := new(runtime.MockedEnvironment)
		env.On("GetPathSeperator", nil).Return("")
		env.On("Getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("Getenv", "CONDA_ENV_PATH").Return("")
		env.On("Getenv", "CONDA_DEFAULT_ENV").Return("")
		env.On("Getenv", "PYENV_VERSION").Return("")
		python := &python{}
		python.init(nil, env)
		python.loadContext()
		assert.Equal(t, tc.Expected, python.inContext())
	}
}
