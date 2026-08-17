package runtime

import (
	"path/filepath"
	"slices"
)

// Activation describes the cheap preconditions under which a segment can
// possibly be enabled, letting the engine skip the (potentially expensive)
// Enabled() probe when none of them hold in the current working directory.
//
// Contract:
//   - Always short-circuits: the segment always executes.
//   - Otherwise conditions are OR'd across ALL kinds: any single match keeps
//     the segment alive. A matching gate does not imply the segment enables;
//     Enabled() still decides. A failing gate MUST imply Enabled() would
//     return false, so a condition list has to be a superset of the triggers
//     the segment's own detection reacts to.
//   - The zero value (no conditions at all) means "no gate - always execute".
//
// Evaluation is deliberately cheap: FileGlobs run against the cached
// directory listing, Folders are a single stat each, EnvVars a getenv, and
// ProjectFiles use the per-invocation HasParentFilePath cache so a later
// Enabled() doing the same search does not pay twice.
type Activation struct {
	// FileGlobs: the cwd contains a file matching one of these globs
	// (HasFiles semantics: cached readdir, case-insensitive, files only).
	FileGlobs []string
	// Folders: the cwd contains one of these directories (stat-based).
	Folders []string
	// ProjectFiles: an upward parent-directory search (cwd included) finds an
	// entry with one of these exact names (HasParentFilePath semantics; both
	// the symlink-following and non-following variants are consulted, so the
	// gate covers writers using either).
	ProjectFiles []string
	// EnvVars: one of these environment variables is non-empty.
	EnvVars []string
	Always  bool
}

// Active reports whether any of the activation's conditions holds. See the
// type documentation for the contract.
func (a *Activation) Active(env Environment) bool {
	if a.Always {
		return true
	}

	// the zero value carries no gate
	if len(a.FileGlobs) == 0 && len(a.Folders) == 0 && len(a.ProjectFiles) == 0 && len(a.EnvVars) == 0 {
		return true
	}

	if slices.ContainsFunc(a.FileGlobs, env.HasFiles) {
		return true
	}

	for _, folder := range a.Folders {
		if env.HasFolder(filepath.Join(env.Pwd(), folder)) {
			return true
		}
	}

	for _, name := range a.ProjectFiles {
		// both variants: writers search with and without symlink resolution,
		// and neither result is a superset of the other. The per-invocation
		// cache makes the double walk (and the writer's own later search)
		// cheap.
		if _, err := env.HasParentFilePath(name, false); err == nil {
			return true
		}

		if _, err := env.HasParentFilePath(name, true); err == nil {
			return true
		}
	}

	for _, envVar := range a.EnvVars {
		if env.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}
