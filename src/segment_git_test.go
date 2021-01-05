package main

import (
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
	env.On("runCommand", "git", []string{"rev-parse", "--is-inside-work-tree"}).Return("true", nil)
	g := &git{
		env: env,
	}
	assert.True(t, g.enabled())
}

func TestGetGitOutputForCommand(t *testing.T) {
	args := []string{"-c", "core.quotepath=false", "-c", "color.status=false"}
	commandArgs := []string{"symbolic-ref", "--short", "HEAD"}
	want := "je suis le output"
	env := new(MockedEnvironment)
	env.On("runCommand", "git", append(args, commandArgs...)).Return(want, nil)
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
	merge         bool
	mergeHEAD     string
}

func setupHEADContextEnv(context *detachedContext) *git {
	env := new(MockedEnvironment)
	env.On("hasFolder", "/.git/rebase-merge").Return(context.rebaseMerge)
	env.On("hasFolder", "/.git/rebase-apply").Return(context.rebaseApply)
	env.On("getFileContent", "/.git/rebase-merge/orig-head").Return(context.origin)
	env.On("getFileContent", "/.git/rebase-merge/onto").Return(context.onto)
	env.On("getFileContent", "/.git/rebase-merge/msgnum").Return(context.step)
	env.On("getFileContent", "/.git/rebase-apply/next").Return(context.step)
	env.On("getFileContent", "/.git/rebase-merge/end").Return(context.total)
	env.On("getFileContent", "/.git/rebase-apply/last").Return(context.total)
	env.On("getFileContent", "/.git/rebase-apply/head-name").Return(context.origin)
	env.On("getFileContent", "/.git/CHERRY_PICK_HEAD").Return(context.cherryPickSHA)
	env.On("getFileContent", "/.git/MERGE_HEAD").Return(context.mergeHEAD)
	env.On("hasFilesInDir", "", ".git/CHERRY_PICK_HEAD").Return(context.cherryPick)
	env.On("hasFilesInDir", "", ".git/MERGE_HEAD").Return(context.merge)
	env.mockGitCommand(context.currentCommit, "rev-parse", "--short", "HEAD")
	env.mockGitCommand(context.tagName, "describe", "--tags", "--exact-match")
	env.mockGitCommand(context.origin, "name-rev", "--name-only", "--exclude=tags/*", context.origin)
	env.mockGitCommand(context.onto, "name-rev", "--name-only", "--exclude=tags/*", context.onto)
	env.mockGitCommand(context.cherryPickSHA, "name-rev", "--name-only", "--exclude=tags/*", context.cherryPickSHA)
	env.mockGitCommand(context.mergeHEAD, "name-rev", "--name-only", "--exclude=tags/*", context.mergeHEAD)
	g := &git{
		env: env,
		repo: &gitRepo{
			root: "",
		},
	}
	return g
}

func (m *MockedEnvironment) mockGitCommand(returnValue string, args ...string) {
	args = append([]string{"-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
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

func TestGetGitHEADContextMerge(t *testing.T) {
	want := "\ue727 \ue0a0feat into \ue0a0main"
	context := &detachedContext{
		merge:     true,
		mergeHEAD: "feat",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextMergeTag(t *testing.T) {
	want := "\ue727 \ue0a0feat into \uf412v3.4.6"
	context := &detachedContext{
		tagName:   "v3.4.6",
		merge:     true,
		mergeHEAD: "feat",
	}
	g := setupHEADContextEnv(context)
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetStashContextZeroEntries(t *testing.T) {
	want := ""
	env := new(MockedEnvironment)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "rev-list", "--walk-reflogs", "--count", "refs/stash"}).Return("", nil)
	g := &git{
		env: env,
	}
	got := g.getStashContext()
	assert.Equal(t, want, got)
}

func TestGetStashContextMultipleEntries(t *testing.T) {
	want := "2"
	env := new(MockedEnvironment)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "rev-list", "--walk-reflogs", "--count", "refs/stash"}).Return("2", nil)
	g := &git{
		env: env,
	}
	got := g.getStashContext()
	assert.Equal(t, want, got)
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
	expected := "<#123456>working: x1</>"
	status := &gitStatus{
		unmerged: 1,
	}
	assert.Equal(t, expected, status.string("working:", "#123456"))
}

func TestGitStatusUnmergedModified(t *testing.T) {
	expected := "<#123456>working: ~3 x1</>"
	status := &gitStatus{
		unmerged: 1,
		modified: 3,
	}
	assert.Equal(t, expected, status.string("working:", "#123456"))
}

func TestGitStatusEmpty(t *testing.T) {
	expected := ""
	status := &gitStatus{}
	assert.Equal(t, expected, status.string("working:", "#123456"))
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
	assert.Equal(t, 1, status.added)
	assert.Equal(t, 1, status.deleted)
	assert.Equal(t, 2, status.untracked)
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
	assert.Equal(t, 0, status.untracked)
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
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "remote", "get-url", "origin"}).Return(upstream, nil)
	props := &properties{
		values: map[Property]interface{}{
			GithubIcon:    "GH",
			GitlabIcon:    "GL",
			BitbucketIcon: "BB",
			GitIcon:       "G",
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
				ColorBackground:   false,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.foreground)
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
				ColorBackground:   true,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.background)
}

func TestGetStatusDetailStringDefault(t *testing.T) {
	expected := "<#111111>icon +1</>"
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

func TestGetStatusDetailStringNoStatus(t *testing.T) {
	expected := "<#111111>icon</>"
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
