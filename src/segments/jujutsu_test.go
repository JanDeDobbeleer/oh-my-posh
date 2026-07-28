package segments

import (
	"errors"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/stretchr/testify/assert"
)

func TestJujutsuEnabledRepositoryNotFound(t *testing.T) {
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasParentFilePath", ".jj", false).Return(&runtime.FileInfo{}, errors.New("not found"))
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)

	jj := &Jujutsu{}
	jj.Init(options.Map{}, env)

	assert.False(t, jj.Enabled())
}

func TestJujutsuEnabledToolNotFound(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasParentFilePath", ".jj", false).Return(fileInfo, nil)
	env.On("GOOS").Return("")
	env.On("HasCommand", "jj").Return(false)

	jj := &Jujutsu{}
	jj.Init(options.Map{FetchStatus: true}, env)

	assert.False(t, jj.Enabled())
	env.AssertNotCalled(t, "RunCommand")
	env.AssertExpectations(t)
}

func TestJujutsuEnabledInWorkingDirectory(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.Environment)
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
	assert.Empty(t, jj.CommitID)
	assert.Empty(t, jj.CommitIDPrefix)
	assert.Empty(t, jj.CommitIDRest)
	env.AssertNotCalled(t, "HasCommand")
	env.AssertNotCalled(t, "RunCommand")
	env.AssertExpectations(t)
}

func TestJujutsuLogTemplate(t *testing.T) {
	jj := &Jujutsu{}
	jj.Init(options.Map{
		ChangeIDMinLen: 8,
		CommitIDMinLen: 12,
	}, new(mock.Environment))

	expected := `"OMPJJ1:5\0" ++ change_id.shortest(8).prefix() ++ "\0" ++ ` +
		`change_id.shortest(8).rest() ++ "\0" ++ commit_id.shortest(12).prefix() ++ "\0" ++ ` +
		`commit_id.shortest(12).rest() ++ "\0" ++ diff.summary() ++ "\0"`
	assert.Equal(
		t,
		expected,
		jj.logTemplate(),
	)
}

func TestJujutsuLogTemplateDefaults(t *testing.T) {
	jj := &Jujutsu{}
	jj.Init(options.Map{}, new(mock.Environment))

	expected := `"OMPJJ1:5\0" ++ change_id.shortest(0).prefix() ++ "\0" ++ ` +
		`change_id.shortest(0).rest() ++ "\0" ++ commit_id.shortest(0).prefix() ++ "\0" ++ ` +
		`commit_id.shortest(0).rest() ++ "\0" ++ diff.summary() ++ "\0"`
	assert.Equal(
		t,
		expected,
		jj.logTemplate(),
	)
}

func TestJujutsuTemplate(t *testing.T) {
	assert.Equal(t, " \uf1fa{{.ChangeID}}{{if .Working.Changed}} \uf044 {{ .Working.String }}{{ end }} ", (&Jujutsu{}).Template())
}

func jujutsuStatusFrame(changeIDPrefix, changeIDRest, commitIDPrefix, commitIDRest, status string) string {
	return jujutsuStatusFrameHeader + "\x00" +
		changeIDPrefix + "\x00" +
		changeIDRest + "\x00" +
		commitIDPrefix + "\x00" +
		commitIDRest + "\x00" +
		status + "\x00"
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
		ExpectedCommitID       string
		ExpectedCommitIDPrefix string
		ExpectedCommitIDRest   string
		ChangeIDMinLen         int
		CommitIDMinLen         int
		ValidFrame             bool
	}{
		{
			Case:                   "clean with minimum-length rests",
			LogOutput:              jujutsuStatusFrame("t", "ususrrr", "a", "bcdefghijk", ""),
			ExpectedChangeID:       "tususrrr",
			ExpectedChangeIDPrefix: "t",
			ExpectedChangeIDRest:   "ususrrr",
			ExpectedCommitID:       "abcdefghijk",
			ExpectedCommitIDPrefix: "a",
			ExpectedCommitIDRest:   "bcdefghijk",
			ChangeIDMinLen:         8,
			CommitIDMinLen:         11,
			ValidFrame:             true,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:                   "clean with empty rests after output trimming",
			LogOutput:              strings.TrimSpace(jujutsuStatusFrame("tususrrr", "", "abcdef12", "", "") + "\n"),
			ExpectedChangeID:       "tususrrr",
			ExpectedChangeIDPrefix: "tususrrr",
			ExpectedChangeIDRest:   "",
			ExpectedCommitID:       "abcdef12",
			ExpectedCommitIDPrefix: "abcdef12",
			ExpectedCommitIDRest:   "",
			ChangeIDMinLen:         4,
			CommitIDMinLen:         4,
			ValidFrame:             true,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case: "changed with empty rests",
			LogOutput: jujutsuStatusFrame("tususrrr", "", "abcdef12", "", `D deleted_file
A added_file
C {copied_file => new_file}
M modified_file
R {renamed_file => new_file}`),
			ExpectedChangeID:       "tususrrr",
			ExpectedChangeIDPrefix: "tususrrr",
			ExpectedChangeIDRest:   "",
			ExpectedCommitID:       "abcdef12",
			ExpectedCommitIDPrefix: "abcdef12",
			ExpectedCommitIDRest:   "",
			ChangeIDMinLen:         4,
			CommitIDMinLen:         4,
			ValidFrame:             true,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  1,
				Added:    2,
				Modified: 1,
				Moved:    1,
			}},
		},
		{
			Case:           "command error",
			LogOutput:      jujutsuStatusFrame("t", "ususrrr", "a", "bcdefghijk", ""),
			CommandError:   errors.New("jj failed"),
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "wrong magic",
			LogOutput:      "OMPJJ2:5\x00t\x00ususrrr\x00a\x00bcdefghijk\x00D deleted_file\x00",
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "wrong field count in magic",
			LogOutput:      "OMPJJ1:4\x00t\x00ususrrr\x00a\x00bcdefghijk\x00D deleted_file\x00",
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "missing terminal sentinel",
			LogOutput:      "OMPJJ1:5\x00t\x00ususrrr\x00a\x00bcdefghijk\x00D deleted_file",
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "missing field",
			LogOutput:      "OMPJJ1:5\x00t\x00ususrrr\x00a\x00D deleted_file\x00",
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "extra field",
			LogOutput:      "OMPJJ1:5\x00t\x00ususrrr\x00a\x00bcdefghijk\x00unexpected\x00D deleted_file\x00",
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "empty change ID prefix",
			LogOutput:      jujutsuStatusFrame("", "tususrrr", "a", "bcdefghijk", "D deleted_file"),
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
			ExpectedWorking: &JujutsuStatus{ScmStatus{
				Deleted:  0,
				Added:    0,
				Modified: 0,
				Moved:    0,
			}},
		},
		{
			Case:           "empty commit ID prefix",
			LogOutput:      jujutsuStatusFrame("t", "ususrrr", "", "abcdefghijk", "D deleted_file"),
			ChangeIDMinLen: 8,
			CommitIDMinLen: 11,
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
				CommitIDMinLen: tc.CommitIDMinLen,
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
			assert.Equal(t, tc.ExpectedCommitID, jj.CommitID)
			assert.Equal(t, tc.ExpectedCommitIDPrefix, jj.CommitIDPrefix)
			assert.Equal(t, tc.ExpectedCommitIDRest, jj.CommitIDRest)
			if tc.ValidFrame {
				assert.Equal(t, jj.ChangeIDPrefix+jj.ChangeIDRest, jj.ChangeID)
				assert.Equal(t, jj.CommitIDPrefix+jj.CommitIDRest, jj.CommitID)
			}
			env.AssertNumberOfCalls(t, "RunCommand", 1)
			env.AssertExpectations(t)
		})
	}
}
