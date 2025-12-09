package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/stretchr/testify/assert"
)

func TestJujutsuEnabledToolNotFound(t *testing.T) {
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasParentFilePath", ".jj", false).Return(&runtime.FileInfo{}, errors.New("not found"))
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)

	jj := &Jujutsu{}
	jj.Init(options.Map{}, env)

	assert.False(t, jj.Enabled())
}

func TestJujutsuEnabledInWorkingDirectory(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "jj").Return(true)
	env.On("HasParentFilePath", ".jj", false).Return(fileInfo, nil)
	env.On("GOOS").Return("")

	jj := &Jujutsu{}
	jj.Init(options.Map{}, env)

	assert.True(t, jj.Enabled())
	assert.Equal(t, fileInfo.Path, jj.mainSCMDir)
	assert.Equal(t, fileInfo.Path, jj.repoRootDir)
}

func TestJujutsuGetIdInfo(t *testing.T) {
	cases := []struct {
		ExpectedWorking  *JujutsuStatus
		Case             string
		LogOutput        string
		ExpectedChangeID string
	}{
		{
			Case:             "nochanges",
			LogOutput:        "a\n\n",
			ExpectedChangeID: "a",
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case: "changed",
			LogOutput: `b
D deleted_file
A added_file
C {copied_file => new_file}
M modified_file
R {renamed_file => new_file}
`,
			ExpectedChangeID: "b",
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  1,
				Added:    2,
				Modified: 1,
				Moved:    1,
			}},
		},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{
			Path:         "/dir/hello",
			ParentFolder: "/dir",
			IsDir:        true,
		}

		props := options.Map{
			FetchStatus: true,
		}

		env := new(mock.Environment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("HasCommand", "jj").Return(true)
		env.On("GOOS").Return("")
		env.On("IsWsl").Return(false)
		env.On("HasParentFilePath", ".jj", false).Return(fileInfo, nil)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return(poshHome)
		env.On("Getenv", poshGitEnv).Return("")

		jj := &Jujutsu{}
		jj.Init(props, env)
		env.MockJjCommand(fileInfo.Path, tc.LogOutput, "log", "-r", "@", "--no-graph", "-T", jj.logTemplate())

		if tc.ExpectedWorking != nil {
			tc.ExpectedWorking.Formats = map[string]string{}
		}

		assert.True(t, jj.Enabled())
		assert.Equal(t, fileInfo.Path, jj.mainSCMDir)
		assert.Equal(t, fileInfo.Path, jj.repoRootDir)
		assert.Equal(t, tc.ExpectedWorking, jj.Working, tc.Case)
		assert.Equal(t, tc.ExpectedChangeID, jj.ChangeID, tc.Case)
	}
}
