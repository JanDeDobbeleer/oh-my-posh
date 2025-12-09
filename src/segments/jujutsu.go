package segments

import (
	"fmt"
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

type Jujutsu struct {
	Working  *JujutsuStatus
	ChangeID string
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

	lines := strings.Split(statusString, "\n")
	return lines[0]
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

	lines := strings.Split(statusString, "\n")
	jj.ChangeID = lines[0]

	for _, line := range lines[1:] {
		if len(line) > 0 {
			jj.Working.add(line[0])
		}
	}
}

func (jj *Jujutsu) logTemplate() string {
	// https://jj-vcs.github.io/jj/latest/templates/#commit-keywords
	return fmt.Sprintf(`change_id.shortest(%d) ++ "\n" ++ diff.summary()`, jj.options.Int(ChangeIDMinLen, 0))
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
