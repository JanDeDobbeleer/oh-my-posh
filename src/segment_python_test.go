package main

import (
	"testing"

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
		DisplayDefaultEnv   bool
	}{
		{Expected: "VENV", VirtualEnvName: "VENV"},
		{Expected: "CONDA", CondaEnvName: "CONDA"},
		{Expected: "CONDA", CondaDefaultEnvName: "CONDA"},
		{Expected: "", CondaDefaultEnvName: "base"},
		{Expected: "base", CondaDefaultEnvName: "base", DisplayDefaultEnv: true},
		{Expected: "PYENV", PyEnvName: "PYENV"},
		{Expected: "PYENV 3.8.4", PyEnvName: "PYENV", DisplayVersion: true},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("hasCommand", "python").Return(true)
		env.On("runCommand", "python", []string{"--version"}).Return("Python 3.8.4", nil)
		env.On("hasFiles", "*.py").Return(true)
		env.On("getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("getenv", "CONDA_ENV_PATH").Return(tc.CondaEnvName)
		env.On("getenv", "CONDA_DEFAULT_ENV").Return(tc.CondaDefaultEnvName)
		env.On("getenv", "PYENV_VERSION").Return(tc.PyEnvName)
		env.On("getPathSeperator", nil).Return("")
		props := &properties{
			values: map[Property]interface{}{
				DisplayVersion:    tc.DisplayVersion,
				DisplayVirtualEnv: true,
				DisplayDefaultEnv: tc.DisplayDefaultEnv,
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
		env := new(MockedEnvironment)
		env.On("getPathSeperator", nil).Return("")
		env.On("getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("getenv", "CONDA_ENV_PATH").Return("")
		env.On("getenv", "CONDA_DEFAULT_ENV").Return("")
		env.On("getenv", "PYENV_VERSION").Return("")
		python := &python{}
		python.init(nil, env)
		python.loadContext()
		assert.Equal(t, tc.Expected, python.inContext())
	}
}
