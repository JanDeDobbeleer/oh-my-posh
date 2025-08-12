package segments

import (
	"errors"
	"testing"

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
func TestJujutsuGetIdInfo(t *testing.T) {
	cases := []struct {
		Case               string
		CmdOutput          string
		LogTemplates       map[string]string
		ExpectedChangeID   string
		ExpectedWorking    *JujutsuStatus
		ExpectedLogResults map[string]string
	}{
		{
			Case:               "nochanges",
			CmdOutput:          "a\n",
			ExpectedChangeID:   "a",
			ExpectedWorking:    &JujutsuStatus{ScmStatus{}},
			ExpectedLogResults: map[string]string{},
		},
		{
			Case: "changed",
			CmdOutput: "b\x00" + `D deleted_file
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
			ExpectedLogResults: map[string]string{},
		},
		{
			Case:      "customTemplates",
			CmdOutput: "kk\x00\x00kk\x00utnvqz",
			LogTemplates: map[string]string{
				"change_id_prefix": "change_id.shortest(8).prefix()",
				"change_id_rest":   "change_id.shortest(8).rest()",
			},
			ExpectedChangeID: "kk",
			ExpectedWorking:  &JujutsuStatus{ScmStatus{}},
			ExpectedLogResults: map[string]string{
				"change_id_prefix": "kk",
				"change_id_rest":   "utnvqz",
			},
		},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{
			Path:         "/dir/hello",
			ParentFolder: "/dir",
			IsDir:        true,
		}

		if tc.LogTemplates == nil {
			tc.LogTemplates = make(map[string]string)
		}

		props := properties.Map{
			FetchStatus:  true,
			LogTemplates: tc.LogTemplates,
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
		logTemplate, _ := jj.logTemplate()
		env.MockJjCommand(fileInfo.Path, tc.CmdOutput, "log", "-r", "@", "--no-graph", "-T", logTemplate)

		if tc.ExpectedWorking != nil {
			tc.ExpectedWorking.Formats = map[string]string{}
		}

		assert.True(t, jj.Enabled())
		assert.Equal(t, fileInfo.Path, jj.mainSCMDir)
		assert.Equal(t, fileInfo.Path, jj.repoRootDir)
		assert.Equal(t, tc.ExpectedWorking, jj.Working, tc.Case)
		assert.Equal(t, tc.ExpectedChangeID, jj.ChangeID, tc.Case)
		assert.Equal(t, tc.ExpectedLogResults, jj.LogResults, tc.Case)
	}
}
