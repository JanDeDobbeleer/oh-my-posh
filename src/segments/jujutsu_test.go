package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
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
				ChangeIDMinLen: tc.ChangeIDMinLen,
			}

			env := new(mock.Environment)
			env.On("InWSLSharedDrive").Return(false)
			env.On("HasCommand", "jj").Return(true)
			env.On("GOOS").Return("")
			env.On("HasParentFilePath", ".jj", false).Return(fileInfo, nil)

			jj := &Jujutsu{}
			jj.Init(props, env)
			// the status probe is derived from template references now
			jj.SetReferencedFields(template.RefSet{Fields: jujutsuStatusFields, Analyzable: true})

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

func TestJujutsuClosestBookmarks(t *testing.T) {
	cases := []struct {
		Options  options.Map
		Case     string
		Output   string
		Expected string
		Error    bool
	}{
		{
			Case:     "undecorated bookmark names",
			Output:   "main feature/x\nnoise",
			Expected: "main feature/x",
		},
		{
			Case:  "command error",
			Error: true,
		},
		{
			Case: "no bookmarks",
		},
		{
			// the fetch_ahead_counter/ahead_icon options no longer exist;
			// leftover keys must parse silently and change nothing
			Case:     "dead option keys are ignored",
			Options:  options.Map{"fetch_ahead_counter": true, "ahead_icon": "⇡"},
			Output:   "main",
			Expected: "main",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		bookmarkArgs := []string{"--repository", "/repo", "--no-pager", "--color", "never", "--ignore-working-copy", "log", "-r", "heads(::@ & bookmarks())", "--no-graph", "-T", "bookmarks"} //nolint:lll
		if tc.Error {
			env.On("RunCommand", "jj", bookmarkArgs).Return("", errors.New("failed")).Once()
		} else {
			env.On("RunCommand", "jj", bookmarkArgs).Return(tc.Output, nil).Once()
		}

		opts := tc.Options
		if opts == nil {
			opts = options.Map{}
		}

		jj := &Jujutsu{
			command:     JUJUTSUCOMMAND,
			repoRootDir: "/repo",
		}
		jj.Init(opts, env)

		// a second call must serve the memo, not a second jj invocation
		// (the .Once() mock above panics otherwise)
		assert.Equal(t, tc.Expected, jj.ClosestBookmarks(), tc.Case)
		assert.Equal(t, tc.Expected, jj.ClosestBookmarks(), tc.Case)
		env.AssertNumberOfCalls(t, "RunCommand", 1)
	}
}

func TestJujutsuAheadCount(t *testing.T) {
	cases := []struct {
		Case          string
		Bookmarks     string
		ExpectedRange string
		AheadOutput   string
		Expected      int
		AheadError    bool
		Invoke        bool
	}{
		{
			// referencing the method is the fetch trigger: without an
			// invocation the distance query must never run
			Case:      "not referenced, not fetched",
			Bookmarks: "main",
			Invoke:    false,
		},
		{
			Case:          "distance to the closest bookmark",
			Bookmarks:     "main feature/x",
			ExpectedRange: "main..@",
			AheadOutput:   "...",
			Expected:      3,
			Invoke:        true,
		},
		{
			// conflicted bookmarks are marked with a trailing *
			Case:          "conflicted bookmark marker is trimmed",
			Bookmarks:     "main*",
			ExpectedRange: "main..@",
			AheadOutput:   ".",
			Expected:      1,
			Invoke:        true,
		},
		{
			Case:      "no bookmarks skips the distance query",
			Bookmarks: "",
			Expected:  0,
			Invoke:    true,
		},
		{
			Case:          "distance query error",
			Bookmarks:     "main",
			ExpectedRange: "main..@",
			AheadError:    true,
			Expected:      0,
			Invoke:        true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		cli := []string{"--repository", "/repo", "--no-pager", "--color", "never", "--ignore-working-copy"}
		bookmarkArgs := append(append([]string{}, cli...), "log", "-r", "heads(::@ & bookmarks())", "--no-graph", "-T", "bookmarks")
		env.On("RunCommand", "jj", bookmarkArgs).Return(tc.Bookmarks, nil).Once()

		aheadCalls := 0
		if tc.ExpectedRange != "" {
			aheadCalls = 1
			aheadArgs := append(append([]string{}, cli...), "log", "--no-graph", "-T", "'.'", "-r", tc.ExpectedRange)

			if tc.AheadError {
				env.On("RunCommand", "jj", aheadArgs).Return("", errors.New("failed")).Once()
			} else {
				env.On("RunCommand", "jj", aheadArgs).Return(tc.AheadOutput, nil).Once()
			}
		}

		jj := &Jujutsu{
			command:     JUJUTSUCOMMAND,
			repoRootDir: "/repo",
		}
		jj.Init(options.Map{}, env)

		// the bookmarks memo is shared: AheadCount reuses ClosestBookmarks'
		// single jj call instead of querying the bookmarks again
		assert.Equal(t, tc.Bookmarks, jj.ClosestBookmarks(), tc.Case)

		if !tc.Invoke {
			env.AssertNumberOfCalls(t, "RunCommand", 1)
			continue
		}

		// a second call must serve the memo (the .Once() mocks panic otherwise)
		assert.Equal(t, tc.Expected, jj.AheadCount(), tc.Case)
		assert.Equal(t, tc.Expected, jj.AheadCount(), tc.Case)
		env.AssertNumberOfCalls(t, "RunCommand", 1+aheadCalls)
	}
}
