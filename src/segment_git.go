package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

// Repo represents a git repository
type Repo struct {
	Working       *GitStatus
	Staging       *GitStatus
	Ahead         int
	Behind        int
	HEAD          string
	Upstream      string
	StashCount    int
	WorktreeCount int
	IsWorkTree    bool

	gitWorkingFolder string // .git working folder, can be different of root if using worktree
	gitRootFolder    string // .git root folder
}

// GitStatus represents part of the status of a git repository
type GitStatus struct {
	Unmerged int
	Deleted  int
	Added    int
	Modified int
	Changed  bool
}

func (s *GitStatus) parse(output []string, working bool) {
	if len(output) <= 1 {
		return
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
				s.Added++
			}
		case "D":
			s.Deleted++
		case "A":
			s.Added++
		case "U":
			s.Unmerged++
		case "M", "R", "C", "m":
			s.Modified++
		}
	}
	s.Changed = s.Added > 0 || s.Deleted > 0 || s.Modified > 0 || s.Unmerged > 0
}

func (s *GitStatus) String() string {
	var status string
	stringIfValue := func(value int, prefix string) string {
		if value > 0 {
			return fmt.Sprintf(" %s%d", prefix, value)
		}
		return ""
	}
	status += stringIfValue(s.Added, "+")
	status += stringIfValue(s.Modified, "~")
	status += stringIfValue(s.Deleted, "-")
	status += stringIfValue(s.Unmerged, "x")
	return status
}

type git struct {
	props *properties
	env   environmentInfo
	repo  *Repo
}

const (
	// DisplayBranchStatus show branch status or not
	DisplayBranchStatus Property = "display_branch_status"
	// DisplayStatus shows the status of the repository
	DisplayStatus Property = "display_status"
	// DisplayStashCount show stash count or not
	DisplayStashCount Property = "display_stash_count"
	// DisplayWorktreeCount show worktree count or not
	DisplayWorktreeCount Property = "display_worktree_count"
	// DisplayUpstreamIcon show or hide the upstream icon
	DisplayUpstreamIcon Property = "display_upstream_icon"

	// BranchIcon the icon to use as branch indicator
	BranchIcon Property = "branch_icon"
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
	// StashCountIcon shows before the stash context
	StashCountIcon Property = "stash_count_icon"
	// StatusSeparatorIcon shows between staging and working area
	StatusSeparatorIcon Property = "status_separator_icon"
	// MergeIcon shows before the merge context
	MergeIcon Property = "merge_icon"
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

	// Deprecated

	// DisplayStatusDetail shows the detailed status of the repository
	DisplayStatusDetail Property = "display_status_detail"
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
	// WorktreeCountIcon shows before the worktree context
	WorktreeCountIcon Property = "worktree_count_icon"
)

func (g *git) enabled() bool {
	if !g.env.hasCommand(g.getGitCommand()) {
		return false
	}
	gitdir, err := g.env.hasParentFilePath(".git")
	if err != nil {
		return false
	}
	if g.shouldIgnoreRootRepository(gitdir.parentFolder) {
		return false
	}

	g.repo = &Repo{
		Staging: &GitStatus{},
		Working: &GitStatus{},
	}
	if gitdir.isDir {
		g.repo.gitWorkingFolder = gitdir.path
		g.repo.gitRootFolder = gitdir.path
		return true
	}
	// handle worktree
	g.repo.gitRootFolder = gitdir.path
	dirPointer := g.env.getFileContent(gitdir.path)
	dirPointer = strings.Trim(dirPointer, " \r\n")
	matches := findNamedRegexMatch(`^gitdir: (?P<dir>.*)$`, dirPointer)
	if matches != nil && matches["dir"] != "" {
		g.repo.gitWorkingFolder = matches["dir"]
		// in worktrees, the path looks like this: gitdir: path/.git/worktrees/branch
		// strips the last .git/worktrees part
		// :ind+5 = index + /.git
		ind := strings.LastIndex(g.repo.gitWorkingFolder, "/.git/worktrees")
		g.repo.gitRootFolder = g.repo.gitWorkingFolder[:ind+5]
		g.repo.IsWorkTree = true
		return true
	}
	return false
}

func (g *git) shouldIgnoreRootRepository(rootDir string) bool {
	if g.props == nil || g.props.values == nil {
		return false
	}
	value, ok := g.props.values[ExcludeFolders]
	if !ok {
		return false
	}
	excludedFolders := parseStringArray(value)
	return dirMatchesOneOf(g.env, rootDir, excludedFolders)
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
	// use template if available
	segmentTemplate := g.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) > 0 {
		template := &textTemplate{
			Template: segmentTemplate,
			Context:  g.repo,
			Env:      g.env,
		}
		text, err := template.render()
		if err != nil {
			return err.Error()
		}
		return text
	}
	// legacy render string	if no template
	// remove this for 6.0
	return g.renderDeprecatedString(displayStatus)
}

func (g *git) renderDeprecatedString(displayStatus bool) string {
	if !displayStatus {
		return g.getPrettyHEADName()
	}
	buffer := new(bytes.Buffer)
	// remote (if available)
	if g.repo.Upstream != "" && g.props.getBool(DisplayUpstreamIcon, false) {
		fmt.Fprintf(buffer, "%s", g.getUpstreamSymbol())
	}
	// branchName
	fmt.Fprintf(buffer, "%s", g.repo.HEAD)
	if g.props.getBool(DisplayBranchStatus, true) {
		buffer.WriteString(g.getBranchStatus())
	}
	if g.repo.Staging.Changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.repo.Staging, StagingColor, LocalStagingIcon, " \uF046"))
	}
	if g.repo.Staging.Changed && g.repo.Working.Changed {
		fmt.Fprint(buffer, g.props.getString(StatusSeparatorIcon, " |"))
	}
	if g.repo.Working.Changed {
		fmt.Fprint(buffer, g.getStatusDetailString(g.repo.Working, WorkingColor, LocalWorkingIcon, " \uF044"))
	}
	if g.repo.StashCount != 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(StashCountIcon, "\uF692 "), g.repo.StashCount)
	}
	if g.repo.WorktreeCount != 0 {
		fmt.Fprintf(buffer, " %s%d", g.props.getString(WorktreeCountIcon, "\uf1bb "), g.repo.WorktreeCount)
	}
	return buffer.String()
}

func (g *git) init(props *properties, env environmentInfo) {
	g.props = props
	g.env = env
}

func (g *git) getBranchStatus() string {
	if g.repo.Ahead > 0 && g.repo.Behind > 0 {
		return fmt.Sprintf(" %s%d %s%d", g.props.getString(BranchAheadIcon, "\u2191"), g.repo.Ahead, g.props.getString(BranchBehindIcon, "\u2193"), g.repo.Behind)
	}
	if g.repo.Ahead > 0 {
		return fmt.Sprintf(" %s%d", g.props.getString(BranchAheadIcon, "\u2191"), g.repo.Ahead)
	}
	if g.repo.Behind > 0 {
		return fmt.Sprintf(" %s%d", g.props.getString(BranchBehindIcon, "\u2193"), g.repo.Behind)
	}
	if g.repo.Behind == 0 && g.repo.Ahead == 0 && g.repo.Upstream != "" {
		return fmt.Sprintf(" %s", g.props.getString(BranchIdenticalIcon, "\u2261"))
	}
	if g.repo.Upstream == "" {
		return fmt.Sprintf(" %s", g.props.getString(BranchGoneIcon, "\u2262"))
	}
	return ""
}

func (g *git) getStatusDetailString(status *GitStatus, color, icon Property, defaultIcon string) string {
	prefix := g.props.getString(icon, defaultIcon)
	foregroundColor := g.props.getColor(color, g.props.foreground)
	if !g.props.getBool(DisplayStatusDetail, true) {
		return g.colorStatusString(prefix, "", foregroundColor)
	}
	return g.colorStatusString(prefix, status.String(), foregroundColor)
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
	upstream := replaceAllString("/.*", g.repo.Upstream, "")
	url := g.getOriginURL(upstream)
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
	g.repo.Working.parse(splittedOutput, true)
	g.repo.Staging.parse(splittedOutput, false)
	status := g.parseGitStatusInfo(splittedOutput[0])
	if status["local"] != "" {
		g.repo.Ahead, _ = strconv.Atoi(status["ahead"])
		g.repo.Behind, _ = strconv.Atoi(status["behind"])
		if status["upstream_status"] != "gone" {
			g.repo.Upstream = status["upstream"]
		}
	}
	g.repo.HEAD = g.getGitHEADContext(status["local"])
	if g.props.getBool(DisplayStashCount, false) {
		g.repo.StashCount = g.getStashContext()
	}
	if g.props.getBool(DisplayWorktreeCount, false) {
		g.repo.WorktreeCount = g.getWorktreeContext()
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
	if g.repo.Staging.Changed || g.repo.Working.Changed {
		return g.props.getColor(LocalChangesColor, defaultValue)
	} else if g.repo.Ahead > 0 && g.repo.Behind > 0 {
		return g.props.getColor(AheadAndBehindColor, defaultValue)
	} else if g.repo.Ahead > 0 {
		return g.props.getColor(AheadColor, defaultValue)
	} else if g.repo.Behind > 0 {
		return g.props.getColor(BehindColor, defaultValue)
	}
	return defaultValue
}

func (g *git) getGitCommand() string {
	inWSLSharedDrive := func(env environmentInfo) bool {
		return env.isWsl() && strings.HasPrefix(env.getcwd(), "/mnt/")
	}
	gitCommand := "git"
	if g.env.getRuntimeGOOS() == windowsPlatform || inWSLSharedDrive(g.env) {
		gitCommand = "git.exe"
	}
	return gitCommand
}

func (g *git) getGitCommandOutput(args ...string) string {
	args = append([]string{"--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	val, _ := g.env.runCommand(g.getGitCommand(), args...)
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
	if g.env.hasFolder(g.repo.gitWorkingFolder + "/rebase-merge") {
		head := g.getGitFileContents(g.repo.gitWorkingFolder, "rebase-merge/head-name")
		origin := strings.Replace(head, "refs/heads/", "", 1)
		origin = g.truncateBranch(origin)
		onto := g.getGitRefFileSymbolicName("rebase-merge/onto")
		onto = g.truncateBranch(onto)
		step := g.getGitFileContents(g.repo.gitWorkingFolder, "rebase-merge/msgnum")
		total := g.getGitFileContents(g.repo.gitWorkingFolder, "rebase-merge/end")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		return fmt.Sprintf("%s%s%s onto %s%s (%s/%s) at %s", icon, branchIcon, origin, branchIcon, onto, step, total, ref)
	}
	if g.env.hasFolder(g.repo.gitWorkingFolder + "/rebase-apply") {
		head := g.getGitFileContents(g.repo.gitWorkingFolder, "rebase-apply/head-name")
		origin := strings.Replace(head, "refs/heads/", "", 1)
		origin = g.truncateBranch(origin)
		step := g.getGitFileContents(g.repo.gitWorkingFolder, "rebase-apply/next")
		total := g.getGitFileContents(g.repo.gitWorkingFolder, "rebase-apply/last")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		return fmt.Sprintf("%s%s%s (%s/%s) at %s", icon, branchIcon, origin, step, total, ref)
	}
	// merge
	if g.hasGitFile("MERGE_MSG") && g.hasGitFile("MERGE_HEAD") {
		icon := g.props.getString(MergeIcon, "\uE727 ")
		mergeContext := g.getGitFileContents(g.repo.gitWorkingFolder, "MERGE_MSG")
		matches := findNamedRegexMatch(`Merge (?P<type>(remote-tracking )?branch|commit|tag) '(?P<head>.*)' into`, mergeContext)

		if matches != nil && matches["head"] != "" {
			var headIcon string
			switch matches["type"] {
			case "tag":
				headIcon = g.props.getString(TagIcon, "\uF412")
			case "commit":
				headIcon = g.props.getString(CommitIcon, "\uF417")
			default:
				headIcon = branchIcon
			}
			head := g.truncateBranch(matches["head"])
			return fmt.Sprintf("%s%s%s into %s", icon, headIcon, head, ref)
		}
	}
	// sequencer status
	// see if a cherry-pick or revert is in progress, if the user has committed a
	// conflict resolution with 'git commit' in the middle of a sequence of picks or
	// reverts then CHERRY_PICK_HEAD/REVERT_HEAD will not exist so we have to read
	// the todo file.
	if g.hasGitFile("CHERRY_PICK_HEAD") {
		sha := g.getGitFileContents(g.repo.gitWorkingFolder, "CHERRY_PICK_HEAD")
		icon := g.props.getString(CherryPickIcon, "\uE29B ")
		return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
	} else if g.hasGitFile("REVERT_HEAD") {
		sha := g.getGitFileContents(g.repo.gitWorkingFolder, "REVERT_HEAD")
		icon := g.props.getString(RevertIcon, "\uF0E2 ")
		return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
	} else if g.hasGitFile("sequencer/todo") {
		todo := g.getGitFileContents(g.repo.gitWorkingFolder, "sequencer/todo")
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
	return g.env.hasFilesInDir(g.repo.gitWorkingFolder, file)
}

func (g *git) getGitFileContents(folder, file string) string {
	return strings.Trim(g.env.getFileContent(folder+"/"+file), " \r\n")
}

func (g *git) getGitRefFileSymbolicName(refFile string) string {
	ref := g.getGitFileContents(g.repo.gitWorkingFolder, refFile)
	return g.getGitCommandOutput("name-rev", "--name-only", "--exclude=tags/*", ref)
}

func (g *git) getPrettyHEADName() string {
	var ref string
	HEAD := g.getGitFileContents(g.repo.gitWorkingFolder, "HEAD")
	branchPrefix := "ref: refs/heads/"
	if strings.HasPrefix(HEAD, branchPrefix) {
		ref = strings.TrimPrefix(HEAD, branchPrefix)
	}
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

func (g *git) getStashContext() int {
	stashContent := g.getGitFileContents(g.repo.gitRootFolder, "logs/refs/stash")
	if stashContent == "" {
		return 0
	}
	lines := strings.Split(stashContent, "\n")
	return len(lines)
}

func (g *git) getWorktreeContext() int {
	if !g.env.hasFolder(g.repo.gitRootFolder + "/worktrees") {
		return 0
	}
	worktreeFolders := g.env.getFoldersList(g.repo.gitRootFolder + "/worktrees")
	return len(worktreeFolders)
}

func (g *git) parseGitStatusInfo(branchInfo string) map[string]string {
	var branchRegex = `^## (?P<local>\S+?)(\.{3}(?P<upstream>\S+?)( \[(?P<upstream_status>(ahead (?P<ahead>\d+)(, )?)?(behind (?P<behind>\d+))?(gone)?)])?)?$`
	return findNamedRegexMatch(branchRegex, branchInfo)
}

func (g *git) getOriginURL(upstream string) string {
	cfg, err := ini.Load(g.repo.gitRootFolder + "/config")
	if err != nil {
		return g.getGitCommandOutput("remote", "get-url", upstream)
	}
	keyVal := cfg.Section("remote \"" + upstream + "\"").Key("url").String()
	if keyVal == "" {
		return g.getGitCommandOutput("remote", "get-url", upstream)
	}
	return keyVal
}
