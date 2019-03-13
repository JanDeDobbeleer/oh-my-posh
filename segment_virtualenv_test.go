package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type venvArgs struct {
	virtualEnvName   string
	condaEnvName     string
	condaDefaultName string
	pathSeparator    string
}

func newVenvArgs() *venvArgs {
	return &venvArgs{
		virtualEnvName:   "",
		condaEnvName:     "",
		condaDefaultName: "",
		pathSeparator:    "/",
	}
}

func bootStrapVenvTest(args *venvArgs) *venv {
	env := new(MockedEnvironment)
	env.On("getenv", "VIRTUAL_ENV").Return(args.virtualEnvName)
	env.On("getenv", "CONDA_ENV_PATH").Return(args.condaEnvName)
	env.On("getenv", "CONDA_DEFAULT_ENV").Return(args.condaDefaultName)
	env.On("getPathSeperator", nil).Return(args.pathSeparator)
	venv := &venv{
		env: env,
	}
	return venv
}

func TestVenvWriterDisabled(t *testing.T) {
	args := newVenvArgs()
	venv := bootStrapVenvTest(args)
	assert.False(t, venv.enabled(), "the virtualenv has no name")
}

func TestVenvWriterEnabledWithVirtualEnv(t *testing.T) {
	args := newVenvArgs()
	args.virtualEnvName = "venv"
	venv := bootStrapVenvTest(args)
	assert.True(t, venv.enabled(), "the virtualenv has a name")
}

func TestVenvWriterEnabledWithCondaEnvPath(t *testing.T) {
	args := newVenvArgs()
	args.condaEnvName = "venv"
	venv := bootStrapVenvTest(args)
	assert.True(t, venv.enabled(), "the virtualenv has a name")
}

func TestVenvWriterEnabledWithCondaDefaultEnv(t *testing.T) {
	args := newVenvArgs()
	args.condaDefaultName = "venv"
	venv := bootStrapVenvTest(args)
	assert.True(t, venv.enabled(), "the virtualenv has a name")
}

func TestVenvWriterEnabledWithTwoValidEnvs(t *testing.T) {
	args := newVenvArgs()
	args.virtualEnvName = "venv"
	args.condaDefaultName = "venv"
	venv := bootStrapVenvTest(args)
	assert.True(t, venv.enabled(), "the virtualenv has a name")
}

func TestVenvWriterNameWithVirtualEnv(t *testing.T) {
	args := newVenvArgs()
	args.virtualEnvName = "venv"
	venv := bootStrapVenvTest(args)
	_ = venv.enabled()
	assert.Equal(t, "venv", venv.venvName)
}

func TestVenvWriterNameWithCondaEnvPath(t *testing.T) {
	args := newVenvArgs()
	args.condaEnvName = "venv"
	venv := bootStrapVenvTest(args)
	_ = venv.enabled()
	assert.Equal(t, "venv", venv.venvName)
}

func TestVenvWriterNameWithCondaDefaultEnv(t *testing.T) {
	args := newVenvArgs()
	args.condaDefaultName = "venv"
	venv := bootStrapVenvTest(args)
	_ = venv.enabled()
	assert.Equal(t, "venv", venv.venvName)
}

func TestVenvWriterNameTrailingSlash(t *testing.T) {
	args := newVenvArgs()
	args.virtualEnvName = "venv/"
	venv := bootStrapVenvTest(args)
	_ = venv.enabled()
	assert.Equal(t, "venv", venv.venvName)
}
