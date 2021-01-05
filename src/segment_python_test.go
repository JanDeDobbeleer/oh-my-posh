package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

type pythonArgs struct {
	virtualEnvName   string
	condaEnvName     string
	condaDefaultName string
	pyEnvName        string
	displayVersion   bool
}

func bootStrapPythonTest(args *pythonArgs) *python {
	env := new(MockedEnvironment)
	env.On("hasCommand", "python").Return(true)
	env.On("runCommand", "python", []string{"--version"}).Return("Python 3.8.4", nil)
	env.On("hasFiles", "*.py").Return(true)
	env.On("getenv", "VIRTUAL_ENV").Return(args.virtualEnvName)
	env.On("getenv", "CONDA_ENV_PATH").Return(args.condaEnvName)
	env.On("getenv", "CONDA_DEFAULT_ENV").Return(args.condaDefaultName)
	env.On("getenv", "PYENV_VERSION").Return(args.pyEnvName)
	env.On("getPathSeperator", nil).Return("")
	props := &properties{
		values: map[Property]interface{}{
			DisplayVersion:    args.displayVersion,
			DisplayVirtualEnv: true,
		},
	}
	python := &python{}
	python.init(props, env)
	return python
}

func TestPythonVertualEnv(t *testing.T) {
	expected := "VENV"
	args := &pythonArgs{
		virtualEnvName: expected,
	}
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonCondaEnv(t *testing.T) {
	expected := "CONDA"
	args := &pythonArgs{
		condaEnvName: expected,
	}
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonCondaDefaultName(t *testing.T) {
	expected := "CONDADEF"
	args := &pythonArgs{
		condaDefaultName: expected,
	}
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonPyEnv(t *testing.T) {
	expected := "PYENV"
	args := &pythonArgs{
		pyEnvName: expected,
	}
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonPyEnvWithVersion(t *testing.T) {
	expected := "PYENV 3.8.4"
	args := &pythonArgs{
		pyEnvName:      "PYENV",
		displayVersion: true,
	}
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
	assert.Equal(t, "3.8.4", python.language.version)
}
