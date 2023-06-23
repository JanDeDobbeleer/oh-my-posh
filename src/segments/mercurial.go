package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

const (
	MERCURIALCOMMAND = "hg"

	hgLogTemplate = "{rev}|{node}|{branch}|{tags}|{bookmarks}"
)

type MercurialStatus struct {
	ScmStatus
}

func (s *MercurialStatus) add(code string) {
	switch code {
	case "R", "!":
		s.Deleted++
	case "A":
		s.Added++
	case "?":
		s.Untracked++
	case "M":
		s.Modified++
	}
}

type Mercurial struct {
	scm

	Working           *MercurialStatus
	IsTip             bool
	LocalCommitNumber string
	ChangeSetID       string
	ChangeSetIDShort  string
	Branch            string
	Bookmarks         []string
	Tags              []string
}

func (hg *Mercurial) Template() string {
	return "hg {{.Branch}} {{if .LocalCommitNumber}}({{.LocalCommitNumber}}:{{.ChangeSetIDShort}}){{end}}{{range .Bookmarks }} \uf02e {{.}}{{end}}{{range .Tags}} \uf02b {{.}}{{end}}{{if .Working.Changed}} \uf044 {{ .Working.String }}{{ end }}" //nolint: lll
}

func (hg *Mercurial) Enabled() bool {
	if !hg.shouldDisplay() {
		return false
	}

	hg.Working = &MercurialStatus{}

	displayStatus := hg.props.GetBool(FetchStatus, false)
	if displayStatus {
		hg.setMercurialStatus()
	}

	return true
}

func (hg *Mercurial) shouldDisplay() bool {
	if !hg.hasCommand(MERCURIALCOMMAND) {
		return false
	}

	hgdir, err := hg.env.HasParentFilePath(".hg")
	if err != nil {
		return false
	}

	if hg.shouldIgnoreRootRepository(hgdir.ParentFolder) {
		return false
	}

	hg.setDir(hgdir.ParentFolder)

	hg.workingDir = hgdir.Path
	hg.rootDir = hgdir.Path
	// convert the worktree file path to a windows one when in a WSL shared folder
	hg.realDir = strings.TrimSuffix(hg.convertToWindowsPath(hgdir.Path), "/.hg")
	return true
}

func (hg *Mercurial) setDir(dir string) {
	dir = platform.ReplaceHomeDirPrefixWithTilde(hg.env, dir) // align with template PWD
	if hg.env.GOOS() == platform.WINDOWS {
		hg.Dir = strings.TrimSuffix(dir, `\.hg`)
		return
	}
	hg.Dir = strings.TrimSuffix(dir, "/.hg")
}

func (hg *Mercurial) setMercurialStatus() {
	hg.Branch = hg.command

	idString := hg.getHgCommandOutput("log", "-r", ".", "--template", hgLogTemplate)
	if len(idString) == 0 {
		return
	}

	idSplit := strings.Split(idString, "|")
	if len(idSplit) != 5 {
		return
	}

	hg.LocalCommitNumber = idSplit[0]
	hg.ChangeSetID = idSplit[1]

	if len(hg.ChangeSetID) >= 12 {
		hg.ChangeSetIDShort = hg.ChangeSetID[:12]
	}
	hg.Branch = idSplit[2]

	hg.Tags = doSplit(idSplit[3])
	hg.Bookmarks = doSplit(idSplit[4])

	hg.IsTip = false
	tipIndex := 0
	for i, tag := range hg.Tags {
		if tag == "tip" {
			hg.IsTip = true
			tipIndex = i
			break
		}
	}

	if hg.IsTip {
		hg.Tags = RemoveAtIndex(hg.Tags, tipIndex)
	}

	statusString := hg.getHgCommandOutput("status")

	if len(statusString) == 0 {
		return
	}

	statusLines := strings.Split(statusString, "\n")

	for _, status := range statusLines {
		hg.Working.add(status[:1])
	}
}

func doSplit(s string) []string {
	if len(s) == 0 {
		return []string{}
	}

	return strings.Split(s, " ")
}

func RemoveAtIndex(s []string, index int) []string {
	ret := make([]string, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func (hg *Mercurial) getHgCommandOutput(command string, args ...string) string {
	args = append([]string{"-R", hg.realDir, command}, args...)
	val, err := hg.env.RunCommand(hg.command, args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(val)
}
