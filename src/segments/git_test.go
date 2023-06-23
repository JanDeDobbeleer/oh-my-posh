package segments

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

const (
	branchName = "main"
)

func TestEnabledGitNotFound(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "git").Return(false)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)
	g := &Git{
		scm: scm{
			env:   env,
			props: properties.Map{},
		},
	}
	assert.False(t, g.Enabled())
}

func TestEnabledInWorkingDirectory(t *testing.T) {
	fileInfo := &platform.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env := new(mock.MockedEnvironment)
	env.On("InWSLSharedDrive").Return(false)
	env.On("HasCommand", "git").Return(true)
	env.On("GOOS").Return("")
	env.On("FileContent", "/dir/hello/HEAD").Return("")
	env.MockGitCommand(fileInfo.Path, "", "describe", "--tags", "--exact-match")
	env.On("IsWsl").Return(false)
	env.On("HasParentFilePath", ".git").Return(fileInfo, nil)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(poshHome)
	env.On("Getenv", poshGitEnv).Return("")
	env.On("DirMatchesOneOf", mock2.Anything, mock2.Anything).Return(false)
	g := &Git{
		scm: scm{
			env:   env,
			props: properties.Map{},
		},
	}
	assert.True(t, g.Enabled())
	assert.Equal(t, fileInfo.Path, g.workingDir)
}

func TestResolveEmptyGitPath(t *testing.T) {
	base := "base"
	assert.Equal(t, base, resolveGitPath(base, ""))
}

func TestEnabledInWorktree(t *testing.T) {
	cases := []struct {
		Case                  string
		ExpectedEnabled       bool
		WorkingFolder         string
		WorkingFolderAddon    string
		WorkingFolderContent  string
		ExpectedRealFolder    string
		ExpectedWorkingFolder string
		ExpectedRootFolder    string
	}{
		{
			Case:                  "worktree",
			ExpectedEnabled:       true,
			WorkingFolder:         TestRootPath + "dev/.git/worktrees/folder_worktree",
			WorkingFolderAddon:    "gitdir",
			WorkingFolderContent:  TestRootPath + "dev/worktree.git\n",
			ExpectedWorkingFolder: TestRootPath + "dev/.git/worktrees/folder_worktree",
			ExpectedRealFolder:    TestRootPath + "dev/worktree",
			ExpectedRootFolder:    TestRootPath + "dev/.git",
		},
		{
			Case:                  "submodule",
			ExpectedEnabled:       true,
			WorkingFolder:         "./.git/modules/submodule",
			ExpectedWorkingFolder: TestRootPath + "dev/.git/modules/submodule",
			ExpectedRealFolder:    TestRootPath + "dev/.git/modules/submodule",
			ExpectedRootFolder:    TestRootPath + "dev/.git/modules/submodule",
		},
		{
			Case:                  "submodule with root working folder",
			ExpectedEnabled:       true,
			WorkingFolder:         TestRootPath + "repo/.git/modules/submodule",
			ExpectedWorkingFolder: TestRootPath + "repo/.git/modules/submodule",
			ExpectedRealFolder:    TestRootPath + "repo/.git/modules/submodule",
			ExpectedRootFolder:    TestRootPath + "repo/.git/modules/submodule",
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
	fileInfo := &platform.FileInfo{
		Path:         TestRootPath + "dev/.git",
		ParentFolder: TestRootPath + "dev",
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("FileContent", TestRootPath+"dev/.git").Return(fmt.Sprintf("gitdir: %s", tc.WorkingFolder))
		env.On("FileContent", filepath.Join(tc.WorkingFolder, tc.WorkingFolderAddon)).Return(tc.WorkingFolderContent)
		env.On("HasFilesInDir", tc.WorkingFolder, tc.WorkingFolderAddon).Return(true)
		env.On("HasFilesInDir", tc.WorkingFolder, "HEAD").Return(true)
		env.On("PathSeparator").Return(string(os.PathSeparator))
		g := &Git{
			scm: scm{
				env:   env,
				props: properties.Map{},
			},
		}
		assert.Equal(t, tc.ExpectedEnabled, g.hasWorktree(fileInfo), tc.Case)
		assert.Equal(t, tc.ExpectedWorkingFolder, g.workingDir, tc.Case)
		assert.Equal(t, tc.ExpectedRealFolder, g.realDir, tc.Case)
		assert.Equal(t, tc.ExpectedRootFolder, g.rootDir, tc.Case)
	}
}

func TestEnabledInBareRepo(t *testing.T) {
	cases := []struct {
		Case            string
		HEAD            string
		IsBare          string
		FetchRemote     bool
		Remote          string
		RemoteURL       string
		ExpectedEnabled bool
		ExpectedHEAD    string
		ExpectedRemote  string
	}{
		{
			Case:            "Bare repo on main",
			IsBare:          "true",
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
			IsBare:          "true",
			HEAD:            "ref: refs/heads/main",
			ExpectedEnabled: true,
			ExpectedHEAD:    "main",
			FetchRemote:     true,
			Remote:          "origin",
			RemoteURL:       "git@github.com:JanDeDobbeleer/oh-my-posh.git",
			ExpectedRemote:  "\uf408 ",
		},
	}
	for _, tc := range cases {
		pwd := "/home/user/bare.git"
		env := new(mock.MockedEnvironment)
		env.On("InWSLSharedDrive").Return(false)
		env.On("GOOS").Return("")
		env.On("HasCommand", "git").Return(true)
		env.On("HasParentFilePath", ".git").Return(&platform.FileInfo{}, errors.New("nope"))
		env.MockGitCommand(pwd, tc.IsBare, "rev-parse", "--is-bare-repository")
		env.On("Pwd").Return(pwd)
		env.On("FileContent", "/home/user/bare.git/HEAD").Return(tc.HEAD)
		env.MockGitCommand(pwd, tc.Remote, "remote")
		env.MockGitCommand(pwd, tc.RemoteURL, "remote", "get-url", tc.Remote)
		g := &Git{
			scm: scm{
				env: env,
				props: properties.Map{
					FetchBareInfo:     true,
					FetchUpstreamIcon: tc.FetchRemote,
				},
			},
		}
		assert.Equal(t, g.Enabled(), tc.ExpectedEnabled, tc.Case)
		assert.Equal(t, g.Ref, tc.ExpectedHEAD, tc.Case)
		assert.Equal(t, g.UpstreamIcon, tc.ExpectedRemote, tc.Case)
	}
}

func TestGetGitOutputForCommand(t *testing.T) {
	args := []string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}
	commandArgs := []string{"symbolic-ref", "--short", "HEAD"}
	want := "je suis le output"
	env := new(mock.MockedEnvironment)
	env.On("IsWsl").Return(false)
	env.On("RunCommand", "git", append(args, commandArgs...)).Return(want, nil)
	env.On("GOOS").Return("unix")
	g := &Git{
		scm: scm{
			env:     env,
			props:   properties.Map{},
			command: GITCOMMAND,
		},
	}
	got := g.getGitCommandOutput(commandArgs...)
	assert.Equal(t, want, got)
}

func TestSetGitHEADContextClean(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Ref         string
		RebaseMerge bool
		RebaseApply bool
		Merge       bool
		CherryPick  bool
		Revert      bool
		Sequencer   bool
		Ours        string
		Theirs      string
		Step        string
		Total       string
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
		env := new(mock.MockedEnvironment)
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

		g := &Git{
			scm: scm{
				env: env,
				props: properties.Map{
					BranchIcon:     "branch ",
					CommitIcon:     "commit ",
					RebaseIcon:     "rebase ",
					MergeIcon:      "merge ",
					CherryPickIcon: "pick ",
					TagIcon:        "tag ",
					RevertIcon:     "revert ",
				},
				command: GITCOMMAND,
			},
			ShortHash: "1234567",
			Ref:       tc.Ref,
		}
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
		env := new(mock.MockedEnvironment)
		env.On("FileContent", "/HEAD").Return(tc.HEAD)
		env.On("GOOS").Return("unix")
		env.On("IsWsl").Return(false)
		env.MockGitCommand("", tc.Tag, "describe", "--tags", "--exact-match")
		g := &Git{
			scm: scm{
				env: env,
				props: properties.Map{
					BranchIcon: "branch ",
					CommitIcon: "commit ",
					TagIcon:    "tag ",
				},
				command: GITCOMMAND,
			},
			ShortHash: tc.ShortHash,
		}
		g.setPrettyHEADName()
		assert.Equal(t, tc.Expected, g.HEAD, tc.Case)
	}
}

func TestSetGitStatus(t *testing.T) {
	cases := []struct {
		Case                 string
		Output               string
		ExpectedWorking      *GitStatus
		ExpectedStaging      *GitStatus
		ExpectedHash         string
		ExpectedRef          string
		ExpectedUpstream     string
		ExpectedUpstreamGone bool
		ExpectedAhead        int
		ExpectedBehind       int
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
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("GOOS").Return("unix")
		env.On("IsWsl").Return(false)
		env.MockGitCommand("", strings.ReplaceAll(tc.Output, "\t", ""), "status", "-unormal", "--branch", "--porcelain=2")
		g := &Git{
			scm: scm{
				env:     env,
				props:   properties.Map{},
				command: GITCOMMAND,
			},
		}
		if tc.ExpectedWorking == nil {
			tc.ExpectedWorking = &GitStatus{}
		}
		if tc.ExpectedStaging == nil {
			tc.ExpectedStaging = &GitStatus{}
		}
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
		Expected     int
		StashContent string
	}{
		{Expected: 0, StashContent: ""},
		{Expected: 2, StashContent: "1\n2\n"},
		{Expected: 4, StashContent: "1\n2\n3\n4\n\n"},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("FileContent", "/logs/refs/stash").Return(tc.StashContent)
		g := &Git{
			scm: scm{
				env:        env,
				workingDir: "",
			},
		}
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
		{Case: "unsupported", Upstream: "\\test\\repo.git"},
	}
	for _, tc := range cases {
		g := &Git{}
		upstreamIcon := g.cleanUpstreamURL(tc.Upstream)
		assert.Equal(t, tc.Expected, upstreamIcon, tc.Case)
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
		{Case: "Gitstash", Expected: "G", Upstream: "gitstash.com/test"},
		{Case: "My custom server", Expected: "CU", Upstream: "mycustom.server/test"},
	}
	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		env.On("IsWsl").Return(false)
		env.On("RunCommand", "git", []string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "remote", "get-url", "origin"}).Return(tc.Upstream, nil)
		env.On("GOOS").Return("unix")
		props := properties.Map{
			GithubIcon:      "GH",
			GitlabIcon:      "GL",
			BitbucketIcon:   "BB",
			AzureDevOpsIcon: "AD",
			GitIcon:         "G",
			UpstreamIcons: map[string]string{
				"mycustom.server": "CU",
				"src.example.com": "EX",
			},
		}
		g := &Git{
			scm: scm{
				env:     env,
				props:   props,
				command: GITCOMMAND,
			},
			Upstream: "origin/main",
		}
		upstreamIcon := g.getUpstreamIcon()
		assert.Equal(t, tc.Expected, upstreamIcon, tc.Case)
	}
}

func TestGetBranchStatus(t *testing.T) {
	cases := []struct {
		Case         string
		Expected     string
		Ahead        int
		Behind       int
		Upstream     string
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
			scm: scm{
				props: props,
			},
			Ahead:        tc.Ahead,
			Behind:       tc.Behind,
			Upstream:     tc.Upstream,
			UpstreamGone: tc.UpstreamGone,
		}
		g.setBranchStatus()
		assert.Equal(t, tc.Expected, g.BranchStatus, tc.Case)
	}
}

func TestGitTemplateString(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
		Git      *Git
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
		env := new(mock.MockedEnvironment)
		tc.Git.env = env
		tc.Git.props = props
		assert.Equal(t, tc.Expected, renderTemplate(env, tc.Template, tc.Git), tc.Case)
	}
}

func TestGitUntrackedMode(t *testing.T) {
	cases := []struct {
		Case           string
		Expected       string
		UntrackedModes map[string]string
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
		g := &Git{
			scm: scm{
				props: properties.Map{
					UntrackedModes: tc.UntrackedModes,
				},
				realDir: "foo",
			},
		}
		got := g.getUntrackedFilesMode()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGitIgnoreSubmodules(t *testing.T) {
	cases := []struct {
		Case             string
		Expected         string
		IgnoreSubmodules map[string]string
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
		g := &Git{
			scm: scm{
				props: properties.Map{
					IgnoreSubmodules: tc.IgnoreSubmodules,
				},
				realDir: "foo",
			},
		}
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
			},
		},
		{
			Case: "No commit output",
			Expected: &Commit{
				Author:    &User{},
				Committer: &User{},
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
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.MockGitCommand("", tc.Output, "log", "-1", "--pretty=format:an:%an%nae:%ae%ncn:%cn%nce:%ce%nat:%at%nsu:%s")
		g := &Git{
			scm: scm{
				env:     env,
				command: "git",
			},
		}
		got := g.Commit()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
