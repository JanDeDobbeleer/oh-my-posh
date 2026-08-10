package segments

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/ini"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	branchName      = "main"
	dotGit          = "dev/.git"
	dotGitSubmodule = "dev/.git/modules/submodule"
)

func TestWorktreeAdminIndex(t *testing.T) {
	cases := []struct {
		Case string
		Path string
		// Expected is the common git directory a caller slices out, or "" when Path is
		// not a worktree administrative directory.
		Expected string
	}{
		{Case: "admin dir", Path: "/repo/.git/worktrees/feat", Expected: "/repo/.git"},
		{Case: "trailing separator", Path: "/repo/.git/worktrees/feat/", Expected: "/repo/.git"},
		{Case: "dot component", Path: "/repo/.git/worktrees/feat/.", Expected: "/repo/.git"},
		{Case: "doubled separator", Path: "/repo/.git/worktrees//feat", Expected: "/repo/.git"},
		{Case: "nested worktrees keeps the last", Path: "/a/.git/worktrees/x/.git/worktrees/y", Expected: "/a/.git/worktrees/x/.git"},
		{Case: "two components after worktrees", Path: "/repo/.git/worktrees/a/b", Expected: ""},
		{Case: "repo under a worktrees component", Path: "/home/me/worktrees/proj/.bare", Expected: ""},
		{Case: "no name after worktrees", Path: "/repo/.git/worktrees/", Expected: ""},
		{Case: "no worktrees segment", Path: "/repo/.git", Expected: ""},
		{Case: "empty", Path: "", Expected: ""},
		// Shape alone cannot reject this one: the remainder is a single component. Task 2's
		// metadata back-reference check is what keeps it out of the worktree branch.
		{Case: "bare layout in a dir named worktrees", Path: "/home/me/worktrees/.bare", Expected: "/home/me"},
	}

	for _, tc := range cases {
		var got string
		if index := worktreeAdminIndex(tc.Path); index > -1 {
			got = filepath.ToSlash(filepath.Clean(tc.Path))[:index]
		}

		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestCommonGitDir(t *testing.T) {
	cases := []struct {
		Case       string
		ScmDir     string
		MainSCMDir string
		Expected   string
	}{
		{
			Case:       "scmDir wins over the worktrees cut",
			ScmDir:     "/repo/.git",
			MainSCMDir: "/repo/.git/worktrees/linked",
			Expected:   "/repo/.git",
		},
		{
			Case:       "checkout path containing worktrees does not fool the cut",
			ScmDir:     "/x/sepdir",
			MainSCMDir: "/x/worktrees/trap/",
			Expected:   "/x/sepdir",
		},
		{
			Case:       "falls back to the cut when scmDir is empty",
			MainSCMDir: "/repo/.git/worktrees/linked",
			Expected:   "/repo/.git",
		},
		{
			Case:     "empty when nothing is known",
			Expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			g := &Git{
				Scm: Scm{
					scmDir:     tc.ScmDir,
					mainSCMDir: tc.MainSCMDir,
				},
			}

			assert.Equal(t, tc.Expected, g.commonGitDir())
		})
	}
}

func TestEnabledGitNotFound(t *testing.T) {
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasParentFilePath", ".git", true).Return((*runtime.FileInfo)(nil), errors.New("no .git found (mock)"))
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)

	g := &Git{}
	g.Init(options.Map{}, env)

	assert.False(t, g.Enabled())
}

func TestEnabledInWorkingDirectory(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "git").Return(true)
	env.On("GOOS").Return("")
	env.On("FileContent", "/dir/hello/HEAD").Return("")
	env.MockGitCommand(fileInfo.Path, "1234567890abcdef1234567890abcdef12345678", "rev-parse", "HEAD")
	env.MockGitCommand(fileInfo.Path, "", "describe", "--tags", "--exact-match")
	env.On("IsWsl").Return(false)
	env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(poshHome)
	env.On("Getenv", poshGitEnv).Return("")
	env.On("DirMatchesOneOf", testify_.Anything, testify_.Anything).Return(false)

	g := &Git{}
	g.Init(options.Map{}, env)

	assert.True(t, g.Enabled())
	assert.Equal(t, fileInfo.Path, g.mainSCMDir)
}

func TestResolveEmptyGitPath(t *testing.T) {
	base := "base"
	assert.Equal(t, base, resolveGitPath(base, ""))
}

func TestEnabledInWorktree(t *testing.T) {
	cases := []struct {
		ExpectedWorkingFolder *string
		ExpectedRootFolder    *string
		ExpectedRealFolder    *string
		DiscoveredParent      string
		MetadataAddon         string
		MetadataContent       string
		ExpectedProbedDir     string
		Case                  string
		Pointer               string
		DiscoveredGitFile     string
		RawGitFile            string
		ExpectedIsWorkTree    bool
		ExpectedEnabled       bool
	}{
		{
			Case:                  "worktree",
			ExpectedEnabled:       true,
			ExpectedIsWorkTree:    true,
			Pointer:               TestRootPath + "dev/.git/worktrees/folder_worktree",
			DiscoveredGitFile:     TestRootPath + "dev/worktree/.git",
			DiscoveredParent:      TestRootPath + "dev/worktree",
			MetadataAddon:         "gitdir",
			MetadataContent:       TestRootPath + "dev/worktree.git\n",
			ExpectedProbedDir:     TestRootPath + "dev/.git/worktrees/folder_worktree",
			ExpectedWorkingFolder: new(TestRootPath + "dev/.git/worktrees/folder_worktree"),
			ExpectedRealFolder:    new(TestRootPath + "dev/worktree"),
			ExpectedRootFolder:    new(TestRootPath + dotGit),
		},
		{
			Case:                  "submodule",
			ExpectedEnabled:       true,
			Pointer:               "./.git/modules/submodule",
			DiscoveredGitFile:     TestRootPath + dotGit,
			DiscoveredParent:      TestRootPath + "dev",
			ExpectedProbedDir:     TestRootPath + dotGitSubmodule,
			ExpectedWorkingFolder: new(TestRootPath + dotGitSubmodule),
			ExpectedRealFolder:    new(TestRootPath + dotGitSubmodule),
			ExpectedRootFolder:    new(TestRootPath + dotGitSubmodule),
		},
		{
			Case:                  "submodule with root working folder",
			ExpectedEnabled:       true,
			Pointer:               TestRootPath + dotGitSubmodule,
			DiscoveredGitFile:     TestRootPath + dotGit,
			DiscoveredParent:      TestRootPath + "dev",
			ExpectedProbedDir:     TestRootPath + dotGitSubmodule,
			ExpectedWorkingFolder: new(TestRootPath + dotGitSubmodule),
			ExpectedRealFolder:    new(TestRootPath + dotGitSubmodule),
			ExpectedRootFolder:    new(TestRootPath + dotGitSubmodule),
		},
		{
			// Directory-role assertions are reserved for a later decision.
			Case:               "submodule with worktrees",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: true,
			Pointer:            TestRootPath + "dev/.git/modules/module/path/worktrees/location",
			DiscoveredGitFile:  TestRootPath + dotGit,
			DiscoveredParent:   TestRootPath + "dev",
			MetadataAddon:      "gitdir",
			MetadataContent:    TestRootPath + "dev/worktree.git\n",
			ExpectedProbedDir:  TestRootPath + "dev/.git/modules/module/path/worktrees/location",
		},
		{
			Case:                  "separate git dir",
			ExpectedEnabled:       true,
			Pointer:               TestRootPath + "dev/separate/.git/posh",
			DiscoveredGitFile:     TestRootPath + dotGit,
			DiscoveredParent:      TestRootPath + "dev",
			ExpectedProbedDir:     TestRootPath + "dev/separate/.git/posh",
			ExpectedWorkingFolder: new(TestRootPath + "dev/separate/.git/posh"),
			ExpectedRealFolder:    new(TestRootPath + "dev/"),
			ExpectedRootFolder:    new(TestRootPath + "dev/separate/.git/posh"),
		},
		{
			Case:                  "bare repo through a .git pointer",
			ExpectedEnabled:       true,
			Pointer:               TestRootPath + "dev/.bare",
			DiscoveredGitFile:     TestRootPath + dotGit,
			DiscoveredParent:      TestRootPath + "dev",
			ExpectedProbedDir:     TestRootPath + "dev/.bare",
			ExpectedWorkingFolder: new(TestRootPath + "dev/.bare"),
			ExpectedRealFolder:    new(TestRootPath + "dev/"),
			ExpectedRootFolder:    new(TestRootPath + "dev/.bare"),
		},
		{
			Case:                  "worktree with relative gitdir path",
			ExpectedEnabled:       true,
			ExpectedIsWorkTree:    true,
			Pointer:               TestRootPath + "dev/.git/worktrees/folder_worktree",
			DiscoveredGitFile:     TestRootPath + "dev/worktree/.git",
			DiscoveredParent:      TestRootPath + "dev/worktree",
			MetadataAddon:         "gitdir",
			MetadataContent:       "../../../worktree/.git\n",
			ExpectedProbedDir:     TestRootPath + "dev/.git/worktrees/folder_worktree",
			ExpectedWorkingFolder: new(TestRootPath + "dev/.git/worktrees/folder_worktree"),
			ExpectedRealFolder:    new(TestRootPath + "dev/worktree"),
			ExpectedRootFolder:    new(TestRootPath + dotGit),
		},
		{
			Case:                  "worktree with relative gitdir path, no trailing newline",
			ExpectedEnabled:       true,
			ExpectedIsWorkTree:    true,
			Pointer:               TestRootPath + "dev/.git/worktrees/folder_worktree",
			DiscoveredGitFile:     TestRootPath + "dev/worktree/.git",
			DiscoveredParent:      TestRootPath + "dev/worktree",
			MetadataAddon:         "gitdir",
			MetadataContent:       "../../../worktree/.git",
			ExpectedProbedDir:     TestRootPath + "dev/.git/worktrees/folder_worktree",
			ExpectedWorkingFolder: new(TestRootPath + "dev/.git/worktrees/folder_worktree"),
			ExpectedRealFolder:    new(TestRootPath + "dev/worktree"),
			ExpectedRootFolder:    new(TestRootPath + dotGit),
		},
		{
			// Shape matches, but metadata does not point back to the discovered .git file.
			Case:              "bare layout in a dir named worktrees",
			ExpectedEnabled:   true,
			Pointer:           TestRootPath + "me/worktrees/.bare",
			DiscoveredGitFile: TestRootPath + "me/worktrees/.git",
			DiscoveredParent:  TestRootPath + "me/worktrees",
			MetadataAddon:     "gitdir",
			ExpectedProbedDir: TestRootPath + "me/worktrees/.bare",
		},
		{
			Case:               "worktree metadata is whitespace only",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: false,
			Pointer:            TestRootPath + "dev/.git/worktrees/folder_worktree",
			DiscoveredGitFile:  TestRootPath + "dev/.git/worktrees/folder_worktree/.git",
			DiscoveredParent:   TestRootPath + "dev/.git/worktrees/folder_worktree",
			MetadataAddon:      "gitdir",
			MetadataContent:    " \n",
			ExpectedProbedDir:  TestRootPath + "dev/.git/worktrees/folder_worktree",
		},
		{
			Case:               "worktree metadata points somewhere else",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: false,
			Pointer:            TestRootPath + "dev/.git/worktrees/folder_worktree",
			DiscoveredGitFile:  TestRootPath + "dev/worktree/.git",
			DiscoveredParent:   TestRootPath + "dev/worktree",
			MetadataAddon:      "gitdir",
			MetadataContent:    TestRootPath + "dev/moved-elsewhere/.git\n",
			ExpectedProbedDir:  TestRootPath + "dev/.git/worktrees/folder_worktree",
		},
		{
			Case:               "worktree metadata spelled with a trailing separator",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: true,
			Pointer:            TestRootPath + "dev/.git/worktrees/folder_worktree",
			DiscoveredGitFile:  TestRootPath + "dev/worktree/.git",
			DiscoveredParent:   TestRootPath + "dev/worktree",
			MetadataAddon:      "gitdir",
			MetadataContent:    TestRootPath + "dev/worktree/.git\n",
			ExpectedProbedDir:  TestRootPath + "dev/.git/worktrees/folder_worktree",
		},
		{
			Case:               "worktree metadata is garbage that resolves nowhere",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: false,
			Pointer:            TestRootPath + "dev/.git/worktrees/folder_worktree",
			DiscoveredGitFile:  TestRootPath + "dev/worktree/.git",
			DiscoveredParent:   TestRootPath + "dev/worktree",
			MetadataAddon:      "gitdir",
			MetadataContent:    "not-a-path-at-all\n",
			ExpectedProbedDir:  TestRootPath + "dev/.git/worktrees/folder_worktree",
		},
		{
			// Both spelling forms must reject a shape match whose metadata does not match.
			Case:              "repo under a worktrees path component, absolute spelling",
			ExpectedEnabled:   true,
			Pointer:           TestRootPath + "me/worktrees/proj/.bare",
			DiscoveredGitFile: TestRootPath + "me/worktrees/proj/.git",
			DiscoveredParent:  TestRootPath + "me/worktrees/proj",
			ExpectedProbedDir: TestRootPath + "me/worktrees/proj/.bare",
		},
		{
			Case:               "genuine worktree whose common dir is under a worktrees component",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: true,
			Pointer:            TestRootPath + "me/worktrees/proj/.git/worktrees/feat",
			DiscoveredGitFile:  TestRootPath + "me/checkouts/feat/.git",
			DiscoveredParent:   TestRootPath + "me/checkouts/feat",
			MetadataAddon:      "gitdir",
			MetadataContent:    TestRootPath + "me/checkouts/feat/.git\n",
			ExpectedProbedDir:  TestRootPath + "me/worktrees/proj/.git/worktrees/feat",
		},
		{
			Case:              "bare layout, dot-slash relative pointer",
			ExpectedEnabled:   true,
			Pointer:           "./.bare",
			DiscoveredGitFile: TestRootPath + "repo/.git",
			DiscoveredParent:  TestRootPath + "repo",
			ExpectedProbedDir: TestRootPath + "repo/.bare",
		},
		{
			Case:              "bare layout, bare relative pointer",
			ExpectedEnabled:   true,
			Pointer:           ".bare",
			DiscoveredGitFile: TestRootPath + "repo/.git",
			DiscoveredParent:  TestRootPath + "repo",
			ExpectedProbedDir: TestRootPath + "repo/.bare",
		},
		{
			Case:              "relative pointer into a subdirectory",
			ExpectedEnabled:   true,
			Pointer:           "sub/git",
			DiscoveredGitFile: TestRootPath + "repo/.git",
			DiscoveredParent:  TestRootPath + "repo",
			ExpectedProbedDir: TestRootPath + "repo/sub/git",
		},
		{
			// Absolute pointers retain their trailing separator; filepath.Join cleans relative ones.
			Case:               "absolute pointer ending in a separator",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: true,
			Pointer:            TestRootPath + "repo/.git/worktrees/feat/",
			DiscoveredGitFile:  TestRootPath + "repo/feat/.git",
			DiscoveredParent:   TestRootPath + "repo/feat",
			MetadataAddon:      "gitdir",
			MetadataContent:    TestRootPath + "repo/feat/.git\n",
			ExpectedProbedDir:  TestRootPath + "repo/.git/worktrees/feat/",
		},
		{
			// This pins the existing trim; Git treats a trailing space as part of the path.
			Case:              "pointer with trailing whitespace is trimmed",
			ExpectedEnabled:   true,
			RawGitFile:        "gitdir: ./.bare \n",
			DiscoveredGitFile: TestRootPath + "repo/.git",
			DiscoveredParent:  TestRootPath + "repo",
			ExpectedProbedDir: TestRootPath + "repo/.bare",
		},
		{
			Case:              "malformed .git file with no gitdir line",
			ExpectedEnabled:   false,
			RawGitFile:        "not a gitdir line",
			DiscoveredGitFile: TestRootPath + "repo/.git",
			DiscoveredParent:  TestRootPath + "repo",
			ExpectedProbedDir: TestRootPath + "repo",
		},
		{
			Case:              "gitdir line with an empty pointer",
			ExpectedEnabled:   false,
			RawGitFile:        "gitdir: ",
			DiscoveredGitFile: TestRootPath + "repo/.git",
			DiscoveredParent:  TestRootPath + "repo",
			ExpectedProbedDir: TestRootPath + "repo",
		},
		{
			// Keep this check on the raw pointer: resolving it first changes the branch.
			Case:               "worktree with a relative pointer under a modules component",
			ExpectedEnabled:    true,
			ExpectedIsWorkTree: true,
			Pointer:            "../.git/worktrees/feat",
			DiscoveredGitFile:  TestRootPath + "modules/repo/feat/.git",
			DiscoveredParent:   TestRootPath + "modules/repo/feat",
			MetadataAddon:      "gitdir",
			MetadataContent:    TestRootPath + "modules/repo/feat/.git\n",
			ExpectedProbedDir:  TestRootPath + "modules/repo/.git/worktrees/feat",
		},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{Path: tc.DiscoveredGitFile, ParentFolder: tc.DiscoveredParent}
		env := new(mock.Environment)
		gitFileContent := fmt.Sprintf("gitdir: %s", tc.Pointer)
		if tc.RawGitFile != "" {
			gitFileContent = tc.RawGitFile
		}
		env.On("FileContent", tc.DiscoveredGitFile).Return(gitFileContent)
		env.On("FileContent", filepath.Join(tc.ExpectedProbedDir, tc.MetadataAddon)).Return(tc.MetadataContent)
		env.On("HasFilesInDir", tc.ExpectedProbedDir, tc.MetadataAddon).Return(tc.MetadataAddon != "")
		env.On("HasFilesInDir", tc.ExpectedProbedDir, "HEAD").Return(true)
		env.On("PathSeparator").Return(string(os.PathSeparator))

		g := &Git{}
		g.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, g.hasWorktree(fileInfo), tc.Case)
		assert.Equal(t, tc.ExpectedIsWorkTree, g.IsWorkTree, tc.Case)

		if tc.ExpectedWorkingFolder != nil {
			assert.Equal(t, *tc.ExpectedWorkingFolder, g.mainSCMDir, tc.Case)
		}

		if tc.ExpectedRealFolder != nil {
			assert.Equal(t, *tc.ExpectedRealFolder, g.repoRootDir, tc.Case)
		}

		if tc.ExpectedRootFolder != nil {
			assert.Equal(t, *tc.ExpectedRootFolder, g.scmDir, tc.Case)
		}

		if tc.ExpectedEnabled {
			assert.True(t, filepath.IsAbs(g.mainSCMDir), tc.Case+": mainSCMDir must be absolute")
			assert.True(t, filepath.IsAbs(g.scmDir), tc.Case+": scmDir must be absolute")
		}
	}
}

func TestEnabledInBareLayout(t *testing.T) {
	cases := []struct {
		Case          string
		FetchBareInfo bool
	}{
		{Case: "bare layout without fetch_bare_info", FetchBareInfo: false},
		{Case: "bare layout with fetch_bare_info", FetchBareInfo: true},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{
			Path:         "/repo/.git",
			ParentFolder: "/repo",
		}

		env := new(mock.Environment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("HasCommand", "git").Return(true)
		env.On("GOOS").Return("")
		env.On("IsWsl").Return(false)
		env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return(poshHome)
		env.On("Getenv", poshGitEnv).Return("")
		env.On("DirMatchesOneOf", testify_.Anything, testify_.Anything).Return(false)
		env.On("FileContent", "/repo/.git").Return("gitdir: ./.bare")
		env.On("HasFilesInDir", "/repo/.bare", "HEAD").Return(true)
		env.On("FileContent", "/repo/.bare/config").Return("[core]\n\tbare = true")
		env.On("FileContent", "/repo/.bare/HEAD").Return("")
		env.MockGitCommand("/repo/", "1234567890abcdef1234567890abcdef12345678", "rev-parse", "HEAD")
		env.MockGitCommand("/repo/", "", "describe", "--tags", "--exact-match")
		env.MockGitCommand("/repo/", "", "remote")

		props := options.Map{}
		if tc.FetchBareInfo {
			props = options.Map{FetchBareInfo: true}
		}

		g := &Git{}
		g.Init(props, env)

		assert.True(t, g.Enabled(), tc.Case)
		assert.Equal(t, tc.FetchBareInfo, g.IsBare, tc.Case)
		assert.True(t, filepath.IsAbs(g.mainSCMDir), tc.Case+": mainSCMDir must be absolute")
		assert.True(t, filepath.IsAbs(g.scmDir), tc.Case+": scmDir must be absolute")
	}
}

func TestIsBareRepoResolvesPointer(t *testing.T) {
	cases := []struct {
		Case              string
		Pointer           string
		ExpectedConfig    string
		ExpectedProbedDir string
		OldConfig         string
		ExpectedIsBare    bool
	}{
		{
			Case:              "relative pointer",
			Pointer:           "./.bare",
			ExpectedConfig:    "/repo/.bare/config",
			ExpectedProbedDir: "/repo/.bare",
			ExpectedIsBare:    true,
		},
		{
			Case:              "absolute pointer",
			Pointer:           "/repo/.bare",
			ExpectedConfig:    "/repo/.bare/config",
			ExpectedProbedDir: "/repo/.bare",
			ExpectedIsBare:    true,
			OldConfig:         "/repo/repo/.bare/config",
		},
		{
			Case:              "absolute pointer to a non-bare git dir",
			Pointer:           "/elsewhere/gitdir",
			ExpectedConfig:    "/elsewhere/gitdir/config",
			ExpectedProbedDir: "/elsewhere/gitdir",
			ExpectedIsBare:    false,
			OldConfig:         "/repo/elsewhere/gitdir/config",
		},
	}

	for _, tc := range cases {
		fileInfo := &runtime.FileInfo{
			Path:         "/repo/.git",
			ParentFolder: "/repo",
		}

		env := new(mock.Environment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("HasCommand", "git").Return(true)
		env.On("GOOS").Return("")
		env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
		env.On("FileContent", "/repo/.git").Return(fmt.Sprintf("gitdir: %s", tc.Pointer))
		env.On("FileContent", tc.ExpectedConfig).Return(fmt.Sprintf("[core]\n\tbare = %t", tc.ExpectedIsBare))
		env.On("HasFilesInDir", tc.ExpectedProbedDir, "HEAD").Return(true)

		if tc.OldConfig != "" {
			env.On("FileContent", tc.OldConfig).Return("")
		}

		g := &Git{}
		g.Init(options.Map{FetchBareInfo: true}, env)

		assert.True(t, g.shouldDisplay(), tc.Case)
		assert.Equal(t, tc.ExpectedIsBare, g.IsBare, tc.Case)
		env.AssertCalled(t, "FileContent", tc.ExpectedConfig)

		if tc.OldConfig != "" {
			env.AssertNotCalled(t, "FileContent", tc.OldConfig)
		}
	}
}

func TestShouldDisplayInitializesWSLBeforeBareRepoDetection(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/repo/.git",
		ParentFolder: "/repo",
	}

	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(true)
	env.On("HasCommand", "git.exe").Return(true)
	env.On("GOOS").Return("")
	env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
	env.On("ConvertToLinuxPath").Return("/mnt/c/repo/.bare")
	env.On("FileContent", "/repo/.git").Return("gitdir: C:/repo/.bare")
	env.On("FileContent", "/repo/C:/repo/.bare/config").Return("")
	env.On("FileContent", "/mnt/c/repo/.bare/config").Return("[core]\n\tbare = true")
	env.On("HasFilesInDir", "/mnt/c/repo/.bare", "HEAD").Return(true)
	env.On("ConvertToWindowsPath", "/repo/").Return("C:/repo")

	g := &Git{}
	g.Init(options.Map{FetchBareInfo: true}, env)

	assert.True(t, g.shouldDisplay())
	assert.True(t, g.IsBare)
	env.AssertNumberOfCalls(t, "ConvertToLinuxPath", 2)
	env.AssertCalled(t, "FileContent", "/mnt/c/repo/.bare/config")
}

func TestEnabledInBareRepo(t *testing.T) {
	cases := []struct {
		Case            string
		HEAD            string
		GitDirPath      string
		GitFileContent  string
		Config          string
		ExpectedRemotes int
		GitDirIsDir     bool
		IsBare          bool
	}{
		{
			Case:        "Bare repo on main",
			IsBare:      true,
			GitDirPath:  "git",
			GitDirIsDir: true,
			HEAD:        "ref: refs/heads/main",
			Config:      "[core]\n\tbare = true",
		},
		{
			Case:        "Not a bare repo",
			IsBare:      false,
			GitDirPath:  "git",
			GitDirIsDir: true,
			HEAD:        "ref: refs/heads/main",
			Config:      "[core]\n\tbare = false",
		},
		{
			Case:            "Linked worktree probe does not poison the remotes memo",
			IsBare:          false,
			GitDirPath:      "/repo/.git",
			GitDirIsDir:     false,
			GitFileContent:  "gitdir: /repo/.git/worktrees/linked",
			Config:          "[remote \"origin\"]\n\turl = git@github.com:JanDeDobbeleer/test.git",
			ExpectedRemotes: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("InWSLSharedDrive").Return(false)
			env.On("GOOS").Return("")
			env.On("HasCommand", "git").Return(true)
			env.On("PathSeparator").Return("/")

			fileInfo := &runtime.FileInfo{
				IsDir:        tc.GitDirIsDir,
				Path:         tc.GitDirPath,
				ParentFolder: "/repo",
			}
			env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
			env.On("FileContent", tc.GitDirPath).Return(tc.GitFileContent)
			env.On("FileContent", "/repo/.git/worktrees/linked/gitdir").Return("/repo/linked/.git\n")
			env.On("FileContent", tc.GitDirPath+"/HEAD").Return(tc.HEAD)
			env.On("FileContent", testify_.Anything).Return(tc.Config)
			env.On("HasFilesInDir", testify_.Anything, testify_.Anything).Return(true)
			env.On("HasFolder", testify_.Anything).Return(false)
			env.On("RunCommand", testify_.Anything, testify_.Anything).Return("", nil)

			props := options.Map{FetchBareInfo: true}

			g := &Git{}
			g.Init(props, env)

			_ = g.Enabled()

			assert.Equal(t, tc.IsBare, g.IsBare, tc.Case)

			if tc.ExpectedRemotes > 0 {
				assert.Len(t, g.Remotes(), tc.ExpectedRemotes, "%s: remotes must survive the bare probe", tc.Case)
			}
		})
	}
}

func TestIsBareRepoDoesNotReuseConfigMemo(t *testing.T) {
	env := new(mock.Environment)
	env.On("FileContent", "first/config").Return("[core]\n\tbare = false")
	env.On("FileContent", "second/config").Return("[core]\n\tbare = true")

	g := &Git{}
	g.Init(options.Map{}, env)

	assert.False(t, g.isBareRepo(&runtime.FileInfo{Path: "first", IsDir: true}))
	assert.True(t, g.isBareRepo(&runtime.FileInfo{Path: "second", IsDir: true}))
}

func TestGetGitOutputForCommand(t *testing.T) {
	args := []string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}
	commandArgs := []string{"symbolic-ref", "--short", "HEAD"}
	want := "je suis le output"
	env := new(mock.Environment)
	env.On("IsWsl").Return(false)
	env.On("RunCommand", "git", append(args, commandArgs...)).Return(want, nil)
	env.On("GOOS").Return("unix")

	g := &Git{
		Scm: Scm{
			command: GITCOMMAND,
		},
	}
	g.Init(options.Map{}, env)

	got := g.getGitCommandOutput(commandArgs...)
	assert.Equal(t, want, got)
}

func TestSetGitHEADContextClean(t *testing.T) {
	cases := []struct {
		Ours        string
		Expected    string
		Ref         string
		Case        string
		Total       string
		Step        string
		Theirs      string
		RebaseMerge bool
		Sequencer   bool
		Revert      bool
		CherryPick  bool
		Merge       bool
		RebaseApply bool
	}{
		{Case: "detached on commit", Ref: DETACHED, Expected: "branch detached at commit 1234567"},
		{Case: "not detached, clean", Ref: "main", Expected: "branch main"},
		{
			Case:        "rebase merge",
			Ref:         DETACHED,
			Expected:    "rebase branch origin/main onto branch main (1/2) at commit 1234567",
			RebaseMerge: true,
			Ours:        "refs/heads/origin/main",
			Theirs:      "main",
			Step:        "1",
			Total:       "2",
		},
		{
			Case:        "rebase apply",
			Ref:         DETACHED,
			Expected:    "rebase branch origin/main (1/2) at commit 1234567",
			RebaseApply: true,
			Ours:        "refs/heads/origin/main",
			Step:        "1",
			Total:       "2",
		},
		{
			Case:     "merge branch",
			Ref:      "main",
			Expected: "merge branch feat-1 into branch main",
			Merge:    true,
			Theirs:   "branch 'feat-1'",
			Ours:     "main",
		},
		{
			Case:     "merge commit",
			Ref:      "main",
			Expected: "merge commit 1234567 into branch main",
			Merge:    true,
			Theirs:   "commit '123456789101112'",
			Ours:     "main",
		},
		{
			Case:     "merge tag",
			Ref:      "main",
			Expected: "merge tag 1.2.4 into branch main",
			Merge:    true,
			Theirs:   "tag '1.2.4'",
			Ours:     "main",
		},
		{
			Case:       "cherry pick",
			Ref:        "main",
			Expected:   "pick commit 1234567 onto branch main",
			CherryPick: true,
			Theirs:     "123456789101012",
			Ours:       "main",
		},
		{
			Case:     "revert",
			Ref:      "main",
			Expected: "revert commit 1234567 onto branch main",
			Revert:   true,
			Theirs:   "123456789101012",
			Ours:     "main",
		},
		{
			Case:      "sequencer cherry",
			Ref:       "main",
			Expected:  "pick commit 1234567 onto branch main",
			Sequencer: true,
			Theirs:    "pick 123456789101012",
			Ours:      "main",
		},
		{
			Case:      "sequencer cherry p",
			Ref:       "main",
			Expected:  "pick commit 1234567 onto branch main",
			Sequencer: true,
			Theirs:    "p 123456789101012",
			Ours:      "main",
		},
		{
			Case:      "sequencer revert",
			Ref:       "main",
			Expected:  "revert commit 1234567 onto branch main",
			Sequencer: true,
			Theirs:    "revert 123456789101012",
			Ours:      "main",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("GOOS").Return("unix")
		env.On("IsWsl").Return(false)
		env.MockGitCommand("", "1234567890abcdef1234567890abcdef12345678", "rev-parse", "HEAD")
		env.MockGitCommand("", "", "describe", "--tags", "--exact-match")
		env.MockGitCommand("", tc.Theirs, "name-rev", "--name-only", "--exclude=tags/*", tc.Theirs)
		env.MockGitCommand("", tc.Ours, "name-rev", "--name-only", "--exclude=tags/*", tc.Ours)
		// rebase merge
		env.On("HasFolder", "/rebase-merge").Return(tc.RebaseMerge)
		env.On("FileContent", "/rebase-merge/head-name").Return(tc.Ours)
		env.On("FileContent", "/rebase-merge/onto").Return(tc.Theirs)
		env.On("FileContent", "/rebase-merge/msgnum").Return(tc.Step)
		env.On("FileContent", "/rebase-merge/end").Return(tc.Total)
		// rebase apply
		env.On("HasFolder", "/rebase-apply").Return(tc.RebaseApply)
		env.On("FileContent", "/rebase-apply/head-name").Return(tc.Ours)
		env.On("FileContent", "/rebase-apply/next").Return(tc.Step)
		env.On("FileContent", "/rebase-apply/last").Return(tc.Total)
		// merge
		env.On("HasFilesInDir", "", "MERGE_MSG").Return(tc.Merge)
		env.On("FileContent", "/MERGE_MSG").Return(fmt.Sprintf("Merge %s into %s", tc.Theirs, tc.Ours))
		// cherry pick
		env.On("HasFilesInDir", "", "CHERRY_PICK_HEAD").Return(tc.CherryPick)
		env.On("FileContent", "/CHERRY_PICK_HEAD").Return(tc.Theirs)
		// revert
		env.On("HasFilesInDir", "", "REVERT_HEAD").Return(tc.Revert)
		env.On("FileContent", "/REVERT_HEAD").Return(tc.Theirs)
		// sequencer
		env.On("HasFilesInDir", "", "sequencer/todo").Return(tc.Sequencer)
		env.On("FileContent", "/sequencer/todo").Return(tc.Theirs)

		props := options.Map{
			BranchIcon:     "branch ",
			CommitIcon:     "commit ",
			RebaseIcon:     "rebase ",
			MergeIcon:      "merge ",
			CherryPickIcon: "pick ",
			TagIcon:        "tag ",
			RevertIcon:     "revert ",
		}

		g := &Git{
			Scm: Scm{
				command: GITCOMMAND,
			},
			ShortHash: "1234567",
			Ref:       tc.Ref,
		}
		g.Init(props, env)
		g.mainSCMDir = ""

		g.setHEADStatus()
		assert.Equal(t, tc.Expected, g.HEAD, tc.Case)
	}
}

func TestSetPrettyHEADName(t *testing.T) {
	cases := []struct {
		Case         string
		Expected     string
		ShortHash    string
		Tag          string
		HEAD         string
		SymbolicName string
	}{
		{Case: "main", Expected: "branch main", HEAD: BRANCHPREFIX + "main"},
		{Case: "no hash", Expected: "commit 1234567", HEAD: "12345678910"},
		{Case: "hash on tag", ShortHash: "132312322321", Expected: "tag tag-1", HEAD: "12345678910", Tag: "tag-1"},
		{Case: "no hash on tag", Expected: "tag tag-1", Tag: "tag-1"},
		{Case: "hash on commit", ShortHash: "1234567", Expected: "commit 1234567"},
		{Case: "no hash on commit", Expected: "commit 1234567", HEAD: "12345678910"},
		{Case: "reftable main branch", Expected: "branch main", HEAD: "ref: refs/heads/.invalid", SymbolicName: "refs/heads/main"},
		{Case: "reftable detached head", Expected: "commit 1234567", HEAD: "ref: refs/heads/.invalid", SymbolicName: "fatal: ref HEAD is not a symbolic ref"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("FileContent", "/HEAD").Return(tc.HEAD)
		env.On("GOOS").Return("unix")
		env.On("IsWsl").Return(false)
		// Mock rev-parse HEAD for detached HEAD cases
		headValue := tc.HEAD
		if headValue == "" || strings.HasSuffix(tc.HEAD, ".invalid") {
			headValue = "12345678910"
		}
		env.MockGitCommand("", headValue, "rev-parse", "HEAD")
		env.MockGitCommand("", tc.Tag, "describe", "--tags", "--exact-match")
		env.MockGitCommand("", tc.SymbolicName, "rev-parse", "--symbolic-full-name", "HEAD")

		props := options.Map{
			BranchIcon: "branch ",
			CommitIcon: "commit ",
			TagIcon:    "tag ",
		}

		g := &Git{
			Scm: Scm{
				command: GITCOMMAND,
			},
			ShortHash: tc.ShortHash,
		}
		g.Init(props, env)
		g.mainSCMDir = ""

		g.updateHEADReference()
		assert.Equal(t, tc.Expected, g.HEAD, tc.Case)
	}
}

func TestSetGitStatus(t *testing.T) {
	cases := []struct {
		ExpectedWorking      *GitStatus
		ExpectedStaging      *GitStatus
		Case                 string
		Output               string
		ExpectedHash         string
		ExpectedRef          string
		ExpectedUpstream     string
		ExpectedAhead        int
		ExpectedBehind       int
		ExpectedUpstreamGone bool
		Rebase               bool
		Merge                bool
	}{
		{
			Case: "all different options on working and staging, no remote",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head rework-git-status
			1 .R N...
			1 .C N...
			1 .M N...
			1 .m N...
			1 .A N...
			1 .D N...
			1 .A N...
			1 .U N...
			1 A. N...
			`,
			ExpectedWorking:      &GitStatus{ScmStatus: ScmStatus{Modified: 4, Added: 2, Deleted: 1, Unmerged: 1}},
			ExpectedStaging:      &GitStatus{ScmStatus: ScmStatus{Added: 1}},
			ExpectedHash:         "1234567",
			ExpectedRef:          "rework-git-status",
			ExpectedUpstreamGone: true,
		},
		{
			Case: "all different options on working and staging, with remote",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head rework-git-status
			# branch.upstream origin/rework-git-status
			# branch.ab +0 -0
			1 .R N...
			1 .C N...
			1 .M N...
			1 .m N...
			1 .A N...
			1 .D N...
			1 .A N...
			1 .U N...
			1 A. N...
			`,
			ExpectedWorking:  &GitStatus{ScmStatus: ScmStatus{Modified: 4, Added: 2, Deleted: 1, Unmerged: 1}},
			ExpectedStaging:  &GitStatus{ScmStatus: ScmStatus{Added: 1}},
			ExpectedUpstream: "origin/rework-git-status",
			ExpectedHash:     "1234567",
			ExpectedRef:      "rework-git-status",
		},
		{
			Case: "remote with equal branch",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head rework-git-status
			# branch.upstream origin/rework-git-status
			# branch.ab +0 -0
			`,
			ExpectedUpstream: "origin/rework-git-status",
			ExpectedHash:     "1234567",
			ExpectedRef:      "rework-git-status",
		},
		{
			Case: "remote with branch status",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head rework-git-status
			# branch.upstream origin/rework-git-status
			# branch.ab +2 -1
			`,
			ExpectedUpstream: "origin/rework-git-status",
			ExpectedHash:     "1234567",
			ExpectedRef:      "rework-git-status",
			ExpectedAhead:    2,
			ExpectedBehind:   1,
		},
		{
			Case: "untracked files",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head main
			# branch.upstream origin/main
			# branch.ab +0 -0
			? q
			? qq
			? qqq
			`,
			ExpectedUpstream: "origin/main",
			ExpectedHash:     "1234567",
			ExpectedRef:      "main",
			ExpectedWorking:  &GitStatus{ScmStatus: ScmStatus{Untracked: 3}},
		},
		{
			Case: "remote branch was deleted",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head branch-is-gone
			# branch.upstream origin/branch-is-gone
			`,
			ExpectedUpstream:     "origin/branch-is-gone",
			ExpectedHash:         "1234567",
			ExpectedRef:          "branch-is-gone",
			ExpectedUpstreamGone: true,
		},
		{
			Case: "rebase with 2 merge conflicts",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head rework-git-status
			# branch.upstream origin/rework-git-status
			# branch.ab +0 -0
			1 AA N...
			1 AA N...
			`,
			ExpectedUpstream: "origin/rework-git-status",
			ExpectedHash:     "1234567",
			ExpectedRef:      "rework-git-status",
			Rebase:           true,
			ExpectedStaging:  &GitStatus{ScmStatus: ScmStatus{Unmerged: 2}},
		},
		{
			Case: "merge with 4 merge conflicts",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head rework-git-status
			# branch.upstream origin/rework-git-status
			# branch.ab +0 -0
			1 AA N...
			1 AA N...
			1 AA N...
			1 AA N...
			`,
			ExpectedUpstream: "origin/rework-git-status",
			ExpectedHash:     "1234567",
			ExpectedRef:      "rework-git-status",
			Merge:            true,
			ExpectedStaging:  &GitStatus{ScmStatus: ScmStatus{Unmerged: 4}},
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return("unix")
		env.On("IsWsl").Return(false)
		env.MockGitCommand("", strings.ReplaceAll(tc.Output, "\t", ""), "status", "-unormal", "--branch", "--porcelain=2")

		g := &Git{
			Scm: Scm{
				command: GITCOMMAND,
			},
		}
		g.Init(options.Map{}, env)

		if tc.ExpectedWorking == nil {
			tc.ExpectedWorking = &GitStatus{}
		}

		if tc.ExpectedStaging == nil {
			tc.ExpectedStaging = &GitStatus{}
		}

		if tc.Rebase {
			g.Rebase = &Rebase{}
		}

		g.Merge = tc.Merge
		tc.ExpectedStaging.Formats = map[string]string{}
		tc.ExpectedWorking.Formats = map[string]string{}
		g.setStatus()
		assert.Equal(t, tc.ExpectedStaging, g.Staging, tc.Case)
		assert.Equal(t, tc.ExpectedWorking, g.Working, tc.Case)
		assert.Equal(t, tc.ExpectedHash, g.ShortHash, tc.Case)
		assert.Equal(t, tc.ExpectedRef, g.Ref, tc.Case)
		assert.Equal(t, tc.ExpectedUpstream, g.Upstream, tc.Case)
		assert.Equal(t, tc.ExpectedUpstreamGone, g.UpstreamGone, tc.Case)
		assert.Equal(t, tc.ExpectedAhead, g.Ahead, tc.Case)
		assert.Equal(t, tc.ExpectedBehind, g.Behind, tc.Case)
	}
}

func TestGetStashContextZeroEntries(t *testing.T) {
	cases := []struct {
		StashContent string
		Expected     int
	}{
		{Expected: 0, StashContent: ""},
		{Expected: 2, StashContent: "1\n2\n"},
		{Expected: 4, StashContent: "1\n2\n3\n4\n\n"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("FileContent", "/logs/refs/stash").Return(tc.StashContent)

		g := &Git{
			Scm: Scm{
				mainSCMDir: "",
			},
		}
		g.Init(options.Map{}, env)

		got := g.StashCount()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGitCleanSSHURL(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Upstream string
	}{
		{Case: "regular URL", Expected: "https://src.example.com/user/repo", Upstream: "/src.example.com/user/repo.git"},
		{Case: "domain:path", Expected: "https://host.xz/path/to/repo", Upstream: "host.xz:/path/to/repo.git/"},
		{Case: "ssh with port", Expected: "https://host.xz/path/to/repo", Upstream: "ssh://user@host.xz:1234/path/to/repo.git"},
		{Case: "ssh with 3-digit port", Expected: "https://host.xz/path/to/repo", Upstream: "ssh://user@host.xz:234/path/to/repo.git"},
		{Case: "ssh with port, trailing slash", Expected: "https://host.xz/path/to/repo", Upstream: "ssh://user@host.xz:1234/path/to/repo.git/"},
		{Case: "ssh without port", Expected: "https://host.xz/path/to/repo", Upstream: "ssh://user@host.xz/path/to/repo.git/"},
		{Case: "ssh port, no user", Expected: "https://host.xz/path/to/repo", Upstream: "ssh://host.xz:1234/path/to/repo.git"},
		{Case: "ssh no port, no user", Expected: "https://host.xz/path/to/repo", Upstream: "ssh://host.xz/path/to/repo.git"},
		{Case: "rsync no port, no user", Expected: "https://host.xz/path/to/repo", Upstream: "rsync://host.xz/path/to/repo.git/"},
		{Case: "git no port, no user", Expected: "https://host.xz/path/to/repo", Upstream: "git://host.xz/path/to/repo.git"},
		{Case: "gitea no port, no user", Expected: "https://src.example.com/user/repo", Upstream: "_gitea@src.example.com:user/repo.git"},
		{Case: "git@ with user", Expected: "https://github.com/JanDeDobbeleer/oh-my-posh", Upstream: "git@github.com:JanDeDobbeleer/oh-my-posh"},
		{Case: "unsupported", Upstream: "\\test\\repo.git"},
		{Case: "Azure DevOps, https", Expected: "https://dev.azure.com/posh/oh-my-posh/_git/website", Upstream: "https://posh@dev.azure.com/posh/oh-my-posh/_git/website"},
		{Case: "Azure DevOps, ssh", Expected: "https://dev.azure.com/posh/oh-my-posh/_git/website", Upstream: "git@ssh.dev.azure.com:v3/posh/oh-my-posh/website"},
	}
	for _, tc := range cases {
		g := &Git{}
		upstreamURL := g.cleanUpstreamURL(tc.Upstream)
		assert.Equal(t, tc.Expected, upstreamURL, tc.Case)
	}
}

func TestGitUpstream(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Upstream string
	}{
		{Case: "No upstream", Expected: "G", Upstream: ""},
		{Case: "SSH url", Expected: "G", Upstream: "ssh://git@git.my.domain:3001/ADIX7/dotconfig.git"},
		{Case: "Gitea", Expected: "EX", Upstream: "_gitea@src.example.com:user/repo.git"},
		{Case: "GitHub", Expected: "GH", Upstream: "github.com/test"},
		{Case: "GitLab", Expected: "GL", Upstream: "gitlab.com/test"},
		{Case: "Bitbucket", Expected: "BB", Upstream: "bitbucket.org/test"},
		{Case: "Azure DevOps", Expected: "AD", Upstream: "dev.azure.com/test"},
		{Case: "Azure DevOps Dos", Expected: "AD", Upstream: "test.visualstudio.com"},
		{Case: "CodeCommit", Expected: "AC", Upstream: "codecommit::eu-west-1://test-repository"},
		{Case: "Codeberg", Expected: "CB", Upstream: "codeberg.org:user/repo.git"},
		{Case: "Gitstash", Expected: "G", Upstream: "gitstash.com/test"},
		{Case: "My custom server", Expected: "CU", Upstream: "mycustom.server/test"},
		{Case: "GitHub with dash", Expected: "GH", Upstream: "github.com:pixel48/custom-reg"},
	}
	for _, tc := range cases {
		env := &mock.Environment{}
		env.On("IsWsl").Return(false)
		env.On("RunCommand", "git", []string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "remote", "get-url", origin}).Return(tc.Upstream, nil)
		env.On("GOOS").Return("unix")
		props := options.Map{
			GithubIcon:      "GH",
			GitlabIcon:      "GL",
			BitbucketIcon:   "BB",
			AzureDevOpsIcon: "AD",
			CodeCommit:      "AC",
			CodebergIcon:    "CB",
			GitIcon:         "G",
			UpstreamIcons: map[string]string{
				"mycustom.server": "CU",
				"src.example.com": "EX",
			},
		}

		g := &Git{
			Scm: Scm{
				command:  GITCOMMAND,
				Upstream: "origin/main",
			},
		}
		g.Init(props, env)

		g.configOnce = sync.Once{}
		g.configOnce.Do(func() {
			g.configErr = errors.New("no config")
		})

		upstreamIcon := g.getUpstreamIcon()
		assert.Equal(t, tc.Expected, upstreamIcon, tc.Case)
	}
}

func TestGetBranchStatus(t *testing.T) {
	cases := []struct {
		Case         string
		Expected     string
		Upstream     string
		Ahead        int
		Behind       int
		UpstreamGone bool
	}{
		{Case: "Equal with remote", Expected: "equal", Upstream: branchName},
		{Case: "Ahead", Expected: "up2", Ahead: 2},
		{Case: "Behind", Expected: "down8", Behind: 8},
		{Case: "Behind and ahead", Expected: "up7 down8", Behind: 8, Ahead: 7},
		{Case: "Gone", Expected: "gone", Upstream: branchName, UpstreamGone: true},
		{Case: "No remote", Expected: "", Upstream: ""},
		{Case: "Default (bug)", Expected: "", Behind: -8, Upstream: "wonky"},
	}

	for _, tc := range cases {
		props := options.Map{
			BranchAheadIcon:     "up",
			BranchBehindIcon:    "down",
			BranchIdenticalIcon: "equal",
			BranchGoneIcon:      "gone",
		}

		g := &Git{
			Scm: Scm{
				Upstream: tc.Upstream,
			},
			Ahead:        tc.Ahead,
			Behind:       tc.Behind,
			UpstreamGone: tc.UpstreamGone,
		}
		g.Init(props, new(mock.Environment))

		g.setBranchStatus()
		assert.Equal(t, tc.Expected, g.BranchStatus, tc.Case)
	}
}

func TestGitTemplateString(t *testing.T) {
	cases := []struct {
		Git      *Git
		Case     string
		Expected string
		Template string
	}{
		{
			Case:     "Only HEAD name",
			Expected: branchName,
			Template: "{{ .HEAD }}",
			Git: &Git{
				HEAD:   branchName,
				Behind: 2,
			},
		},
		{
			Case:     "Working area changes",
			Expected: "main \uF044 +2 ~3",
			Template: "{{ .HEAD }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Git: &Git{
				HEAD: branchName,
				Working: &GitStatus{
					ScmStatus: ScmStatus{
						Added:    2,
						Modified: 3,
					},
				},
			},
		},
		{
			Case:     "No working area changes",
			Expected: branchName,
			Template: "{{ .HEAD }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Git: &Git{
				HEAD:    branchName,
				Working: &GitStatus{},
			},
		},
		{
			Case:     "Working and staging area changes",
			Expected: "main \uF046 +5 ~1 \uF044 +2 ~3",
			Template: "{{ .HEAD }}{{ if .Staging.Changed }} \uF046 {{ .Staging.String }}{{ end }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Git: &Git{
				HEAD: branchName,
				Working: &GitStatus{
					ScmStatus: ScmStatus{
						Added:    2,
						Modified: 3,
					},
				},
				Staging: &GitStatus{
					ScmStatus: ScmStatus{
						Added:    5,
						Modified: 1,
					},
				},
			},
		},
		{
			Case:     "Working and staging area changes with separator",
			Expected: "main \uF046 +5 ~1 | \uF044 +2 ~3",
			Template: "{{ .HEAD }}{{ if .Staging.Changed }} \uF046 {{ .Staging.String }}{{ end }}{{ if and (.Working.Changed) (.Staging.Changed) }} |{{ end }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}", //nolint:lll
			Git: &Git{
				HEAD: branchName,
				Working: &GitStatus{
					ScmStatus: ScmStatus{
						Added:    2,
						Modified: 3,
					},
				},
				Staging: &GitStatus{
					ScmStatus: ScmStatus{
						Added:    5,
						Modified: 1,
					},
				},
			},
		},
		{
			Case:     "Working and staging area changes with separator and stash count",
			Expected: "main \uF046 +5 ~1 | \uF044 +2 ~3 \ueb4b 3",
			Template: "{{ .HEAD }}{{ if .Staging.Changed }} \uF046 {{ .Staging.String }}{{ end }}{{ if and (.Working.Changed) (.Staging.Changed) }} |{{ end }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}{{ if gt .StashCount 0 }} \ueb4b {{ .StashCount }}{{ end }}", //nolint:lll
			Git: &Git{
				HEAD: branchName,
				Working: &GitStatus{
					ScmStatus: ScmStatus{
						Added:    2,
						Modified: 3,
					},
				},
				Staging: &GitStatus{
					ScmStatus: ScmStatus{
						Added:    5,
						Modified: 1,
					},
				},
				stashCount: 3,
				poshgit:    true,
			},
		},
		{
			Case:     "No local changes",
			Expected: branchName,
			Template: "{{ .HEAD }}{{ if .Staging.Changed }} \uF046{{ .Staging.String }}{{ end }}{{ if .Working.Changed }} \uF044{{ .Working.String }}{{ end }}",
			Git: &Git{
				HEAD:    branchName,
				Staging: &GitStatus{},
				Working: &GitStatus{},
			},
		},
		{
			Case:     "Upstream Icon",
			Expected: "from GitHub on main",
			Template: "from {{ .UpstreamIcon }} on {{ .HEAD }}",
			Git: &Git{
				HEAD:         branchName,
				Staging:      &GitStatus{},
				Working:      &GitStatus{},
				UpstreamIcon: "GitHub",
			},
		},
	}

	for _, tc := range cases {
		props := options.Map{
			FetchStatus: true,
		}
		env := new(mock.Environment)
		tc.Git.env = env
		tc.Git.options = props
		assert.Equal(t, tc.Expected, renderTemplate(env, tc.Template, tc.Git), tc.Case)
	}
}

func TestGitUntrackedMode(t *testing.T) {
	cases := []struct {
		UntrackedModes map[string]string
		Case           string
		Expected       string
	}{
		{
			Case:     "Default mode - no map",
			Expected: "-unormal",
		},
		{
			Case:     "Default mode - no match",
			Expected: "-unormal",
			UntrackedModes: map[string]string{
				"bar": "no",
			},
		},
		{
			Case:     "No mode - match",
			Expected: "-uno",
			UntrackedModes: map[string]string{
				"foo": "no",
				"bar": "normal",
			},
		},
		{
			Case:     "Global mode",
			Expected: "-uno",
			UntrackedModes: map[string]string{
				"*": "no",
			},
		},
	}

	for _, tc := range cases {
		props := options.Map{
			UntrackedModes: tc.UntrackedModes,
		}

		g := &Git{
			Scm: Scm{
				repoRootDir: "foo",
			},
		}
		g.Init(props, new(mock.Environment))

		got := g.getUntrackedFilesMode()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGitIgnoreSubmodules(t *testing.T) {
	cases := []struct {
		IgnoreSubmodules map[string]string
		Case             string
		Expected         string
	}{
		{
			Case:     "Override",
			Expected: "--ignore-submodules=all",
			IgnoreSubmodules: map[string]string{
				"foo": "all",
			},
		},
		{
			Case: "Default mode - empty",
			IgnoreSubmodules: map[string]string{
				"bar": "no",
			},
		},
		{
			Case:     "Global mode",
			Expected: "--ignore-submodules=dirty",
			IgnoreSubmodules: map[string]string{
				"*": "dirty",
			},
		},
	}

	for _, tc := range cases {
		props := options.Map{
			IgnoreSubmodules: tc.IgnoreSubmodules,
		}

		g := &Git{
			Scm: Scm{
				repoRootDir: "foo",
			},
		}
		g.Init(props, new(mock.Environment))

		got := g.getIgnoreSubmodulesMode()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGitCommit(t *testing.T) {
	cases := []struct {
		Case     string
		Expected *Commit
		Output   string
	}{
		{
			Case: "Clean commit",
			Output: `
			an:Jan De Dobbeleer
			ae:jan@ohmyposh.dev
			cn:Jan De Dobbeleer
			ce:jan@ohmyposh.dev
			at:1673176335
			su:docs(error): you can't use cross segment properties
			ha:1234567891011121314
			rf:HEAD -> refs/heads/main, tag: refs/tags/tag-1, tag: refs/tags/0.3.4, refs/remotes/origin/main, refs/remotes/origin/dev, refs/heads/dev, refs/remotes/origin/HEAD
			`,
			Expected: &Commit{
				Author: &User{
					Name:  "Jan De Dobbeleer",
					Email: "jan@ohmyposh.dev",
				},
				Committer: &User{
					Name:  "Jan De Dobbeleer",
					Email: "jan@ohmyposh.dev",
				},
				Subject:   "docs(error): you can't use cross segment properties",
				Timestamp: time.Unix(1673176335, 0),
				Refs: &Refs{
					Tags:    []string{"tag-1", "0.3.4"},
					Heads:   []string{"main", "dev"},
					Remotes: []string{"origin/main", "origin/dev"},
				},
				Sha: "1234567891011121314",
			},
		},
		{
			Case: "No commit output",
			Expected: &Commit{
				Author:    &User{},
				Committer: &User{},
				Refs:      &Refs{},
			},
		},
		{
			Case: "No author",
			Output: `
			an:
			ae:
			cn:Jan De Dobbeleer
			ce:jan@ohmyposh.dev
			at:1673176335
			su:docs(error): you can't use cross segment properties
			`,
			Expected: &Commit{
				Author: &User{},
				Committer: &User{
					Name:  "Jan De Dobbeleer",
					Email: "jan@ohmyposh.dev",
				},
				Subject:   "docs(error): you can't use cross segment properties",
				Timestamp: time.Unix(1673176335, 0),
				Refs:      &Refs{},
			},
		},
		{
			Case: "No refs",
			Output: `
			rf:HEAD
			`,
			Expected: &Commit{
				Author:    &User{},
				Committer: &User{},
				Refs:      &Refs{},
			},
		},
		{
			Case: "Just tag ref",
			Output: `
			rf:HEAD, tag: refs/tags/tag-1
			`,
			Expected: &Commit{
				Author:    &User{},
				Committer: &User{},
				Refs: &Refs{
					Tags: []string{"tag-1"},
				},
			},
		},
		{
			Case: "Feature branch including slash",
			Output: `
			rf:HEAD, tag: refs/tags/feat/feat-1
			`,
			Expected: &Commit{
				Author:    &User{},
				Committer: &User{},
				Refs: &Refs{
					Tags: []string{"feat/feat-1"},
				},
			},
		},
		{
			Case: "Bad timestamp",
			Output: `
			at:err
			`,
			Expected: &Commit{
				Author:    &User{},
				Committer: &User{},
				Refs:      &Refs{},
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.MockGitCommand("", tc.Output, "log", "-1", "--pretty=format:an:%an%nae:%ae%ncn:%cn%nce:%ce%nat:%at%nsu:%s%nha:%H%nrf:%D", "--decorate=full")

		g := &Git{
			Scm: Scm{
				command: GITCOMMAND,
			},
		}
		g.Init(options.Map{}, env)

		got := g.Commit()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGitRemotes(t *testing.T) {
	cases := []struct {
		ExpectedRemotes map[string]string
		Case            string
		Config          string
		ScmDir          string
		Expected        int
	}{
		{
			Case:            "Empty config file",
			ScmDir:          "/repo/.git",
			Expected:        0,
			ExpectedRemotes: map[string]string{},
		},
		{
			Case:     "Two remotes",
			ScmDir:   "/repo/.git",
			Expected: 2,
			Config: `
[remote "origin"]
	url = git@github.com:JanDeDobbeleer/test.git
[remote "upstream"]
	url = git@github.com:microsoft/test.git
`,
			ExpectedRemotes: map[string]string{
				"origin":   "https://github.com/JanDeDobbeleer/test",
				"upstream": "https://github.com/microsoft/test",
			},
		},
		{
			Case:     "One remote",
			ScmDir:   "/repo/.git",
			Expected: 1,
			Config: `
[remote "origin"]
	url = git@github.com:JanDeDobbeleer/test.git
`,
			ExpectedRemotes: map[string]string{
				"origin": "https://github.com/JanDeDobbeleer/test",
			},
		},
		{
			Case:            "Broken config",
			ScmDir:          "/repo/.git",
			Expected:        0,
			Config:          "{{}}",
			ExpectedRemotes: map[string]string{},
		},
		{
			Case:     "Three remotes with different URL formats",
			ScmDir:   "/repo/.git",
			Expected: 3,
			Config: `
[remote "origin"]
	url = git@github.com:JanDeDobbeleer/test.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[remote "upstream"]
	url = https://github.com/microsoft/test.git
	fetch = +refs/heads/*:refs/remotes/upstream/*
[remote "fork"]
	url = git@gitlab.com:user/test.git
	fetch = +refs/heads/*:refs/remotes/fork/*
`,
			ExpectedRemotes: map[string]string{
				"origin":   "https://github.com/JanDeDobbeleer/test",
				"upstream": "https://github.com/microsoft/test.git",
				"fork":     "https://gitlab.com/user/test",
			},
		},
		{
			Case:     "Linked worktree reads the common config",
			ScmDir:   "/repo/.git",
			Expected: 1,
			Config: `
[remote "origin"]
	url = git@github.com:JanDeDobbeleer/test.git
`,
			ExpectedRemotes: map[string]string{
				"origin": "https://github.com/JanDeDobbeleer/test",
			},
		},
		{
			Case:            "Empty common dir returns empty without reading",
			ScmDir:          "",
			Expected:        0,
			ExpectedRemotes: map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			if tc.ScmDir != "" {
				env.On("FileContent", tc.ScmDir+"/config").Return(tc.Config)
			}

			g := &Git{
				Scm: Scm{
					repoRootDir: "foo",
					scmDir:      tc.ScmDir,
				},
			}
			g.Init(options.Map{}, env)

			got := g.Remotes()
			assert.Equal(t, tc.Expected, len(got), tc.Case)

			for name, expectedURL := range tc.ExpectedRemotes {
				actualURL, exists := got[name]
				assert.True(t, exists, "%s: expected remote '%s' to exist", tc.Case, name)
				assert.Equal(t, expectedURL, actualURL, "%s: remote '%s' URL mismatch", tc.Case, name)
			}

			if tc.ScmDir == "" {
				env.AssertNotCalled(t, "FileContent", testify_.Anything)
			}
		})
	}
}

func TestGitRepoName(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   string
		WorkingDir string
		RealDir    string
		IsWorkTree bool
	}{
		{
			Case:       "In worktree",
			Expected:   "oh-my-posh",
			IsWorkTree: true,
			WorkingDir: "/Users/jan/Code/oh-my-posh/.git/worktrees/oh-my-posh2",
		},
		{
			Case:       "Not in worktree",
			Expected:   "oh-my-posh",
			IsWorkTree: false,
			RealDir:    "/Users/jan/Code/oh-my-posh",
		},
		{
			Case:       "In worktree, unexpected dir",
			Expected:   "",
			IsWorkTree: true,
			WorkingDir: "/Users/jan/Code/oh-my-posh2",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return("/")
		env.On("GOOS").Return(runtime.LINUX)

		g := &Git{
			Scm: Scm{
				repoRootDir: tc.RealDir,
				mainSCMDir:  tc.WorkingDir,
			},
			IsWorkTree: tc.IsWorkTree,
		}
		g.Init(options.Map{}, env)

		got := g.repoName()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestParseMainWorktree(t *testing.T) {
	cases := []struct {
		Case     string
		Output   string
		Expected string
		Valid    bool
	}{
		{
			Case: "main worktree",
			Output: "worktree /repo/main\x00HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00" +
				"worktree /repo/linked\x00HEAD abcdef1234567890\x00branch refs/heads/feature\x00\x00",
			Expected: "/repo/main",
			Valid:    true,
		},
		{
			Case:     "path with spaces and newline",
			Output:   "worktree /repo/main path\nwith newline\x00HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00",
			Expected: "/repo/main path\nwith newline",
			Valid:    true,
		},
		{
			Case:   "bare main repository",
			Output: "worktree /repo/main.git\x00bare\x00\x00worktree /repo/linked\x00HEAD 1234567890abcdef\x00\x00",
			Valid:  true,
		},
		{
			Case: "empty output",
		},
		{
			Case:   "missing record terminator",
			Output: "worktree /repo/main\x00HEAD 1234567890abcdef",
		},
		{
			Case:   "missing worktree field",
			Output: "HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00",
		},
		{
			Case:   "empty worktree path",
			Output: "worktree \x00HEAD 1234567890abcdef\x00\x00",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			got, valid := parseMainWorktree(tc.Output)
			assert.Equal(t, tc.Expected, got)
			assert.Equal(t, tc.Valid, valid)
		})
	}
}

func TestGitMainWorktree(t *testing.T) {
	cases := []struct {
		Case          string
		Output        string
		CommandError  error
		Expected      string
		Linked        bool
		ExpectedCalls int
	}{
		{
			Case: "not a linked worktree",
		},
		{
			Case:          "linked worktree",
			Output:        "worktree /repo/main\x00HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00",
			Expected:      "/repo/main",
			Linked:        true,
			ExpectedCalls: 1,
		},
		{
			Case:          "command failure",
			CommandError:  errors.New("git failed"),
			Linked:        true,
			ExpectedCalls: 1,
		},
		{
			Case:          "malformed output",
			Output:        "worktree /repo/main",
			Linked:        true,
			ExpectedCalls: 1,
		},
	}

	for index, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			commonDir := fmt.Sprintf("/repo/%d/.git", index)
			key := fmt.Sprintf("%s@%s", mainWorktreeCacheKey, commonDir)
			cache.Delete(cache.Session, key)
			t.Cleanup(func() {
				cache.Delete(cache.Session, key)
			})

			env := new(mock.Environment)
			if tc.ExpectedCalls > 0 {
				args := []string{
					"-C", "/repo/linked",
					"--no-optional-locks",
					"-c", "core.quotepath=false",
					"-c", "color.status=false",
					"worktree", "list", "--porcelain", "-z",
				}
				env.On("RunCommand", GITCOMMAND, args).Return(tc.Output, tc.CommandError).Once()
			}

			g := &Git{
				Scm: Scm{
					command:     GITCOMMAND,
					repoRootDir: "/repo/linked",
					scmDir:      commonDir,
				},
				IsWorkTree: tc.Linked,
			}
			g.Init(options.Map{}, env)

			got := renderTemplate(env, "{{ .MainWorktree }}|{{ .MainWorktree }}", g)

			assert.Equal(t, tc.Expected+"|"+tc.Expected, got)
			env.AssertNumberOfCalls(t, "RunCommand", tc.ExpectedCalls)
		})
	}
}

func TestGitMainWorktreeSessionCache(t *testing.T) {
	const (
		mainWorktree = TestRootPath + "repo/main"
		firstRoot    = TestRootPath + "repo/linked-one"
		secondRoot   = TestRootPath + "repo/linked-two"
		commonDir    = mainWorktree + "/.git"
	)

	key := fmt.Sprintf("%s@%s", mainWorktreeCacheKey, commonDir)
	cache.Delete(cache.Session, key)
	t.Cleanup(func() {
		cache.Delete(cache.Session, key)
	})

	firstEnv := new(mock.Environment)
	firstGitFile := &runtime.FileInfo{
		Path:         firstRoot + "/.git",
		ParentFolder: firstRoot,
	}
	firstAdminDir := commonDir + "/worktrees/linked-one"
	firstEnv.On("FileContent", firstGitFile.Path).Return("gitdir: " + firstAdminDir)
	firstEnv.On("FileContent", filepath.Join(firstAdminDir, "gitdir")).Return(firstRoot + "/.git")
	firstEnv.MockGitCommand(
		firstRoot+"/",
		"worktree "+mainWorktree+"\x00HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00",
		"worktree", "list", "--porcelain", "-z",
	)
	first := &Git{}
	first.Init(options.Map{}, firstEnv)
	require.True(t, first.hasWorktree(firstGitFile))
	first.command = GITCOMMAND

	secondEnv := new(mock.Environment)
	secondGitFile := &runtime.FileInfo{
		Path:         secondRoot + "/.git",
		ParentFolder: secondRoot,
	}
	secondAdminDir := commonDir + "/worktrees/linked-two"
	secondEnv.On("FileContent", secondGitFile.Path).Return("gitdir: ../main/.git/worktrees/linked-two")
	secondEnv.On("FileContent", filepath.Join(secondAdminDir, "gitdir")).Return("../../../../linked-two/.git")
	second := &Git{}
	second.Init(options.Map{}, secondEnv)
	require.True(t, second.hasWorktree(secondGitFile))
	second.command = GITCOMMAND

	assert.Equal(t, firstAdminDir, first.mainSCMDir)
	assert.Equal(t, secondAdminDir, second.mainSCMDir)
	assert.Equal(t, commonDir, first.commonGitDir())
	assert.Equal(t, commonDir, second.commonGitDir())

	assert.Equal(t, mainWorktree, first.MainWorktree())
	assert.Equal(t, mainWorktree, second.MainWorktree())
	firstEnv.AssertNumberOfCalls(t, "RunCommand", 1)
	secondEnv.AssertNotCalled(t, "RunCommand", testify_.Anything, testify_.Anything)
}

func TestGitMainWorktreeConvertsWSLPath(t *testing.T) {
	const (
		commonDir      = "C:/repo/main/.git"
		windowsPath    = "C:/repo/main"
		linuxPath      = "/mnt/c/repo/main"
		linkedWorktree = "C:/repo/linked"
	)

	key := fmt.Sprintf("%s@%s", mainWorktreeCacheKey, commonDir)
	cache.Delete(cache.Session, key)
	t.Cleanup(func() {
		cache.Delete(cache.Session, key)
	})

	env := new(mock.Environment)
	env.On("RunCommand", "git.exe", []string{
		"-C", linkedWorktree,
		"--no-optional-locks",
		"-c", "core.quotepath=false",
		"-c", "color.status=false",
		"worktree", "list", "--porcelain", "-z",
	}).Return("worktree "+windowsPath+"\x00HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00", nil).Once()
	env.On("ConvertToLinuxPath").Return(linuxPath).Once()

	g := &Git{
		Scm: Scm{
			command:         "git.exe",
			repoRootDir:     linkedWorktree,
			mainSCMDir:      commonDir + "/worktrees/linked",
			IsWslSharedPath: true,
		},
		IsWorkTree: true,
	}
	g.Init(options.Map{}, env)

	assert.Equal(t, linuxPath, g.MainWorktree())
	env.AssertExpectations(t)
}

func TestGitMainWorktreeFromWindowsGitInWSL(t *testing.T) {
	const (
		windowsMain     = "D:/repo/main"
		windowsLinked   = "D:/repo/linked"
		linuxMain       = "/mnt/d/repo/main"
		linuxLinked     = "/mnt/d/repo/linked"
		linuxCommonDir  = linuxMain + "/.git"
		linuxAdminDir   = linuxCommonDir + "/worktrees/linked"
		windowsAdminDir = windowsMain + "/.git/worktrees/linked"
	)

	key := fmt.Sprintf("%s@%s", mainWorktreeCacheKey, linuxCommonDir)
	cache.Delete(cache.Session, key)
	t.Cleanup(func() {
		cache.Delete(cache.Session, key)
	})

	env := new(mock.Environment)
	gitFile := &runtime.FileInfo{
		Path:         linuxLinked + "/.git",
		ParentFolder: linuxLinked,
	}
	env.On("GOOS").Return("")
	env.On("FileContent", gitFile.Path).Return("gitdir: " + windowsAdminDir).Once()
	env.On("ConvertToLinuxPath").Return(linuxAdminDir).Once()
	env.On("FileContent", filepath.Join(linuxAdminDir, "gitdir")).Return(windowsLinked + "/.git").Once()
	env.On("ConvertToLinuxPath").Return(linuxLinked).Once()
	env.On("ConvertToWindowsPath", linuxLinked).Return(windowsLinked).Once()
	env.On("RunCommand", "git.exe", []string{
		"-C", windowsLinked,
		"--no-optional-locks",
		"-c", "core.quotepath=false",
		"-c", "color.status=false",
		"worktree", "list", "--porcelain", "-z",
	}).Return("worktree "+windowsMain+"\x00HEAD 1234567890abcdef\x00branch refs/heads/main\x00\x00", nil).Once()
	env.On("ConvertToLinuxPath").Return(linuxMain).Once()

	g := &Git{
		Scm: Scm{
			command:         "git.exe",
			IsWslSharedPath: true,
		},
	}
	g.Init(options.Map{}, env)

	require.True(t, g.isRepo(gitFile))
	assert.Equal(t, windowsLinked, g.repoRootDir)
	assert.Equal(t, linuxCommonDir, g.commonGitDir())
	assert.Equal(t, linuxMain, g.MainWorktree())
	env.AssertExpectations(t)
}

func TestGitMainWorktreeIsLazy(t *testing.T) {
	env := new(mock.Environment)
	g := &Git{
		Working: &GitStatus{},
		Staging: &GitStatus{},
		Scm: Scm{
			command:     GITCOMMAND,
			repoRootDir: "/repo/linked",
			scmDir:      "/repo/.git",
		},
		IsWorkTree: true,
	}
	g.Init(options.Map{}, env)

	got := renderTemplateNoTrimSpace(env, g.Template(), g)

	assert.NotEmpty(t, got)
	env.AssertNotCalled(t, "RunCommand", testify_.Anything, testify_.Anything)
}

func TestDisableWithJJEnabled(t *testing.T) {
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)
	// Mock .jj directory exists
	env.On("HasParentFilePath", ".jj", false).Return(&runtime.FileInfo{Path: "/dir/.jj", IsDir: true}, nil)

	g := &Git{}
	props := options.Map{
		DisableWithJJ: true,
	}
	g.Init(props, env)

	assert.False(t, g.Enabled())
}

func TestDisableWithJJDisabled(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/dir/.git",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "git").Return(true)
	env.On("GOOS").Return("")
	env.On("FileContent", "/dir/.git/HEAD").Return("")
	env.MockGitCommand("/dir", "1234567890abcdef1234567890abcdef12345678", "rev-parse", "HEAD")
	env.MockGitCommand("/dir", "", "describe", "--tags", "--exact-match") // Use repo root, not .git dir
	env.On("IsWsl").Return(false)
	// Mock .jj directory exists
	env.On("HasParentFilePath", ".jj", false).Return(&runtime.FileInfo{Path: "/dir/.jj", IsDir: true}, nil)
	env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(poshHome)
	env.On("Getenv", poshGitEnv).Return("")
	env.On("DirMatchesOneOf", testify_.Anything, testify_.Anything).Return(false)

	g := &Git{}
	props := options.Map{
		DisableWithJJ: false, // Property is disabled
	}
	g.Init(props, env)

	assert.True(t, g.Enabled()) // Should still be enabled since disable_with_jj is false
}

func TestDisableWithJJNoJJDirectory(t *testing.T) {
	fileInfo := &runtime.FileInfo{
		Path:         "/dir/.git",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "git").Return(true)
	env.On("GOOS").Return("")
	env.On("FileContent", "/dir/.git/HEAD").Return("")
	env.MockGitCommand("/dir", "1234567890abcdef1234567890abcdef12345678", "rev-parse", "HEAD")
	env.MockGitCommand("/dir", "", "describe", "--tags", "--exact-match") // Use repo root, not .git dir
	env.On("IsWsl").Return(false)
	// Mock .jj directory does not exist
	env.On("HasParentFilePath", ".jj", false).Return((*runtime.FileInfo)(nil), errors.New("no .jj found"))
	env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(poshHome)
	env.On("Getenv", poshGitEnv).Return("")
	env.On("DirMatchesOneOf", testify_.Anything, testify_.Anything).Return(false)

	g := &Git{}
	props := options.Map{
		DisableWithJJ: true, // Property is enabled but no .jj directory
	}
	g.Init(props, env)

	assert.True(t, g.Enabled()) // Should be enabled since .jj directory doesn't exist
}

func TestPushStatusAheadAndBehind(t *testing.T) {
	cases := []struct {
		Case               string
		PushAheadCount     string
		PushBehindCount    string
		Config             string
		ExpectedPushAhead  int
		ExpectedPushBehind int
	}{
		{
			Case:               "ahead and behind",
			PushAheadCount:     "3",
			PushBehindCount:    "5",
			ExpectedPushAhead:  3,
			ExpectedPushBehind: 5,
		},
		{
			Case:               "only ahead",
			PushAheadCount:     "2",
			PushBehindCount:    "0",
			ExpectedPushAhead:  2,
			ExpectedPushBehind: 0,
		},
		{
			Case:               "only behind",
			PushAheadCount:     "0",
			PushBehindCount:    "7",
			ExpectedPushAhead:  0,
			ExpectedPushBehind: 7,
		},
		{
			Case:               "up to date",
			PushAheadCount:     "0",
			PushBehindCount:    "0",
			ExpectedPushAhead:  0,
			ExpectedPushBehind: 0,
		},
		{
			Case:               "remote from config",
			PushAheadCount:     "2",
			PushBehindCount:    "0",
			ExpectedPushAhead:  2,
			ExpectedPushBehind: 0,
			Config: `
			[branch "main"]
				remote = origin
				merge = refs/heads/main
			`,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("RunCommand", "git", []string{"-C", "/dir", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "rev-parse", "--abbrev-ref", "@{push}"}).Return("", nil)
		env.On("RunCommand", "git", []string{"-C", "/dir", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "config", "--get", "branch.main.pushRemote"}).Return("", nil)
		env.On("RunCommand", "git", []string{"-C", "/dir", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "config", "--get", "remote.pushDefault"}).Return("", nil)
		env.On("RunCommand", "git", []string{"-C", "/dir", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "rev-list", "--count", "origin/main..HEAD"}).Return(tc.PushAheadCount, nil)
		env.On("RunCommand", "git", []string{"-C", "/dir", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "rev-list", "--count", "HEAD..origin/main"}).Return(tc.PushBehindCount, nil)
		env.On("FileContent", "/dir/.git/config").Return("")

		g := &Git{
			Scm: Scm{
				command:     "git",
				repoRootDir: "/dir",
				scmDir:      "/dir/.git",
				Upstream:    "origin/main",
			},
			Ref: "main",
		}

		props := options.Map{
			FetchPushStatus: true,
		}

		g.Init(props, env)

		g.configOnce = sync.Once{}
		g.configOnce.Do(func() {
			if len(tc.Config) > 0 {
				g.config, g.configErr = ini.Load(tc.Config)
				return
			}

			g.configErr = errors.New("no config")
		})

		g.setPushStatus()

		assert.Equal(t, tc.ExpectedPushAhead, g.PushAhead, tc.Case)
		assert.Equal(t, tc.ExpectedPushBehind, g.PushBehind, tc.Case)
	}
}

func TestPushRef(t *testing.T) {
	args := func(extra ...string) []string {
		return append([]string{
			"-C", "/repo", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false",
		}, extra...)
	}

	t.Run("uses @{push} when it resolves", func(t *testing.T) {
		env := new(mock.Environment)
		env.On("IsWsl").Return(false)
		env.On("GOOS").Return("unix")
		env.On("RunCommand", GITCOMMAND, args("rev-parse", "--abbrev-ref", "@{push}")).Return("origin/main", nil)

		g := &Git{
			Scm: Scm{command: GITCOMMAND, repoRootDir: "/repo"},
			Ref: "main",
		}
		g.Init(options.Map{}, env)

		assert.Equal(t, "origin/main", g.pushRef())
		env.AssertNotCalled(t, "RunCommand", GITCOMMAND, args("config", "--get", "remote.pushDefault"))
	})

	t.Run("falls back to per-branch pushRemote when @{push} is unresolvable", func(t *testing.T) {
		env := new(mock.Environment)
		env.On("IsWsl").Return(false)
		env.On("GOOS").Return("unix")
		env.On("RunCommand", GITCOMMAND, args("rev-parse", "--abbrev-ref", "@{push}")).Return("", errors.New("cannot resolve 'simple' push to a single destination"))
		env.On("RunCommand", GITCOMMAND, args("config", "--get", "branch.main.pushRemote")).Return("fork", nil)

		g := &Git{
			Scm: Scm{command: GITCOMMAND, repoRootDir: "/repo"},
			Ref: "main",
		}
		g.Init(options.Map{}, env)

		assert.Equal(t, "fork/main", g.pushRef())
	})
}

// TestSetStatusNative builds a real temp repo with the git CLI (skipped when
// git isn't on PATH), then asserts setStatusNative populates the same
// fields as the existing exec+porcelain path for that exact repo. The
// porcelain text fed to the exec path is captured from a real `git status`
// call, so both sides are exercised against genuine, non-trivial repo state.
func TestSetStatusNative(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	dir := t.TempDir()
	runRealGit(t, dir, "init", "-q", "-b", "main", ".")
	runRealGit(t, dir, "config", "user.email", "test@example.com")
	runRealGit(t, dir, "config", "user.name", "Test")
	runRealGit(t, dir, "config", "core.autocrlf", "false")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644))
	runRealGit(t, dir, "add", ".")
	runRealGit(t, dir, "commit", "-q", "-m", "init")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("changed\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("u\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "staged-add.txt"), []byte("new\n"), 0o644))
	runRealGit(t, dir, "add", "staged-add.txt")

	worktreeGitDir := realGitPath(t, dir, "--git-dir")
	commonGitDir := realGitPath(t, dir, "--git-common-dir")
	repoRoot := realGitPath(t, dir, "--show-toplevel")

	// Capture the exact porcelain text a real `git status` produces for this
	// repo, then feed it through the existing exec parsing path via the
	// mock environment, following the mocking pattern the other setStatus
	// tests already use.
	porcelain := runRealGit(t, dir, "status", "-unormal", "--branch", "--porcelain=2")

	env := new(mock.Environment)
	env.MockGitCommand(repoRoot, porcelain, "status", "-unormal", "--branch", "--porcelain=2")

	gExec := &Git{Scm: Scm{command: GITCOMMAND, repoRootDir: repoRoot}}
	gExec.Init(options.Map{}, env)
	gExec.setStatus()

	// A bare mock.Environment (no expectations configured) means any
	// accidental exec fallback would either panic on an unmet expectation
	// or, since Scm.command is unset here, silently return empty output —
	// either way the sanity check below would catch it.
	gNative := &Git{Scm: Scm{
		mainSCMDir:  worktreeGitDir,
		scmDir:      commonGitDir,
		repoRootDir: repoRoot,
	}}
	gNative.Init(options.Map{NativeStatus: true}, new(mock.Environment))
	gNative.setStatus()

	// sanity: make sure this scenario actually exercises non-trivial status
	// before comparing, so a broken fixture (or a silent fallback) can't
	// pass by both sides being all-zero.
	require.True(t, gNative.Working.Modified > 0 && gNative.Working.Untracked > 0 && gNative.Staging.Added > 0)

	assert.Equal(t, gExec.Working, gNative.Working)
	assert.Equal(t, gExec.Staging, gNative.Staging)
	assert.Equal(t, gExec.Hash, gNative.Hash)
	assert.Equal(t, gExec.ShortHash, gNative.ShortHash)
	assert.Equal(t, gExec.Ref, gNative.Ref)
	assert.Equal(t, gExec.Upstream, gNative.Upstream)
	assert.Equal(t, gExec.Ahead, gNative.Ahead)
	assert.Equal(t, gExec.Behind, gNative.Behind)
	assert.Equal(t, gExec.UpstreamGone, gNative.UpstreamGone)
}

func runRealGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %s failed: %s", strings.Join(args, " "), out)
	return string(out)
}

func realGitPath(t *testing.T, dir, arg string) string {
	t.Helper()
	out := runRealGit(t, dir, "rev-parse", "--path-format=absolute", arg)
	return filepath.FromSlash(strings.TrimSpace(out))
}

func TestGetRemoteURL(t *testing.T) {
	const configWithOrigin = "[remote \"origin\"]\n\turl = gh:example/repo.git"

	cases := []struct {
		Case      string
		Upstream  string
		CLIOutput string
		CLIError  error
		Config    string
		ScmDir    string
		Expected  string
	}{
		{
			Case:      "CLI wins over the raw config value (insteadOf is applied)",
			Upstream:  "origin/main",
			CLIOutput: "https://github.com/example/repo.git",
			Config:    configWithOrigin,
			ScmDir:    "/repo/.git",
			Expected:  "https://github.com/example/repo.git",
		},
		{
			Case:     "falls back to the common config when the CLI fails",
			Upstream: "origin/main",
			CLIError: errors.New("git unavailable"),
			Config:   configWithOrigin,
			ScmDir:   "/repo/.git",
			Expected: "gh:example/repo.git",
		},
		{
			Case:     "empty when the CLI fails and there is no common config",
			Upstream: "origin/main",
			CLIError: errors.New("git unavailable"),
			Config:   configWithOrigin,
			ScmDir:   "",
			Expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("IsWsl").Return(false)
			env.On("GOOS").Return("unix")
			env.On("RunCommand", GITCOMMAND, []string{
				"-C", "/repo", "--no-optional-locks", "-c", "core.quotepath=false",
				"-c", "color.status=false", "remote", "get-url", "origin",
			}).Return(tc.CLIOutput, tc.CLIError)

			if tc.ScmDir != "" {
				env.On("FileContent", tc.ScmDir+"/config").Return(tc.Config)
			}

			g := &Git{
				Scm: Scm{
					command:     GITCOMMAND,
					repoRootDir: "/repo",
					scmDir:      tc.ScmDir,
					Upstream:    tc.Upstream,
				},
			}
			g.Init(options.Map{}, env)

			assert.Equal(t, tc.Expected, g.getRemoteURL())
		})
	}
}

func TestWorktreeCount(t *testing.T) {
	cases := []struct {
		Case       string
		ScmDir     string
		MainSCMDir string
		Entries    []fs.DirEntry
		Expected   int
		HasFolder  bool
	}{
		{
			Case:       "plain clone with no registry",
			ScmDir:     "/repo/.git",
			MainSCMDir: "/repo/.git",
			HasFolder:  false,
			Expected:   0,
		},
		{
			Case:       "linked worktree counts the common registry",
			ScmDir:     "/repo/.git",
			MainSCMDir: "/repo/.git/worktrees/linked",
			HasFolder:  true,
			Entries: []fs.DirEntry{
				&MockDirEntry{name: "linked", isDir: true},
				&MockDirEntry{name: "other", isDir: true},
			},
			Expected: 2,
		},
		{
			Case:       "files in the registry are not counted",
			ScmDir:     "/repo/.git",
			MainSCMDir: "/repo/.git",
			HasFolder:  true,
			Entries: []fs.DirEntry{
				&MockDirEntry{name: "linked", isDir: true},
				&MockDirEntry{name: "README", isDir: false},
			},
			Expected: 1,
		},
		{
			Case:       "separate git dir counts its external registry",
			ScmDir:     "/x/sepdir",
			MainSCMDir: "/x/worktrees/trap/",
			HasFolder:  true,
			Entries:    []fs.DirEntry{&MockDirEntry{name: "wt", isDir: true}},
			Expected:   1,
		},
		{
			Case:     "unknown common dir touches nothing",
			Expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)

			if tc.ScmDir != "" {
				worktrees := filepath.Join(tc.ScmDir, "worktrees")
				env.On("HasFolder", worktrees).Return(tc.HasFolder)
				env.On("LsDir", worktrees).Return(tc.Entries)
			}

			g := &Git{
				Scm: Scm{
					scmDir:     tc.ScmDir,
					mainSCMDir: tc.MainSCMDir,
				},
			}
			g.Init(options.Map{}, env)

			assert.Equal(t, tc.Expected, g.WorktreeCount())

			if tc.ScmDir == "" {
				env.AssertNotCalled(t, "HasFolder", testify_.Anything)
				env.AssertNotCalled(t, "LsDir", testify_.Anything)
			}
		})
	}
}
