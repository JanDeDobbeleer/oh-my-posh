package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

// SaplingStatus represents part of the status of a Sapling repository
type SaplingStatus struct {
	ScmStatus
}

func (s *SaplingStatus) add(code string) {
	// M = modified
	// A = added
	// R = removed/deleted
	// C = clean
	// ! = missing (deleted by a non-sl command, but still tracked)
	// ? = not tracked
	// I = ignored
	//   = origin of the previous file (with --copies)
	switch code {
	case "M":
		s.Modified++
	case "A":
		s.Added++
	case "R":
		s.Deleted++
	case "C":
		s.Clean++
	case "!":
		s.Missing++
	case "?":
		s.Untracked++
	case "I":
		s.Ignored++
	}
}

const (
	SAPLINGCOMMAND   = "sl"
	SLCOMMITTEMPLATE = "no:{node}\nns:{sl_node}\nnd:{sl_date}\nun:{sl_user}\nbm:{activebookmark}\ndn:{desc|firstline}"
)

type Sapling struct {
	scm

	ShortHash   string
	Hash        string
	When        string
	Author      string
	Bookmark    string
	Description string

	Working *SaplingStatus
}

func (sl *Sapling) Template() string {
	return " {{ if .Bookmark }}\uf097 {{ .Bookmark }}*{{ else }}\ue729 {{ .ShortHash }}{{ end }}{{ if .Working.Changed }} \uf044 {{ .Working.String }}{{ end }} "
}

func (sl *Sapling) Enabled() bool {
	if !sl.shouldDisplay() {
		return false
	}

	sl.setHeadContext()

	return true
}

func (sl *Sapling) shouldDisplay() bool {
	if !sl.hasCommand(SAPLINGCOMMAND) {
		return false
	}

	slDir, err := sl.env.HasParentFilePath(".sl")
	if err != nil {
		return false
	}

	if sl.shouldIgnoreRootRepository(slDir.ParentFolder) {
		return false
	}

	sl.workingDir = slDir.Path
	sl.rootDir = slDir.Path
	// convert the worktree file path to a windows one when in a WSL shared folder
	sl.realDir = strings.TrimSuffix(sl.convertToWindowsPath(slDir.Path), "/.sl")
	sl.RepoName = platform.Base(sl.env, sl.convertToLinuxPath(sl.realDir))
	sl.setDir(slDir.Path)
	return true
}

func (sl *Sapling) setDir(dir string) {
	dir = platform.ReplaceHomeDirPrefixWithTilde(sl.env, dir) // align with template PWD
	if sl.env.GOOS() == platform.WINDOWS {
		sl.Dir = strings.TrimSuffix(dir, `\.sl`)
		return
	}
	sl.Dir = strings.TrimSuffix(dir, "/.sl")
}

func (sl *Sapling) setHeadContext() {
	sl.setCommitContext()

	sl.Working = &SaplingStatus{}

	displayStatus := sl.props.GetBool(FetchStatus, true)
	if !displayStatus {
		return
	}

	changes := sl.getSaplingCommandOutput("status")
	if len(changes) == 0 {
		return
	}
	lines := strings.Split(changes, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		// element is the element from someSlice for where we are
		sl.Working.add(line[0:1])
	}
}

func (sl *Sapling) setCommitContext() {
	body := sl.getSaplingCommandOutput("log", "--limit", "1", "--template", SLCOMMITTEMPLATE)
	splitted := strings.Split(strings.TrimSpace(body), "\n")
	for _, line := range splitted {
		line = strings.TrimSpace(line)
		if len(line) <= 3 {
			continue
		}
		anchor := line[:3]
		line = line[3:]
		switch anchor {
		case "no:":
			sl.Hash = line
		case "ns:":
			sl.ShortHash = line
		case "nd:":
			sl.When = line
		case "un:":
			sl.Author = line
		case "bm:":
			sl.Bookmark = line
		case "dn:":
			sl.Description = line
		}
	}
}

func (sl *Sapling) getSaplingCommandOutput(command string, args ...string) string {
	args = append([]string{command}, args...)
	val, err := sl.env.RunCommand(sl.command, args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(val)
}
