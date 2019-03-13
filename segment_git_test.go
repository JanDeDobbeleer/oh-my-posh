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
	got := g.getGitOutputForCommand(commandArgs...)
	assert.Equal(t, want, got)
}

func TestGetGitDetachedBranch(t *testing.T) {
	want := "master"
	env := new(MockedEnvironment)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "symbolic-ref", "--short", "HEAD"}).Return(want)
	g := &git{
		env: env,
	}
	got := g.getGitDetachedBranch()
	assert.Equal(t, want, got)
}

func TestGetGitDetachedBranchEmpty(t *testing.T) {
	want := "unknown"
	env := new(MockedEnvironment)
	env.On("runCommand", "git", []string{"-c", "core.quotepath=false", "-c", "color.status=false", "symbolic-ref", "--short", "HEAD"}).Return("")
	g := &git{
		env: env,
	}
	got := g.getGitDetachedBranch()
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
	got := g.parseGitBranchInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Empty(t, got["ahead"])
	assert.Empty(t, got["behind"])
}

func TestParseGitBranchInfoAhead(t *testing.T) {
	g := git{}
	branchInfo := "## master...origin/master [ahead 1]"
	got := g.parseGitBranchInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Equal(t, "1", got["ahead"])
	assert.Empty(t, got["behind"])
}

func TestParseGitBranchInfoBehind(t *testing.T) {
	g := git{}
	branchInfo := "## master...origin/master [behind 1]"
	got := g.parseGitBranchInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Equal(t, "1", got["behind"])
	assert.Empty(t, got["ahead"])
}

func TestParseGitBranchInfoBehindandAhead(t *testing.T) {
	g := git{}
	branchInfo := "## master...origin/master [ahead 1, behind 2]"
	got := g.parseGitBranchInfo(branchInfo)
	assert.Equal(t, "master", got["local"])
	assert.Equal(t, "origin/master", got["upstream"])
	assert.Equal(t, "2", got["behind"])
	assert.Equal(t, "1", got["ahead"])
}

// func TestGetGitStatus(t *testing.T) {
// 	env := new(environment)
// 	writer := gitWriter{
// 		env: env,
// 	}
// 	writer.getGitStatus()
// 	assert.NotEmpty(t, writer.repo)
// }
