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
	HEAD       string
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
	//RebaseIcon shows before the rebase context
	RebaseIcon Property = "rebase_icon"
	//CherryPickIcon shows before the cherry-pick context
	CherryPickIcon Property = "cherry_pick_icon"
	//CommitIcon shows before the detached context
	CommitIcon Property = "commit_icon"
	//TagIcon shows before the tag context
	TagIcon Property = "tag_icon"
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
	// branchName
	fmt.Fprintf(buffer, "%s", g.repo.HEAD)
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
	output := g.getGitCommandOutput("status", "--porcelain", "-b", "--ignore-submodules")
	splittedOutput := strings.Split(output, "\n")
	g.repo.working = g.parseGitStats(splittedOutput, true)
	g.repo.staging = g.parseGitStats(splittedOutput, false)
	status := g.parseGitStatusInfo(splittedOutput[0])
	if status["local"] != "" {
		g.repo.ahead, _ = strconv.Atoi(status["ahead"])
		g.repo.behind, _ = strconv.Atoi(status["behind"])
		g.repo.upstream = status["upstream"]
	}
	g.repo.HEAD = g.getGitHEADContext(status["local"])
	g.repo.stashCount = g.getStashContext()
}

func (g *git) getGitCommandOutput(args ...string) string {
	args = append([]string{"-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	return g.env.runCommand("git", args...)
}

func (g *git) getGitHEADContext(ref string) string {
	branchIcon := g.props.getString(BranchIcon, "BRANCH:")
	if ref == "" {
		ref = g.getPrettyHEADName()
	} else {
		ref = fmt.Sprintf("%s%s", branchIcon, ref)
	}
	// rebase
	if g.env.hasFolder(".git/rebase-merge") {
		origin := g.getGitRefFileSymbolicName("rebase-merge/orig-head")
		onto := g.getGitRefFileSymbolicName("rebase-merge/onto")
		step := g.getGitFileContents("rebase-merge/msgnum")
		total := g.getGitFileContents("rebase-merge/end")
		icon := g.props.getString(RebaseIcon, "REBASE:")
		return fmt.Sprintf("%s%s%s onto %s%s (%s/%s) at %s", icon, branchIcon, origin, branchIcon, onto, step, total, ref)
	}
	if g.env.hasFolder(".git/rebase-apply") {
		head := g.getGitFileContents("rebase-apply/head-name")
		origin := strings.Replace(head, "refs/heads/", "", 1)
		step := g.getGitFileContents("rebase-apply/next")
		total := g.getGitFileContents("rebase-apply/last")
		icon := g.props.getString(RebaseIcon, "REBASING:")
		return fmt.Sprintf("%s%s%s (%s/%s) at %s", icon, branchIcon, origin, step, total, ref)
	}
	// cherry-pick
	if g.env.hasFiles(".git/CHERRY_PICK_HEAD") {
		sha := g.getGitRefFileSymbolicName("CHERRY_PICK_HEAD")
		icon := g.props.getString(CherryPickIcon, "CHERRY PICK:")
		return fmt.Sprintf("%s%s onto %s", icon, sha, ref)
	}
	return ref
}

func (g *git) getPrettyHEADName() string {
	// check for tag
	ref := g.getGitCommandOutput("describe", "--tags", "--exact-match")
	if ref != "" {
		return fmt.Sprintf("%s%s", g.props.getString(TagIcon, "TAG:"), ref)
	}
	// fallback to commit
	ref = g.getGitCommandOutput("rev-parse", "--short", "HEAD")
	return fmt.Sprintf("%s%s", g.props.getString(CommitIcon, "COMMIT:"), ref)
}

func (g *git) getGitFileContents(file string) string {
	content := g.env.getFileContent(fmt.Sprintf(".git/%s", file))
	return strings.Trim(content, " \r\n")
}

func (g *git) getGitRefFileSymbolicName(refFile string) string {
	ref := g.getGitFileContents(refFile)
	return g.getGitCommandOutput("name-rev", "--name-only", "--exclude=tags/*", ref)
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
	stash := g.getGitCommandOutput("stash", "list")
	return numberOfLinesInString(stash)
}

func (g *git) hasStaging() bool {
	return g.repo.staging.deleted > 0 || g.repo.staging.added > 0 || g.repo.staging.unmerged > 0 || g.repo.staging.modified > 0
}

func (g *git) hasWorking() bool {
	return g.repo.working.deleted > 0 || g.repo.working.added > 0 || g.repo.working.unmerged > 0 || g.repo.working.modified > 0 || g.repo.working.untracked > 0
}

func (g *git) parseGitStatusInfo(branchInfo string) map[string]string {
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
