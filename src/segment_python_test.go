package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestPythonVirtualEnv(t *testing.T) {
	cases := []struct {
		Case                string
		Expected            string
		ExpectedDisabled    bool
		VirtualEnvName      string
		CondaEnvName        string
		CondaDefaultEnvName string
		PyEnvName           string
		DisplayVersion      bool
		DisplayDefault      bool
	}{
		{Case: "VENV", Expected: "VENV", VirtualEnvName: "VENV"},
		{Case: "CONDA", Expected: "CONDA", CondaEnvName: "CONDA"},
		{Case: "CONDA default", Expected: "CONDA", CondaDefaultEnvName: "CONDA"},
		{Case: "Display Base", Expected: "base", CondaDefaultEnvName: "base", DisplayDefault: true},
		{Case: "Hide base", Expected: "", CondaDefaultEnvName: "base", ExpectedDisabled: true},
		{Case: "PYENV", Expected: "PYENV", PyEnvName: "PYENV"},
		{Case: "PYENV Version", Expected: "PYENV 3.8.4", PyEnvName: "PYENV", DisplayVersion: true},
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
		env.On("getcwd", nil).Return("/usr/home/project")
		env.On("homeDir", nil).Return("/usr/home")
		var props properties = map[Property]interface{}{
			DisplayVersion:    tc.DisplayVersion,
			DisplayVirtualEnv: true,
			DisplayDefault:    tc.DisplayDefault,
		}
		python := &python{}
		python.init(props, env)
		assert.Equal(t, !tc.ExpectedDisabled, python.enabled(), tc.Case)
		assert.Equal(t, tc.Expected, python.string(), tc.Case)
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
