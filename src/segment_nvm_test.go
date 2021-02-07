package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type nvmArgs struct {
	nvmContent    string
	nodeVersion   string
	hasFiles      bool
	expectedError error
}

func setupNvm(args *nvmArgs) (nvm, *properties) {
	env := new(MockedEnvironment)

	for _, extension := range []string{"*.js", "*.ts", "package.json", ".nvmrc", "jsx"} {
		env.On("hasFiles", extension).Return(args.hasFiles)
	}
	env.On("getFileContent", ".nvmrc").Return(args.nvmContent)
	env.On("runCommand", "node", []string{"--version"}).Return(args.nodeVersion, args.expectedError)
	props := &properties{
		foreground: "#fff",
		background: "#000",
	}
	n := nvm{
		env:   env,
		props: props,
	}
	return n, props
}

func testNvmSegment(args *nvmArgs) (bool, string, *properties) {
	n, p := setupNvm(args)
	enabled := n.enabled()
	value := ""
	if enabled {
		value = n.string()
	}
	return n.enabled(), value, p
}

func TestSegmentNotEnabled(t *testing.T) {
	args := &nvmArgs{
		hasFiles: false,
	}
	enabled, _, _ := testNvmSegment(args)
	assert.EqualValues(t, false, enabled)
}

func TestNoNode(t *testing.T) {
	want := fmt.Sprintf("%s %s", "", "---")
	args := &nvmArgs{
		nvmContent:    "",
		hasFiles:      true,
		nodeVersion:   "",
		expectedError: &commandError{exitCode: 1, err: ""},
	}
	_, got, props := testNvmSegment(args)
	assert.EqualValues(t, want, got)
	assert.EqualValues(t, "#448C42", props.background)
	assert.EqualValues(t, "#FFFFFF", props.foreground)
}

func TestOnlyNodeVersion(t *testing.T) {
	want := fmt.Sprintf("%s %s", "", "v10.15.1")
	args := &nvmArgs{
		nvmContent:  "",
		hasFiles:    true,
		nodeVersion: "v10.15.1",
	}
	_, got, props := testNvmSegment(args)
	assert.EqualValues(t, want, got)
	assert.EqualValues(t, "#448C42", props.background)
	assert.EqualValues(t, "#FFFFFF", props.foreground)
}

func TestNodeVersionMatchNvmVersion(t *testing.T) {
	want := fmt.Sprintf("%s %s", "", "v10.15.1")
	args := &nvmArgs{
		nvmContent:  "v10.15.1",
		hasFiles:    true,
		nodeVersion: "v10.15.1",
	}

	_, got, props := testNvmSegment(args)
	assert.EqualValues(t, want, got)
	assert.EqualValues(t, "#448C42", props.background)
	assert.EqualValues(t, "#FFFFFF", props.foreground)
}

func TestNodeVersionNoMatchNvmVersion(t *testing.T) {
	want := fmt.Sprintf("%s %s", "", "v12.20.1")
	args := &nvmArgs{
		nvmContent:  "v12.20.1",
		hasFiles:    true,
		nodeVersion: "v10.15.1",
	}

	_, got, props := testNvmSegment(args)
	assert.EqualValues(t, want, got)
	assert.EqualValues(t, "red", props.background)
	assert.EqualValues(t, "#FFFFFF", props.foreground)
}
