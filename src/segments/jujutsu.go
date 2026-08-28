package segments

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	JUJUTSUCOMMAND = "jj"

	IgnoreWorkingCopy options.Option = "ignore_working_copy"
	ChangeIDMinLen    options.Option = "change_id_min_len"
)

type JujutsuStatus struct {
	ScmStatus
}

func (s *JujutsuStatus) add(code byte) {
	switch code {
	case 'D':
		s.Deleted++
	case 'A', 'C': // added, copied
		s.Added++
	case 'M':
		s.Modified++
	case 'R': // renamed
		s.Moved++
	}
}

// jujutsuStatusFields lists what setJujutsuStatus populates: the single
// probe this segment derives from its templates (see FieldRefs).
var jujutsuStatusFields = []string{workingField, "ChangeID", "ChangeIDPrefix", "ChangeIDRest"}

type Jujutsu struct {
	Working          *JujutsuStatus
	ChangeID         string
	ChangeIDPrefix   string
	ChangeIDRest     string
	closestBookmarks string
	Scm
	FieldRefs
	aheadCount          int
	closestBookmarksSet bool
	aheadCountSet       bool
}

func (jj *Jujutsu) Template() string {
	return " \uf1fa{{.ChangeID}}{{if .Working.Changed}} \uf044 {{ .Working.String }}{{ end }} "
}

// Activation gates on the repository marker shouldDisplay searches for.
func (jj *Jujutsu) Activation() Activation {
	return Activation{ProjectFiles: []string{".jj"}}
}

func (jj *Jujutsu) Enabled() bool {
	displayStatus := jj.fetchUnit(jujutsuStatusFields...)

	if !jj.shouldDisplay(displayStatus) {
		return false
	}

	statusFormats := jj.options.KeyValueMap(StatusFormats, map[string]string{})
	jj.Working = &JujutsuStatus{Formats: statusFormats}

	if displayStatus {
		jj.setJujutsuStatus()
	}

	return true
}

func (jj *Jujutsu) CacheKey() (string, bool) {
	dir, err := jj.env.HasParentFilePath(".jj", true)
	if err != nil {
		return "", false
	}

	return dir.Path, true
}

// ClosestBookmarks returns the bookmark name(s) on the closest bookmarked
// ancestor(s) of the working copy, undecorated. Resolved lazily on first
// template use and memoized, so a template can reference it more than once
// per render at the cost of a single jj call.
func (jj *Jujutsu) ClosestBookmarks() string {
	if jj.closestBookmarksSet {
		return jj.closestBookmarks
	}

	jj.closestBookmarksSet = true

	statusString, err := jj.getJujutsuCommandOutput("log", "-r", "heads(::@ & bookmarks())", "--no-graph", "-T", "bookmarks")
	if err != nil {
		return ""
	}

	jj.closestBookmarks, _, _ = strings.Cut(statusString, "\n")

	return jj.closestBookmarks
}

// AheadCount returns the number of changes between the working copy and the
// closest bookmark (see ClosestBookmarks). Referencing it in a template is
// what triggers the extra jj call: it runs on first use, memoized, and
// returns 0 when there is no bookmark or the call fails. Templates compose
// their own decoration, e.g. {{ if gt .AheadCount 0 }}\u21e1{{ .AheadCount }}{{ end }}.
func (jj *Jujutsu) AheadCount() int {
	if jj.aheadCountSet {
		return jj.aheadCount
	}

	jj.aheadCountSet = true

	// closest bookmarks all share the same distance from the working copy,
	// so the first one measures for all of them - reusing ClosestBookmarks'
	// memoized call instead of querying the bookmarks twice
	line := jj.ClosestBookmarks()
	if line == "" {
		return 0
	}

	marks := strings.Split(line, " ")
	rangeString := strings.Trim(marks[0], "*") + "..@"

	aheadString, err := jj.getJujutsuCommandOutput("log", "--no-graph", "-T", "'.'", "-r", rangeString)
	if err != nil {
		return 0
	}

	jj.aheadCount = len(aheadString)

	log.Debug("distance to nearest jj bookmark: " + strconv.Itoa(jj.aheadCount))

	return jj.aheadCount
}

func (jj *Jujutsu) shouldDisplay(displayStatus bool) bool {
	jjdir, err := jj.env.HasParentFilePath(".jj", false)
	if err != nil {
		log.Debug("Jujutsu directory not found")
		return false
	}

	if displayStatus && !jj.hasCommand(JUJUTSUCOMMAND) {
		log.Debug("Jujutsu command not found, skipping segment")
		return false
	}

	jj.setDir(jjdir.ParentFolder)

	jj.mainSCMDir = jjdir.Path
	jj.scmDir = jjdir.Path
	// convert the worktree file path to a windows one when in a WSL shared folder
	jj.repoRootDir = strings.TrimSuffix(jj.convertToWindowsPath(jjdir.Path), "/.jj")

	return true
}

func (jj *Jujutsu) setDir(dir string) {
	dir = path.ReplaceHomeDirPrefixWithTilde(dir) // align with template PWD
	if jj.env.GOOS() == runtime.WINDOWS {
		jj.Dir = strings.TrimSuffix(dir, `\.jj`)
		return
	}

	jj.Dir = strings.TrimSuffix(dir, "/.jj")
}

func (jj *Jujutsu) setJujutsuStatus() {
	statusString, err := jj.getJujutsuCommandOutput("log", "-r", "@", "--no-graph", "-T", jj.logTemplate())
	if err != nil {
		return
	}

	header, statusString, _ := strings.Cut(statusString, "\n")
	prefix, rest, found := strings.Cut(header, "|")
	// Jujutsu change IDs contain only canonical ID characters, so "|" is a safe
	// separator that survives RunCommand's trimming when rest is empty.
	if !found || len(prefix) == 0 || strings.Contains(rest, "|") {
		return
	}

	jj.ChangeIDPrefix = prefix
	jj.ChangeIDRest = rest
	jj.ChangeID = prefix + rest

	for line := range strings.SplitSeq(statusString, "\n") {
		if len(line) > 0 {
			jj.Working.add(line[0])
		}
	}
}

func (jj *Jujutsu) logTemplate() string {
	// https://jj-vcs.github.io/jj/latest/templates/#commit-keywords
	minLength := jj.options.Int(ChangeIDMinLen, 0)
	return fmt.Sprintf(
		`change_id.shortest(%d).prefix() ++ "|" ++ change_id.shortest(%d).rest() ++ "\n" ++ diff.summary()`,
		minLength,
		minLength,
	)
}

func (jj *Jujutsu) getJujutsuCommandOutput(command string, args ...string) (string, error) {
	cli := []string{"--repository", jj.repoRootDir, "--no-pager", "--color", "never"}

	if jj.options.Bool(IgnoreWorkingCopy, true) {
		cli = append(cli, "--ignore-working-copy")
	}

	cli = append(cli, command)
	cli = append(cli, args...)

	return jj.env.RunCommand(jj.command, cli...)
}
