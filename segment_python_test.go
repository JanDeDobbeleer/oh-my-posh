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
	hasPyFiles       bool
	hasNotebookFiles bool
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
		hasPyFiles:       true,
		hasNotebookFiles: true,
	}
}

func bootStrapPythonTest(args *pythonArgs) *python {
	env := new(MockedEnvironment)
	env.On("hasFiles", "*.py").Return(args.hasPyFiles)
	env.On("hasFiles", "*.ipynb").Return(args.hasNotebookFiles)
	env.On("runCommand", "python", []string{"--version"}).Return(args.pythonVersion, nil)
	env.On("runCommand", "python3", []string{"--version"}).Return(args.python3Version, nil)
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
	args.hasPyFiles = false
	args.hasNotebookFiles = false
	args.python3Version = "Python 3.4.5"
	python := bootStrapPythonTest(args)
	assert.False(t, python.enabled(), "there are no Python files in the current folder")
}

func TestPythonWriterDisabledHasPythonFiles(t *testing.T) {
	args := newPythonArgs()
	args.hasPyFiles = true
	args.hasNotebookFiles = false
	args.python3Version = "Python 3.4.5"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled(), "there should be a Python file in the current folder")
}

func TestPythonWriterDisabledHasJupyterNotebookFiles(t *testing.T) {
	args := newPythonArgs()
	args.hasPyFiles = false
	args.hasNotebookFiles = true
	args.python3Version = "Python 3.4.5"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled(), "there should be a Jupyter Notebook file in the current folder")
}

func TestPythonWriterDisabledHasPyAndJupyterNotebookFiles(t *testing.T) {
	args := newPythonArgs()
	args.hasPyFiles = true
	args.hasNotebookFiles = true
	args.python3Version = "Python 3.4.5"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled(), "there should be a Jupyter Notebook file in the current folder")
}

func TestPythonWriterDisabledHasPyAndJupyterNotebookFilesButNoVersion(t *testing.T) {
	args := newPythonArgs()
	args.hasPyFiles = true
	args.hasNotebookFiles = true
	python := bootStrapPythonTest(args)
	assert.False(t, python.enabled(), "there should be a Jupyter Notebook file in the current folder")
}

func TestPythonWriterDisabledNoPythonInstalled(t *testing.T) {
	args := newPythonArgs()
	python := bootStrapPythonTest(args)
	assert.False(t, python.enabled(), "Python isn't installed")
}

func TestPythonWriterEnabledNoVirtualEnv(t *testing.T) {
	args := newPythonArgs()
	args.python3Version = "Python 3.4.5"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, "3.4.5", python.string())
}

func TestPythonWriterEnabledVirtualEnvOverrule(t *testing.T) {
	args := newPythonArgs()
	args.python3Version = "Python 3.4.5"
	args.condaEnvName = "myenv"
	props := &properties{
		values: map[Property]interface{}{
			DisplayVirtualEnv: false,
		},
	}
	python := bootStrapPythonTest(args)
	python.props = props
	assert.True(t, python.enabled())
	assert.Equal(t, "3.4.5", python.string())
}

func TestPythonWriterEnabledVirtualEnv(t *testing.T) {
	args := newPythonArgs()
	args.python3Version = "Python 3.4.5"
	args.condaEnvName = "myenv"
	expected := fmt.Sprintf("%s %s", args.condaEnvName, "3.4.5")
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
	args.python3Version = "Python 3.4.5"
	expected := fmt.Sprintf("%s %s", args.virtualEnvName, "3.4.5")
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithCondaEnvPath(t *testing.T) {
	args := newPythonArgs()
	args.condaEnvName = "conda"
	args.python3Version = "Python 3.4.5 something off about this one"
	expected := fmt.Sprintf("%s %s", args.condaEnvName, "3.4.5")
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithCondaDefaultEnv(t *testing.T) {
	args := newPythonArgs()
	args.condaDefaultName = "conda2"
	args.python3Version = "Python 3.4.5"
	expected := fmt.Sprintf("%s %s", args.condaDefaultName, "3.4.5")
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithCondaDefaultEnvAnacondaInc(t *testing.T) {
	args := newPythonArgs()
	args.condaDefaultName = "flatland_rl"
	args.pythonVersion = "Python 3.6.8 :: Anaconda, Inc."
	expected := "flatland_rl 3.6.8"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterEnabledWithTwoValidEnvs(t *testing.T) {
	args := newPythonArgs()
	args.condaEnvName = "conda"
	args.condaDefaultName = "conda2"
	args.python3Version = "Python 3.4.5"
	expected := fmt.Sprintf("%s %s", args.condaEnvName, "3.4.5")
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, expected, python.string())
}

func TestPythonWriterNameTrailingSlash(t *testing.T) {
	args := newPythonArgs()
	args.virtualEnvName = "python/"
	args.pythonVersion = "Python 2.7.3"
	python := bootStrapPythonTest(args)
	assert.True(t, python.enabled())
	assert.Equal(t, "python", python.venvName)
}
