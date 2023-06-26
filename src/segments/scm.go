package segments

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

const (
	// Fallback to native command
	NativeFallback properties.Property = "native_fallback"
	// Override the built-in status formats
	StatusFormats properties.Property = "status_formats"
)

// ScmStatus represents part of the status of a repository
type ScmStatus struct {
	Unmerged   int
	Deleted    int
	Added      int
	Modified   int
	Moved      int
	Conflicted int
	Untracked  int
	Clean      int
	Missing    int
	Ignored    int

	Formats map[string]string
}

func (s *ScmStatus) Changed() bool {
	return s.Unmerged > 0 ||
		s.Added > 0 ||
		s.Deleted > 0 ||
		s.Modified > 0 ||
		s.Moved > 0 ||
		s.Conflicted > 0 ||
		s.Untracked > 0 ||
		s.Clean > 0 ||
		s.Missing > 0 ||
		s.Ignored > 0
}

func (s *ScmStatus) String() string {
	var status strings.Builder

	if s.Formats == nil {
		s.Formats = make(map[string]string)
	}

	stringIfValue := func(value int, name, prefix string) {
		if value <= 0 {
			return
		}

		// allow user override for prefix
		if _, ok := s.Formats[name]; ok {
			status.WriteString(fmt.Sprintf(s.Formats[name], value))
			return
		}

		status.WriteString(fmt.Sprintf(" %s%d", prefix, value))
	}

	stringIfValue(s.Untracked, "Untracked", "?")
	stringIfValue(s.Added, "Added", "+")
	stringIfValue(s.Modified, "Modified", "~")
	stringIfValue(s.Deleted, "Deleted", "-")
	stringIfValue(s.Moved, "Moved", ">")
	stringIfValue(s.Unmerged, "Unmerged", "x")
	stringIfValue(s.Conflicted, "Conflicted", "!")
	stringIfValue(s.Missing, "Missing", "!")
	stringIfValue(s.Clean, "Clean", "=")
	stringIfValue(s.Ignored, "Ignored", "Ø")

	return strings.TrimSpace(status.String())
}

type scm struct {
	props properties.Properties
	env   platform.Environment

	IsWslSharedPath bool
	CommandMissing  bool
	Dir             string // actual repo root directory
	RepoName        string

	workingDir string
	rootDir    string
	realDir    string // real directory (can be different from current path when in worktrees)
	command    string
}

const (
	// BranchMaxLength truncates the length of the branch name
	BranchMaxLength properties.Property = "branch_max_length"
	// TruncateSymbol appends the set symbol to a truncated branch name
	TruncateSymbol properties.Property = "truncate_symbol"
	// FullBranchPath displays the full path of a branch
	FullBranchPath properties.Property = "full_branch_path"
)

func (s *scm) Init(props properties.Properties, env platform.Environment) {
	s.props = props
	s.env = env
}

func (s *scm) truncateBranch(branch string) string {
	fullBranchPath := s.props.GetBool(FullBranchPath, true)
	maxLength := s.props.GetInt(BranchMaxLength, 0)
	if !fullBranchPath && strings.Contains(branch, "/") {
		index := strings.LastIndex(branch, "/")
		branch = branch[index+1:]
	}
	if maxLength == 0 || len(branch) <= maxLength {
		return branch
	}
	symbol := s.props.GetString(TruncateSymbol, "")
	return branch[0:maxLength] + symbol
}

func (s *scm) shouldIgnoreRootRepository(rootDir string) bool {
	excludedFolders := s.props.GetStringArray(properties.ExcludeFolders, []string{})
	if len(excludedFolders) == 0 {
		return false
	}
	return s.env.DirMatchesOneOf(rootDir, excludedFolders)
}

func (s *scm) FileContents(folder, file string) string {
	return strings.Trim(s.env.FileContent(folder+"/"+file), " \r\n")
}

func (s *scm) convertToWindowsPath(path string) string {
	if s.env.GOOS() == platform.WINDOWS || s.IsWslSharedPath {
		return s.env.ConvertToWindowsPath(path)
	}
	return path
}

func (s *scm) convertToLinuxPath(path string) string {
	if !s.IsWslSharedPath {
		return path
	}
	return s.env.ConvertToLinuxPath(path)
}

func (s *scm) hasCommand(command string) bool {
	if len(s.command) > 0 {
		return true
	}
	// when in a WSL shared folder, we must use command.exe and convert paths accordingly
	// for worktrees, stashes, and path to work
	s.IsWslSharedPath = s.env.InWSLSharedDrive()
	if s.env.GOOS() == platform.WINDOWS || s.IsWslSharedPath {
		command += ".exe"
	}
	if s.env.HasCommand(command) {
		s.command = command
		return true
	}
	s.CommandMissing = true
	// only use the native fallback when set by the user
	if s.IsWslSharedPath && s.props.GetBool(NativeFallback, false) {
		command = strings.TrimSuffix(command, ".exe")
		if s.env.HasCommand(command) {
			s.command = command
			return true
		}
	}
	return false
}
