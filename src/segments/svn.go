package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

// SvnStatus represents part of the status of a Svn repository
type SvnStatus struct {
	ScmStatus
}

func (s *SvnStatus) add(code string) {
	switch code {
	case "?":
		s.Untracked++
	case "C":
		s.Conflicted++
	case "D":
		s.Deleted++
	case "A":
		s.Added++
	case "M":
		s.Modified++
	case "R", "!":
		s.Moved++
	}
}

func (s *SvnStatus) HasConflicts() bool {
	return s.Conflicted > 0
}

const (
	SVNCOMMAND = "svn"
)

type Svn struct {
	scm

	Working *SvnStatus
	BaseRev int
	Branch  string
}

func (s *Svn) Template() string {
	return " \ue0a0{{.Branch}} r{{.BaseRev}} {{.Working.String}} "
}

func (s *Svn) Enabled() bool {
	if !s.shouldDisplay() {
		return false
	}

	s.setSvnStatus()

	return true
}

func (s *Svn) shouldDisplay() bool {
	if !s.hasCommand(SVNCOMMAND) {
		return false
	}
	Svndir, err := s.env.HasParentFilePath(".svn")
	if err != nil {
		return false
	}
	if s.shouldIgnoreRootRepository(Svndir.ParentFolder) {
		return false
	}

	if Svndir.IsDir {
		s.workingDir = Svndir.Path
		s.rootDir = Svndir.Path
		// convert the worktree file path to a windows one when in a WSL shared folder
		s.realDir = strings.TrimSuffix(s.convertToWindowsPath(Svndir.Path), "/.svn")
		return true
	}
	// handle worktree
	s.rootDir = Svndir.Path
	dirPointer := strings.Trim(s.env.FileContent(Svndir.Path), " \r\n")
	matches := regex.FindNamedRegexMatch(`^Svndir: (?P<dir>.*)$`, dirPointer)
	if matches != nil && matches["dir"] != "" {
		// if we open a worktree file in a WSL shared folder, we have to convert it back
		// to the mounted path
		s.workingDir = s.convertToLinuxPath(matches["dir"])
	}
	return false
}

func (s *Svn) setSvnStatus() {
	s.BaseRev, _ = strconv.Atoi(s.getSvnCommandOutput("info", "--show-item", "revision"))

	branch := s.getSvnCommandOutput("info", "--show-item", "relative-url")
	if len(branch) > 2 {
		s.Branch = branch[2:]
	}

	s.Working = &SvnStatus{}

	displayStatus := s.props.GetBool(FetchStatus, false)
	if !displayStatus {
		return
	}

	changes := s.getSvnCommandOutput("status")
	if len(changes) == 0 {
		return
	}
	lines := strings.Split(changes, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// element is the element from someSlice for where we are
		s.Working.add(line[0:1])
	}
}

func (s *Svn) getSvnCommandOutput(command string, args ...string) string {
	args = append([]string{command, s.realDir}, args...)
	val, err := s.env.RunCommand(s.command, args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(val)
}
