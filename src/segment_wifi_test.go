package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func bootStrapWifiTest() *wifi {
	env := new(MockedEnvironment)
	// env.On("hasCommand", "terraform").Return(args.hasTfCommand)
	// env.On("hasFolder", ".terraform").Return(args.hasTfFolder)
	// env.On("runCommand", "terraform", []string{"workspace", "show"}).Return(args.workspaceName, nil)
	k := &wifi{
		env:   env,
		props: &properties{},
	}
	return k
}

func TestSomething(t *testing.T) {
	wifi := bootStrapWifiTest()
	assert.True(t, wifi.enabled())
}

func TestInterfaces(t *testing.T) {
	wifi := bootStrapWifiTest()
	assert.NotNil(t, wifi.string())
}
