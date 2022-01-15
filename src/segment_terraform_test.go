package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type terraformArgs struct {
	hasTfCommand  bool
	hasTfFolder   bool
	workspaceName string
}

func bootStrapTerraformTest(args *terraformArgs) *terraform {
	env := new(MockedEnvironment)
	env.On("hasCommand", "terraform").Return(args.hasTfCommand)
	env.On("hasFolder", "/.terraform").Return(args.hasTfFolder)
	env.On("getcwd").Return("")
	env.On("runCommand", "terraform", []string{"workspace", "show"}).Return(args.workspaceName, nil)
	env.onTemplate()
	k := &terraform{
		env:   env,
		props: properties{},
	}
	return k
}

func TestTerraformWriterDisabled(t *testing.T) {
	args := &terraformArgs{
		hasTfCommand: false,
		hasTfFolder:  false,
	}
	terraform := bootStrapTerraformTest(args)
	assert.False(t, terraform.enabled())
}

func TestTerraformMissingDir(t *testing.T) {
	args := &terraformArgs{
		hasTfCommand: true,
		hasTfFolder:  false,
	}
	terraform := bootStrapTerraformTest(args)
	assert.False(t, terraform.enabled())
}

func TestTerraformMissingBinary(t *testing.T) {
	args := &terraformArgs{
		hasTfCommand: false,
		hasTfFolder:  true,
	}
	terraform := bootStrapTerraformTest(args)
	assert.False(t, terraform.enabled())
}

func TestTerraformEnabled(t *testing.T) {
	expected := "default"
	args := &terraformArgs{
		hasTfCommand:  true,
		hasTfFolder:   true,
		workspaceName: expected,
	}
	terraform := bootStrapTerraformTest(args)
	assert.True(t, terraform.enabled())
	assert.Equal(t, expected, terraform.string())
}
