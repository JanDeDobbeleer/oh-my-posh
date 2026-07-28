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
	CommitIDMinLen    options.Option = "commit_id_min_len"
	FetchAhead        options.Option = "fetch_ahead_counter"
	AheadIcon         options.Option = "ahead_icon"

	jujutsuStatusFrameHeader = "OMPJJ1:5"
	jujutsuStatusFieldCount  = 5
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

type Jujutsu struct {
	Working        *JujutsuStatus
	ChangeID       string
	ChangeIDPrefix string
	ChangeIDRest   string
	CommitID       string
	CommitIDPrefix string
	CommitIDRest   string
	Scm
}

func (jj *Jujutsu) Template() string {
	return " \uf1fa{{.ChangeID}}{{if .Working.Changed}} \uf044 {{ .Working.String }}{{ end }} "
}

func (jj *Jujutsu) Enabled() bool {
	displayStatus := jj.options.Bool(FetchStatus, false)

	if !jj.shouldDisplay(displayStatus) {
		return false
	}

	statusFormats := jj.options.KeyValueMap(StatusFormats, map[string]string{})
	jj.Working = &JujutsuStatus{ScmStatus: ScmStatus{Formats: statusFormats}}

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

func (jj *Jujutsu) ClosestBookmarks() string {
	statusString, err := jj.getJujutsuCommandOutput("log", "-r", "heads(::@ & bookmarks())", "--no-graph", "-T", "bookmarks")
	if err != nil {
		return ""
	}

	line, _, _ := strings.Cut(statusString, "\n")

	if !jj.options.Bool(FetchAhead, false) || len(line) == 0 {
		return line
	}

	aheadIcon := jj.options.String(AheadIcon, "\u21e1")
	marks := strings.Split(line, " ")
	// String to return for status
	var endString strings.Builder

	// Closest bookmarks are all the same distance away from the working copy
	// so retrieve the distance to the first one and use it for all of them

	rangeString := strings.Trim(marks[0], "*") + "..@"

	aheadString, err := jj.getJujutsuCommandOutput("log", "--no-graph", "-T", "'.'", "-r", rangeString)
	if err != nil {
		return line
	}

	aheadCounter := len(aheadString)
	aheadCounterString := ""

	if aheadCounter != 0 {
		aheadCounterString = aheadIcon + strconv.Itoa(aheadCounter)
	}

	log.Debug("distance to nearest jj bookmark:" + aheadCounterString)

	// Loop through each bookmark
	for index, mark := range marks {
		if index > 0 {
			endString.WriteString(" ")
		}

		endString.WriteString(mark + aheadCounterString)
	}

	return endString.String()
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

	fields := strings.Split(statusString, "\x00")
	if len(fields) != jujutsuStatusFieldCount+2 ||
		fields[0] != jujutsuStatusFrameHeader ||
		fields[len(fields)-1] != "" ||
		fields[1] == "" ||
		fields[3] == "" {
		return
	}

	changeIDPrefix := fields[1]
	changeIDRest := fields[2]
	commitIDPrefix := fields[3]
	commitIDRest := fields[4]
	statusString = fields[5]

	for line := range strings.SplitSeq(statusString, "\n") {
		if len(line) > 0 {
			jj.Working.add(line[0])
		}
	}

	jj.ChangeIDPrefix = changeIDPrefix
	jj.ChangeIDRest = changeIDRest
	jj.ChangeID = changeIDPrefix + changeIDRest
	jj.CommitIDPrefix = commitIDPrefix
	jj.CommitIDRest = commitIDRest
	jj.CommitID = commitIDPrefix + commitIDRest
}

func (jj *Jujutsu) logTemplate() string {
	// https://jj-vcs.github.io/jj/latest/templates/#commit-keywords
	changeIDMinLength := jj.options.Int(ChangeIDMinLen, 0)
	commitIDMinLength := jj.options.Int(CommitIDMinLen, 0)
	template := `"%s\0" ++ change_id.shortest(%d).prefix() ++ "\0" ++ ` +
		`change_id.shortest(%d).rest() ++ "\0" ++ commit_id.shortest(%d).prefix() ++ "\0" ++ ` +
		`commit_id.shortest(%d).rest() ++ "\0" ++ diff.summary() ++ "\0"`
	return fmt.Sprintf(
		template,
		jujutsuStatusFrameHeader,
		changeIDMinLength,
		changeIDMinLength,
		commitIDMinLength,
		commitIDMinLength,
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
