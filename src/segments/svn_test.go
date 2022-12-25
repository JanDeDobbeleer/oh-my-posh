package segments

import (
	"oh-my-posh/mock"
	"oh-my-posh/platform"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSvnEnabledToolNotFound(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "svn").Return(false)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)
	s := &Svn{
		scm: scm{
			env:   env,
			props: properties.Map{},
		},
	}
	assert.False(t, s.Enabled())
}

func TestSvnEnabledInWorkingDirectory(t *testing.T) {
	fileInfo := &platform.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.MockedEnvironment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "svn").Return(true)
	env.On("GOOS").Return("")
	env.On("FileContent", "/dir/hello/trunk").Return("")
	env.MockSvnCommand(fileInfo.Path, "", "info", "--tags", "--exact-match")
	env.On("IsWsl").Return(false)
	env.On("HasParentFilePath", ".svn").Return(fileInfo, nil)
	s := &Svn{
		scm: scm{
			env:   env,
			props: properties.Map{},
		},
	}
	assert.True(t, s.Enabled())
	assert.Equal(t, fileInfo.Path, s.workingDir)
	assert.Equal(t, fileInfo.Path, s.realDir)
}

func TestSvnTemplateString(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
		Svn      *Svn
	}{
		{
			Case:     "Default template",
			Expected: "\ue0a0trunk r2 ?9 +2 ~3 -7 >13 x5 !1",
			Template: " \ue0a0{{.Branch}} r{{.BaseRev}} {{.Working.String}} ",
			Svn: &Svn{
				Branch:  "trunk",
				BaseRev: 2,
				Working: &SvnStatus{
					ScmStatus: ScmStatus{
						Untracked:  9,
						Added:      2,
						Conflicted: 1,
						Deleted:    7,
						Modified:   3,
						Moved:      13,
						Unmerged:   5,
					},
				},
			},
		},
		{
			Case:     "Only Branch name",
			Expected: "trunk",
			Template: "{{ .Branch }}",
			Svn: &Svn{
				Branch:  "trunk",
				BaseRev: 2,
			},
		},
		{
			Case:     "Working area changes",
			Expected: "trunk \uF044 +2 ~3",
			Template: "{{ .Branch }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Svn: &Svn{
				Branch: "trunk",
				Working: &SvnStatus{
					ScmStatus: ScmStatus{
						Added:    2,
						Modified: 3,
					},
				},
			},
		},
		{
			Case:     "No working area changes (using changed flag)",
			Expected: "trunk",
			Template: "{{ .Branch }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Svn: &Svn{
				Branch:  "trunk",
				Working: &SvnStatus{},
			},
		},
		{
			Case:     "No working area changes",
			Expected: "trunk",
			Template: "{{ .Branch }}{{ .Working.String }}",
			Svn: &Svn{
				Branch:  "trunk",
				Working: &SvnStatus{},
			},
		},
		{
			Case:     "Base revision with Working changes",
			Expected: "trunk - 2 \uF044 +2 ~3",
			Template: "{{ .Branch }} - {{ .BaseRev }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Svn: &Svn{
				Branch:  "trunk",
				BaseRev: 2,
				Working: &SvnStatus{
					ScmStatus: ScmStatus{
						Added:    2,
						Modified: 3,
					},
				},
			},
		},
		{
			Case:     "Working and staging area changes with separator and stash count",
			Expected: "trunk CONFLICTED \uF044 +2 ~3 !7",
			Template: "{{ .Branch }}{{ if .Working.HasConflicts }} CONFLICTED{{ end }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Svn: &Svn{
				Branch:  "trunk",
				BaseRev: 2,
				Working: &SvnStatus{
					ScmStatus: ScmStatus{
						Added:      2,
						Modified:   3,
						Conflicted: 7,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		props := properties.Map{
			FetchStatus: true,
		}
		env := new(mock.MockedEnvironment)
		tc.Svn.env = env
		tc.Svn.props = props
		assert.Equal(t, tc.Expected, renderTemplate(env, tc.Template, tc.Svn), tc.Case)
	}
}

func TestSetSvnStatus(t *testing.T) {
	cases := []struct {
		Case              string
		StatusOutput      string
		RefOutput         string
		BranchOutput      string
		ExpectedWorking   *SvnStatus
		ExpectedBranch    string
		ExpectedRef       int
		ExpectedConflicts bool
		ExpectedChanged   bool
	}{
		{
			Case: "changed",
			StatusOutput: `
?       Untracked.File
!       Missing.File
A       FileHasBeen.Added
D       FileMarkedAs.Deleted
M       Modified.File
C       Conflicted.File
R       Moved.File`,
			ExpectedWorking: &SvnStatus{ScmStatus: ScmStatus{
				Modified:   1,
				Added:      1,
				Deleted:    1,
				Moved:      2,
				Untracked:  1,
				Conflicted: 1,
			}},
			RefOutput:         "1133",
			ExpectedRef:       1133,
			BranchOutput:      "^/trunk",
			ExpectedBranch:    "trunk",
			ExpectedChanged:   true,
			ExpectedConflicts: true,
		},
		{
			Case:         "conflict",
			StatusOutput: `C       build.cake`,
			ExpectedWorking: &SvnStatus{ScmStatus: ScmStatus{
				Conflicted: 1,
			}},
			ExpectedChanged:   true,
			ExpectedConflicts: true,
		},
		{
			Case:            "no change",
			ExpectedWorking: &SvnStatus{ScmStatus: ScmStatus{}},
			ExpectedChanged: false,
		},
		{
			Case:            "not an integer ref",
			ExpectedWorking: &SvnStatus{ScmStatus: ScmStatus{}},
			ExpectedChanged: false,
			RefOutput:       "not an integer",
		},
	}
	for _, tc := range cases {
		fileInfo := &platform.FileInfo{
			Path:         "/dir/hello",
			ParentFolder: "/dir",
			IsDir:        true,
		}
		env := new(mock.MockedEnvironment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("IsWsl").Return(false)
		env.On("HasCommand", "svn").Return(true)
		env.On("GOOS").Return("")
		env.On("FileContent", "/dir/hello/trunk").Return("")
		env.MockSvnCommand(fileInfo.Path, "", "info", "--tags", "--exact-match")
		env.On("HasParentFilePath", ".svn").Return(fileInfo, nil)
		env.On("RunCommand", "svn", []string{"info", "", "--show-item", "revision"}).Return(tc.RefOutput, nil)
		env.On("RunCommand", "svn", []string{"info", "", "--show-item", "relative-url"}).Return(tc.BranchOutput, nil)
		env.On("RunCommand", "svn", []string{"status", ""}).Return(tc.StatusOutput, nil)

		s := &Svn{
			scm: scm{
				env:     env,
				props:   properties.Map{},
				command: SVNCOMMAND,
			},
		}
		s.setSvnStatus()
		if tc.ExpectedWorking == nil {
			tc.ExpectedWorking = &SvnStatus{}
		}
		assert.Equal(t, tc.ExpectedWorking, s.Working, tc.Case)
		assert.Equal(t, tc.ExpectedRef, s.BaseRev, tc.Case)
		assert.Equal(t, tc.ExpectedBranch, s.Branch, tc.Case)
		assert.Equal(t, tc.ExpectedChanged, s.Working.Changed(), tc.Case)
		assert.Equal(t, tc.ExpectedConflicts, s.Working.HasConflicts(), tc.Case)
	}
}
