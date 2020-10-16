package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type nodeArgs struct {
	enabled        bool
	nodeVersion    string
	hasJS          bool
	hasTS          bool
	displayVersion bool
}

func bootStrapNodeTest(args *nodeArgs) *node {
	env := new(MockedEnvironment)
	env.On("hasCommand", "node").Return(args.enabled)
	env.On("runCommand", "node", []string{"--version"}).Return(args.nodeVersion, nil)
	env.On("hasFiles", "*.js").Return(args.hasJS)
	env.On("hasFiles", "*.ts").Return(args.hasTS)
	props := &properties{
		values: map[Property]interface{}{
			DisplayVersion: args.displayVersion,
		},
	}
	n := &node{
		env:   env,
		props: props,
	}
	return n
}

func TestNodeWriterDisabled(t *testing.T) {
	args := &nodeArgs{
		enabled: false,
	}
	node := bootStrapNodeTest(args)
	assert.False(t, node.enabled(), "node is not available")
}

func TestNodeWriterDisabledNoJSorTSFiles(t *testing.T) {
	args := &nodeArgs{
		enabled: true,
	}
	node := bootStrapNodeTest(args)
	assert.False(t, node.enabled(), "no JS or TS files in the current directory")
}

func TestNodeEnabledJSFiles(t *testing.T) {
	expected := "1.14"
	args := &nodeArgs{
		enabled:        true,
		nodeVersion:    expected,
		hasJS:          true,
		displayVersion: true,
	}
	node := bootStrapNodeTest(args)
	assert.True(t, node.enabled())
	assert.Equal(t, expected, node.string(), "node is available and JS files are found")
}

func TestNodeEnabledTsFiles(t *testing.T) {
	expected := "1.14"
	args := &nodeArgs{
		enabled:        true,
		nodeVersion:    expected,
		hasTS:          true,
		displayVersion: true,
	}
	node := bootStrapNodeTest(args)
	assert.True(t, node.enabled())
	assert.Equal(t, expected, node.string(), "node is available and TS files are found")
}

func TestNodeEnabledJsAndTsFiles(t *testing.T) {
	expected := "1.14"
	args := &nodeArgs{
		enabled:        true,
		nodeVersion:    expected,
		hasJS:          true,
		hasTS:          true,
		displayVersion: true,
	}
	node := bootStrapNodeTest(args)
	assert.True(t, node.enabled())
	assert.Equal(t, expected, node.string(), "node is available and JS and TS files are found")
}

func TestNodeEnabledNoVersion(t *testing.T) {
	expected := ""
	args := &nodeArgs{
		enabled:        true,
		nodeVersion:    "1.14",
		hasJS:          true,
		displayVersion: false,
	}
	node := bootStrapNodeTest(args)
	assert.True(t, node.enabled())
	assert.Equal(t, expected, node.string(), "we don't expect a version")
}

func TestNodeEnabledNodeVersion(t *testing.T) {
	expected := "1.14"
	args := &nodeArgs{
		enabled:        true,
		nodeVersion:    expected,
		hasJS:          true,
		displayVersion: true,
	}
	node := bootStrapNodeTest(args)
	assert.True(t, node.enabled())
	assert.Equal(t, expected, node.string(), "we expect a version")
}
