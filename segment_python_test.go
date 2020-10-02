package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type pythonArgs struct {
	virtualEnvName   string
	condaEnvName     string
	condaDefaultName string
	pyEnvName        string
	pathSeparator    string
	pythonVersion    string
	python3Version   string
	hasFiles         bool
}

func newPythonArgs() *pythonArgs {
	return &pythonArgs{
		virtualEnvName:   "",
		condaEnvName:     "",
		condaDefaultName: "",
		pyEnvName:        "",
		pathSeparator:    "/",
		pythonVersion:    "",
		python3Version:   "",
		hasFiles:         true,
	}
}

func bootStrapPythonTest(args *pythonArgs) *python {
	env := new(MockedEnvironment)
	env.On("hasFiles", "*.py").Return(args.hasFiles)
	env.On("runCommand", "python", []string{"--version"}).Return(args.pythonVersion)
	env.On("runCommand", "python3", []string{"--version"}).Return(args.python3Version)
	env.On("getenv", "VIRTUAL_ENV").Return(args.virtualEnvName)
	env.On("getenv", "CONDA_ENV_PATH").Return(args.condaEnvName)
	env.On("getenv", "CONDA_DEFAULT_ENV").Return(args.condaDefaultName)
	env.On("getenv", "PYENV_VERSION").Return(args.pyEnvName)
	env.On("getPathSeperator", nil).Return(args.pathSeparator)
	python := &python{
		env: env,
	}
	return python
}

func TestPythonWriterDisabledNoPythonFiles(t *testing.T) {
	args := newPythonArgs()
	args.hasFiles = false
	python := bootStrapPythonTest(args)
	assert.False(t, python.enabled(), "there are no Python files in the current folder")
}

func TestPythonWriterDisabledNoPythonInstalled(t *testing.T) {
	args := newPythonArgs()
	python := bootStrapPythonTest(args)
	assert.False(t, python.enabled(), "Python isn't installed")
}

func TestPythonWriterEnabledNoVirtualEnv(t *testing.T) {
	args := newPythonArgs()
	args.python3Version = "3.4.5"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, args.python3Version, python.string())
}

func TestPythonWriterEnabledVirtualEnvOverrule(t *testing.T) {
	args := newPythonArgs()
	args.python3Version = "3.4.5"
	args.condaEnvName = "myenv"
	props := &properties{
		values: map[Property]interface{}{
			DisplayVirtualEnv: false,
		},
	}
	python := bootStrapPythonTest(args)
	python.props = props
	assert.True(t, python.enabled())
	assert.Equal(t, args.python3Version, python.string())
}

func TestPythonWriterEnabledVirtualEnv(t *testing.T) {
	args := newPythonArgs()
	args.python3Version = "3.4.5"
	args.condaEnvName = "myenv"
	expected := fmt.Sprintf("%s %s", args.condaEnvName, args.python3Version)
	props := &properties{
		values: map[Property]interface{}{
			DisplayVirtualEnv: true,
		},
	}
	python := bootStrapPythonTest(args)
	python.props = props
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithVirtualEnv(t *testing.T) {
	args := newPythonArgs()
	args.virtualEnvName = "venv"
	args.python3Version = "3.4.5"
	expected := fmt.Sprintf("%s %s", args.virtualEnvName, args.python3Version)
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithCondaEnvPath(t *testing.T) {
	args := newPythonArgs()
	args.condaEnvName = "conda"
	args.python3Version = "3.4.5"
	expected := fmt.Sprintf("%s %s", args.condaEnvName, args.python3Version)
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithCondaDefaultEnv(t *testing.T) {
	args := newPythonArgs()
	args.condaDefaultName = "conda2"
	args.python3Version = "3.4.5"
	expected := fmt.Sprintf("%s %s", args.condaDefaultName, args.python3Version)
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithTwoValidEnvs(t *testing.T) {
	args := newPythonArgs()
	args.condaEnvName = "conda"
	args.condaDefaultName = "conda2"
	args.python3Version = "3.4.5"
	expected := fmt.Sprintf("%s %s", args.condaEnvName, args.python3Version)
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterNameTrailingSlash(t *testing.T) {
	args := newPythonArgs()
	args.virtualEnvName = "python/"
	args.pythonVersion = "2.7.3"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, "python", python.venvName)
}
