package main

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

// GitStatus represents part of the status of a git repository
type GitStatus struct {
	ScmStatus
}

func (s *GitStatus) add(code string) {
	switch code {
	case ".":
		return
	case "D":
		s.Deleted++
	case "A", "?":
		s.Added++
	case "U":
		s.Unmerged++
	case "M", "R", "C", "m":
		s.Modified++
	}
}

type git struct {
	scm

	Working       *GitStatus
	Staging       *GitStatus
	Ahead         int
	Behind        int
	HEAD          string
	Ref           string
	Hash          string
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

	DETACHED     = "(detached)"
	BRANCHPREFIX = "ref: refs/heads/"
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

func (g *git) string() string {
	statusColorsEnabled := g.props.getBool(StatusColorsEnabled, false)
	displayStatus := g.props.getOneOfBool(FetchStatus, DisplayStatus, false)
	if !displayStatus {
		g.setPrettyHEADName()
	}
	if displayStatus || statusColorsEnabled {
		g.setGitStatus()
		g.setGitHEADContext()
		g.setBranchStatus()
	} else {
		g.Working = &GitStatus{}
		g.Staging = &GitStatus{}
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

func (g *git) setBranchStatus() {
	getBranchStatus := func() string {
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
	g.BranchStatus = getBranchStatus()
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
	addToStatus := func(status string) {
		if len(status) <= 4 {
			return
		}
		const UNTRACKED = "?"
		if strings.HasPrefix(status, UNTRACKED) {
			g.Working.add(UNTRACKED)
			return
		}
		workingCode := status[3:4]
		stagingCode := status[2:3]
		g.Working.add(workingCode)
		g.Staging.add(stagingCode)
	}
	const (
		HASH         = "# branch.oid "
		REF          = "# branch.head "
		UPSTREAM     = "# branch.upstream "
		BRANCHSTATUS = "# branch.ab "
	)
	g.Working = &GitStatus{}
	g.Staging = &GitStatus{}
	output := g.getGitCommandOutput("status", "-unormal", "--branch", "--porcelain=2")
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, HASH) {
			g.Hash = line[len(HASH) : len(HASH)+7]
			continue
		}
		if strings.HasPrefix(line, REF) {
			g.Ref = line[len(REF):]
			continue
		}
		if strings.HasPrefix(line, UPSTREAM) {
			g.Upstream = line[len(UPSTREAM):]
			continue
		}
		if strings.HasPrefix(line, BRANCHSTATUS) {
			status := line[len(BRANCHSTATUS):]
			splitted := strings.Split(status, " ")
			g.Ahead, _ = strconv.Atoi(splitted[0])
			behind, _ := strconv.Atoi(splitted[1])
			g.Behind = -behind
			continue
		}
		addToStatus(line)
	}
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

func (g *git) setGitHEADContext() {
	branchIcon := g.props.getString(BranchIcon, "\uE0A0")
	if g.Ref == DETACHED {
		g.setPrettyHEADName()
	} else {
		head := g.formatHEAD(g.Ref)
		g.HEAD = fmt.Sprintf("%s%s", branchIcon, head)
	}

	formatDetached := func() string {
		if g.Ref == DETACHED {
			return fmt.Sprintf("%sdetached at %s", branchIcon, g.HEAD)
		}
		return g.HEAD
	}

	getPrettyNameOrigin := func(file string) string {
		var origin string
		head := g.getFileContents(g.gitWorkingFolder, file)
		if head == "detached HEAD" {
			origin = formatDetached()
		} else {
			head = strings.Replace(head, "refs/heads/", "", 1)
			origin = branchIcon + g.formatHEAD(head)
		}
		return origin
	}

	if g.env.hasFolder(g.gitWorkingFolder + "/rebase-merge") {
		origin := getPrettyNameOrigin("rebase-merge/head-name")
		onto := g.getGitRefFileSymbolicName("rebase-merge/onto")
		onto = g.formatHEAD(onto)
		step := g.getFileContents(g.gitWorkingFolder, "rebase-merge/msgnum")
		total := g.getFileContents(g.gitWorkingFolder, "rebase-merge/end")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		g.HEAD = fmt.Sprintf("%s%s onto %s%s (%s/%s) at %s", icon, origin, branchIcon, onto, step, total, g.HEAD)
		return
	}
	if g.env.hasFolder(g.gitWorkingFolder + "/rebase-apply") {
		origin := getPrettyNameOrigin("rebase-apply/head-name")
		step := g.getFileContents(g.gitWorkingFolder, "rebase-apply/next")
		total := g.getFileContents(g.gitWorkingFolder, "rebase-apply/last")
		icon := g.props.getString(RebaseIcon, "\uE728 ")
		g.HEAD = fmt.Sprintf("%s%s (%s/%s) at %s", icon, origin, step, total, g.HEAD)
		return
	}
	// merge
	commitIcon := g.props.getString(CommitIcon, "\uF417")
	if g.hasGitFile("MERGE_MSG") {
		icon := g.props.getString(MergeIcon, "\uE727 ")
		mergeContext := g.getFileContents(g.gitWorkingFolder, "MERGE_MSG")
		matches := findNamedRegexMatch(`Merge (remote-tracking )?(?P<type>branch|commit|tag) '(?P<theirs>.*)'`, mergeContext)
		// head := g.getGitRefFileSymbolicName("ORIG_HEAD")
		if matches != nil && matches["theirs"] != "" {
			var headIcon, theirs string
			switch matches["type"] {
			case "tag":
				headIcon = g.props.getString(TagIcon, "\uF412")
				theirs = matches["theirs"]
			case "commit":
				headIcon = commitIcon
				theirs = g.formatSHA(matches["theirs"])
			default:
				headIcon = branchIcon
				theirs = g.formatHEAD(matches["theirs"])
			}
			g.HEAD = fmt.Sprintf("%s%s%s into %s", icon, headIcon, theirs, formatDetached())
			return
		}
	}
	// sequencer status
	// see if a cherry-pick or revert is in progress, if the user has committed a
	// conflict resolution with 'git commit' in the middle of a sequence of picks or
	// reverts then CHERRY_PICK_HEAD/REVERT_HEAD will not exist so we have to read
	// the todo file.
	if g.hasGitFile("CHERRY_PICK_HEAD") {
		sha := g.getFileContents(g.gitWorkingFolder, "CHERRY_PICK_HEAD")
		cherry := g.props.getString(CherryPickIcon, "\uE29B ")
		g.HEAD = fmt.Sprintf("%s%s%s onto %s", cherry, commitIcon, g.formatSHA(sha), formatDetached())
		return
	}
	if g.hasGitFile("REVERT_HEAD") {
		sha := g.getFileContents(g.gitWorkingFolder, "REVERT_HEAD")
		revert := g.props.getString(RevertIcon, "\uF0E2 ")
		g.HEAD = fmt.Sprintf("%s%s%s onto %s", revert, commitIcon, g.formatSHA(sha), formatDetached())
		return
	}
	if g.hasGitFile("sequencer/todo") {
		todo := g.getFileContents(g.gitWorkingFolder, "sequencer/todo")
		matches := findNamedRegexMatch(`^(?P<action>p|pick|revert)\s+(?P<sha>\S+)`, todo)
		if matches != nil && matches["sha"] != "" {
			action := matches["action"]
			sha := matches["sha"]
			switch action {
			case "p", "pick":
				cherry := g.props.getString(CherryPickIcon, "\uE29B ")
				g.HEAD = fmt.Sprintf("%s%s%s onto %s", cherry, commitIcon, g.formatSHA(sha), formatDetached())
				return
			case "revert":
				revert := g.props.getString(RevertIcon, "\uF0E2 ")
				g.HEAD = fmt.Sprintf("%s%s%s onto %s", revert, commitIcon, g.formatSHA(sha), formatDetached())
				return
			}
		}
	}
	g.HEAD = formatDetached()
}

func (g *git) formatHEAD(head string) string {
	maxLength := g.props.getInt(BranchMaxLength, 0)
	if maxLength == 0 || len(head) < maxLength {
		return head
	}
	symbol := g.props.getString(TruncateSymbol, "")
	return head[0:maxLength] + symbol
}

func (g *git) formatSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[0:7]
}

func (g *git) hasGitFile(file string) bool {
	return g.env.hasFilesInDir(g.gitWorkingFolder, file)
}

func (g *git) getGitRefFileSymbolicName(refFile string) string {
	ref := g.getFileContents(g.gitWorkingFolder, refFile)
	return g.getGitCommandOutput("name-rev", "--name-only", "--exclude=tags/*", ref)
}

func (g *git) setPrettyHEADName() {
	// we didn't fetch status, fallback to parsing the HEAD file
	if len(g.Hash) == 0 {
		HEADRef := g.getFileContents(g.gitWorkingFolder, "HEAD")
		if strings.HasPrefix(HEADRef, BRANCHPREFIX) {
			branchName := strings.TrimPrefix(HEADRef, BRANCHPREFIX)
			g.HEAD = fmt.Sprintf("%s%s", g.props.getString(BranchIcon, "\uE0A0"), g.formatHEAD(branchName))
			return
		}
		// no branch, points to commit
		if len(HEADRef) >= 7 {
			g.Hash = HEADRef[0:7]
		}
	}
	// check for tag
	tagName := g.getGitCommandOutput("describe", "--tags", "--exact-match")
	if len(tagName) > 0 {
		g.HEAD = fmt.Sprintf("%s%s", g.props.getString(TagIcon, "\uF412"), tagName)
		return
	}
	// fallback to commit
	if len(g.Hash) == 0 {
		g.HEAD = g.props.getString(NoCommitsIcon, "\uF594 ")
		return
	}
	g.HEAD = fmt.Sprintf("%s%s", g.props.getString(CommitIcon, "\uF417"), g.Hash)
}

func (g *git) getStashContext() int {
	stashContent := g.getFileContents(g.gitRootFolder, "logs/refs/stash")
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
