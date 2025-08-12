package segments

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

const (
	JUJUTSUCOMMAND = "jj"

	IgnoreWorkingCopy properties.Property = "ignore_working_copy"
	MinChangeIDLen    properties.Property = "min_change_id_len"
	MinCommitIDLen    properties.Property = "min_commit_id_len"
)

// JujutsuID is used for both change_id and commit_id
type JujutsuID struct {
	Shortest string
	Full     string
	Rest     string
}

func (id *JujutsuID) String() string { return id.Shortest + id.Rest }

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
	Working         *JujutsuStatus
	ChangeID        string
	Working         JujutsuStatus
	ChangeID        JujutsuID
	CommitID        JujutsuID
	LocalBookmarks  []string
	RemoteBookmarks []string
	Description     string

	Conflict  bool
	Immutable bool
	Empty     bool
	Divergent bool
	Hidden    bool
	Mine      bool

	AuthorID    User // defined in git.go
	CommitterID User // defined in git.go

	Scm
}

func (jj *Jujutsu) Template() string {
	return " \uf1fa{{ .ChangeID }}{{if .Working.Changed}} \uf044 {{ .Working.String }}{{ end }} "
}

func (jj *Jujutsu) Enabled() bool {
	displayStatus := jj.props.GetBool(FetchStatus, false)

	if !jj.shouldDisplay(displayStatus) {
		return false
	}

	statusFormats := jj.props.GetKeyValueMap(StatusFormats, nil)
	jj.Working = JujutsuStatus{ScmStatus: ScmStatus{Formats: statusFormats}}

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

type jjTemplate struct {
	template string
	action   func(jj *Jujutsu, result string)
}

func (jj *Jujutsu) jjTemplates() []jjTemplate {
	minChangeIDLen := jj.props.GetString(MinChangeIDLen, "8")
	minCommitIDLen := jj.props.GetString(MinCommitIDLen, "8")

	// https://jj-vcs.github.io/jj/latest/templates/
	return []jjTemplate{
		{
			template: "diff.summary()",
			action: func(jj *Jujutsu, v string) {
				for _, line := range strings.Split(v, "\n") {
					if len(line) == 0 {
						continue
					}
					jj.Working.add(line[0])
				}
			},
		},
		{
			template: "change_id",
			action:   func(jj *Jujutsu, v string) { jj.ChangeID.Full = v },
		},
		{
			template: fmt.Sprintf("change_id.shortest(%s).prefix()", minChangeIDLen),
			action:   func(jj *Jujutsu, v string) { jj.ChangeID.Shortest = v },
		},
		{
			template: fmt.Sprintf("change_id.shortest(%s).rest()", minChangeIDLen),
			action:   func(jj *Jujutsu, v string) { jj.ChangeID.Rest = v },
		},
		{
			template: "commit_id",
			action:   func(jj *Jujutsu, v string) { jj.CommitID.Full = v },
		},
		{
			template: fmt.Sprintf("commit_id.shortest(%s).prefix()", minCommitIDLen),
			action:   func(jj *Jujutsu, v string) { jj.CommitID.Shortest = v },
		},
		{
			template: fmt.Sprintf("commit_id.shortest(%s).rest()", minCommitIDLen),
			action:   func(jj *Jujutsu, v string) { jj.CommitID.Rest = v },
		},
		{
			template: "local_bookmarks.join('\n')",
			action: func(jj *Jujutsu, v string) {
				if strings.TrimSpace(v) == "" {
					return
				}
				jj.LocalBookmarks = strings.Split(v, "\n")
			},
		},
		{
			template: "remote_bookmarks.join('\n')",
			action: func(jj *Jujutsu, v string) {
				if strings.TrimSpace(v) == "" {
					return
				}
				jj.RemoteBookmarks = strings.Split(v, "\n")
			},
		},
		{
			template: "divergent",
			action:   func(jj *Jujutsu, v string) { jj.Divergent, _ = strconv.ParseBool(v) },
		},
		{
			template: "hidden",
			action:   func(jj *Jujutsu, v string) { jj.Hidden, _ = strconv.ParseBool(v) },
		},
		{
			template: "immutable",
			action:   func(jj *Jujutsu, v string) { jj.Immutable, _ = strconv.ParseBool(v) },
		},
		{
			template: "empty",
			action:   func(jj *Jujutsu, v string) { jj.Empty, _ = strconv.ParseBool(v) },
		},
		{
			template: "mine",
			action:   func(jj *Jujutsu, v string) { jj.Mine, _ = strconv.ParseBool(v) },
		},
		{
			template: "author().name",
			action:   func(jj *Jujutsu, v string) { jj.AuthorID.Name = v },
		},
		{
			template: "author().email",
			action:   func(jj *Jujutsu, v string) { jj.AuthorID.Email = v },
		},
		{
			template: "commiter().name",
			action:   func(jj *Jujutsu, v string) { jj.CommitterID.Name = v },
		},
		{
			template: "commiter().email",
			action:   func(jj *Jujutsu, v string) { jj.CommitterID.Email = v },
		},
	}
}

// logTemplate will create a jj log template string with each template separated
// by a newline.  Returns the template string and the keys in the order they
// were specified.
func logTemplate(templates []jjTemplate) string {
	var sb strings.Builder
	for i, tmpl := range templates {
		sb.WriteString(tmpl.template)
		if i < len(templates)-1 {
			sb.WriteString(` ++ "\0" ++ `)
		}
	}
	return sb.String()
}

func (jj *Jujutsu) setJujutsuStatus() {
	templates := jj.jjTemplates()
	fmt.Println(templates)

	statusString, err := jj.getJujutsuCommandOutput("log", "-r", "@", "--no-graph", "-T", logTemplate(templates))
	if err != nil {
		return
	}

	statusString = strings.TrimSuffix(statusString, "\n")
	for i, result := range strings.Split(statusString, string(rune(0))) {
		templates[i].action(jj, result)
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
