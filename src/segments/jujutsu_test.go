package segments

import (
	"errors"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"
)

func TestJujutsuEnabledToolNotFound(t *testing.T) {
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasParentFilePath", ".jj", false).Return(&runtime.FileInfo{}, errors.New("not found"))
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)

	jj := &Jujutsu{}
	jj.Init(properties.Map{}, env)

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
	jj.Init(properties.Map{}, env)

	assert.True(t, jj.Enabled())
	assert.Equal(t, fileInfo.Path, jj.mainSCMDir)
	assert.Equal(t, fileInfo.Path, jj.repoRootDir)
}

// genJJOutput is a test helper to generate the output for specific JJ templates seperated by NUL characters.
func genJJOutput(jj *Jujutsu, outputMap map[string]string) string {
	jjTemplates := jj.jjTemplates()

	tmplIndexes := make(map[string]int)
	for i, tmpl := range jjTemplates {
		tmplIndexes[tmpl.template] = i
	}

	outputs := make([]string, len(jjTemplates))
	for tmplString, output := range outputMap {
		tmplIndex, ok := tmplIndexes[tmplString]
		if !ok {
			continue
		}
		outputs[tmplIndex] = output
	}
	return strings.Join(outputs, "\x00")
}

func TestJujutsuOutyt(t *testing.T) {
	cases := []struct {
		Case      string
		CmdOutput map[string]string

		ExpectedScmStatus       ScmStatus
		ExpectedChangeID        JujutsuID
		ExpectedCommitID        JujutsuID
		ExpectedLocalBookmarks  []string
		ExpectedRemoteBookmarks []string
		ExpectedDescription     string
		ExpectedConflict        bool
		ExpectedImmutable       bool
		ExpectedEmpty           bool
		ExpectedDivergent       bool
		ExpectedHidden          bool
		ExpectedMine            bool
		ExpectedAuthorID        User
		ExpectedCommitterID     User
	}{
		{
			Case: "working_status",
			CmdOutput: map[string]string{
				"diff.summary()": `D deleted_file
A added_file
C {copied_file => new_file}
M modified_file
R {renamed_file => new_file}
`,
			},
			ExpectedScmStatus: ScmStatus{
				Deleted:  1,
				Added:    2,
				Modified: 1,
				Moved:    1,
			},
		},
		{
			Case: "change_id",
			CmdOutput: map[string]string{
				"change_id":                      "snwyymunypszptwmpnqqktoukulnrslv",
				"change_id.shortest(8).prefix()": "s",
				"change_id.shortest(8).rest()":   "nwyymun",
			},
			ExpectedChangeID: JujutsuID{
				Full:     "snwyymunypszptwmpnqqktoukulnrslv",
				Shortest: "s",
				Rest:     "nwyymun",
			},
		},
		{
			Case: "commit_id",
			CmdOutput: map[string]string{
				"commit_id":                      "26dbb0c48661c8b827a868fb10b2dd3666811e3a",
				"commit_id.shortest(8).prefix()": "26",
				"commit_id.shortest(8).rest()":   "dbb0c4",
			},
			ExpectedCommitID: JujutsuID{
				Full:     "26dbb0c48661c8b827a868fb10b2dd3666811e3a",
				Shortest: "26",
				Rest:     "dbb0c4",
			},
		},
		{
			Case: "bookmarks",
			CmdOutput: map[string]string{
				"local_bookmarks.join('\n')":  "main\nfeature-branch",
				"remote_bookmarks.join('\n')": "origin/main\norigin/feature-branch",
			},
			ExpectedLocalBookmarks:  []string{"main", "feature-branch"},
			ExpectedRemoteBookmarks: []string{"origin/main", "origin/feature-branch"},
		},
		{
			Case: "booleans",
			CmdOutput: map[string]string{
				"divergent": "true",
				"hidden":    "false",
				"immutable": "true",
				"empty":     "false",
				"mine":      "true",
			},
			ExpectedDivergent: true,
			ExpectedHidden:    false,
			ExpectedImmutable: true,
			ExpectedEmpty:     false,
			ExpectedMine:      true,
		},
		{
			Case: "author/commiter",
			CmdOutput: map[string]string{
				"author().name":    "John Doe",
				"author().email":   "john@example.com",
				"commiter().name":  "Jane Doe",
				"commiter().email": "jane@example.com",
			},
			ExpectedAuthorID:    User{Name: "John Doe", Email: "john@example.com"},
			ExpectedCommitterID: User{Name: "Jane Doe", Email: "jane@example.com"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			fileInfo := &runtime.FileInfo{
				Path:         "/dir/hello",
				ParentFolder: "/dir",
				IsDir:        true,
			}

			props := properties.Map{
				FetchStatus:    true,
				MinChangeIDLen: "8",
				MinCommitIDLen: "8",
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
			templates := jj.jjTemplates()
			jjTmplString := logTemplate(templates)
			log.Debugf("Using Jujutsu log template: %s", jjTmplString)
			cmdOutput := genJJOutput(jj, tc.CmdOutput)
			env.MockJjCommand(fileInfo.Path, cmdOutput, "log", "-r", "@", "--no-graph", "-T", jjTmplString)

			assert.True(t, jj.Enabled(), tc.Case)
			assert.Equal(t, fileInfo.Path, jj.mainSCMDir)
			assert.Equal(t, fileInfo.Path, jj.repoRootDir)
			assert.Equal(t, tc.ExpectedScmStatus, jj.Working.ScmStatus, tc.Case)
			assert.Equal(t, tc.ExpectedChangeID, jj.ChangeID, tc.Case)
			assert.Equal(t, tc.ExpectedCommitID, jj.CommitID, tc.Case)
			assert.Equal(t, tc.ExpectedLocalBookmarks, jj.LocalBookmarks, tc.Case)
			assert.Equal(t, tc.ExpectedRemoteBookmarks, jj.RemoteBookmarks, tc.Case)
			assert.Equal(t, tc.ExpectedDescription, jj.Description, tc.Case)
			assert.Equal(t, tc.ExpectedConflict, jj.Conflict, tc.Case)
			assert.Equal(t, tc.ExpectedImmutable, jj.Immutable, tc.Case)
			assert.Equal(t, tc.ExpectedEmpty, jj.Empty, tc.Case)
			assert.Equal(t, tc.ExpectedDivergent, jj.Divergent, tc.Case)
			assert.Equal(t, tc.ExpectedHidden, jj.Hidden, tc.Case)
			assert.Equal(t, tc.ExpectedMine, jj.Mine, tc.Case)
			assert.Equal(t, tc.ExpectedAuthorID, jj.AuthorID, tc.Case)
			assert.Equal(t, tc.ExpectedCommitterID, jj.CommitterID, tc.Case)
		})
	}

}
