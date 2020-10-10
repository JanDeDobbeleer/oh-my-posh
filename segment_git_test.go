package main

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
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
	env.On("runCommand", "git", []string{"rev-parse", "--is-inside-work-tree"}).Return("true")
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
	env.On("runCommand", "git", append(args, commandArgs...)).Return(want)
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
}

func setupHEADContextEnv(context *detachedContext) environmentInfo {
	env := new(MockedEnvironment)
	env.On("hasFolder", ".git/rebase-merge").Return(context.rebaseMerge)
	env.On("hasFolder", ".git/rebase-apply").Return(context.rebaseApply)
	env.On("getFileContent", ".git/rebase-merge/orig-head").Return(context.origin)
	env.On("getFileContent", ".git/rebase-merge/onto").Return(context.onto)
	env.On("getFileContent", ".git/rebase-merge/msgnum").Return(context.step)
	env.On("getFileContent", ".git/rebase-apply/next").Return(context.step)
	env.On("getFileContent", ".git/rebase-merge/end").Return(context.total)
	env.On("getFileContent", ".git/rebase-apply/last").Return(context.total)
	env.On("getFileContent", ".git/rebase-apply/head-name").Return(context.origin)
	env.On("getFileContent", ".git/CHERRY_PICK_HEAD").Return(context.cherryPickSHA)
	env.On("hasFiles", ".git/CHERRY_PICK_HEAD").Return(context.cherryPick)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "rev-parse", "--short", "HEAD"}).Return(context.currentCommit)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "describe", "--tags", "--exact-match"}).Return(context.tagName)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "name-rev", "--name-only", "--exclude=tags/*", context.origin}).Return(context.origin)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "name-rev", "--name-only", "--exclude=tags/*", context.onto}).Return(context.onto)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "name-rev", "--name-only", "--exclude=tags/*", context.cherryPickSHA}).Return(context.cherryPickSHA)
	return env
}

func TestGetGitDetachedCommitHash(t *testing.T) {
	want := "COMMIT:lalasha1"
	context := &detachedContext{
		currentCommit: "lalasha1",
	}
	env := setupHEADContextEnv(context)
	g := &git{
		env: env,
	}
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextTagName(t *testing.T) {
	want := "TAG:lalasha1"
	context := &detachedContext{
		currentCommit: "whatever",
		tagName:       "lalasha1",
	}
	env := setupHEADContextEnv(context)
	g := &git{
		env: env,
	}
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRebaseMerge(t *testing.T) {
	want := "REBASE:BRANCH:cool-feature-bro onto BRANCH:main (2/3) at COMMIT:whatever"
	context := &detachedContext{
		currentCommit: "whatever",
		rebase:        "true",
		rebaseMerge:   true,
		origin:        "cool-feature-bro",
		onto:          "main",
		step:          "2",
		total:         "3",
	}
	env := setupHEADContextEnv(context)
	g := &git{
		env: env,
	}
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRebaseApply(t *testing.T) {
	want := "REBASING:BRANCH:cool-feature-bro (2/3) at COMMIT:whatever"
	context := &detachedContext{
		currentCommit: "whatever",
		rebase:        "true",
		rebaseApply:   true,
		origin:        "cool-feature-bro",
		step:          "2",
		total:         "3",
	}
	env := setupHEADContextEnv(context)
	g := &git{
		env: env,
	}
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextRebaseUnknown(t *testing.T) {
	want := "COMMIT:whatever"
	context := &detachedContext{
		currentCommit: "whatever",
		rebase:        "true",
	}
	env := setupHEADContextEnv(context)
	g := &git{
		env: env,
	}
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextCherryPickOnBranch(t *testing.T) {
	want := "CHERRY PICK:pickme onto BRANCH:main"
	context := &detachedContext{
		currentCommit: "whatever",
		branchName:    "main",
		cherryPick:    true,
		cherryPickSHA: "pickme",
	}
	env := setupHEADContextEnv(context)
	g := &git{
		env: env,
	}
	got := g.getGitHEADContext("main")
	assert.Equal(t, want, got)
}

func TestGetGitHEADContextCherryPickOnTag(t *testing.T) {
	want := "CHERRY PICK:pickme onto TAG:v3.4.6"
	context := &detachedContext{
		currentCommit: "whatever",
		tagName:       "v3.4.6",
		cherryPick:    true,
		cherryPickSHA: "pickme",
	}
	env := setupHEADContextEnv(context)
	g := &git{
		env: env,
	}
	got := g.getGitHEADContext("")
	assert.Equal(t, want, got)
}

func TestGetStashContextZeroEntries(t *testing.T) {
	want := 0
	env := new(MockedEnvironment)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "stash", "list"}).Return("")
	g := &git{
		env: env,
	}
	got := g.getStashContext()
	assert.Equal(t, want, got)
}

func TestGetStashContextMultipleEntries(t *testing.T) {
	want := rand.Intn(100)
	var response string
	for i := 0; i < want; i++ {
		response += "I'm a stash entry\n"
	}
	env := new(MockedEnvironment)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "stash", "list"}).Return(response)
	g := &git{
		env: env,
	}
	got := g.getStashContext()
	assert.Equal(t, want, got)
}

func TestGetStashContextOneEntry(t *testing.T) {
	want := 1
	env := new(MockedEnvironment)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "stash", "list"}).Return("stash entry")
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
