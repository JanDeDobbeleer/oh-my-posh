package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type golangArgs struct {
	enabled        bool
	goVersion      string
	hasFiles       bool
	displayVersion bool
}

func bootStrapGolangTest(args *golangArgs) *golang {
	env := new(MockedEnvironment)
	env.On("hasCommand", "go").Return(args.enabled)
	env.On("runCommand", "go", []string{"version"}).Return(args.goVersion, nil)
	env.On("hasFiles", "*.go").Return(args.hasFiles)
	props := &properties{
		values: map[Property]interface{}{
			DisplayVersion: args.displayVersion,
		},
	}
	g := &golang{
		env:   env,
		props: props,
	}
	return g
}

func TestGolangNoGoInstalled(t *testing.T) {
	args := &golangArgs{
		enabled: false,
	}
	golang := bootStrapGolangTest(args)
	assert.False(t, golang.enabled())
}

func TestGolangGoInstalledNoFiles(t *testing.T) {
	args := &golangArgs{
		enabled:  true,
		hasFiles: false,
	}
	golang := bootStrapGolangTest(args)
	assert.False(t, golang.enabled())
}

func TestGolangFilesNoGo(t *testing.T) {
	args := &golangArgs{
		enabled:  false,
		hasFiles: true,
	}
	golang := bootStrapGolangTest(args)
	assert.False(t, golang.enabled())
}

func TestGolangGoEnabled(t *testing.T) {
	args := &golangArgs{
		enabled:  true,
		hasFiles: true,
	}
	golang := bootStrapGolangTest(args)
	assert.True(t, golang.enabled())
}

func TestGolangGoEnabledWithVersion(t *testing.T) {
	args := &golangArgs{
		enabled:        true,
		hasFiles:       true,
		displayVersion: true,
		goVersion:      "go version go1.15.3 darwin/amd64",
	}
	golang := bootStrapGolangTest(args)
	assert.True(t, golang.enabled())
	assert.Equal(t, "1.15.3", golang.string())
}

func TestGolangGoEnabledWithoutVersion(t *testing.T) {
	args := &golangArgs{
		enabled:        true,
		hasFiles:       true,
		displayVersion: false,
		goVersion:      "go version go1.15.3 darwin/amd64",
	}
	golang := bootStrapGolangTest(args)
	assert.True(t, golang.enabled())
	assert.Equal(t, "", golang.string())
}
