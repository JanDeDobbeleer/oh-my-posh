package segments

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"
)

func TestAzdSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		Template        string
		ExpectedEnabled bool
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
		env := new(mock.Environment)
		env.On("Flags").Return(&runtime.Flags{})

		if tc.IsInited {
			fileInfo := &runtime.FileInfo{
				Path:         "test/.azure",
				ParentFolder: "test",
				IsDir:        true,
			}
			env.On("HasParentFilePath", ".azure", false).Return(fileInfo, nil)
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
			env.On("HasParentFilePath", ".azure", false).Return(&runtime.FileInfo{}, errors.New("no such file or directory"))
		}

		azd := Azd{}
		azd.Init(properties.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, azd.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, azd), tc.Case)
	}
}
