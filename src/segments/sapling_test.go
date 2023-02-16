package segments

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

func TestSetDir(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Path     string
		GOOS     string
	}{
		{
			Case:     "In home folder",
			Expected: "~/sapling",
			Path:     "/usr/home/sapling/.sl",
			GOOS:     platform.LINUX,
		},
		{
			Case:     "Outside home folder",
			Expected: "/usr/sapling/repo",
			Path:     "/usr/sapling/repo/.sl",
			GOOS:     platform.LINUX,
		},
		{
			Case:     "Windows home folder",
			Expected: "~\\sapling",
			Path:     "\\usr\\home\\sapling\\.sl",
			GOOS:     platform.WINDOWS,
		},
		{
			Case:     "Windows outside home folder",
			Expected: "\\usr\\sapling\\repo",
			Path:     "\\usr\\sapling\\repo\\.sl",
			GOOS:     platform.WINDOWS,
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return(tc.GOOS)
		home := "/usr/home"
		if tc.GOOS == platform.WINDOWS {
			home = "\\usr\\home"
		}
		env.On("Home").Return(home)
		sl := &Sapling{
			scm: scm{
				env: env,
			},
		}
		sl.setDir(tc.Path)
		assert.Equal(t, tc.Expected, sl.Dir, tc.Case)
	}
}

func TestSetCommitContext(t *testing.T) {
	cases := []struct {
		Case   string
		Output string
		Error  error

		ExpectedHash      string
		ExpectedShortHash string
		ExpectedWhen      string
		ExpectedAuthor    string
		ExpectedBookmark  string
	}{
		{
			Case:  "Error",
			Error: errors.New("error"),
		},
		{
			Case: "No output",
		},
		{
			Case: "All output",
			Output: `
			no:734349e9f1abd229ec6e9bbebed35aed56b26a9e
    		ns:734349e9f
    		nd:23 minutes ago
    		un:jan
    		bm:sapling-segment
			`,
			ExpectedHash:      "734349e9f1abd229ec6e9bbebed35aed56b26a9e",
			ExpectedShortHash: "734349e9f",
			ExpectedWhen:      "23 minutes ago",
			ExpectedAuthor:    "jan",
			ExpectedBookmark:  "sapling-segment",
		},
		{
			Case:   "Short line",
			Output: "er",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("RunCommand", "sl", []string{"log", "--limit", "1", "--template", SLCOMMITTEMPLATE}).Return(tc.Output, tc.Error)
		sl := &Sapling{
			scm: scm{
				env:     env,
				command: SAPLINGCOMMAND,
			},
		}
		sl.setCommitContext()
		assert.Equal(t, tc.ExpectedHash, sl.Hash, tc.Case)
		assert.Equal(t, tc.ExpectedShortHash, sl.ShortHash, tc.Case)
		assert.Equal(t, tc.ExpectedWhen, sl.When, tc.Case)
		assert.Equal(t, tc.ExpectedAuthor, sl.Author, tc.Case)
		assert.Equal(t, tc.ExpectedBookmark, sl.Bookmark, tc.Case)
	}
}

func TestShouldDisplay(t *testing.T) {
	cases := []struct {
		Case       string
		HasSapling bool
		InRepo     bool
		Expected   bool
		Excluded   bool
	}{
		{
			Case: "Sapling not installed",
		},
		{
			Case:       "Sapling installed, not in repo",
			HasSapling: true,
		},
		{
			Case:       "Sapling installed, in repo but ignored",
			HasSapling: true,
			InRepo:     true,
			Excluded:   true,
		},
		{
			Case:       "Sapling installed, in repo",
			HasSapling: true,
			InRepo:     true,
			Expected:   true,
		},
	}
	fileInfo := &platform.FileInfo{
		Path:         "/sapling/repo/.sl",
		ParentFolder: "/sapling/repo",
		IsDir:        true,
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("HasCommand", "sl").Return(tc.HasSapling)
		env.On("InWSLSharedDrive").Return(false)
		env.On("GOOS").Return(platform.LINUX)
		env.On("Home").Return("/usr/home/sapling")
		env.On("DirMatchesOneOf", fileInfo.ParentFolder, []string{"/sapling/repo"}).Return(tc.Excluded)
		if tc.InRepo {
			env.On("HasParentFilePath", ".sl").Return(fileInfo, nil)
		} else {
			env.On("HasParentFilePath", ".sl").Return(&platform.FileInfo{}, errors.New("error"))
		}
		sl := &Sapling{
			scm: scm{
				env: env,
				props: &properties.Map{
					properties.ExcludeFolders: []string{"/sapling/repo"},
				},
			},
		}
		got := sl.shouldDisplay()
		assert.Equal(t, tc.Expected, got, tc.Case)
		if tc.Expected {
			assert.Equal(t, "/sapling/repo/.sl", sl.workingDir, tc.Case)
			assert.Equal(t, "/sapling/repo/.sl", sl.rootDir, tc.Case)
			assert.Equal(t, "/sapling/repo", sl.realDir, tc.Case)
			assert.Equal(t, "repo", sl.RepoName, tc.Case)
		}
	}
}

func TestSetHeadContext(t *testing.T) {
	cases := []struct {
		Case        string
		FetchStatus bool
		Output      string
		Expected    string
	}{
		{
			Case: "Do not fetch status",
		},
		{
			Case:        "Fetch status, no output",
			FetchStatus: true,
		},
		{
			Case:        "Fetch status, changed files",
			FetchStatus: true,
			Output: `
			M file.go
			M file2.go
			`,
			Expected: "~2",
		},
		{
			Case:        "Fetch status, all cases",
			FetchStatus: true,
			Output: `
			M file.go
			R file2.go
			A file3.go
			C file4.go
			! missing.go
			? untracked.go
			? untracked.go
			I ignored.go
			I ignored.go
			`,
			Expected: "?2 +1 ~1 -1 !1 =1 Ø2",
		},
	}
	output := `
	no:734349e9f1abd229ec6e9bbebed35aed56b26a9e
	ns:734349e9f
	nd:23 minutes ago
	un:jan
	bm:sapling-segment
	`
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("RunCommand", "sl", []string{"log", "--limit", "1", "--template", SLCOMMITTEMPLATE}).Return(output, nil)
		env.On("RunCommand", "sl", []string{"status"}).Return(tc.Output, nil)
		sl := &Sapling{
			scm: scm{
				env: env,
				props: &properties.Map{
					FetchStatus: tc.FetchStatus,
				},
				command: SAPLINGCOMMAND,
			},
		}
		sl.setHeadContext()
		got := sl.Working.String()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
