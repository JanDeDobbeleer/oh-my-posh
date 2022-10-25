package segments

import (
	"io/fs"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DotnetProjectFile struct{}

func (d DotnetProjectFile) Name() string { return "Foo.csproj" }
func (d DotnetProjectFile) IsDir() bool  { return false }
func (d DotnetProjectFile) Type() fs.FileMode {
	var mode fs.FileMode
	return mode
}
func (d DotnetProjectFile) Info() (fs.FileInfo, error) {
	var info fs.FileInfo
	return info, nil
}

func TestDotnetTargetSegment(t *testing.T) {
	cwd := "/usr/home/project"
	file := DotnetProjectFile{}

	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		FileContent     string
		DisplayError    bool
	}{
		{
			Case:            "No project files found",
			ExpectedEnabled: false,
		},
		{
			Case:            "Empty project file",
			ExpectedString:  "cannot extract TFM from " + file.Name(),
			ExpectedEnabled: true,
			FileContent:     "",
			DisplayError:    true,
		},
		{
			Case:            "Regular project file",
			ExpectedString:  "netcoreapp3.1",
			ExpectedEnabled: true,
			FileContent:     "...<TargetFramework>netcoreapp3.1</TargetFramework>...",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Pwd").Return(cwd)
		if tc.ExpectedEnabled {
			env.On("LsDir", cwd).Return([]fs.DirEntry{file})
		} else {
			env.On("LsDir", cwd).Return([]fs.DirEntry{})
		}
		env.On("FileContent", file.Name()).Return(tc.FileContent)
		props := properties.Map{
			properties.DisplayError: tc.DisplayError,
		}
		dotnettarget := &DotnetTarget{}
		dotnettarget.Init(props, env)
		enabled := dotnettarget.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, dotnettarget.Template(), dotnettarget), tc.Case)
	}
}
