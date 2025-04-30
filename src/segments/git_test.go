package segments

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	branchName      = "main"
	dotGit          = "dev/.git"
	dotGitSubmodule = "dev/.git/modules/submodule"
)

func TestEnabledGitNotFound(t *testing.T) {
	env := new(mock.Environment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "git").Return(false)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)

	g := &Git{}
	g.Init(properties.Map{}, env)

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
	env.MockGitCommand(fileInfo.Path, "", "describe", "--tags", "--exact-match")
	env.On("IsWsl").Return(false)
	env.On("HasParentFilePath", ".git", true).Return(fileInfo, nil)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(poshHome)
	env.On("Getenv", poshGitEnv).Return("")
	env.On("DirMatchesOneOf", testify_.Anything, testify_.Anything).Return(false)

	g := &Git{}
	g.Init(properties.Map{}, env)

	assert.True(t, g.Enabled())
	assert.Equal(t, fileInfo.Path, g.mainSCMDir)
}

func TestResolveEmptyGitPath(t *testing.T) {
	base := "base"
	assert.Equal(t, base, resolveGitPath(base, ""))
}

func TestEnabledInWorktree(t *testing.T) {
	cases := []struct {
		Case                  string
		WorkingFolder         string
		WorkingFolderAddon    string
		WorkingFolderContent  string
		ExpectedRealFolder    string
		ExpectedWorkingFolder string
		ExpectedRootFolder    string
		ExpectedEnabled       bool
	}{
		{
			Case:                  "worktree",
			ExpectedEnabled:       true,
			WorkingFolder:         TestRootPath + "dev/.git/worktrees/folder_worktree",
			WorkingFolderAddon:    "gitdir",
			WorkingFolderContent:  TestRootPath + "dev/worktree.git\n",
			ExpectedWorkingFolder: TestRootPath + "dev/.git/worktrees/folder_worktree",
			ExpectedRealFolder:    TestRootPath + "dev/worktree",
			ExpectedRootFolder:    TestRootPath + dotGit,
		},
		{
			Case:                  "submodule",
			ExpectedEnabled:       true,
			WorkingFolder:         "./.git/modules/submodule",
			ExpectedWorkingFolder: TestRootPath + dotGitSubmodule,
			ExpectedRealFolder:    TestRootPath + dotGitSubmodule,
			ExpectedRootFolder:    TestRootPath + dotGitSubmodule,
		},
		{
			Case:                  "submodule with root working folder",
			ExpectedEnabled:       true,
			WorkingFolder:         TestRootPath + dotGitSubmodule,
			ExpectedWorkingFolder: TestRootPath + dotGitSubmodule,
			ExpectedRealFolder:    TestRootPath + dotGitSubmodule,
			ExpectedRootFolder:    TestRootPath + dotGitSubmodule,
		},
		{
			Case:                  "submodule with worktrees",
			ExpectedEnabled:       true,
			WorkingFolder:         TestRootPath + "dev/.git/modules/module/path/worktrees/location",
			WorkingFolderAddon:    "gitdir",
			WorkingFolderContent:  TestRootPath + "dev/worktree.git\n",
			ExpectedWorkingFolder: TestRootPath + "dev/.git/modules/module/path",
			ExpectedRealFolder:    TestRootPath + "dev/worktree",
			ExpectedRootFolder:    TestRootPath + "dev/.git/modules/module/path",
		},
		{
			Case:                  "separate git dir",
			ExpectedEnabled:       true,
			WorkingFolder:         TestRootPath + "dev/separate/.git/posh",
			ExpectedWorkingFolder: TestRootPath + "dev/",
			ExpectedRealFolder:    TestRootPath + "dev/",
			ExpectedRootFolder:    TestRootPath + "dev/separate/.git/posh",
		},
	}
	fileInfo := &runtime.FileInfo{
		Path:         TestRootPath + dotGit,
		ParentFolder: TestRootPath + "dev",
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("FileContent", TestRootPath+dotGit).Return(fmt.Sprintf("gitdir: %s", tc.WorkingFolder))
		env.On("FileContent", filepath.Join(tc.WorkingFolder, tc.WorkingFolderAddon)).Return(tc.WorkingFolderContent)
		env.On("HasFilesInDir", tc.WorkingFolder, tc.WorkingFolderAddon).Return(true)
		env.On("HasFilesInDir", tc.WorkingFolder, "HEAD").Return(true)
		env.On("PathSeparator").Return(string(os.PathSeparator))

		g := &Git{}
		g.Init(properties.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, g.hasWorktree(fileInfo), tc.Case)
		assert.Equal(t, tc.ExpectedWorkingFolder, g.mainSCMDir, tc.Case)
		assert.Equal(t, tc.ExpectedRealFolder, g.repoRootDir, tc.Case)
		assert.Equal(t, tc.ExpectedRootFolder, g.scmDir, tc.Case)
	}
}

func TestEnabledInBareRepo(t *testing.T) {
	cases := []struct {
		Case            string
		HEAD            string
		IsBare          string
		Remote          string
		RemoteURL       string
		ExpectedHEAD    string
		ExpectedRemote  string
		FetchRemote     bool
		ExpectedEnabled bool
	}{
		{
			Case:            "Bare repo on main",
			IsBare:          trueStr,
			HEAD:            "ref: refs/heads/main",
			ExpectedEnabled: true,
			ExpectedHEAD:    "main",
		},
		{
			Case:   "Not a bare repo",
			IsBare: "false",
		},
		{
			Case:            "Bare repo on main remote enabled",
			IsBare:          trueStr,
			HEAD:            "ref: refs/heads/main",
			ExpectedEnabled: true,
			ExpectedHEAD:    "main",
			FetchRemote:     true,
			Remote:          "origin",
			RemoteURL:       "git@github.com:JanDeDobbeleer/oh-my-posh.git",
			ExpectedRemote:  "\uf408",
		},
	}
	for _, tc := range cases {
		pwd := "/home/user/bare.git"
		env := new(mock.Environment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("GOOS").Return("")
		env.On("HasCommand", "git").Return(true)
		env.On("HasParentFilePath", ".git", true).Return(&runtime.FileInfo{}, errors.New("nope"))
		env.MockGitCommand(pwd, tc.IsBare, "rev-parse", "--is-bare-repository")
		env.On("Pwd").Return(pwd)
		env.On("FileContent", "/home/user/bare.git/HEAD").Return(tc.HEAD)
		env.MockGitCommand(pwd, tc.Remote, "remote")
		env.MockGitCommand(pwd, tc.RemoteURL, "remote", "get-url", tc.Remote)

		props := properties.Map{
			FetchBareInfo:     true,
			FetchUpstreamIcon: tc.FetchRemote,
		}

		g := &Git{}
		g.Init(props, env)

		assert.Equal(t, tc.ExpectedEnabled, g.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedHEAD, g.Ref, tc.Case)
		assert.Equal(t, tc.ExpectedRemote, g.UpstreamIcon, tc.Case)
	}
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
		scm: scm{
			command: GITCOMMAND,
		},
	}
	g.Init(properties.Map{}, env)

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

		props := properties.Map{
			BranchIcon:     "branch ",
			CommitIcon:     "commit ",
			RebaseIcon:     "rebase ",
			MergeIcon:      "merge ",
			CherryPickIcon: "pick ",
			TagIcon:        "tag ",
			RevertIcon:     "revert ",
		}

		g := &Git{
			scm: scm{
				command: GITCOMMAND,
			},
			ShortHash: "1234567",
			Ref:       tc.Ref,
		}
		g.Init(props, env)

		g.setGitHEADContext()
		assert.Equal(t, tc.Expected, g.HEAD, tc.Case)
	}
}

func TestSetPrettyHEADName(t *testing.T) {
	cases := []struct {
		Case      string
		Expected  string
		ShortHash string
		Tag       string
		HEAD      string
	}{
		{Case: "main", Expected: "branch main", HEAD: BRANCHPREFIX + "main"},
		{Case: "no hash", Expected: "commit 1234567", HEAD: "12345678910"},
		{Case: "hash on tag", ShortHash: "132312322321", Expected: "tag tag-1", HEAD: "12345678910", Tag: "tag-1"},
		{Case: "no hash on tag", Expected: "tag tag-1", Tag: "tag-1"},
		{Case: "hash on commit", ShortHash: "1234567", Expected: "commit 1234567"},
		{Case: "no hash on commit", Expected: "commit 1234567", HEAD: "12345678910"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("FileContent", "/HEAD").Return(tc.HEAD)
		env.On("GOOS").Return("unix")
		env.On("IsWsl").Return(false)
		env.MockGitCommand("", tc.Tag, "describe", "--tags", "--exact-match")

		props := properties.Map{
			BranchIcon: "branch ",
			CommitIcon: "commit ",
			TagIcon:    "tag ",
		}

		g := &Git{
			scm: scm{
				command: GITCOMMAND,
			},
			ShortHash: tc.ShortHash,
		}
		g.Init(props, env)

		g.setPrettyHEADName()
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
			scm: scm{
				command: GITCOMMAND,
			},
		}
		g.Init(properties.Map{}, env)

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
		g.setGitStatus()
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
			scm: scm{
				mainSCMDir: "",
			},
		}
		g.Init(properties.Map{}, env)

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
		{Case: "No upstream", Expected: "", Upstream: ""},
		{Case: "SSH url", Expected: "G", Upstream: "ssh://git@git.my.domain:3001/ADIX7/dotconfig.git"},
		{Case: "Gitea", Expected: "EX", Upstream: "_gitea@src.example.com:user/repo.git"},
		{Case: "GitHub", Expected: "GH", Upstream: "github.com/test"},
		{Case: "Gitlab", Expected: "GL", Upstream: "gitlab.com/test"},
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
			"-c", "color.status=false", "remote", "get-url", "origin"}).Return(tc.Upstream, nil)
		env.On("GOOS").Return("unix")
		props := properties.Map{
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
			scm: scm{
				command: GITCOMMAND,
			},
			Upstream: "origin/main",
		}
		g.Init(props, env)

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
		props := properties.Map{
			BranchAheadIcon:     "up",
			BranchBehindIcon:    "down",
			BranchIdenticalIcon: "equal",
			BranchGoneIcon:      "gone",
		}

		g := &Git{
			Ahead:        tc.Ahead,
			Behind:       tc.Behind,
			Upstream:     tc.Upstream,
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
		props := properties.Map{
			FetchStatus: true,
		}
		env := new(mock.Environment)
		tc.Git.env = env
		tc.Git.props = props
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
		props := properties.Map{
			UntrackedModes: tc.UntrackedModes,
		}

		g := &Git{
			scm: scm{
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
			Case:     "Overide",
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
		props := properties.Map{
			IgnoreSubmodules: tc.IgnoreSubmodules,
		}

		g := &Git{
			scm: scm{
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
			scm: scm{
				command: GITCOMMAND,
			},
		}
		g.Init(properties.Map{}, env)

		got := g.Commit()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGitRemotes(t *testing.T) {
	cases := []struct {
		Case     string
		Config   string
		Expected int
	}{
		{
			Case:     "Empty config file",
			Expected: 0,
		},
		{
			Case:     "Two remotes",
			Expected: 2,
			Config: `
[remote "origin"]
	url = git@github.com:JanDeDobbeleer/test.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[remote "upstream"]
	url = git@github.com:microsoft/test.git
	fetch = +refs/heads/*:refs/remotes/upstream/*
`,
		},
		{
			Case:     "One remote",
			Expected: 1,
			Config: `
[remote "origin"]
	url = git@github.com:JanDeDobbeleer/test.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`,
		},
		{
			Case:     "Broken config",
			Expected: 0,
			Config:   "{{}}",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("FileContent", "config").Return(tc.Config)

		g := &Git{
			scm: scm{
				repoRootDir: "foo",
			},
		}
		g.Init(properties.Map{}, env)

		got := g.Remotes()
		assert.Equal(t, tc.Expected, len(got), tc.Case)
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
			scm: scm{
				repoRootDir: tc.RealDir,
				mainSCMDir:  tc.WorkingDir,
			},
			IsWorkTree: tc.IsWorkTree,
		}
		g.Init(properties.Map{}, env)

		got := g.repoName()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
