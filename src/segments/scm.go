package segments

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

const (
	// Fallback to native command
	NativeFallback options.Option = "native_fallback"
	// Override the built-in status formats
	StatusFormats options.Option = "status_formats"
)

// ScmStatus represents part of the status of a repository
type ScmStatus struct {
	Formats    map[string]string
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
	status := text.NewBuilder()

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

type Scm struct {
	Base

	Dir             string
	RepoName        string
	Upstream        string
	mainSCMDir      string
	scmDir          string
	repoRootDir     string
	command         string
	IsWslSharedPath bool
	CommandMissing  bool
	nativeFallback  bool
}

const (
	// BranchTemplate allows to specify a template for the branch name
	BranchTemplate options.Option = "branch_template"
)

func (s *Scm) RelativeDir() string {
	if s.repoRootDir == "" {
		return ""
	}

	pwd := s.env.Pwd()
	log.Debug("repo root dir:", s.repoRootDir, "pwd:", pwd)

	rel, err := filepath.Rel(s.repoRootDir, pwd)
	if err != nil {
		log.Error(err)
	}

	if rel == "." || rel == "" {
		log.Debug("repo root dir is the same as the current working directory, returning empty string")
		return ""
	}

	return rel
}

func (s *Scm) formatBranch(branch string) string {
	mappedBranches := s.options.KeyValueMap(MappedBranches, make(map[string]string))

	// sort the keys alphabetically
	keys := make([]string, 0, len(mappedBranches))
	for k := range mappedBranches {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	const wildcard = "*"

	for _, key := range keys {
		if key == wildcard {
			branch = mappedBranches[key]
			break
		}

		matchSubFolders := strings.HasSuffix(key, wildcard)
		subfolderKey := strings.TrimSuffix(key, wildcard)

		if matchSubFolders && strings.HasPrefix(branch, subfolderKey) {
			branch = strings.Replace(branch, subfolderKey, mappedBranches[key], 1)
			break
		}

		if matchSubFolders || branch != key {
			continue
		}

		branch = strings.Replace(branch, key, mappedBranches[key], 1)
		break
	}

	branchTemplate := s.options.String(BranchTemplate, "")
	if branchTemplate == "" {
		return branch
	}

	txt, err := template.Render(branchTemplate, struct{ Branch, Upstream string }{Branch: branch, Upstream: s.Upstream})
	if err != nil {
		return branch
	}

	return txt
}

func (s *Scm) fileContent(folder, file string) string {
	return strings.Trim(s.env.FileContent(folder+"/"+file), " \r\n")
}

func (s *Scm) convertToWindowsPath(path string) string {
	// only convert when in Windows, or when in a WSL shared folder and not using the native fallback
	if s.env.GOOS() == runtime.WINDOWS || (s.IsWslSharedPath && !s.nativeFallback) {
		return s.env.ConvertToWindowsPath(path)
	}

	return path
}

func (s *Scm) convertToLinuxPath(path string) string {
	if !s.IsWslSharedPath {
		return path
	}

	return s.env.ConvertToLinuxPath(path)
}

func (s *Scm) hasCommand(command string) bool {
	if len(s.command) > 0 {
		return true
	}

	// when in a WSL shared folder, we must use command.exe and convert paths accordingly
	// for worktrees, stashes, and path to work, except when native_fallback is set
	s.IsWslSharedPath = s.env.InWSLSharedDrive()
	if s.env.GOOS() == runtime.WINDOWS || s.IsWslSharedPath {
		command += ".exe"
	}

	if s.env.HasCommand(command) {
		s.command = command
		return true
	}

	s.CommandMissing = true

	// only use the native fallback when set by the user
	if s.IsWslSharedPath && s.options.Bool(NativeFallback, false) {
		command = strings.TrimSuffix(command, ".exe")
		if s.env.HasCommand(command) {
			s.command = command
			s.nativeFallback = true
			return true
		}
	}

	return false
}
