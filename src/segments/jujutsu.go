package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

const (
	JUJUTSUCOMMAND = "jj"

	jjLogTemplate = `change_id.shortest() ++ "\n" ++ diff.summary()`

	IgnoreWorkingCopy properties.Property = "ignore_working_copy"
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
	scm
}

func (jj *Jujutsu) Template() string {
	return " \uf1fa{{.ChangeID}}{{if .Working.Changed}} \uf044 {{ .Working.String }}{{ end }} "
}

func (jj *Jujutsu) Enabled() bool {
	displayStatus := jj.props.GetBool(FetchStatus, false)

	if !jj.shouldDisplay(displayStatus) {
		return false
	}

	statusFormats := jj.props.GetKeyValueMap(StatusFormats, map[string]string{})
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
	// https://jj-vcs.github.io/jj/latest/templates/#commit-keywords
	statusString, err := jj.getJujutsuCommandOutput("log", "-r", "@", "--no-graph", "-T", jjLogTemplate)
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

func (jj *Jujutsu) getJujutsuCommandOutput(command string, args ...string) (string, error) {
	cli := []string{"--repository", jj.repoRootDir, "--no-pager", "--color", "never"}

	if jj.props.GetBool(IgnoreWorkingCopy, true) {
		cli = append(cli, "--ignore-working-copy")
	}

	cli = append(cli, command)
	cli = append(cli, args...)

	return jj.env.RunCommand(jj.command, cli...)
}
