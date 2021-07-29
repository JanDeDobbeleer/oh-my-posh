package main

import (
	"bytes"
	"fmt"
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
	gitFolder  string
}

type gitStatus struct {
	unmerged int
	deleted  int
	added    int
	modified int
	changed  bool
}

func (s *gitStatus) string() string {
	var status string
	stringIfValue := func(value int, prefix string) string {
		if value > 0 {
			return fmt.Sprintf(" %s%d", prefix, value)
		}
		return ""
	}
	status += stringIfValue(s.added, "+")
	status += stringIfValue(s.modified, "~")
	status += stringIfValue(s.deleted, "-")
	status += stringIfValue(s.unmerged, "x")
	return status
}

type git struct {
	props *properties
	env   environmentInfo
	repo  *gitRepo
}

const (
	// BranchIcon the icon to use as branch indicator
	BranchIcon Property = "branch_icon"
	// DisplayBranchStatus show branch status or not
	DisplayBranchStatus Property = "display_branch_status"
	// BranchIdenticalIcon the icon to display when the remote and local branch are identical
	BranchIdenticalIcon Property = "branch_identical_icon"
	// BranchAheadIcon the icon to display when the local branch is ahead of the remote
	BranchAheadIcon Property = "branch_ahead_icon"
	// BranchBehindIcon the icon to display when the local branch is behind the remote
	BranchBehindIcon Property = "branch_behind_icon"
	// BranchGoneIcon the icon to use when ther's no remote
	BranchGoneIcon Property = "branch_gone_icon"
	// LocalWorkingIcon the icon to use as the local working area changes indicator
	LocalWorkingIcon Property = "local_working_icon"
	// LocalStagingIcon the icon to use as the local staging area changes indicator
	LocalStagingIcon Property = "local_staged_icon"
	// DisplayStatus shows the status of the repository
	DisplayStatus Property = "display_status"
	// DisplayStatusDetail shows the detailed status of the repository
	DisplayStatusDetail Property = "display_status_detail"
	// RebaseIcon shows before the rebase context
	RebaseIcon Property = "rebase_icon"
	// CherryPickIcon shows before the cherry-pick context
	CherryPickIcon Property = "cherry_pick_icon"
	// RevertIcon shows before the revert context
	RevertIcon Property = "revert_icon"
	// CommitIcon shows before the detached context
	CommitIcon Property = "commit_icon"
	// NoCommitsIcon shows when there are no commits in the repo yet
	NoCommitsIcon Property = "no_commits_icon"
	// TagIcon shows before the tag context
	TagIcon Property = "tag_icon"
	// DisplayStashCount show stash count or not
	DisplayStashCount Property = "display_stash_count"
	// StashCountIcon shows before the stash context
	StashCountIcon Property = "stash_count_icon"
	// StatusSeparatorIcon shows between staging and working area
	StatusSeparatorIcon Property = "status_separator_icon"
	// MergeIcon shows before the merge context
	MergeIcon Property = "merge_icon"
	// DisplayUpstreamIcon show or hide the upstream icon
	DisplayUpstreamIcon Property = "display_upstream_icon"
	// GithubIcon shows√ when upstream is github
	GithubIcon Property = "github_icon"
	// BitbucketIcon shows  when upstream is bitbucket
	BitbucketIcon Property = "bitbucket_icon"
	// AzureDevOpsIcon shows  when upstream is azure devops
	AzureDevOpsIcon Property = "azure_devops_icon"
	// GitlabIcon shows when upstream is gitlab
	GitlabIcon Property = "gitlab_icon"
	// GitIcon shows when the upstream can't be identified
	GitIcon Property = "git_icon"
	// WorkingColor if set, the color to use on the working area
	WorkingColor Property = "working_color"
	// StagingColor if set, the color to use on the staging area
	StagingColor Property = "staging_color"
	// StatusColorsEnabled enables status colors
	StatusColorsEnabled Property = "status_colors_enabled"
	// LocalChangesColor if set, the color to use when there are local changes
	LocalChangesColor Property = "local_changes_color"
	// AheadAndBehindColor if set, the color to use when the branch is ahead and behind the remote
	AheadAndBehindColor Property = "ahead_and_behind_color"
	// BehindColor if set, the color to use when the branch is ahead and behind the remote
	BehindColor Property = "behind_color"
	// AheadColor if set, the color to use when the branch is ahead and behind the remote
	AheadColor Property = "ahead_color"
	// BranchMaxLength truncates the length of the branch name
	BranchMaxLength Property = "branch_max_length"
)

func (g *git) enabled() bool {
	if !g.env.hasCommand("git") {
		return false
	}
	gitdir, err := g.env.hasParentFilePath(".git")
	if err != nil {
		return false
	}
	g.repo = &gitRepo{}
	if gitdir.isDir {
		g.repo.gitFolder = gitdir.path
		return true
	}
	// handle worktree
	dirPointer := g.env.getFileContent(gitdir.path)
	dirPointer = strings.Trim(dirPointer, " \r\n")
	matches := findNamedRegexMatch(`^gitdir: (?P<dir>.*)$`, dirPointer)
	if matches != nil && matches["dir"] != "" {
		g.repo.gitFolder = matches["dir"]
		return true
	}
	return false
}

func (g *git) string() string {
	statusColorsEnabled := g.props.getBool(StatusColorsEnabled, false)
	displayStatus := g.props.getBool(DisplayStatus, false)

	if displayStatus || statusColorsEnabled {
		g.setGitStatus()
	}
	if statusColorsEnabled {
		g.SetStatusColor()
	}
	if !displayStatus {
		return g.getPrettyHEADName()
	}
	buffer := new(bytes.Buffer)
	// remote (if available)
	if g.repo.upstream != "" && g.props.getBool(DisplayUpstreamIcon, false) {
		fmt.Fprintf(buffer, "%s", g.getUpstreamSymbol())
	}
	// branchName
	fmt.Fprintf(buffer, "%s", g.repo.HEAD)
	if g.props.getBool(DisplayBranchStatus, true) {
		buffer.WriteString(g.getBranchStatus())
	}
	if g.repo.staging.changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.repo.staging, StagingColor, LocalStagingIcon, " \uF046"))
	}
	if g.repo.staging.changed && g.repo.working.changed {
		fmt.Fprint(buffer, g.props.getString(StatusSeparatorIcon, " |"))
	}
	if g.repo.working.changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.repo.working, WorkingColor, LocalWorkingIcon, " \uF044"))
	}
	if g.repo.stashCount != 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(StashCountIcon, "\uF692 "), g.repo.stashCount)
	}
	return buffer.String()
}

func (g *git) init(props *properties, env environmentInfo) {
	g.props = props
	g.env = env
}

func (g *git) getBranchStatus() string {
	if g.repo.ahead > 0 && g.repo.behind > 0 {
		return fmt.Sprintf(" %s%d %s%d", g.props.getString(BranchAheadIcon, "\u2191"), g.repo.ahead, g.props.getString(BranchBehindIcon, "\u2193"), g.repo.behind)
	}
	if g.repo.ahead > 0 {
		return fmt.Sprintf(" %s%d", g.props.getString(BranchAheadIcon, "\u2191"), g.repo.ahead)
	}
	if g.repo.behind > 0 {
		return fmt.Sprintf(" %s%d", g.props.getString(BranchBehindIcon, "\u2193"), g.repo.behind)
	}
	if g.repo.behind == 0 && g.repo.ahead == 0 && g.repo.upstream != "" {
		return fmt.Sprintf(" %s", g.props.getString(BranchIdenticalIcon, "\u2261"))
	}
	if g.repo.upstream == "" {
		return fmt.Sprintf(" %s", g.props.getString(BranchGoneIcon, "\u2262"))
	}
	return ""
}

func (g *git) getStatusDetailString(status *gitStatus, color, icon Property, defaultIcon string) string {
	prefix := g.props.getString(icon, defaultIcon)
	foregroundColor := g.props.getColor(color, g.props.foreground)
	if !g.props.getBool(DisplayStatusDetail, true) {
		return g.colorStatusString(prefix, "", foregroundColor)
	}
	return g.colorStatusString(prefix, status.string(), foregroundColor)
}

func (g *git) colorStatusString(prefix, status, color string) string {
	if color == g.props.foreground {
		return fmt.Sprintf("%s%s", prefix, status)
	}
	if strings.Contains(prefix, "</>") {
		return fmt.Sprintf("%s<%s>%s</>", prefix, color, status)
	}
	return fmt.Sprintf("<%s>%s%s</>", color, prefix, status)
}

func (g *git) getUpstreamSymbol() string {
	upstream := replaceAllString("/.*", g.repo.upstream, "")
	url := g.getGitCommandOutput("remote", "get-url", upstream)
	if strings.Contains(url, "github") {
		return g.props.getString(GithubIcon, "\uF408 ")
	}
	if strings.Contains(url, "gitlab") {
		return g.props.getString(GitlabIcon, "\uF296 ")
	}
	if strings.Contains(url, "bitbucket") {
		return g.props.getString(BitbucketIcon, "\uF171 ")
	}
	if strings.Contains(url, "dev.azure.com") || strings.Contains(url, "visualstudio.com") {
		return g.props.getString(AzureDevOpsIcon, "\uFD03 ")
	}
	return g.props.getString(GitIcon, "\uE5FB ")
}

func (g *git) setGitStatus() {
	output := g.getGitCommandOutput("status", "-unormal", "--short", "--branch")
	splittedOutput := strings.Split(output, "\n")
	g.repo.working = g.parseGitStats(splittedOutput, true)
	g.repo.staging = g.parseGitStats(splittedOutput, false)
	status := g.parseGitStatusInfo(splittedOutput[0])
	if status["local"] != "" {
		g.repo.ahead, _ = strconv.Atoi(status["ahead"])
		g.repo.behind, _ = strconv.Atoi(status["behind"])
		if status["upstream_status"] != "gone" {
			g.repo.upstream = status["upstream"]
		}
	}
	g.repo.HEAD = g.getGitHEADContext(status["local"])
	if g.props.getBool(DisplayStashCount, false) {
		g.repo.stashCount = g.getStashContext()
	}
}

func (g *git) SetStatusColor() {
	if g.props.getBool(ColorBackground, true) {
		g.props.background = g.getStatusColor(g.props.background)
	} else {
		g.props.foreground = g.getStatusColor(g.props.foreground)
	}
}

func (g *git) getStatusColor(defaultValue string) string {
	if g.repo.staging.changed || g.repo.working.changed {
		return g.props.getColor(LocalChangesColor, defaultValue)
	} else if g.repo.ahead > 0 && g.repo.behind > 0 {
		return g.props.getColor(AheadAndBehindColor, defaultValue)
	} else if g.repo.ahead > 0 {
		return g.props.getColor(AheadColor, defaultValue)
	} else if g.repo.behind > 0 {
		return g.props.getColor(BehindColor, defaultValue)
	}
	return defaultValue
}

func (g *git) getGitCommandOutput(args ...string) string {
	inWSLSharedDrive := func(env environmentInfo) bool {
		return env.isWsl() && strings.HasPrefix(env.getcwd(), "/mnt/")
	}
	gitCommand := "git"
	if g.env.getRuntimeGOOS() == windowsPlatform || inWSLSharedDrive(g.env) {
		gitCommand = "git.exe"
	}
	args = append([]string{"--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	val, _ := g.env.runCommand(gitCommand, args...)
	return val
}

func (g *git) getGitHEADContext(ref string) string {
	branchIcon := g.props.getString(BranchIcon, "\uE0A0")
	if ref == "" {
		ref = g.getPrettyHEADName()
	} else {
		ref = g.truncateBranch(ref)
		ref = fmt.Sprintf("%s%s", branchIcon, ref)
	}
	// rebase
	if g.hasGitFolder("rebase-merge") {
		head := g.getGitFileContents("rebase-merge/head-name")
		origin := strings.Replace(head, "refs/heads/", "", 1)
		origin = g.truncateBranch(origin)
		onto := g.getGitRefFileSymbolicName("rebase-merge/onto")
		onto = g.truncateBranch(onto)
		step := g.getGitFileContents("rebase-merge/msgnum")
		total := g.getGitFileContents("rebase-merge/end")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		return fmt.Sprintf("%s%s%s onto %s%s (%s/%s) at %s", icon, branchIcon, origin, branchIcon, onto, step, total, ref)
	}
	if g.hasGitFolder("rebase-apply") {
		head := g.getGitFileContents("rebase-apply/head-name")
		origin := strings.Replace(head, "refs/heads/", "", 1)
		origin = g.truncateBranch(origin)
		step := g.getGitFileContents("rebase-apply/next")
		total := g.getGitFileContents("rebase-apply/last")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		return fmt.Sprintf("%s%s%s (%s/%s) at %s", icon, branchIcon, origin, step, total, ref)
	}
	// merge
	if g.hasGitFile("MERGE_MSG") && g.hasGitFile("MERGE_HEAD") {
		icon := g.props.getString(MergeIcon, "\uE727 ")
		mergeContext := g.getGitFileContents("MERGE_MSG")
		matches := findNamedRegexMatch(`Merge branch '(?P<head>.*)' into`, mergeContext)
		if matches != nil && matches["head"] != "" {
			branch := g.truncateBranch(matches["head"])
			return fmt.Sprintf("%s%s%s into %s", icon, branchIcon, branch, ref)
		}
	}
	// sequencer status
	// see if a cherry-pick or revert is in progress, if the user has committed a
	// conflict resolution with 'git commit' in the middle of a sequence of picks or
	// reverts then CHERRY_PICK_HEAD/REVERT_HEAD will not exist so we have to read
	// the todo file.
	if g.hasGitFile("CHERRY_PICK_HEAD") {
		sha := g.getGitFileContents("CHERRY_PICK_HEAD")
		icon := g.props.getString(CherryPickIcon, "\uE29B ")
		return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
	} else if g.hasGitFile("REVERT_HEAD") {
		sha := g.getGitFileContents("REVERT_HEAD")
		icon := g.props.getString(RevertIcon, "\uF0E2 ")
		return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
	} else if g.hasGitFile("sequencer/todo") {
		todo := g.getGitFileContents("sequencer/todo")
		matches := findNamedRegexMatch(`^(?P<action>p|pick|revert)\s+(?P<sha>\S+)`, todo)
		if matches != nil && matches["sha"] != "" {
			action := matches["action"]
			sha := matches["sha"]
			switch action {
			case "p", "pick":
				icon := g.props.getString(CherryPickIcon, "\uE29B ")
				return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
			case "revert":
				icon := g.props.getString(RevertIcon, "\uF0E2 ")
				return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
			}
		}
	}
	return ref
}

func (g *git) truncateBranch(branch string) string {
	maxLength := g.props.getInt(BranchMaxLength, 0)
	if maxLength == 0 || len(branch) < maxLength {
		return branch
	}
	return branch[0:maxLength]
}

func (g *git) hasGitFile(file string) bool {
	return g.env.hasFilesInDir(g.repo.gitFolder, file)
}

func (g *git) hasGitFolder(folder string) bool {
	path := g.repo.gitFolder + "/" + folder
	return g.env.hasFolder(path)
}

func (g *git) getGitFileContents(file string) string {
	path := g.repo.gitFolder + "/" + file
	content := g.env.getFileContent(path)
	return strings.Trim(content, " \r\n")
}

func (g *git) getGitRefFileSymbolicName(refFile string) string {
	ref := g.getGitFileContents(refFile)
	return g.getGitCommandOutput("name-rev", "--name-only", "--exclude=tags/*", ref)
}

func (g *git) getPrettyHEADName() string {
	ref := g.getGitCommandOutput("branch", "--show-current")
	if ref != "" {
		ref = g.truncateBranch(ref)
		return fmt.Sprintf("%s%s", g.props.getString(BranchIcon, "\uE0A0"), ref)
	}
	// check for tag
	ref = g.getGitCommandOutput("describe", "--tags", "--exact-match")
	if ref != "" {
		return fmt.Sprintf("%s%s", g.props.getString(TagIcon, "\uF412"), ref)
	}
	// fallback to commit
	ref = g.getGitCommandOutput("rev-parse", "--short", "HEAD")
	if ref == "" {
		return g.props.getString(NoCommitsIcon, "\uF594 ")
	}
	return fmt.Sprintf("%s%s", g.props.getString(CommitIcon, "\uF417"), ref)
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
			if working {
				status.added++
			}
		case "D":
			status.deleted++
		case "A":
			status.added++
		case "U":
			status.unmerged++
		case "M", "R", "C", "m":
			status.modified++
		}
	}
	status.changed = status.added > 0 || status.deleted > 0 || status.modified > 0 || status.unmerged > 0
	return &status
}

func (g *git) getStashContext() int {
	stashContent := g.getGitFileContents("logs/refs/stash")
	if stashContent == "" {
		return 0
	}
	lines := strings.Split(stashContent, "\n")
	return len(lines)
}

func (g *git) parseGitStatusInfo(branchInfo string) map[string]string {
	var branchRegex = `^## (?P<local>\S+?)(\.{3}(?P<upstream>\S+?)( \[(?P<upstream_status>(ahead (?P<ahead>\d+)(, )?)?(behind (?P<behind>\d+))?(gone)?)])?)?$`
	return findNamedRegexMatch(branchRegex, branchInfo)
}
