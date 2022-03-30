package segments

import (
	"oh-my-posh/regex"
	"strconv"
	"strings"
)

// SvnStatus represents part of the status of a Svn repository
type SvnStatus struct {
	ScmStatus
}

func (s *SvnStatus) add(code string) {
	switch code {
	case "C":
		s.Conflicted++
	case "D":
		s.Deleted++
	case "A":
		s.Added++
	case "M":
		s.Modified++
	case "R":
		s.Moved++
	default:
		s.Unmerged++
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
	displayStatus := s.props.GetBool(FetchStatus, false)
	if displayStatus {
		s.setSvnStatus()
	} else {
		s.Working = &SvnStatus{}
	}
	return true
}

func (s *Svn) shouldDisplay() bool {
	// when in wsl/wsl2 and in a windows shared folder
	// we must use Svn.exe and convert paths accordingly
	// for worktrees, stashes, and path to work
	s.IsWslSharedPath = s.env.InWSLSharedDrive()
	if !s.env.HasCommand(s.getCommand(SVNCOMMAND)) {
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
		s.workingFolder = Svndir.Path
		s.rootFolder = Svndir.Path
		// convert the worktree file path to a windows one when in wsl 2 shared folder
		s.realFolder = strings.TrimSuffix(s.convertToWindowsPath(Svndir.Path), ".svn")
		return true
	}
	// handle worktree
	s.rootFolder = Svndir.Path
	dirPointer := strings.Trim(s.env.FileContent(Svndir.Path), " \r\n")
	matches := regex.FindNamedRegexMatch(`^Svndir: (?P<dir>.*)$`, dirPointer)
	if matches != nil && matches["dir"] != "" {
		// if we open a worktree file in a shared wsl2 folder, we have to convert it back
		// to the mounted path
		s.workingFolder = s.convertToLinuxPath(matches["dir"])
		return false
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
	args = append([]string{command, s.realFolder}, args...)
	val, err := s.env.RunCommand(s.getCommand(SVNCOMMAND), args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(val)
}
