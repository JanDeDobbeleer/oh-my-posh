package segments

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestAzdSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		Template        string
		IsInited        bool
	}{
		{
			Case:            "no .azure directory found",
			ExpectedEnabled: false,
		},
		{
			Case:            "Environment located",
			ExpectedEnabled: true,
			ExpectedString:  "TestEnvironment",
			Template:        "{{ .DefaultEnvironment }}",
			IsInited:        true,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Debug", mock2.Anything)

		if tc.IsInited {
			fileInfo := &platform.FileInfo{
				Path:         "test/.azure",
				ParentFolder: "test",
				IsDir:        true,
			}
			env.On("HasParentFilePath", ".azure").Return(fileInfo, nil)
			dirEntries := []fs.DirEntry{
				&MockDirEntry{
					name:  "config.json",
					isDir: false,
				}, &MockDirEntry{
					name:  "TestEnvironment",
					isDir: true,
				},
			}
			env.On("LsDir", filepath.Join("test", ".azure")).Return(dirEntries, nil)

			env.On("FileContent", filepath.Join("test", ".azure", "config.json")).Return(`{"version": 1, "defaultEnvironment": "TestEnvironment"}`, nil)
		} else {
			env.On("HasParentFilePath", ".azure").Return(&platform.FileInfo{}, errors.New("no such file or directory"))
		}

		azd := Azd{
			env:   env,
			props: properties.Map{},
		}

		assert.Equal(t, tc.ExpectedEnabled, azd.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, azd), tc.Case)
	}
}
