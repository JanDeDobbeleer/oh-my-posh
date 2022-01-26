package main

import (
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

type terraformArgs struct {
	hasTfCommand  bool
	hasTfFolder   bool
	workspaceName string
}

func bootStrapTerraformTest(args *terraformArgs) *Terraform {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "terraform").Return(args.hasTfCommand)
	env.On("HasFolder", "/.terraform").Return(args.hasTfFolder)
	env.On("Pwd").Return("")
	env.On("RunCommand", "terraform", []string{"workspace", "show"}).Return(args.workspaceName, nil)
	k := &Terraform{
		env:   env,
		props: properties.Map{},
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
}
