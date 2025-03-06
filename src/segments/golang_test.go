package segments

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/stretchr/testify/assert"
)

func TestGolang(t *testing.T) {
	cases := []struct {
		Case                     string
		ExpectedString           string
		Version                  string
		ParseModFile             bool
		HasModFileInParentDir    bool
		InvalidModfile           bool
		ParseGoWorkFile          bool
		HasGoWorkFileInParentDir bool
		InvalidGoWorkFile        bool
	}{
		{Case: "Go 1.15", ExpectedString: "1.15.8", Version: "go version go1.15.8 darwin/amd64"},
		{Case: "Go 1.16", ExpectedString: "1.16", Version: "go version go1.16 darwin/amd64"},
		{Case: "go.mod 1.24.1", ParseModFile: true, HasModFileInParentDir: true, ExpectedString: "1.24.1"},
		{Case: "no go.mod file fallback", ParseModFile: true, ExpectedString: "1.16", Version: "go version go1.16 darwin/amd64"},
		{
			Case:                  "invalid go.mod file fallback",
			ParseModFile:          true,
			HasModFileInParentDir: true,
			InvalidModfile:        true,
			ExpectedString:        "1.16",
			Version:               "go version go1.16 darwin/amd64",
		},
		{Case: "go.work file", ParseGoWorkFile: true, HasGoWorkFileInParentDir: true, ExpectedString: "1.21"},
		{
			Case:                     "invalid go.work file fallback",
			ParseGoWorkFile:          true,
			HasGoWorkFileInParentDir: true,
			InvalidGoWorkFile:        true,
			ExpectedString:           "1.16",
			Version:                  "go version go1.16 darwin/amd64",
		},
		{
			Case:                     "go.work file with go.mod file uses go.mod's version",
			ParseModFile:             true,
			HasModFileInParentDir:    true,
			ParseGoWorkFile:          true,
			HasGoWorkFileInParentDir: true,
			ExpectedString:           "1.24.1",
		},
		{
			Case:            "missing both go.mod and go.work file fallback",
			ParseModFile:    true,
			ParseGoWorkFile: true,
			ExpectedString:  "1.16",
			Version:         "go version go1.16 darwin/amd64",
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
			fileInfo := &runtime.FileInfo{
				Path:         "../go.mod",
				ParentFolder: "./",
				IsDir:        false,
			}

			var err error
			if !tc.HasModFileInParentDir {
				err = errors.New("no match")
			}
			env.On("HasParentFilePath", "go.mod", false).Return(fileInfo, err)

			var content string
			if tc.InvalidModfile {
				content = "invalid go.mod file"
			} else {
				tmp, _ := os.ReadFile(fileInfo.Path)
				content = string(tmp)
			}

			env.On("FileContent", fileInfo.Path).Return(content)
		}

		if tc.ParseGoWorkFile {
			props[ParseWorkFile] = tc.ParseGoWorkFile
			fileInfo := &runtime.FileInfo{
				Path:         "../test/go.work",
				ParentFolder: "./",
				IsDir:        false,
			}

			var err error
			if !tc.HasGoWorkFileInParentDir {
				err = errors.New("no match")
			}

			env.On("HasParentFilePath", "go.work", false).Return(fileInfo, err)
			var content string
			if tc.InvalidGoWorkFile {
				content = "invalid go.work file"
			} else {
				tmp, _ := os.ReadFile(fileInfo.Path)
				content = string(tmp)
			}

			env.On("FileContent", fileInfo.Path).Return(content)
		}

		g := &Golang{}
		g.Init(props, env)

		assert.True(t, g.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, g.Template(), g), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
