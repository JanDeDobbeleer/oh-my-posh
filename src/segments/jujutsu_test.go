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
	assert.Empty(t, jj.ChangeID)
	assert.Empty(t, jj.ChangeIDPrefix)
	assert.Empty(t, jj.ChangeIDRest)
	env.AssertNotCalled(t, "RunCommand")
}

func TestJujutsuLogTemplate(t *testing.T) {
	jj := &Jujutsu{}
	jj.Init(options.Map{ChangeIDMinLen: 8}, new(mock.Environment))

	assert.Equal(
		t,
		`change_id.shortest(8).prefix() ++ "|" ++ change_id.shortest(8).rest() ++ "\n" ++ diff.summary()`,
		jj.logTemplate(),
	)
}

func TestJujutsuTemplate(t *testing.T) {
	assert.Equal(t, " \uf1fa{{.ChangeID}}{{if .Working.Changed}} \uf044 {{ .Working.String }}{{ end }} ", (&Jujutsu{}).Template())
}

func TestJujutsuGetIdInfo(t *testing.T) {
	cases := []struct {
		ExpectedWorking        *JujutsuStatus
		CommandError           error
		Case                   string
		LogOutput              string
		ExpectedChangeID       string
		ExpectedChangeIDPrefix string
		ExpectedChangeIDRest   string
		ChangeIDMinLen         int
		ValidHeader            bool
	}{
		{
			Case:                   "clean with minimum-length rest",
			LogOutput:              "t|ususrrr",
			ExpectedChangeID:       "tususrrr",
			ExpectedChangeIDPrefix: "t",
			ExpectedChangeIDRest:   "ususrrr",
			ChangeIDMinLen:         8,
			ValidHeader:            true,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:                   "clean with empty rest after output trimming",
			LogOutput:              "tususrrr|",
			ExpectedChangeID:       "tususrrr",
			ExpectedChangeIDPrefix: "tususrrr",
			ExpectedChangeIDRest:   "",
			ChangeIDMinLen:         4,
			ValidHeader:            true,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case: "changed with empty rest",
			LogOutput: `tususrrr|
D deleted_file
A added_file
C {copied_file => new_file}
M modified_file
R {renamed_file => new_file}`,
			ExpectedChangeID:       "tususrrr",
			ExpectedChangeIDPrefix: "tususrrr",
			ExpectedChangeIDRest:   "",
			ChangeIDMinLen:         4,
			ValidHeader:            true,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  1,
				Added:    2,
				Modified: 1,
				Moved:    1,
			}},
		},
		{
			Case:           "command error",
			LogOutput:      "ignored",
			CommandError:   errors.New("jj failed"),
			ChangeIDMinLen: 8,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "missing delimiter",
			LogOutput:      "tususrrr\nD deleted_file",
			ChangeIDMinLen: 8,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "multiple delimiters",
			LogOutput:      "t|usus|rrr\nD deleted_file",
			ChangeIDMinLen: 8,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			fileInfo := &runtime.FileInfo{
				Path:         "/dir/hello",
				ParentFolder: "/dir",
				IsDir:        true,
			}

			props := options.Map{
				FetchStatus:    true,
				ChangeIDMinLen: tc.ChangeIDMinLen,
			}

			env := new(mock.Environment)
			env.On("InWSLSharedDrive").Return(false)
			env.On("HasCommand", "jj").Return(true)
			env.On("GOOS").Return("")
			env.On("HasParentFilePath", ".jj", false).Return(fileInfo, nil)

			jj := &Jujutsu{}
			jj.Init(props, env)

			tc.ExpectedWorking.Formats = map[string]string{}

			commandArgs := []string{
				"--repository",
				fileInfo.Path,
				"--no-pager",
				"--color",
				"never",
				"--ignore-working-copy",
				"log",
				"-r",
				"@",
				"--no-graph",
				"-T",
				jj.logTemplate(),
			}
			env.On("RunCommand", "jj", commandArgs).Return(tc.LogOutput, tc.CommandError)

			assert.True(t, jj.Enabled())
			assert.Equal(t, fileInfo.Path, jj.mainSCMDir)
			assert.Equal(t, fileInfo.Path, jj.repoRootDir)
			assert.Equal(t, tc.ExpectedWorking, jj.Working)
			assert.Equal(t, tc.ExpectedChangeID, jj.ChangeID)
			assert.Equal(t, tc.ExpectedChangeIDPrefix, jj.ChangeIDPrefix)
			assert.Equal(t, tc.ExpectedChangeIDRest, jj.ChangeIDRest)
			if tc.ValidHeader {
				assert.Equal(t, jj.ChangeIDPrefix+jj.ChangeIDRest, jj.ChangeID)
			}
			env.AssertNumberOfCalls(t, "RunCommand", 1)
			env.AssertExpectations(t)
		})
	}
}
