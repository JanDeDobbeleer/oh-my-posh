package segments

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	fileInfo := &environment.FileInfo{
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
	env.On("Home").Return("/Users/posh")
	env.On("Getenv", poshGitEnv).Return("")
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
	fileInfo := &environment.FileInfo{
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
			env:   env,
			props: properties.Map{},
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
				env:   env,
				props: properties.Map{},
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
		got := g.getStashContext()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestGitUpstream(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Upstream string
	}{
		{Case: "GitHub", Expected: "GH", Upstream: "github.com/test"},
		{Case: "Gitlab", Expected: "GL", Upstream: "gitlab.com/test"},
		{Case: "Bitbucket", Expected: "BB", Upstream: "bitbucket.org/test"},
		{Case: "Azure DevOps", Expected: "AD", Upstream: "dev.azure.com/test"},
		{Case: "Azure DevOps Dos", Expected: "AD", Upstream: "test.visualstudio.com"},
		{Case: "Gitstash", Expected: "G", Upstream: "gitstash.com/test"},
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
		}
		g := &Git{
			scm: scm{
				env:   env,
				props: props,
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

func TestGetGitCommand(t *testing.T) {
	cases := []struct {
		Case            string
		Expected        string
		IsWSL           bool
		IsWSL1          bool
		GOOS            string
		CWD             string
		IsWslSharedPath bool
	}{
		{Case: "On Windows", Expected: "git.exe", GOOS: environment.WINDOWS},
		{Case: "Non Windows", Expected: "git"},
		{Case: "Iside WSL2, non shared", IsWSL: true, Expected: "git"},
		{Case: "Iside WSL2, shared", Expected: "git.exe", IsWSL: true, IsWslSharedPath: true, CWD: "/mnt/bill"},
		{Case: "Iside WSL1, shared", Expected: "git", IsWSL: true, IsWSL1: true, CWD: "/mnt/bill"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("IsWsl").Return(tc.IsWSL)
		env.On("GOOS").Return(tc.GOOS)
		env.On("Pwd").Return(tc.CWD)
		wslUname := "5.10.60.1-microsoft-standard-WSL2"
		if tc.IsWSL1 {
			wslUname = "4.4.0-19041-Microsoft"
		}
		env.On("RunCommand", "uname", []string{"-r"}).Return(wslUname, nil)
		g := &Git{
			scm: scm{
				env: env,
			},
		}
		if tc.IsWslSharedPath {
			env.On("InWSLSharedDrive").Return(true)
			g.IsWslSharedPath = tc.IsWslSharedPath
		} else {
			env.On("InWSLSharedDrive").Return(false)
		}
		assert.Equal(t, tc.Expected, g.getCommand(GITCOMMAND), tc.Case)
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
			Expected: "main \uF046 +5 ~1 | \uF044 +2 ~3 \uf692 3",
			Template: "{{ .HEAD }}{{ if .Staging.Changed }} \uF046 {{ .Staging.String }}{{ end }}{{ if and (.Working.Changed) (.Staging.Changed) }} |{{ end }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}{{ if gt .StashCount 0 }} \uF692 {{ .StashCount }}{{ end }}", //nolint:lll
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
				StashCount: 3,
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
