package main

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

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
	return strings.TrimSpace(status)
}

type git struct {
	props properties
	env   environmentInfo

	Working       *GitStatus
	Staging       *GitStatus
	Ahead         int
	Behind        int
	HEAD          string
	BranchStatus  string
	Upstream      string
	UpstreamIcon  string
	StashCount    int
	WorktreeCount int
	IsWorkTree    bool

	gitWorkingFolder  string // .git working folder, can be different of root if using worktree
	gitRootFolder     string // .git root folder
	gitWorktreeFolder string // .git real worktree path

	gitCommand string
}

const (
	// FetchStatus fetches the status of the repository
	FetchStatus Property = "fetch_status"
	// FetchStashCount fetches the stash count
	FetchStashCount Property = "fetch_stash_count"
	// FetchWorktreeCount fetches the worktree count
	FetchWorktreeCount Property = "fetch_worktree_count"
	// FetchUpstreamIcon fetches the upstream icon
	FetchUpstreamIcon Property = "fetch_upstream_icon"

	// BranchMaxLength truncates the length of the branch name
	BranchMaxLength Property = "branch_max_length"
	// TruncateSymbol appends the set symbol to a truncated branch name
	TruncateSymbol Property = "truncate_symbol"
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

	g.Staging = &GitStatus{}
	g.Working = &GitStatus{}

	if gitdir.isDir {
		g.gitWorkingFolder = gitdir.path
		g.gitRootFolder = gitdir.path
		return true
	}
	// handle worktree
	g.gitRootFolder = gitdir.path
	dirPointer := strings.Trim(g.env.getFileContent(gitdir.path), " \r\n")
	matches := findNamedRegexMatch(`^gitdir: (?P<dir>.*)$`, dirPointer)
	if matches != nil && matches["dir"] != "" {
		g.gitWorkingFolder = matches["dir"]
		// in worktrees, the path looks like this: gitdir: path/.git/worktrees/branch
		// strips the last .git/worktrees part
		// :ind+5 = index + /.git
		ind := strings.LastIndex(g.gitWorkingFolder, "/.git/worktrees")
		g.gitRootFolder = g.gitWorkingFolder[:ind+5]
		g.gitWorktreeFolder = strings.TrimSuffix(g.env.getFileContent(g.gitWorkingFolder+"/gitdir"), ".git\n")
		g.IsWorkTree = true
		return true
	}
	return false
}

func (g *git) shouldIgnoreRootRepository(rootDir string) bool {
	value, ok := g.props[ExcludeFolders]
	if !ok {
		return false
	}
	excludedFolders := parseStringArray(value)
	return dirMatchesOneOf(g.env, rootDir, excludedFolders)
}

func (g *git) string() string {
	statusColorsEnabled := g.props.getBool(StatusColorsEnabled, false)
	displayStatus := g.props.getOneOfBool(FetchStatus, DisplayStatus, false)
	if !displayStatus {
		g.HEAD = g.getPrettyHEADName()
	}
	if displayStatus || statusColorsEnabled {
		g.setGitStatus()
	}
	if g.Upstream != "" && g.props.getOneOfBool(FetchUpstreamIcon, DisplayUpstreamIcon, false) {
		g.UpstreamIcon = g.getUpstreamIcon()
	}
	if g.props.getOneOfBool(FetchStashCount, DisplayStashCount, false) {
		g.StashCount = g.getStashContext()
	}
	if g.props.getOneOfBool(FetchWorktreeCount, DisplayWorktreeCount, false) {
		g.WorktreeCount = g.getWorktreeContext()
	}
	// use template if available
	segmentTemplate := g.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) > 0 {
		return g.templateString(segmentTemplate)
	}
	// legacy render string	if no template
	// remove this for 6.0
	return g.deprecatedString(statusColorsEnabled)
}

func (g *git) templateString(segmentTemplate string) string {
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  g,
		Env:      g.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (g *git) init(props properties, env environmentInfo) {
	g.props = props
	g.env = env
}

func (g *git) getBranchStatus() string {
	if g.Ahead > 0 && g.Behind > 0 {
		return fmt.Sprintf(" %s%d %s%d", g.props.getString(BranchAheadIcon, "\u2191"), g.Ahead, g.props.getString(BranchBehindIcon, "\u2193"), g.Behind)
	}
	if g.Ahead > 0 {
		return fmt.Sprintf(" %s%d", g.props.getString(BranchAheadIcon, "\u2191"), g.Ahead)
	}
	if g.Behind > 0 {
		return fmt.Sprintf(" %s%d", g.props.getString(BranchBehindIcon, "\u2193"), g.Behind)
	}
	if g.Behind == 0 && g.Ahead == 0 && g.Upstream != "" {
		return fmt.Sprintf(" %s", g.props.getString(BranchIdenticalIcon, "\u2261"))
	}
	if g.Upstream == "" {
		return fmt.Sprintf(" %s", g.props.getString(BranchGoneIcon, "\u2262"))
	}
	return ""
}

func (g *git) getUpstreamIcon() string {
	upstream := replaceAllString("/.*", g.Upstream, "")
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
	g.Working.parse(splittedOutput, true)
	g.Staging.parse(splittedOutput, false)
	status := g.parseGitStatusInfo(splittedOutput[0])
	if status["local"] != "" {
		g.Ahead, _ = strconv.Atoi(status["ahead"])
		g.Behind, _ = strconv.Atoi(status["behind"])
		if status["upstream_status"] != "gone" {
			g.Upstream = status["upstream"]
		}
	}
	g.HEAD = g.getGitHEADContext(status["local"])
	g.BranchStatus = g.getBranchStatus()
}

func (g *git) getGitCommand() string {
	if len(g.gitCommand) > 0 {
		return g.gitCommand
	}
	inWSL2SharedDrive := func(env environmentInfo) bool {
		if !env.isWsl() {
			return false
		}
		if !strings.HasPrefix(env.getcwd(), "/mnt/") {
			return false
		}
		uname, _ := g.env.runCommand("uname", "-r")
		return strings.Contains(uname, "WSL2")
	}
	g.gitCommand = "git"
	if g.env.getRuntimeGOOS() == windowsPlatform || inWSL2SharedDrive(g.env) {
		g.gitCommand = "git.exe"
	}
	return g.gitCommand
}

func (g *git) getGitCommandOutput(args ...string) string {
	args = append([]string{"-C", g.gitWorktreeFolder, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
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
	if g.env.hasFolder(g.gitWorkingFolder + "/rebase-merge") {
		head := g.getGitFileContents(g.gitWorkingFolder, "rebase-merge/head-name")
		origin := strings.Replace(head, "refs/heads/", "", 1)
		origin = g.truncateBranch(origin)
		onto := g.getGitRefFileSymbolicName("rebase-merge/onto")
		onto = g.truncateBranch(onto)
		step := g.getGitFileContents(g.gitWorkingFolder, "rebase-merge/msgnum")
		total := g.getGitFileContents(g.gitWorkingFolder, "rebase-merge/end")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		return fmt.Sprintf("%s%s%s onto %s%s (%s/%s) at %s", icon, branchIcon, origin, branchIcon, onto, step, total, ref)
	}
	if g.env.hasFolder(g.gitWorkingFolder + "/rebase-apply") {
		head := g.getGitFileContents(g.gitWorkingFolder, "rebase-apply/head-name")
		origin := strings.Replace(head, "refs/heads/", "", 1)
		origin = g.truncateBranch(origin)
		step := g.getGitFileContents(g.gitWorkingFolder, "rebase-apply/next")
		total := g.getGitFileContents(g.gitWorkingFolder, "rebase-apply/last")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		return fmt.Sprintf("%s%s%s (%s/%s) at %s", icon, branchIcon, origin, step, total, ref)
	}
	// merge
	if g.hasGitFile("MERGE_MSG") && g.hasGitFile("MERGE_HEAD") {
		icon := g.props.getString(MergeIcon, "\uE727 ")
		mergeContext := g.getGitFileContents(g.gitWorkingFolder, "MERGE_MSG")
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
		sha := g.getGitFileContents(g.gitWorkingFolder, "CHERRY_PICK_HEAD")
		icon := g.props.getString(CherryPickIcon, "\uE29B ")
		return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
	} else if g.hasGitFile("REVERT_HEAD") {
		sha := g.getGitFileContents(g.gitWorkingFolder, "REVERT_HEAD")
		icon := g.props.getString(RevertIcon, "\uF0E2 ")
		return fmt.Sprintf("%s%s onto %s", icon, sha[0:6], ref)
	} else if g.hasGitFile("sequencer/todo") {
		todo := g.getGitFileContents(g.gitWorkingFolder, "sequencer/todo")
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
	symbol := g.props.getString(TruncateSymbol, "")
	return branch[0:maxLength] + symbol
}

func (g *git) hasGitFile(file string) bool {
	return g.env.hasFilesInDir(g.gitWorkingFolder, file)
}

func (g *git) getGitFileContents(folder, file string) string {
	return strings.Trim(g.env.getFileContent(folder+"/"+file), " \r\n")
}

func (g *git) getGitRefFileSymbolicName(refFile string) string {
	ref := g.getGitFileContents(g.gitWorkingFolder, refFile)
	return g.getGitCommandOutput("name-rev", "--name-only", "--exclude=tags/*", ref)
}

func (g *git) getPrettyHEADName() string {
	var ref string
	HEAD := g.getGitFileContents(g.gitWorkingFolder, "HEAD")
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
	stashContent := g.getGitFileContents(g.gitRootFolder, "logs/refs/stash")
	if stashContent == "" {
		return 0
	}
	lines := strings.Split(stashContent, "\n")
	return len(lines)
}

func (g *git) getWorktreeContext() int {
	if !g.env.hasFolder(g.gitRootFolder + "/worktrees") {
		return 0
	}
	worktreeFolders := g.env.getFoldersList(g.gitRootFolder + "/worktrees")
	return len(worktreeFolders)
}

func (g *git) parseGitStatusInfo(branchInfo string) map[string]string {
	var branchRegex = `^## (?P<local>\S+?)(\.{3}(?P<upstream>\S+?)( \[(?P<upstream_status>(ahead (?P<ahead>\d+)(, )?)?(behind (?P<behind>\d+))?(gone)?)])?)?$`
	return findNamedRegexMatch(branchRegex, branchInfo)
}

func (g *git) getOriginURL(upstream string) string {
	cfg, err := ini.Load(g.gitRootFolder + "/config")
	if err != nil {
		return g.getGitCommandOutput("remote", "get-url", upstream)
	}
	keyVal := cfg.Section("remote \"" + upstream + "\"").Key("url").String()
	if keyVal == "" {
		return g.getGitCommandOutput("remote", "get-url", upstream)
	}
	return keyVal
}
