package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type nvmArgs struct {
	enabled     bool
	nodeVersion string
}

func bootStrapNVMTest(args *nvmArgs) *nvm {
	env := new(MockedEnvironment)
	env.On("hasCommand", "node").Return(args.enabled)
	env.On("runCommand", "node", []string{"--version"}).Return(args.nodeVersion)
	nvm := &nvm{
		env: env,
	}
	return nvm
}

func TestNVMWriterDisabled(t *testing.T) {
	args := &nvmArgs{
		enabled: false,
	}
	nvm := bootStrapNVMTest(args)
	assert.False(t, nvm.enabled(), "the nvm command is not available")
}

func TestNVMWriterEnabled(t *testing.T) {
	expected := "1.14"
	args := &nvmArgs{
		enabled:     true,
		nodeVersion: expected,
	}
	nvm := bootStrapNVMTest(args)
	assert.True(t, nvm.enabled(), "the nvm command is available")
	assert.Equal(t, expected, nvm.string(), "the nvm command is available")
}
