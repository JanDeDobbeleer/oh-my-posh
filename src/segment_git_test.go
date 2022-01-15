package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	changesColor = "#BD8BDE"
	branchName   = "main"
)

func TestEnabledGitNotFound(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("inWSLSharedDrive").Return(false)
	env.On("hasCommand", "git").Return(false)
	env.On("getRuntimeGOOS").Return("")
	env.On("isWsl").Return(false)
	g := &git{
		scm: scm{
			env:   env,
			props: properties{},
		},
	}
	assert.False(t, g.enabled())
}

func TestEnabledInWorkingDirectory(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("inWSLSharedDrive").Return(false)
	env.On("hasCommand", "git").Return(true)
	env.On("getRuntimeGOOS").Return("")
	env.On("isWsl").Return(false)
	fileInfo := &fileInfo{
		path:         "/dir/hello",
		parentFolder: "/dir",
		isDir:        true,
	}
	env.On("hasParentFilePath", ".git").Return(fileInfo, nil)
	g := &git{
		scm: scm{
			env:   env,
			props: properties{},
		},
	}
	assert.True(t, g.enabled())
	assert.Equal(t, fileInfo.path, g.gitWorkingFolder)
}

func TestEnabledInWorkingTree(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("inWSLSharedDrive").Return(false)
	env.On("hasCommand", "git").Return(true)
	env.On("getRuntimeGOOS").Return("")
	env.On("isWsl").Return(false)
	fileInfo := &fileInfo{
		path:         "/dev/folder_worktree/.git",
		parentFolder: "/dev/folder_worktree",
		isDir:        false,
	}
	env.On("hasParentFilePath", ".git").Return(fileInfo, nil)
	env.On("getFileContent", "/dev/folder_worktree/.git").Return("gitdir: /dev/real_folder/.git/worktrees/folder_worktree")
	env.On("getFileContent", "/dev/real_folder/.git/worktrees/folder_worktree/gitdir").Return("/dev/folder_worktree.git\n")
	g := &git{
		scm: scm{
			env:   env,
			props: properties{},
		},
	}
	assert.True(t, g.enabled())
	assert.Equal(t, "/dev/real_folder/.git/worktrees/folder_worktree", g.gitWorkingFolder)
	assert.Equal(t, "/dev/folder_worktree", g.gitRealFolder)
}

func TestEnabledInSubmodule(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("inWSLSharedDrive").Return(false)
	env.On("hasCommand", "git").Return(true)
	env.On("getRuntimeGOOS").Return("")
	env.On("isWsl").Return(false)
	fileInfo := &fileInfo{
		path:         "/dev/parent/test-submodule/.git",
		parentFolder: "/dev/parent/test-submodule",
		isDir:        false,
	}
	env.On("hasParentFilePath", ".git").Return(fileInfo, nil)
	env.On("getFileContent", "/dev/parent/test-submodule/.git").Return("gitdir: ../.git/modules/test-submodule")
	env.On("getFileContent", "/dev/parent/.git/modules/test-submodule").Return("/dev/folder_worktree.git\n")
	g := &git{
		scm: scm{
			env:   env,
			props: properties{},
		},
	}
	assert.True(t, g.enabled())
	assert.Equal(t, "/dev/parent/test-submodule/../.git/modules/test-submodule", g.gitWorkingFolder)
	assert.Equal(t, "/dev/parent/test-submodule/../.git/modules/test-submodule", g.gitRealFolder)
	assert.Equal(t, "/dev/parent/test-submodule/../.git/modules/test-submodule", g.gitRootFolder)
}

func TestGetGitOutputForCommand(t *testing.T) {
	args := []string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}
	commandArgs := []string{"symbolic-ref", "--short", "HEAD"}
	want := "je suis le output"
	env := new(MockedEnvironment)
	env.On("isWsl").Return(false)
	env.On("runCommand", "git", append(args, commandArgs...)).Return(want, nil)
	env.On("getRuntimeGOOS").Return("unix")
	g := &git{
		scm: scm{
			env:   env,
			props: properties{},
		},
	}
	got := g.getGitCommandOutput(commandArgs...)
	assert.Equal(t, want, got)
}

func (m *MockedEnvironment) mockGitCommand(returnValue string, args ...string) {
	args = append([]string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	m.On("runCommand", "git", args).Return(returnValue, nil)
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
		env := new(MockedEnvironment)
		env.On("inWSLSharedDrive").Return(false)
		env.On("getRuntimeGOOS").Return("unix")
		env.On("isWsl").Return(false)
		env.mockGitCommand("", "describe", "--tags", "--exact-match")
		env.mockGitCommand(tc.Theirs, "name-rev", "--name-only", "--exclude=tags/*", tc.Theirs)
		env.mockGitCommand(tc.Ours, "name-rev", "--name-only", "--exclude=tags/*", tc.Ours)
		// rebase merge
		env.On("hasFolder", "/rebase-merge").Return(tc.RebaseMerge)
		env.On("getFileContent", "/rebase-merge/head-name").Return(tc.Ours)
		env.On("getFileContent", "/rebase-merge/onto").Return(tc.Theirs)
		env.On("getFileContent", "/rebase-merge/msgnum").Return(tc.Step)
		env.On("getFileContent", "/rebase-merge/end").Return(tc.Total)
		// rebase apply
		env.On("hasFolder", "/rebase-apply").Return(tc.RebaseApply)
		env.On("getFileContent", "/rebase-apply/head-name").Return(tc.Ours)
		env.On("getFileContent", "/rebase-apply/next").Return(tc.Step)
		env.On("getFileContent", "/rebase-apply/last").Return(tc.Total)
		// merge
		env.On("hasFilesInDir", "", "MERGE_MSG").Return(tc.Merge)
		env.On("getFileContent", "/MERGE_MSG").Return(fmt.Sprintf("Merge %s into %s", tc.Theirs, tc.Ours))
		// cherry pick
		env.On("hasFilesInDir", "", "CHERRY_PICK_HEAD").Return(tc.CherryPick)
		env.On("getFileContent", "/CHERRY_PICK_HEAD").Return(tc.Theirs)
		// revert
		env.On("hasFilesInDir", "", "REVERT_HEAD").Return(tc.Revert)
		env.On("getFileContent", "/REVERT_HEAD").Return(tc.Theirs)
		// sequencer
		env.On("hasFilesInDir", "", "sequencer/todo").Return(tc.Sequencer)
		env.On("getFileContent", "/sequencer/todo").Return(tc.Theirs)

		g := &git{
			scm: scm{
				env: env,
				props: properties{
					BranchIcon:     "branch ",
					CommitIcon:     "commit ",
					RebaseIcon:     "rebase ",
					MergeIcon:      "merge ",
					CherryPickIcon: "pick ",
					TagIcon:        "tag ",
					RevertIcon:     "revert ",
				},
			},
			Hash: "1234567",
			Ref:  tc.Ref,
		}
		g.setGitHEADContext()
		assert.Equal(t, tc.Expected, g.HEAD, tc.Case)
	}
}

func TestSetPrettyHEADName(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Hash     string
		Tag      string
		HEAD     string
	}{
		{Case: "main", Expected: "branch main", HEAD: BRANCHPREFIX + "main"},
		{Case: "no hash", Expected: "commit 1234567", HEAD: "12345678910"},
		{Case: "hash on tag", Hash: "132312322321", Expected: "tag tag-1", HEAD: "12345678910", Tag: "tag-1"},
		{Case: "no hash on tag", Expected: "tag tag-1", Tag: "tag-1"},
		{Case: "hash on commit", Hash: "1234567", Expected: "commit 1234567"},
		{Case: "no hash on commit", Expected: "commit 1234567", HEAD: "12345678910"},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getFileContent", "/HEAD").Return(tc.HEAD)
		env.On("getRuntimeGOOS").Return("unix")
		env.On("isWsl").Return(false)
		env.mockGitCommand(tc.Tag, "describe", "--tags", "--exact-match")
		g := &git{
			scm: scm{
				env: env,
				props: properties{
					BranchIcon: "branch ",
					CommitIcon: "commit ",
					TagIcon:    "tag ",
				},
			},
			Hash: tc.Hash,
		}
		g.setPrettyHEADName()
		assert.Equal(t, tc.Expected, g.HEAD, tc.Case)
	}
}

func TestSetGitStatus(t *testing.T) {
	cases := []struct {
		Case             string
		Output           string
		ExpectedWorking  *GitStatus
		ExpectedStaging  *GitStatus
		ExpectedHash     string
		ExpectedRef      string
		ExpectedUpstream string
		ExpectedAhead    int
		ExpectedBehind   int
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
			1 .? N...
			1 .D N...
			1 .A N...
			1 .U N...
			1 A. N...
			`,
			ExpectedWorking: &GitStatus{ScmStatus: ScmStatus{Modified: 4, Added: 2, Deleted: 1, Unmerged: 1}},
			ExpectedStaging: &GitStatus{ScmStatus: ScmStatus{Added: 1}},
			ExpectedHash:    "1234567",
			ExpectedRef:     "rework-git-status",
		},
		{
			Case: "all different options on working and staging, with remote",
			Output: `
			# branch.oid 1234567891011121314
			# branch.head rework-git-status
			# branch.upstream origin/rework-git-status
			1 .R N...
			1 .C N...
			1 .M N...
			1 .m N...
			1 .? N...
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
			ExpectedWorking:  &GitStatus{ScmStatus: ScmStatus{Added: 3}},
		},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getRuntimeGOOS").Return("unix")
		env.On("isWsl").Return(false)
		env.mockGitCommand(strings.ReplaceAll(tc.Output, "\t", ""), "status", "-unormal", "--branch", "--porcelain=2")
		g := &git{
			scm: scm{
				env: env,
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
		assert.Equal(t, tc.ExpectedHash, g.Hash, tc.Case)
		assert.Equal(t, tc.ExpectedRef, g.Ref, tc.Case)
		assert.Equal(t, tc.ExpectedUpstream, g.Upstream, tc.Case)
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
		env := new(MockedEnvironment)
		env.On("getFileContent", "/logs/refs/stash").Return(tc.StashContent)
		g := &git{
			scm: scm{
				env: env,
			},
			gitWorkingFolder: "",
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
		env := &MockedEnvironment{}
		env.On("isWsl").Return(false)
		env.On("runCommand", "git", []string{"-C", "", "--no-optional-locks", "-c", "core.quotepath=false",
			"-c", "color.status=false", "remote", "get-url", "origin"}).Return(tc.Upstream, nil)
		env.On("getRuntimeGOOS").Return("unix")
		props := properties{
			GithubIcon:      "GH",
			GitlabIcon:      "GL",
			BitbucketIcon:   "BB",
			AzureDevOpsIcon: "AD",
			GitIcon:         "G",
		}
		g := &git{
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
		Case     string
		Expected string
		Ahead    int
		Behind   int
		Upstream string
	}{
		{Case: "Equal with remote", Expected: " equal", Upstream: branchName},
		{Case: "Ahead", Expected: " up2", Ahead: 2},
		{Case: "Behind", Expected: " down8", Behind: 8},
		{Case: "Behind and ahead", Expected: " up7 down8", Behind: 8, Ahead: 7},
		{Case: "Gone", Expected: " gone"},
		{Case: "Default (bug)", Expected: "", Behind: -8, Upstream: "wonky"},
	}

	for _, tc := range cases {
		props := properties{
			BranchAheadIcon:     "up",
			BranchBehindIcon:    "down",
			BranchIdenticalIcon: "equal",
			BranchGoneIcon:      "gone",
		}
		g := &git{
			scm: scm{
				props: props,
			},
			Ahead:    tc.Ahead,
			Behind:   tc.Behind,
			Upstream: tc.Upstream,
		}
		g.setBranchStatus()
		assert.Equal(t, tc.Expected, g.BranchStatus, tc.Case)
	}
}

func TestShouldIgnoreRootRepository(t *testing.T) {
	cases := []struct {
		Case     string
		Dir      string
		Expected bool
	}{
		{Case: "inside excluded", Dir: "/home/bill/repo"},
		{Case: "oustide excluded", Dir: "/home/melinda"},
		{Case: "excluded exact match", Dir: "/home/gates", Expected: true},
		{Case: "excluded inside match", Dir: "/home/gates/bill", Expected: true},
	}

	for _, tc := range cases {
		props := properties{
			ExcludeFolders: []string{
				"/home/bill",
				"/home/gates.*",
			},
		}
		env := new(MockedEnvironment)
		env.On("homeDir").Return("/home/bill")
		env.On("getRuntimeGOOS").Return(windowsPlatform)
		git := &git{
			scm: scm{
				props: props,
				env:   env,
			},
		}
		got := git.shouldIgnoreRootRepository(tc.Dir)
		assert.Equal(t, tc.Expected, got, tc.Case)
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
		{Case: "On Windows", Expected: "git.exe", GOOS: windowsPlatform},
		{Case: "Non Windows", Expected: "git"},
		{Case: "Iside WSL2, non shared", IsWSL: true, Expected: "git"},
		{Case: "Iside WSL2, shared", Expected: "git.exe", IsWSL: true, IsWslSharedPath: true, CWD: "/mnt/bill"},
		{Case: "Iside WSL1, shared", Expected: "git", IsWSL: true, IsWSL1: true, CWD: "/mnt/bill"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("isWsl").Return(tc.IsWSL)
		env.On("getRuntimeGOOS").Return(tc.GOOS)
		env.On("getcwd").Return(tc.CWD)
		wslUname := "5.10.60.1-microsoft-standard-WSL2"
		if tc.IsWSL1 {
			wslUname = "4.4.0-19041-Microsoft"
		}
		env.On("runCommand", "uname", []string{"-r"}).Return(wslUname, nil)
		g := &git{
			scm: scm{
				env: env,
			},
		}
		if tc.IsWslSharedPath {
			env.On("inWSLSharedDrive").Return(true)
			g.IsWslSharedPath = tc.IsWslSharedPath
		} else {
			env.On("inWSLSharedDrive").Return(false)
		}
		assert.Equal(t, tc.Expected, g.getGitCommand(), tc.Case)
	}
}

func TestGitTemplateString(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
		Git      *git
	}{
		{
			Case:     "Only HEAD name",
			Expected: branchName,
			Template: "{{ .HEAD }}",
			Git: &git{
				HEAD:   branchName,
				Behind: 2,
			},
		},
		{
			Case:     "Working area changes",
			Expected: "main \uF044 +2 ~3",
			Template: "{{ .HEAD }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Git: &git{
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
			Git: &git{
				HEAD:    branchName,
				Working: &GitStatus{},
			},
		},
		{
			Case:     "Working and staging area changes",
			Expected: "main \uF046 +5 ~1 \uF044 +2 ~3",
			Template: "{{ .HEAD }}{{ if .Staging.Changed }} \uF046 {{ .Staging.String }}{{ end }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}",
			Git: &git{
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
			Git: &git{
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
			Git: &git{
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
			Git: &git{
				HEAD:    branchName,
				Staging: &GitStatus{},
				Working: &GitStatus{},
			},
		},
		{
			Case:     "Upstream Icon",
			Expected: "from GitHub on main",
			Template: "from {{ .UpstreamIcon }} on {{ .HEAD }}",
			Git: &git{
				HEAD:         branchName,
				Staging:      &GitStatus{},
				Working:      &GitStatus{},
				UpstreamIcon: "GitHub",
			},
		},
	}

	for _, tc := range cases {
		props := properties{
			FetchStatus: true,
		}
		tc.Git.props = props
		assert.Equal(t, tc.Expected, tc.Git.templateString(tc.Template), tc.Case)
	}
}
