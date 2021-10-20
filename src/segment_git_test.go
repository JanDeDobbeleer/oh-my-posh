package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	changesColor = "#BD8BDE"
)

func TestEnabledGitNotFound(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "git").Return(false)
	g := &git{
		env: env,
	}
	assert.False(t, g.enabled())
}

func TestEnabledInWorkingDirectory(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "git").Return(true)
	fileInfo := &fileInfo{
		path:         "/dir/hello",
		parentFolder: "/dir",
		isDir:        true,
	}
	env.On("hasParentFilePath", ".git").Return(fileInfo, nil)
	g := &git{
		env: env,
	}
	assert.True(t, g.enabled())
	assert.Equal(t, fileInfo.path, g.repo.gitWorkingFolder)
}

func TestEnabledInWorkingTree(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "git").Return(true)
	fileInfo := &fileInfo{
		path:         "/dir/hello",
		parentFolder: "/dir",
		isDir:        false,
	}
	env.On("hasParentFilePath", ".git").Return(fileInfo, nil)
	env.On("getFileContent", "/dir/hello").Return("gitdir: /dir/hello/burp/burp")
	g := &git{
		env: env,
	}
	assert.True(t, g.enabled())
	assert.Equal(t, "/dir/hello/burp/burp", g.repo.gitWorkingFolder)
}

func TestGetGitOutputForCommand(t *testing.T) {
	args := []string{"--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}
	commandArgs := []string{"symbolic-ref", "--short", "HEAD"}
	want := "je suis le output"
	env := new(MockedEnvironment)
	env.On("isWsl", nil).Return(false)
	env.On("runCommand", "git", append(args, commandArgs...)).Return(want, nil)
	env.On("getRuntimeGOOS", nil).Return("unix")
	g := &git{
		env: env,
	}
	got := g.getGitCommandOutput(commandArgs...)
	assert.Equal(t, want, got)
}

type detachedContext struct {
	currentCommit string
	rebase        string
	rebaseMerge   bool
	rebaseApply   bool
	origin        string
	onto          string
	step          string
	total         string
	branchName    string
	tagName       string
	cherryPick    bool
	cherryPickSHA string
	revert        bool
	revertSHA     string
	sequencer     bool
	sequencerTodo string
	merge         bool
	mergeHEAD     string
	mergeMsgStart string
	status        string
}

func setupHEADContextEnv(context *detachedContext) *git {
	env := new(MockedEnvironment)
	env.On("isWsl", nil).Return(false)
	env.On("hasFolder", "/rebase-merge").Return(context.rebaseMerge)
	env.On("hasFolder", "/rebase-apply").Return(context.rebaseApply)
	env.On("hasFolder", "/sequencer").Return(context.sequencer)
	env.On("getFileContent", "/rebase-merge/head-name").Return(context.origin)
	env.On("getFileContent", "/rebase-merge/onto").Return(context.onto)
	env.On("getFileContent", "/rebase-merge/msgnum").Return(context.step)
	env.On("getFileContent", "/rebase-apply/next").Return(context.step)
	env.On("getFileContent", "/rebase-merge/end").Return(context.total)
	env.On("getFileContent", "/rebase-apply/last").Return(context.total)
	env.On("getFileContent", "/rebase-apply/head-name").Return(context.origin)
	env.On("getFileContent", "/CHERRY_PICK_HEAD").Return(context.cherryPickSHA)
	env.On("getFileContent", "/REVERT_HEAD").Return(context.revertSHA)
	env.On("getFileContent", "/MERGE_MSG").Return(fmt.Sprintf("%s '%s' into %s", context.mergeMsgStart, context.mergeHEAD, context.onto))
	env.On("getFileContent", "/sequencer/todo").Return(context.sequencerTodo)
	env.On("getFileContent", "/HEAD").Return(context.branchName)
	env.On("hasFilesInDir", "", "CHERRY_PICK_HEAD").Return(context.cherryPick)
	env.On("hasFilesInDir", "", "REVERT_HEAD").Return(context.revert)
	env.On("hasFilesInDir", "", "MERGE_MSG").Return(context.merge)
	env.On("hasFilesInDir", "", "MERGE_HEAD").Return(context.merge)
	env.On("hasFilesInDir", "", "sequencer/todo").Return(context.sequencer)
	env.mockGitCommand(context.currentCommit, "rev-parse", "--short", "HEAD")
	env.mockGitCommand(context.tagName, "describe", "--tags", "--exact-match")
	env.mockGitCommand(context.origin, "name-rev", "--name-only", "--exclude=tags/*", context.origin)
	env.mockGitCommand(context.onto, "name-rev", "--name-only", "--exclude=tags/*", context.onto)
	env.mockGitCommand(context.branchName, "branch", "--show-current")
	env.mockGitCommand(context.status, "status", "-unormal", "--short", "--branch")
	env.On("getRuntimeGOOS", nil).Return("unix")
	g := &git{
		env: env,
		repo: &gitRepo{
			gitWorkingFolder: "",
		},
	}
	return g
}

func (m *MockedEnvironment) mockGitCommand(returnValue string, args ...string) {
	args = append([]string{"--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	m.On("runCommand", "git", args).Return(returnValue, nil)
}

func TestGetGitDetachedCommitHash(t *testing.T) {
	want := "\uf417lalasha1"
	context := &detachedContext{
		currentCommit: "lalasha1",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextTagName(t *testing.T) {
	want := "\uf412lalasha1"
	context := &detachedContext{
		currentCommit: "whatever",
		tagName:       "lalasha1",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRebaseMerge(t *testing.T) {
	want := "\ue728 \ue0a0cool-feature-bro onto \ue0a0main (2/3) at \uf417whatever"
	context := &detachedContext{
		currentCommit: "whatever",
		rebase:        "true",
		rebaseMerge:   true,
		origin:        "cool-feature-bro",
		onto:          "main",
		step:          "2",
		total:         "3",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRebaseApply(t *testing.T) {
	want := "\ue728 \ue0a0cool-feature-bro (2/3) at \uf417whatever"
	context := &detachedContext{
		currentCommit: "whatever",
		rebase:        "true",
		rebaseApply:   true,
		origin:        "cool-feature-bro",
		step:          "2",
		total:         "3",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRebaseUnknown(t *testing.T) {
	want := "\uf417whatever"
	context := &detachedContext{
		currentCommit: "whatever",
		rebase:        "true",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextCherryPickOnBranch(t *testing.T) {
	want := "\ue29b pickme onto \ue0a0main"
	context := &detachedContext{
		currentCommit: "whatever",
		branchName:    "main",
		cherryPick:    true,
		cherryPickSHA: "pickme",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextCherryPickOnTag(t *testing.T) {
	want := "\ue29b pickme onto \uf412v3.4.6"
	context := &detachedContext{
		currentCommit: "whatever",
		tagName:       "v3.4.6",
		cherryPick:    true,
		cherryPickSHA: "pickme",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRevertOnBranch(t *testing.T) {
	want := "\uf0e2 012345 onto \ue0a0main"
	context := &detachedContext{
		currentCommit: "whatever",
		branchName:    "main",
		revert:        true,
		revertSHA:     "01234567",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRevertOnTag(t *testing.T) {
	want := "\uf0e2 012345 onto \uf412v3.4.6"
	context := &detachedContext{
		currentCommit: "whatever",
		tagName:       "v3.4.6",
		revert:        true,
		revertSHA:     "01234567",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextSequencerCherryPickOnBranch(t *testing.T) {
	want := "\ue29b pickme onto \ue0a0main"
	context := &detachedContext{
		currentCommit: "whatever",
		branchName:    "main",
		sequencer:     true,
		sequencerTodo: "pick pickme message\npick notme message",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextSequencerCherryPickOnTag(t *testing.T) {
	want := "\ue29b pickme onto \uf412v3.4.6"
	context := &detachedContext{
		currentCommit: "whatever",
		tagName:       "v3.4.6",
		sequencer:     true,
		sequencerTodo: "pick pickme message\npick notme message",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextSequencerRevertOnBranch(t *testing.T) {
	want := "\uf0e2 012345 onto \ue0a0main"
	context := &detachedContext{
		currentCommit: "whatever",
		branchName:    "main",
		sequencer:     true,
		sequencerTodo: "revert 01234567 message\nrevert notme message",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextSequencerRevertOnTag(t *testing.T) {
	want := "\uf0e2 012345 onto \uf412v3.4.6"
	context := &detachedContext{
		currentCommit: "whatever",
		tagName:       "v3.4.6",
		sequencer:     true,
		sequencerTodo: "revert 01234567 message\nrevert notme message",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextMerge(t *testing.T) {
	want := "\ue727 \ue0a0feat into \ue0a0main"
	context := &detachedContext{
		merge:         true,
		mergeHEAD:     "feat",
		mergeMsgStart: "Merge branch",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextMergeRemote(t *testing.T) {
	want := "\ue727 \ue0a0feat into \ue0a0main"
	context := &detachedContext{
		merge:         true,
		mergeHEAD:     "feat",
		mergeMsgStart: "Merge remote-tracking branch",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextMergeTag(t *testing.T) {
	want := "\ue727 \uf412v7.8.9 into \ue0a0main"
	context := &detachedContext{
		merge:         true,
		mergeHEAD:     "v7.8.9",
		mergeMsgStart: "Merge tag",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextMergeCommit(t *testing.T) {
	want := "\ue727 \uf4178d7e869 into \ue0a0main"
	context := &detachedContext{
		merge:         true,
		mergeHEAD:     "8d7e869",
		mergeMsgStart: "Merge commit",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextMergeIntoTag(t *testing.T) {
	want := "\ue727 \ue0a0feat into \uf412v3.4.6"
	context := &detachedContext{
		tagName:       "v3.4.6",
		merge:         true,
		mergeHEAD:     "feat",
		mergeMsgStart: "Merge branch",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
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
			repo: &gitRepo{
				gitWorkingFolder: "",
			},
			env: env,
		}
		got := g.getStashContext()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestParseGitBranchInfoEqual(t *testing.T) {
	g := git{}
	branchInfo := "## master...origin/master"
	got := g.parseGitStatusInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Empty(t, got["ahead"])
	assert.Empty(t, got["behind"])
}

func TestParseGitBranchInfoAhead(t *testing.T) {
	g := git{}
	branchInfo := "## master...origin/master [ahead 1]"
	got := g.parseGitStatusInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Equal(t, "1", got["ahead"])
	assert.Empty(t, got["behind"])
}

func TestParseGitBranchInfoBehind(t *testing.T) {
	g := git{}
	branchInfo := "## master...origin/master [behind 1]"
	got := g.parseGitStatusInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Equal(t, "1", got["behind"])
	assert.Empty(t, got["ahead"])
}

func TestParseGitBranchInfoBehindandAhead(t *testing.T) {
	g := git{}
	branchInfo := "## master...origin/master [ahead 1, behind 2]"
	got := g.parseGitStatusInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Equal(t, "2", got["behind"])
	assert.Equal(t, "1", got["ahead"])
}

func TestParseGitBranchInfoNoRemote(t *testing.T) {
	g := git{}
	branchInfo := "## master"
	got := g.parseGitStatusInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Empty(t, got["upstream"])
}

func TestParseGitBranchInfoRemoteGone(t *testing.T) {
	g := git{}
	branchInfo := "## test-branch...origin/test-branch [gone]"
	got := g.parseGitStatusInfo(branchInfo)
	assert.Equal(t, "test-branch", got["local"])
	assert.Equal(t, "gone", got["upstream_status"])
}

func TestGitStatusUnmerged(t *testing.T) {
	expected := " x1"
	status := &gitStatus{
		unmerged: 1,
	}
	assert.Equal(t, expected, status.string())
}

func TestGitStatusUnmergedModified(t *testing.T) {
	expected := " ~3 x1"
	status := &gitStatus{
		unmerged: 1,
		modified: 3,
	}
	assert.Equal(t, expected, status.string())
}

func TestGitStatusEmpty(t *testing.T) {
	expected := ""
	status := &gitStatus{}
	assert.Equal(t, expected, status.string())
}

func TestParseGitStatsWorking(t *testing.T) {
	g := &git{}
	output := []string{
		"## amazing-feat",
		" M change.go",
		"DD change.go",
		" ? change.go",
		" ? change.go",
		" A change.go",
		" U change.go",
		" R change.go",
		" C change.go",
	}
	status := g.parseGitStats(output, true)
	assert.Equal(t, 3, status.modified)
	assert.Equal(t, 1, status.unmerged)
	assert.Equal(t, 3, status.added)
	assert.Equal(t, 1, status.deleted)
	assert.True(t, status.changed)
}

func TestParseGitStatsStaging(t *testing.T) {
	g := &git{}
	output := []string{
		"## amazing-feat",
		" M change.go",
		"DD change.go",
		" ? change.go",
		"?? change.go",
		" A change.go",
		"DU change.go",
		"MR change.go",
		"AC change.go",
	}
	status := g.parseGitStats(output, false)
	assert.Equal(t, 1, status.modified)
	assert.Equal(t, 0, status.unmerged)
	assert.Equal(t, 1, status.added)
	assert.Equal(t, 2, status.deleted)
	assert.True(t, status.changed)
}

func TestParseGitStatsNoChanges(t *testing.T) {
	g := &git{}
	expected := &gitStatus{}
	output := []string{
		"## amazing-feat",
	}
	status := g.parseGitStats(output, false)
	assert.Equal(t, expected, status)
	assert.False(t, status.changed)
}

func TestParseGitStatsInvalidLine(t *testing.T) {
	g := &git{}
	expected := &gitStatus{}
	output := []string{
		"## amazing-feat",
		"#",
	}
	status := g.parseGitStats(output, false)
	assert.Equal(t, expected, status)
	assert.False(t, status.changed)
}

func bootstrapUpstreamTest(upstream string) *git {
	env := &MockedEnvironment{}
	env.On("isWsl", nil).Return(false)
	env.On("runCommand", "git", []string{"--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false", "remote", "get-url", "origin"}).Return(upstream, nil)
	env.On("getRuntimeGOOS", nil).Return("unix")
	props := &properties{
		values: map[Property]interface{}{
			GithubIcon:      "GH",
			GitlabIcon:      "GL",
			BitbucketIcon:   "BB",
			AzureDevOpsIcon: "AD",
			GitIcon:         "G",
		},
	}
	g := &git{
		env: env,
		repo: &gitRepo{
			upstream: "origin/main",
		},
		props: props,
	}
	return g
}

func TestGetUpstreamSymbolGitHub(t *testing.T) {
	g := bootstrapUpstreamTest("github.com/test")
	upstreamIcon := g.getUpstreamSymbol()
	assert.Equal(t, "GH", upstreamIcon)
}

func TestGetUpstreamSymbolGitLab(t *testing.T) {
	g := bootstrapUpstreamTest("gitlab.com/test")
	upstreamIcon := g.getUpstreamSymbol()
	assert.Equal(t, "GL", upstreamIcon)
}

func TestGetUpstreamSymbolBitBucket(t *testing.T) {
	g := bootstrapUpstreamTest("bitbucket.org/test")
	upstreamIcon := g.getUpstreamSymbol()
	assert.Equal(t, "BB", upstreamIcon)
}

func TestGetUpstreamSymbolAzureDevOps(t *testing.T) {
	g := bootstrapUpstreamTest("dev.azure.com/test")
	upstreamIcon := g.getUpstreamSymbol()
	assert.Equal(t, "AD", upstreamIcon)

	g = bootstrapUpstreamTest("test.visualstudio.com")
	upstreamIcon = g.getUpstreamSymbol()
	assert.Equal(t, "AD", upstreamIcon)
}

func TestGetUpstreamSymbolGit(t *testing.T) {
	g := bootstrapUpstreamTest("gitstash.com/test")
	upstreamIcon := g.getUpstreamSymbol()
	assert.Equal(t, "G", upstreamIcon)
}

func TestGetStatusColorLocalChangesStaging(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{
			changed: true,
		},
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorLocalChangesWorking(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{},
		working: &gitStatus{
			changed: true,
		},
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAheadAndBehind(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{},
		working: &gitStatus{},
		ahead:   1,
		behind:  3,
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				AheadAndBehindColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAhead(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{},
		working: &gitStatus{},
		ahead:   1,
		behind:  0,
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				AheadColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorBehind(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{},
		working: &gitStatus{},
		ahead:   0,
		behind:  5,
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				BehindColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorDefault(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{},
		working: &gitStatus{},
		ahead:   0,
		behind:  0,
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				BehindColor: changesColor,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor(expected))
}

func TestSetStatusColorForeground(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{
			changed: true,
		},
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: changesColor,
				ColorBackground:   false,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.foreground)
}

func TestSetStatusColorBackground(t *testing.T) {
	expected := changesColor
	repo := &gitRepo{
		staging: &gitStatus{
			changed: true,
		},
	}
	g := &git{
		repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: changesColor,
				ColorBackground:   true,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.background)
}

func TestStatusColorsWithoutDisplayStatus(t *testing.T) {
	expected := changesColor
	context := &detachedContext{
		status: "## main...origin/main [ahead 33]\n M myfile",
	}
	g := setupHEADContextEnv(context)
	g.props = &properties{
		values: map[Property]interface{}{
			DisplayStatus:       false,
			StatusColorsEnabled: true,
			LocalChangesColor:   expected,
		},
	}
	g.string()
	assert.Equal(t, expected, g.props.background)
}

func TestGetStatusDetailStringDefault(t *testing.T) {
	expected := "icon +1"
	status := &gitStatus{
		changed: true,
		added:   1,
	}
	g := &git{
		props: &properties{
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverride(t *testing.T) {
	expected := "<#123456>icon +1</>"
	status := &gitStatus{
		changed: true,
		added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor: "#123456",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideAndIconColorOverride(t *testing.T) {
	expected := "<#789123>work</><#123456> +1</>"
	status := &gitStatus{
		changed: true,
		added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor:     "#123456",
				LocalWorkingIcon: "<#789123>work</>",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideNoIconColorOverride(t *testing.T) {
	expected := "<#123456>work +1</>"
	status := &gitStatus{
		changed: true,
		added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor:     "#123456",
				LocalWorkingIcon: "work",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatus(t *testing.T) {
	expected := "icon"
	status := &gitStatus{
		changed: true,
		added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				DisplayStatusDetail: false,
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatusColorOverride(t *testing.T) {
	expected := "<#123456>icon</>"
	status := &gitStatus{
		changed: true,
		added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				DisplayStatusDetail: false,
				WorkingColor:        "#123456",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetBranchStatus(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Ahead    int
		Behind   int
		Upstream string
	}{
		{Case: "Equal with remote", Expected: " equal", Upstream: "main"},
		{Case: "Ahead", Expected: " up2", Ahead: 2},
		{Case: "Behind", Expected: " down8", Behind: 8},
		{Case: "Behind and ahead", Expected: " up7 down8", Behind: 8, Ahead: 7},
		{Case: "Gone", Expected: " gone"},
		{Case: "Default (bug)", Expected: "", Behind: -8, Upstream: "wonky"},
	}

	for _, tc := range cases {
		g := &git{
			props: &properties{
				values: map[Property]interface{}{
					BranchAheadIcon:     "up",
					BranchBehindIcon:    "down",
					BranchIdenticalIcon: "equal",
					BranchGoneIcon:      "gone",
				},
			},
			repo: &gitRepo{
				ahead:    tc.Ahead,
				behind:   tc.Behind,
				upstream: tc.Upstream,
			},
		}
		assert.Equal(t, tc.Expected, g.getBranchStatus(), tc.Case)
	}
}

func TestTruncateBranch(t *testing.T) {
	cases := []struct {
		Case      string
		Expected  string
		Branch    string
		MaxLength interface{}
	}{
		{Case: "No limit", Expected: "all-your-base-are-belong-to-us", Branch: "all-your-base-are-belong-to-us"},
		{Case: "No limit - larger", Expected: "all-your-base", Branch: "all-your-base-are-belong-to-us", MaxLength: 13.0},
		{Case: "No limit - smaller", Expected: "all-your-base", Branch: "all-your-base", MaxLength: 13.0},
		{Case: "Invalid setting", Expected: "all-your-base", Branch: "all-your-base", MaxLength: "burp"},
		{Case: "Lower than limit", Expected: "all-your-base", Branch: "all-your-base", MaxLength: 20.0},
	}

	for _, tc := range cases {
		g := &git{
			props: &properties{
				values: map[Property]interface{}{
					BranchMaxLength: tc.MaxLength,
				},
			},
		}
		assert.Equal(t, tc.Expected, g.truncateBranch(tc.Branch), tc.Case)
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
		props := map[Property]interface{}{
			ExcludeFolders: []string{
				"/home/bill",
				"/home/gates.*",
			},
		}
		env := new(MockedEnvironment)
		env.On("homeDir", nil).Return("/home/bill")
		env.On("getRuntimeGOOS", nil).Return(windowsPlatform)
		git := &git{
			props: &properties{
				values: props,
			},
			env: env,
		}
		got := git.shouldIgnoreRootRepository(tc.Dir)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
