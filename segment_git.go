package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type gitRepo struct {
	working    *gitStatus
	staging    *gitStatus
	ahead      int
	behind     int
	branch     string
	upstream   string
	stashCount int
}

type gitStatus struct {
	unmerged  int
	deleted   int
	added     int
	modified  int
	untracked int
}

type git struct {
	props *properties
	env   environmentInfo
	repo  *gitRepo
}

const (
	//BranchIcon the icon to use as branch indicator
	BranchIcon Property = "branch_icon"
	//BranchIdenticalIcon the icon to display when the remote and local branch are identical
	BranchIdenticalIcon Property = "branch_identical_icon"
	//BranchAheadIcon the icon to display when the local branch is ahead of the remote
	BranchAheadIcon Property = "branch_ahead_icon"
	//BranchBehindIcon the icon to display when the local branch is behind the remote
	BranchBehindIcon Property = "branch_behind_icon"
	//LocalWorkingIcon the icon to use as the local working area changes indicator
	LocalWorkingIcon Property = "local_working_icon"
	//LocalStagingIcon the icon to use as the local staging area changes indicator
	LocalStagingIcon Property = "local_staged_icon"
	//DisplayStatus shows the status of the repository
	DisplayStatus Property = "display_status"
)

func (g *git) enabled() bool {
	if !g.env.hasCommand("git") {
		return false
	}
	output := g.env.runCommand("git", "rev-parse", "--is-inside-work-tree")
	return output == "true"
}

func (g *git) string() string {
	g.getGitStatus()
	buffer := new(bytes.Buffer)
	// branchsymbol
	buffer.WriteString(g.props.getString(BranchIcon, "Branch: "))
	// branchName
	fmt.Fprintf(buffer, "%s", g.repo.branch)
	displayStatus := g.props.getBool(DisplayStatus, true)
	if !displayStatus {
		return buffer.String()
	}
	// TODO: Add upstream gone icon
	// if ahead, print with symbol
	if g.repo.ahead > 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(BranchAheadIcon, "+"), g.repo.ahead)
	}
	// if behind, print with symbol
	if g.repo.behind > 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(BranchBehindIcon, "-"), g.repo.behind)
	}
	if g.repo.behind == 0 && g.repo.ahead == 0 {
		fmt.Fprintf(buffer, " %s", g.props.getString(BranchIdenticalIcon, "="))
	}
	// if staging, print that part
	if g.hasStaging() {
		fmt.Fprintf(buffer, " %s +%d ~%d -%d", g.props.getString(LocalStagingIcon, "~"), g.repo.staging.added, g.repo.staging.modified, g.repo.staging.deleted)
	}
	// if working, print that part
	if g.hasWorking() {
		fmt.Fprintf(buffer, " %s +%d ~%d -%d", g.props.getString(LocalWorkingIcon, "#"), g.repo.working.added+g.repo.working.untracked, g.repo.working.modified, g.repo.working.deleted)
	}
	// TODO: Add stash entries
	return buffer.String()
}

func (g *git) init(props *properties, env environmentInfo) {
	g.props = props
	g.env = env
}

func (g *git) getGitStatus() {
	g.repo = &gitRepo{}
	output := g.getGitOutputForCommand("status", "--porcelain", "-b", "--ignore-submodules")
	splittedOutput := strings.Split(output, "\n")
	g.repo.working = g.parseGitStats(splittedOutput, true)
	g.repo.staging = g.parseGitStats(splittedOutput, false)
	branchInfo := g.parseGitBranchInfo(splittedOutput[0])
	if branchInfo["local"] != "" {
		g.repo.ahead, _ = strconv.Atoi(branchInfo["ahead"])
		g.repo.behind, _ = strconv.Atoi(branchInfo["behind"])
		g.repo.branch = branchInfo["local"]
		g.repo.upstream = branchInfo["upstream"]
	} else {
		g.repo.branch = g.getGitDetachedBranch()
	}
	g.repo.stashCount = g.getStashContext()
}

func (g *git) getGitOutputForCommand(args ...string) string {
	args = append([]string{"-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	return g.env.runCommand("git", args...)
}

func (g *git) getGitDetachedBranch() string {
	ref := g.getGitOutputForCommand("symbolic-ref", "--short", "HEAD")
	if ref == "" {
		return "unknown"
	}
	return ref
}

func (g *git) parseGitStats(output []string, working bool) *gitStatus {
	status := gitStatus{}
	if len(output) <= 1 {
		return &status
	}
	for _, line := range output[1:] {
		if len(line) < 2 {
			continue
		}
		code := line[0:1]
		if working {
			code = line[1:2]
		}
		switch code {
		case "?":
			status.untracked++
		case "D":
			status.deleted++
		case "A":
			status.added++
		case "U":
			status.unmerged++
		case "M", "R", "C":
			status.modified++
		}
	}
	return &status
}

func (g *git) getStashContext() int {
	stash := g.getGitOutputForCommand("stash", "list")
	return numberOfLinesInString(stash)
}

func (g *git) hasStaging() bool {
	return g.repo.staging.deleted > 0 || g.repo.staging.added > 0 || g.repo.staging.unmerged > 0 || g.repo.staging.modified > 0
}

func (g *git) hasWorking() bool {
	return g.repo.working.deleted > 0 || g.repo.working.added > 0 || g.repo.working.unmerged > 0 || g.repo.working.modified > 0 || g.repo.working.untracked > 0
}

func (g *git) parseGitBranchInfo(branchInfo string) map[string]string {
	var branchRegex = regexp.MustCompile(`^## (?P<local>\S+?)(\.{3}(?P<upstream>\S+?)( \[(ahead (?P<ahead>\d+)(, )?)?(behind (?P<behind>\d+))?])?)?$`)
	return groupDict(branchRegex, branchInfo)
}

func groupDict(pattern *regexp.Regexp, haystack string) map[string]string {
	match := pattern.FindStringSubmatch(haystack)
	result := make(map[string]string)
	if len(match) > 0 {
		for i, name := range pattern.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
	}
	return result
}

func numberOfLinesInString(s string) int {
	n := 0
	for _, r := range s {
		if r == '\n' {
			n++
		}
	}
	if len(s) > 0 && !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}
