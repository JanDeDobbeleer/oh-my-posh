package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockedLanguageParams struct {
	cmd           string
	versionParam  string
	versionOutput string
	extension     string
}

func getMockedLanguageEnv(params *mockedLanguageParams) (*MockedEnvironment, properties) {
	env := new(MockedEnvironment)
	env.On("hasCommand", params.cmd).Return(true)
	env.On("runCommand", params.cmd, []string{params.versionParam}).Return(params.versionOutput, nil)
	env.On("hasFiles", params.extension).Return(true)
	env.On("getcwd", nil).Return("/usr/home/project")
	env.On("homeDir", nil).Return("/usr/home")
	var props properties = map[Property]interface{}{
		DisplayVersion: true,
	}
	return env, props
}

func TestGolang(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Go 1.15", ExpectedString: "1.15.8", Version: "go version go1.15.8 darwin/amd64"},
		{Case: "Go 1.16", ExpectedString: "1.16", Version: "go version go1.16 darwin/amd64"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "go",
			versionParam:  "version",
			versionOutput: tc.Version,
			extension:     "*.go",
		}
		env, props := getMockedLanguageEnv(params)
		g := &golang{}
		g.init(props, env)
		assert.True(t, g.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, g.string(), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
