package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type kubectlArgs struct {
	enabled     bool
	contextName string
}

func bootStrapKubectlTest(args *kubectlArgs) *kubectl {
	env := new(MockedEnvironment)
	env.On("hasCommand", "kubectl").Return(args.enabled)
	env.On("runCommand", "kubectl", []string{"config", "current-context"}).Return(args.contextName, nil)
	k := &kubectl{
		env:   env,
		props: &properties{},
	}
	return k
}

func TestKubectlWriterDisabled(t *testing.T) {
	args := &kubectlArgs{
		enabled: false,
	}
	kubectl := bootStrapKubectlTest(args)
	assert.False(t, kubectl.enabled())
}

func TestKubectlEnabled(t *testing.T) {
	expected := "context-name"
	args := &kubectlArgs{
		enabled:     true,
		contextName: expected,
	}
	kubectl := bootStrapKubectlTest(args)
	assert.True(t, kubectl.enabled())
	assert.Equal(t, expected, kubectl.string())
}

func TestKubectlNoContext(t *testing.T) {
	args := &kubectlArgs{
		enabled:     true,
		contextName: "",
	}
	kubectl := bootStrapKubectlTest(args)
	assert.False(t, kubectl.enabled())
}
