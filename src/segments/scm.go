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
	NativeFallback options.Option = "native_fallback"
	StatusFormats  options.Option = "status_formats"

	// workingField is the template field name shared by every SCM status
	// unit table below (see FieldRefs).
	workingField = "Working"
)

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

// String composes the status from config formats (status_formats) and counts,
// so the result is trusted markup: the `>` moved-prefix and any anchors in a
// configured format must survive rendering.
func (s *ScmStatus) String() template.Markup {
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

	return template.RawMarkup(strings.TrimSpace(status.String()))
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

// formatBranch returns the branch name as markup: the branch itself is
// untrusted VCS data (a repository controls it), so its chevrons are escaped,
// while mapped_branches values and the branch_template render are user
// configuration and keep their anchors.
func (s *Scm) formatBranch(branch string) template.Markup {
	mappedBranches := s.options.KeyValueMap(MappedBranches, make(map[string]string))

	// sort the keys alphabetically
	keys := make([]string, 0, len(mappedBranches))
	for k := range mappedBranches {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	const wildcard = "*"

	display := template.EscapeMarkup(branch)

	for _, key := range keys {
		mappedBranch := mappedBranches[key]

		if key == wildcard || branch == key {
			display = template.RawMarkup(mappedBranch)
			break
		}

		matchSubFolders := strings.HasSuffix(key, wildcard)
		subfolderKey := strings.TrimSuffix(key, wildcard)

		if matchSubFolders && strings.HasPrefix(branch, subfolderKey) {
			remainder := strings.TrimPrefix(branch, subfolderKey)
			display = template.JoinMarkup(template.RawMarkup(mappedBranch), template.EscapeMarkup(remainder))
			break
		}
	}

	branchTemplate := s.options.String(BranchTemplate, "")
	if branchTemplate == "" {
		return display
	}

	context := struct {
		Branch   template.Markup
		Upstream string
	}{Branch: display, Upstream: s.Upstream}

	txt, err := template.RenderTrusted(branchTemplate, context)
	if err != nil {
		return display
	}

	return template.RawMarkup(txt)
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
