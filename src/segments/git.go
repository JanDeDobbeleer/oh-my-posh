package segments

import (
	"fmt"
	url2 "net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"gopkg.in/ini.v1"
)

type Commit struct {
	Timestamp time.Time
	Author    *User
	Committer *User
	Refs      *Refs
	Subject   string
	Sha       string
}

type Refs struct {
	Heads   []string
	Tags    []string
	Remotes []string
}

type User struct {
	Name  string
	Email string
}

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
	case "A":
		s.Added++
	case "?":
		s.Untracked++
	case "U", "AA":
		s.Unmerged++
	case "M", "R", "C", "m":
		s.Modified++
	}
}

const (
	// FetchStatus fetches the status of the repository
	FetchStatus options.Option = "fetch_status"
	// FetchPushStatus fetches the push-remote status
	FetchPushStatus options.Option = "fetch_push_status"
	// IgnoreStatus allows to ignore certain repo's for status information
	IgnoreStatus options.Option = "ignore_status"
	// FetchUpstreamIcon fetches the upstream icon
	FetchUpstreamIcon options.Option = "fetch_upstream_icon"
	// FetchBareInfo fetches the bare repo status
	FetchBareInfo options.Option = "fetch_bare_info"
	// FetchUser fetches the current user for the repo
	FetchUser options.Option = "fetch_user"
	// UntrackedModes list the optional untracked files mode per repo
	UntrackedModes options.Option = "untracked_modes"
	// IgnoreSubmodules list the optional ignore-submodules mode per repo
	IgnoreSubmodules options.Option = "ignore_submodules"
	// MappedBranches allows overriding certain branches with an icon/text
	MappedBranches options.Option = "mapped_branches"
	// DisableWithJJ disables the git segment when there's a .jj directory in the parent file path
	DisableWithJJ options.Option = "disable_with_jj"

	// BranchIcon the icon to use as branch indicator
	BranchIcon options.Option = "branch_icon"
	// BranchIdenticalIcon the icon to display when the remote and local branch are identical
	BranchIdenticalIcon options.Option = "branch_identical_icon"
	// BranchAheadIcon the icon to display when the local branch is ahead of the remote
	BranchAheadIcon options.Option = "branch_ahead_icon"
	// BranchBehindIcon the icon to display when the local branch is behind the remote
	BranchBehindIcon options.Option = "branch_behind_icon"
	// BranchGoneIcon the icon to use when ther's no remote
	BranchGoneIcon options.Option = "branch_gone_icon"
	// RebaseIcon shows before the rebase context
	RebaseIcon options.Option = "rebase_icon"
	// CherryPickIcon shows before the cherry-pick context
	CherryPickIcon options.Option = "cherry_pick_icon"
	// RevertIcon shows before the revert context
	RevertIcon options.Option = "revert_icon"
	// CommitIcon shows before the detached context
	CommitIcon options.Option = "commit_icon"
	// NoCommitsIcon shows when there are no commits in the repo yet
	NoCommitsIcon options.Option = "no_commits_icon"
	// TagIcon shows before the tag context
	TagIcon options.Option = "tag_icon"
	// MergeIcon shows before the merge context
	MergeIcon options.Option = "merge_icon"
	// UpstreamIcons allows to add custom upstream icons
	UpstreamIcons options.Option = "upstream_icons"
	// GithubIcon shows when upstream is github
	GithubIcon options.Option = "github_icon"
	// BitbucketIcon shows  when upstream is bitbucket
	BitbucketIcon options.Option = "bitbucket_icon"
	// AzureDevOpsIcon shows  when upstream is azure devops
	AzureDevOpsIcon options.Option = "azure_devops_icon"
	// CodeCommit shows  when upstream is aws codecommit
	CodeCommit options.Option = "codecommit_icon"
	// CodebergIcon shows when upstream is codeberg
	CodebergIcon options.Option = "codeberg_icon"
	// GitlabIcon shows when upstream is gitlab
	GitlabIcon options.Option = "gitlab_icon"
	// GitIcon shows when the upstream can't be identified
	GitIcon options.Option = "git_icon"

	DETACHED     = "(detached)"
	BRANCHPREFIX = "ref: refs/heads/"
	GITCOMMAND   = "git"

	trueStr = "true"
	origin  = "origin"
)

type Rebase struct {
	HEAD    string
	Onto    string
	Current int
	Total   int
}

type Git struct {
	configErr      error
	config         *ini.File
	Working        *GitStatus
	Staging        *GitStatus
	commit         *Commit
	Rebase         *Rebase
	User           *User
	ShortHash      string
	Hash           string
	BranchStatus   string
	HEAD           string
	UpstreamIcon   string
	UpstreamURL    string
	Ref            string
	RawUpstreamURL string
	Scm
	stashCount    int
	Ahead         int
	PushAhead     int
	PushBehind    int
	Behind        int
	worktreeCount int
	configOnce    sync.Once
	IsWorkTree    bool
	Merge         bool
	CherryPick    bool
	Revert        bool
	poshgit       bool
	Detached      bool
	IsBare        bool
	UpstreamGone  bool
}

func (g *Git) Template() string {
	return " {{ .HEAD }}{{if .BranchStatus }} {{ .BranchStatus }}{{ end }}{{ if .Working.Changed }} \uF044 {{ .Working.String }}{{ end }}{{ if and (.Staging.Changed) (.Working.Changed) }} |{{ end }}{{ if .Staging.Changed }} \uF046 {{ .Staging.String }}{{ end }} " //nolint: lll
}

func (g *Git) Enabled() bool {
	g.User = &User{}
	g.Working = &GitStatus{}
	g.Staging = &GitStatus{}

	if !g.shouldDisplay() {
		return false
	}

	fetchUser := g.options.Bool(FetchUser, false)
	if fetchUser {
		g.setUser()
	}

	g.RepoName = g.repoName()

	if g.IsBare {
		g.getBareRepoInfo()
		return true
	}

	source := g.options.String(Source, Cli)
	if source == Pwsh && g.hasPoshGitStatus() {
		return true
	}

	displayStatus := g.options.Bool(FetchStatus, false)
	if displayStatus && g.shouldIgnoreStatus() {
		displayStatus = false
	}

	if displayStatus {
		g.setStatus()
		g.setHEADStatus()
		g.setBranchStatus()
		g.setPushStatus()
	} else {
		g.updateHEADReference()
	}

	if g.options.Bool(FetchUpstreamIcon, false) {
		g.UpstreamIcon = g.getUpstreamIcon()
	}

	return true
}

func (g *Git) CacheKey() (string, bool) {
	dir, err := g.env.HasParentFilePath(".git", true)
	if err != nil {
		return "", false
	}

	if !g.isRepo(dir) {
		return "", false
	}

	ref := g.fileContent(g.mainSCMDir, "HEAD")
	ref = strings.Replace(ref, "ref: refs/heads/", "", 1)

	// Use the repo clone in the cache key so the mapped path is consistent
	// for primary and worktree repos.
	return fmt.Sprintf("%s@%s", dir.Path, ref), true
}

func (g *Git) Commit() *Commit {
	if g.commit != nil {
		return g.commit
	}

	g.commit = &Commit{
		Author:    &User{},
		Committer: &User{},
		Refs:      &Refs{},
	}

	commitBody := g.getGitCommandOutput("log", "-1", "--pretty=format:an:%an%nae:%ae%ncn:%cn%nce:%ce%nat:%at%nsu:%s%nha:%H%nrf:%D", "--decorate=full")
	splitted := strings.SplitSeq(strings.TrimSpace(commitBody), "\n")
	for line := range splitted {
		line = strings.TrimSpace(line)
		if len(line) <= 3 {
			continue
		}
		anchor := line[:3]
		line = line[3:]
		switch anchor {
		case "an:":
			g.commit.Author.Name = line
		case "ae:":
			g.commit.Author.Email = line
		case "cn:":
			g.commit.Committer.Name = line
		case "ce:":
			g.commit.Committer.Email = line
		case "at:":
			if t, err := strconv.ParseInt(line, 10, 64); err == nil {
				g.commit.Timestamp = time.Unix(t, 0)
			}
		case "su:":
			g.commit.Subject = line
		case "ha:":
			g.commit.Sha = line
		case "rf:":
			refs := strings.SplitSeq(line, ", ")
			for ref := range refs {
				ref = strings.TrimSpace(ref)
				switch {
				case strings.HasSuffix(ref, "HEAD"):
					continue
				case strings.HasPrefix(ref, "tag: refs/tags/"):
					g.commit.Refs.Tags = append(g.commit.Refs.Tags, strings.TrimPrefix(ref, "tag: refs/tags/"))
				case strings.HasPrefix(ref, "refs/remotes/"):
					g.commit.Refs.Remotes = append(g.commit.Refs.Remotes, strings.TrimPrefix(ref, "refs/remotes/"))
				case strings.HasPrefix(ref, "HEAD -> refs/heads/"):
					g.commit.Refs.Heads = append(g.commit.Refs.Heads, strings.TrimPrefix(ref, "HEAD -> refs/heads/"))
				case strings.HasPrefix(ref, "refs/heads/"):
					g.commit.Refs.Heads = append(g.commit.Refs.Heads, strings.TrimPrefix(ref, "refs/heads/"))
				default:
					g.commit.Refs.Heads = append(g.commit.Refs.Heads, ref)
				}
			}
		}
	}
	return g.commit
}

func (g *Git) StashCount() int {
	if g.poshgit || g.stashCount != 0 {
		return g.stashCount
	}

	stashContent := g.fileContent(g.scmDir, "logs/refs/stash")
	if stashContent == "" {
		return 0
	}

	lines := strings.Split(stashContent, "\n")
	g.stashCount = len(lines)
	return g.stashCount
}

func (g *Git) Kraken() string {
	root := g.getGitCommandOutput("rev-list", "--max-parents=0", "HEAD")
	root, _, _ = strings.Cut(root, "\n")

	if g.RawUpstreamURL == "" {
		if g.Upstream == "" {
			g.Upstream = origin
		}
		g.RawUpstreamURL = g.getRemoteURL()
	}

	if g.Hash == "" {
		g.Hash = g.getGitCommandOutput("rev-parse", "HEAD")
	}

	return fmt.Sprintf("gitkraken://repolink/%s/commit/%s?url=%s", root, g.Hash, url2.QueryEscape(g.RawUpstreamURL))
}

func (g *Git) LatestTag() string {
	return g.getGitCommandOutput("describe", "--tags", "--abbrev=0")
}

func (g *Git) shouldDisplay() bool {
	// Check if disable_with_jj is enabled and .jj directory exists
	if g.options.Bool(DisableWithJJ, false) {
		if _, err := g.env.HasParentFilePath(".jj", false); err == nil {
			return false
		}
	}

	gitdir, err := g.env.HasParentFilePath(".git", true)
	if err != nil {
		return false
	}

	if g.options.Bool(FetchBareInfo, false) {
		g.IsBare = g.isBareRepo(gitdir)
	}

	if !g.hasCommand(GITCOMMAND) {
		return false
	}

	return g.isRepo(gitdir)
}

func (g *Git) isRepo(gitdir *runtime.FileInfo) bool {
	g.setDir(gitdir.Path)

	if !gitdir.IsDir {
		if g.hasWorktree(gitdir) {
			g.repoRootDir = g.convertToWindowsPath(g.repoRootDir)
			return true
		}

		return false
	}

	g.mainSCMDir = gitdir.Path
	g.scmDir = gitdir.Path
	// convert the worktree file path to a windows one when in a WSL shared folder
	g.repoRootDir = strings.TrimSuffix(g.convertToWindowsPath(gitdir.Path), "/.git")
	return true
}

func (g *Git) setUser() {
	g.User.Name = g.getGitCommandOutput("config", "user.name")
	g.User.Email = g.getGitCommandOutput("config", "user.email")
}

func (g *Git) isBareRepo(gitDir *runtime.FileInfo) bool {
	defer log.Trace(time.Now())

	if gitDir.IsDir {
		g.mainSCMDir = gitDir.Path
	} else {
		content := g.fileContent(gitDir.ParentFolder, ".git")
		dir := strings.TrimPrefix(content, "gitdir: ")
		g.mainSCMDir = filepath.Join(gitDir.ParentFolder, dir)
	}

	cfg, err := g.getGitConfig()
	if err != nil {
		log.Error(err)
		return false
	}

	coreSection := cfg.Section("core")
	if coreSection == nil {
		log.Debug("Git core section not found, not a bare repo")
		return false
	}

	bare := coreSection.Key("bare").String()

	return bare == trueStr
}

func (g *Git) getBareRepoInfo() {
	head := g.fileContent(g.mainSCMDir, "HEAD")
	branchIcon := g.options.String(BranchIcon, "\uE0A0")
	g.Ref = strings.Replace(head, "ref: refs/heads/", "", 1)
	g.HEAD = fmt.Sprintf("%s%s", branchIcon, g.formatBranch(g.Ref))
	if !g.options.Bool(FetchUpstreamIcon, false) {
		return
	}

	g.Upstream = g.getGitCommandOutput("remote")
	if len(g.Upstream) != 0 {
		g.UpstreamIcon = g.getUpstreamIcon()
	}
}

func (g *Git) setDir(dir string) {
	dir = path.ReplaceHomeDirPrefixWithTilde(dir) // align with template PWD
	if g.env.GOOS() == runtime.WINDOWS {
		g.Dir = strings.TrimSuffix(dir, `\.git`)
		return
	}

	g.Dir = strings.TrimSuffix(dir, "/.git")
}

func (g *Git) hasWorktree(gitdir *runtime.FileInfo) bool {
	g.scmDir = gitdir.Path
	content := g.env.FileContent(gitdir.Path)
	content = strings.Trim(content, " \r\n")
	matches := regex.FindNamedRegexMatch(`^gitdir: (?P<dir>.*)$`, content)

	if len(matches) == 0 {
		log.Debug("no matches found, directory isn't a worktree")
		return false
	}

	// if we open a worktree file in a WSL shared folder, we have to convert it back
	// to the mounted path
	g.mainSCMDir = g.convertToLinuxPath(matches["dir"])

	// in worktrees, the path looks like this: gitdir: path/.git/worktrees/branch
	// scmDir needs to become path/.git
	// repoRootDir needs to become path
	worktreeIndex := strings.LastIndex(g.mainSCMDir, "/worktrees/")

	// in submodules, the path looks like this: gitdir: ../.git/modules/test-submodule
	// we need the parent folder to detect where the real .git folder is
	if strings.Contains(g.mainSCMDir, "/modules/") {
		g.scmDir = resolveGitPath(gitdir.ParentFolder, g.mainSCMDir)
		// this might be both a worktree and a submodule, where the path would look like
		// this: path/.git/modules/module/path/worktrees/location. We cannot distinguish
		// between worktree and a module path containing the word 'worktree,' however.
		worktreeIndex = strings.LastIndex(g.scmDir, "/worktrees/")
		if worktreeIndex > -1 && g.env.HasFilesInDir(g.scmDir, "gitdir") {
			gitDir := filepath.Join(g.scmDir, "gitdir")
			realGitFolder := g.env.FileContent(gitDir)
			g.repoRootDir = strings.TrimSuffix(realGitFolder, ".git\n")
			// resolve relative paths (worktree.useRelativePaths = true)
			g.repoRootDir = resolveGitPath(g.scmDir, g.repoRootDir)
			g.scmDir = g.scmDir[:worktreeIndex]
			g.mainSCMDir = g.scmDir
			g.IsWorkTree = true
			return true
		}

		g.repoRootDir = g.scmDir
		g.mainSCMDir = g.scmDir
		return true
	}

	// convert to absolute path for worktrees only
	if strings.HasPrefix(g.mainSCMDir, "..") {
		g.mainSCMDir = filepath.Join(gitdir.ParentFolder, g.mainSCMDir)
	}

	if worktreeIndex > -1 {
		gitDir := filepath.Join(g.mainSCMDir, "gitdir")
		g.scmDir = g.mainSCMDir[:worktreeIndex]
		gitDirContent := g.env.FileContent(gitDir)
		g.repoRootDir = strings.TrimSuffix(gitDirContent, ".git\n")
		// resolve relative paths (worktree.useRelativePaths = true)
		g.repoRootDir = resolveGitPath(g.mainSCMDir, g.repoRootDir)
		g.IsWorkTree = true
		return true
	}

	// check for separate git folder(--separate-git-dir)
	// check if the folder contains a HEAD file
	if g.env.HasFilesInDir(g.mainSCMDir, "HEAD") {
		gitFolder := strings.TrimSuffix(g.scmDir, ".git")
		g.scmDir = g.mainSCMDir
		g.mainSCMDir = gitFolder
		g.repoRootDir = gitFolder
		return true
	}

	return false
}

func (g *Git) shouldIgnoreStatus() bool {
	list := g.options.StringArray(IgnoreStatus, []string{})
	return g.env.DirMatchesOneOf(g.repoRootDir, list)
}

func (g *Git) setBranchStatus() {
	getBranchStatus := func() string {
		if g.Ahead > 0 && g.Behind > 0 {
			return fmt.Sprintf("%s%d %s%d", g.options.String(BranchAheadIcon, "\u2191"), g.Ahead, g.options.String(BranchBehindIcon, "\u2193"), g.Behind)
		}
		if g.Ahead > 0 {
			return fmt.Sprintf("%s%d", g.options.String(BranchAheadIcon, "\u2191"), g.Ahead)
		}
		if g.Behind > 0 {
			return fmt.Sprintf("%s%d", g.options.String(BranchBehindIcon, "\u2193"), g.Behind)
		}
		if g.UpstreamGone {
			return g.options.String(BranchGoneIcon, "\u2262")
		}
		if g.Behind == 0 && g.Ahead == 0 && g.Upstream != "" {
			return g.options.String(BranchIdenticalIcon, "\u2261")
		}
		return ""
	}
	g.BranchStatus = getBranchStatus()
}

func (g *Git) setPushStatus() {
	if !g.options.Bool(FetchPushStatus, false) {
		return
	}

	if g.Ref == "" || g.Ref == DETACHED {
		return
	}

	pushRemote := g.getPushRemote()
	if pushRemote == "" {
		return
	}

	ahead := g.getGitCommandOutput("rev-list", "--count", pushRemote+"..HEAD")
	if ahead != "" {
		g.PushAhead, _ = strconv.Atoi(strings.TrimSpace(ahead))
	}

	behind := g.getGitCommandOutput("rev-list", "--count", "HEAD.."+pushRemote)
	if behind != "" {
		g.PushBehind, _ = strconv.Atoi(strings.TrimSpace(behind))
	}
}

func (g *Git) getPushRemote() string {
	upstream := g.Upstream
	if idx := strings.Index(upstream, "/"); idx != -1 {
		upstream = upstream[:idx]
	}

	if upstream == "" {
		upstream = origin
	}

	branch := g.Ref
	if branch == "" {
		return ""
	}

	cfg, err := g.getGitConfig()
	if err != nil {
		pushRemote := g.getGitCommandOutput("config", "--get", "remote.pushDefault")
		if pushRemote == "" {
			pushRemote = upstream
		}

		return strings.TrimSpace(pushRemote) + "/" + branch
	}

	sectionName := fmt.Sprintf(`branch "%s"`, branch)
	section := cfg.Section(sectionName)
	pushRemote := section.Key("pushRemote").String()
	if pushRemote == "" {
		pushRemote = cfg.Section("remote").Key("pushDefault").String()
	}

	if pushRemote == "" {
		pushRemote = upstream
	}

	return pushRemote + "/" + branch
}

func (g *Git) getGitConfig() (*ini.File, error) {
	g.configOnce.Do(func() {
		configData := g.fileContent(g.mainSCMDir, "config")
		if configData == "" {
			log.Debug("git config file not found")
			g.configErr = fmt.Errorf("git config file not found")
			return
		}

		// ini.Load expects []byte to parse content, not a file path
		cfg, err := ini.Load([]byte(configData))
		if err != nil {
			g.configErr = err
			return
		}

		g.config = cfg
	})

	return g.config, g.configErr
}

func (g *Git) cleanUpstreamURL(url string) string {
	// Azure DevOps
	if strings.Contains(url, "dev.azure.com") {
		match := regex.FindNamedRegexMatch(`^.*@(ssh.)?dev\.azure\.com(:v3)?/(?P<ORGANIZATION>[A-Za-z0-9_-]+)/(?P<PROJECT>[A-Za-z0-9_-]+)/(_git/)?(?P<REPOSITORY>[A-Za-z0-9_-]+)$`, url)
		if len(match) == 4 {
			return fmt.Sprintf("https://dev.azure.com/%s/%s/_git/%s", match["ORGANIZATION"], match["PROJECT"], match["REPOSITORY"])
		}
	}

	if strings.HasPrefix(url, "http") {
		return url
	}

	// /path/to/repo.git/
	match := regex.FindNamedRegexMatch(`^(?P<URL>[a-z0-9./]+)$`, url)
	if len(match) != 0 {
		url := strings.Trim(match["URL"], "/")
		url = strings.TrimSuffix(url, ".git")
		return fmt.Sprintf("https://%s", strings.TrimPrefix(url, "/"))
	}

	// ssh://user@host.xz:1234/path/to/repo.git/
	match = regex.FindNamedRegexMatch(`(ssh|ftp|git|rsync)://(.*@)?(?P<URL>[a-z0-9.-]+)(:[0-9]{1,5})?/(?P<PATH>.*).git`, url)
	if len(match) == 0 {
		// host.xz:/path/to/repo.git/
		match = regex.FindNamedRegexMatch(`^(?P<URL>[a-z0-9.-]+):(?P<PATH>[\w.\-~/@]+)$`, url)
	}

	if len(match) != 0 {
		repoPath := strings.Trim(match["PATH"], "/")
		repoPath = strings.TrimSuffix(repoPath, ".git")
		return fmt.Sprintf("https://%s/%s", match["URL"], repoPath)
	}

	// codecommit::region-identifier-id://repo-name
	match = regex.FindNamedRegexMatch(`codecommit::(?P<URL>[a-z0-9-]+)://(?P<PATH>[\w\.@\:/\-~]+)`, url)
	if len(match) != 0 {
		return fmt.Sprintf("https://%s.console.aws.amazon.com/codesuite/codecommit/repositories/%s/browse?region=%s", match["URL"], match["PATH"], match["URL"])
	}

	// user@host.xz:/path/to/repo.git
	match = regex.FindNamedRegexMatch(`.*@(?P<URL>.*):(?P<PATH>.*)`, url)
	if len(match) == 0 {
		return ""
	}

	return fmt.Sprintf("https://%s/%s", match["URL"], strings.TrimSuffix(match["PATH"], ".git"))
}

func (g *Git) getUpstreamIcon() string {
	fallback := g.options.String(GitIcon, "\uE5FB ")

	g.RawUpstreamURL = g.getRemoteURL()
	if g.RawUpstreamURL == "" {
		return fallback
	}

	g.UpstreamURL = g.cleanUpstreamURL(g.RawUpstreamURL)

	// allow overrides first
	custom := g.options.KeyValueMap(UpstreamIcons, map[string]string{})
	for key, value := range custom {
		if strings.Contains(g.UpstreamURL, key) {
			return value
		}
	}

	defaults := map[string]struct {
		Icon    options.Option
		Default string
	}{
		"github":           {GithubIcon, "\uF408"},
		"gitlab":           {GitlabIcon, "\uF296"},
		"bitbucket":        {BitbucketIcon, "\uF171"},
		"dev.azure.com":    {AzureDevOpsIcon, "\uEBE8"},
		"visualstudio.com": {AzureDevOpsIcon, "\uEBE8"},
		"codecommit":       {CodeCommit, "\uF270"},
		"codeberg":         {CodebergIcon, "\uF330"},
	}

	for key, value := range defaults {
		if strings.Contains(g.UpstreamURL, key) {
			return g.options.String(value.Icon, value.Default)
		}
	}

	return fallback
}

func (g *Git) setStatus() {
	addToStatus := func(status string) {
		const UNTRACKED = "?"
		if strings.HasPrefix(status, UNTRACKED) {
			g.Working.add(UNTRACKED)
			return
		}
		if len(status) <= 4 {
			return
		}

		// map conflicts separately when in a merge or rebase
		if g.Rebase != nil || g.Merge {
			conflict := "AA"
			full := status[2:4]
			if full == conflict {
				g.Staging.add(conflict)
				return
			}
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

	// firstly assume that upstream is gone
	g.UpstreamGone = true
	statusFormats := g.options.KeyValueMap(StatusFormats, map[string]string{})

	g.Working = &GitStatus{ScmStatus: ScmStatus{Formats: statusFormats}}
	g.Staging = &GitStatus{ScmStatus: ScmStatus{Formats: statusFormats}}

	untrackedMode := g.getUntrackedFilesMode()
	args := []string{"status", untrackedMode, "--branch", "--porcelain=2"}
	ignoreSubmodulesMode := g.getIgnoreSubmodulesMode()
	if len(ignoreSubmodulesMode) > 0 {
		args = append(args, ignoreSubmodulesMode)
	}

	output := g.getGitCommandOutput(args...)
	for line := range strings.SplitSeq(output, "\n") {
		if strings.HasPrefix(line, HASH) && len(line) >= len(HASH)+7 {
			g.ShortHash = line[len(HASH) : len(HASH)+7]
			g.Hash = line[len(HASH):]
			continue
		}

		if strings.HasPrefix(line, REF) && len(line) > len(REF) {
			g.Ref = line[len(REF):]
			continue
		}

		if strings.HasPrefix(line, UPSTREAM) && len(line) > len(UPSTREAM) {
			// status reports upstream, but upstream may be gone (must check BRANCHSTATUS)
			g.Upstream = line[len(UPSTREAM):]
			g.UpstreamGone = true
			continue
		}

		if strings.HasPrefix(line, BRANCHSTATUS) && len(line) > len(BRANCHSTATUS) {
			status := line[len(BRANCHSTATUS):]
			splitted := strings.Split(status, " ")
			if len(splitted) >= 2 {
				g.Ahead, _ = strconv.Atoi(splitted[0])
				behind, _ := strconv.Atoi(splitted[1])
				g.Behind = -behind
			}
			// confirmed: upstream exists
			g.UpstreamGone = false
			continue
		}

		addToStatus(line)
	}
}

func (g *Git) getGitCommandOutput(args ...string) string {
	if g.command == "" {
		return ""
	}

	args = append([]string{"-C", g.repoRootDir, "--no-optional-locks", "-c", "core.quotepath=false", "-c", "color.status=false"}, args...)
	val, err := g.env.RunCommand(g.command, args...)
	if err != nil {
		return ""
	}

	return val
}

func (g *Git) setHEADStatus() {
	branchIcon := g.options.String(BranchIcon, "\uE0A0")
	if g.Ref == DETACHED {
		g.Detached = true
		g.resolveDetachedHEAD()
	} else {
		head := g.formatBranch(g.Ref)
		g.HEAD = fmt.Sprintf("%s%s", branchIcon, head)
	}

	formatDetached := func() string {
		if g.Detached {
			return fmt.Sprintf("%sdetached at %s", branchIcon, g.HEAD)
		}
		return g.HEAD
	}

	getPrettyNameOrigin := func(file string) string {
		var origin string
		head := g.fileContent(g.mainSCMDir, file)
		if head == "detached HEAD" {
			origin = formatDetached()
		} else {
			head = strings.Replace(head, "refs/heads/", "", 1)
			origin = branchIcon + g.formatBranch(head)
		}
		return origin
	}

	parseInt := func(file string) int {
		val, _ := strconv.Atoi(g.fileContent(g.mainSCMDir, file))
		return val
	}

	if g.env.HasFolder(g.mainSCMDir + "/rebase-merge") {
		head := getPrettyNameOrigin("rebase-merge/head-name")
		onto := g.getGitRefFileSymbolicName("rebase-merge/onto")
		onto = g.formatBranch(onto)
		current := parseInt("rebase-merge/msgnum")
		total := parseInt("rebase-merge/end")
		icon := g.options.String(RebaseIcon, "\uE728 ")

		g.Rebase = &Rebase{
			HEAD:    head,
			Onto:    onto,
			Current: current,
			Total:   total,
		}

		g.HEAD = fmt.Sprintf("%s%s onto %s%s (%d/%d) at %s", icon, head, branchIcon, onto, current, total, g.HEAD)
		return
	}

	if g.env.HasFolder(g.mainSCMDir + "/rebase-apply") {
		head := getPrettyNameOrigin("rebase-apply/head-name")
		current := parseInt("rebase-apply/next")
		total := parseInt("rebase-apply/last")
		icon := g.options.String(RebaseIcon, "\uE728 ")

		g.Rebase = &Rebase{
			HEAD:    head,
			Current: current,
			Total:   total,
		}

		g.HEAD = fmt.Sprintf("%s%s (%d/%d) at %s", icon, head, current, total, g.HEAD)
		return
	}

	// merge
	commitIcon := g.options.String(CommitIcon, "\uF417")

	if g.hasGitFile("MERGE_MSG") {
		g.Merge = true
		icon := g.options.String(MergeIcon, "\uE727 ")
		mergeContext := g.fileContent(g.mainSCMDir, "MERGE_MSG")
		matches := regex.FindNamedRegexMatch(`Merge (remote-tracking )?(?P<type>branch|commit|tag) '(?P<theirs>.*)'`, mergeContext)
		// head := g.getGitRefFileSymbolicName("ORIG_HEAD")
		if matches != nil && matches["theirs"] != "" {
			var headIcon, theirs string
			switch matches["type"] {
			case "tag":
				headIcon = g.options.String(TagIcon, "\uF412")
				theirs = matches["theirs"]
			case "commit":
				headIcon = commitIcon
				theirs = g.formatSHA(matches["theirs"])
			default:
				headIcon = branchIcon
				theirs = g.formatBranch(matches["theirs"])
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
		g.CherryPick = true
		sha := g.fileContent(g.mainSCMDir, "CHERRY_PICK_HEAD")
		cherry := g.options.String(CherryPickIcon, "\uE29B ")
		g.HEAD = fmt.Sprintf("%s%s%s onto %s", cherry, commitIcon, g.formatSHA(sha), formatDetached())
		return
	}

	if g.hasGitFile("REVERT_HEAD") {
		g.Revert = true
		sha := g.fileContent(g.mainSCMDir, "REVERT_HEAD")
		revert := g.options.String(RevertIcon, "\uF0E2 ")
		g.HEAD = fmt.Sprintf("%s%s%s onto %s", revert, commitIcon, g.formatSHA(sha), formatDetached())
		return
	}

	if g.hasGitFile("sequencer/todo") {
		todo := g.fileContent(g.mainSCMDir, "sequencer/todo")
		matches := regex.FindNamedRegexMatch(`^(?P<action>p|pick|revert)\s+(?P<sha>\S+)`, todo)
		if matches != nil && matches["sha"] != "" {
			action := matches["action"]
			sha := matches["sha"]
			switch action {
			case "p", "pick":
				g.CherryPick = true
				cherry := g.options.String(CherryPickIcon, "\uE29B ")
				g.HEAD = fmt.Sprintf("%s%s%s onto %s", cherry, commitIcon, g.formatSHA(sha), formatDetached())
				return
			case "revert":
				g.Revert = true
				revert := g.options.String(RevertIcon, "\uF0E2 ")
				g.HEAD = fmt.Sprintf("%s%s%s onto %s", revert, commitIcon, g.formatSHA(sha), formatDetached())
				return
			}
		}
	}

	g.HEAD = formatDetached()
}

func (g *Git) formatSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[0:7]
}

func (g *Git) hasGitFile(file string) bool {
	return g.env.HasFilesInDir(g.mainSCMDir, file)
}

func (g *Git) getGitRefFileSymbolicName(refFile string) string {
	ref := g.fileContent(g.mainSCMDir, refFile)
	return g.getGitCommandOutput("name-rev", "--name-only", "--exclude=tags/*", ref)
}

func (g *Git) updateHEADReference() {
	HEADRef := g.fileContent(g.mainSCMDir, "HEAD")
	log.Debug("HEADRef:", HEADRef)

	// check if we are in a repo using reftables
	if HEADRef == "ref: refs/heads/.invalid" {
		log.Debug("repo is using reftables")

		HEADRef = g.getGitCommandOutput("rev-parse", "--symbolic-full-name", "HEAD")

		// this is a detached head
		if strings.HasPrefix(HEADRef, "fatal:") {
			log.Debug("detached HEAD detected")
			g.Detached = true
			g.resolveDetachedHEAD()
			return
		}

		if strings.HasPrefix(HEADRef, "refs/heads/") {
			HEADRef = "ref: " + HEADRef
		}

		log.Debug("resolved HEADRef:", HEADRef)
	}

	g.Detached = !strings.HasPrefix(HEADRef, "ref:")
	if branchName, ok := strings.CutPrefix(HEADRef, BRANCHPREFIX); ok {
		log.Debug("current HEAD is a branch:", branchName)

		g.Ref = branchName
		g.HEAD = fmt.Sprintf("%s%s", g.options.String(BranchIcon, "\uE0A0"), g.formatBranch(branchName))

		return
	}

	g.resolveDetachedHEAD()
}

func (g *Git) resolveDetachedHEAD() {
	HEADRef := g.getGitCommandOutput("rev-parse", "HEAD")

	if len(HEADRef) >= 7 {
		g.ShortHash = HEADRef[0:7]
		g.Hash = HEADRef[0:]
	}
	g.Ref = g.ShortHash

	// check for tag
	tagName := g.getGitCommandOutput("describe", "--tags", "--exact-match")
	if len(tagName) > 0 {
		g.Ref = tagName
		g.HEAD = fmt.Sprintf("%s%s", g.options.String(TagIcon, "\uF412"), tagName)
		return
	}

	// fallback to no commits found
	if g.ShortHash == "" {
		g.HEAD = g.options.String(NoCommitsIcon, "\uF594 ")
		return
	}

	g.HEAD = fmt.Sprintf("%s%s", g.options.String(CommitIcon, "\uF417"), g.ShortHash)
}

func (g *Git) WorktreeCount() int {
	if g.worktreeCount > 0 {
		return g.worktreeCount
	}

	worktreesFolder := filepath.Join(g.mainSCMDir, "worktrees")

	if !g.env.HasFolder(worktreesFolder) {
		return 0
	}

	worktreeFolders := g.env.LsDir(worktreesFolder)
	var count int
	for _, folder := range worktreeFolders {
		if folder.IsDir() {
			count++
		}
	}

	return count
}

func (g *Git) getRemoteURL() string {
	upstream := regex.ReplaceAllString("/.*", g.Upstream, "")
	if upstream == "" {
		upstream = origin
	}

	cfg, err := g.getGitConfig()
	if err != nil {
		return g.getGitCommandOutput("remote", "get-url", upstream)
	}

	url := cfg.Section("remote \"" + upstream + "\"").Key("url").String()
	if len(url) != 0 {
		log.Debug("remote url found in config:", url)
		return url
	}

	return g.getGitCommandOutput("remote", "get-url", upstream)
}

func (g *Git) Remotes() map[string]string {
	var remotes = make(map[string]string)

	cfg, err := g.getGitConfig()
	if err != nil {
		return remotes
	}

	for _, section := range cfg.Sections() {
		if !strings.HasPrefix(section.Name(), "remote ") {
			continue
		}

		name := strings.TrimPrefix(section.Name(), "remote ")
		name = strings.Trim(name, "\"")
		url := section.Key("url").String()
		url = g.cleanUpstreamURL(url)
		remotes[name] = url
	}
	return remotes
}

func (g *Git) getUntrackedFilesMode() string {
	return g.getSwitchMode(UntrackedModes, "-u", "normal")
}

func (g *Git) getIgnoreSubmodulesMode() string {
	return g.getSwitchMode(IgnoreSubmodules, "--ignore-submodules=", "")
}

func (g *Git) getSwitchMode(property options.Option, gitSwitch, mode string) string {
	repoModes := g.options.KeyValueMap(property, map[string]string{})
	// make use of a wildcard for all repo's
	if val := repoModes["*"]; len(val) != 0 {
		mode = val
	}
	// get the specific repo mode
	if val := repoModes[g.repoRootDir]; len(val) != 0 {
		mode = val
	}
	if mode == "" {
		return ""
	}
	return fmt.Sprintf("%s%s", gitSwitch, mode)
}

func (g *Git) repoName() string {
	if !g.IsWorkTree {
		return path.Base(g.convertToLinuxPath(g.repoRootDir))
	}

	ind := strings.LastIndex(g.mainSCMDir, ".git/worktrees")
	if ind > -1 {
		return path.Base(g.mainSCMDir[:ind])
	}

	return ""
}
