package main

import (
	"errors"
	"fmt"
	"io/ioutil"
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
		FetchVersion: true,
	}
	return env, props
}

func TestGolang(t *testing.T) {
	cases := []struct {
		Case                  string
		ExpectedString        string
		Version               string
		ParseModFile          bool
		HasModFileInParentDir bool
		InvalidModfile        bool
	}{
		{Case: "Go 1.15", ExpectedString: "1.15.8", Version: "go version go1.15.8 darwin/amd64"},
		{Case: "Go 1.16", ExpectedString: "1.16", Version: "go version go1.16 darwin/amd64"},
		{Case: "go.mod 1.17", ParseModFile: true, HasModFileInParentDir: true, ExpectedString: "1.17"},
		{Case: "no go.mod file fallback", ParseModFile: true, ExpectedString: "1.16", Version: "go version go1.16 darwin/amd64"},
		{
			Case:                  "invalid go.mod file fallback",
			ParseModFile:          true,
			HasModFileInParentDir: true,
			InvalidModfile:        true,
			ExpectedString:        "./go.mod:1: unknown directive: invalid",
			Version:               "go version go1.16 darwin/amd64",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "go",
			versionParam:  "version",
			versionOutput: tc.Version,
			extension:     "*.go",
		}
		env, props := getMockedLanguageEnv(params)
		if tc.ParseModFile {
			props[ParseModFile] = tc.ParseModFile
			fileInfo := &fileInfo{
				path:         "./go.mod",
				parentFolder: "./",
				isDir:        false,
			}
			var err error
			if !tc.HasModFileInParentDir {
				err = errors.New("no match")
			}
			env.On("hasParentFilePath", "go.mod").Return(fileInfo, err)
			var content string
			if tc.InvalidModfile {
				content = "invalid go.mod file"
			} else {
				tmp, _ := ioutil.ReadFile(fileInfo.path)
				content = string(tmp)
			}
			env.On("getFileContent", fileInfo.path).Return(content)
		}
		g := &golang{}
		g.init(props, env)
		assert.True(t, g.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, g.string(), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
