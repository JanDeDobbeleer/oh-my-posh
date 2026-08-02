package gitstatus

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"

	"github.com/jandedobbeleer/oh-my-posh/src/ini"
)

// loadBasePatterns returns the ignore patterns that apply repo-wide, before
// any per-directory .gitignore is layered on during the worktree walk:
// CommonGitDir/info/exclude, then the resolved global excludes file.
func loadBasePatterns(opts Options, cfg *ini.File) []gitignore.Pattern {
	var patterns []gitignore.Pattern

	if ps := readIgnoreFile(filepath.Join(opts.CommonGitDir, "info", "exclude"), nil); ps != nil {
		patterns = append(patterns, ps...)
	}

	if global := resolveGlobalExcludesFile(cfg); global != "" {
		if ps := readIgnoreFile(global, nil); ps != nil {
			patterns = append(patterns, ps...)
		}
	}

	return patterns
}

// resolveGlobalExcludesFile resolves core.excludesfile in the order git
// itself uses: the repo's own config, then ~/.gitconfig, then
// $XDG_CONFIG_HOME/git/config; falling back to the default
// $XDG_CONFIG_HOME/git/ignore (or ~/.config/git/ignore) location when none
// of those set it explicitly.
func resolveGlobalExcludesFile(cfg *ini.File) string {
	if cfg != nil {
		if v := cfg.Section("core").Key("excludesfile").String(); v != "" {
			return expandHome(v)
		}
	}

	home, homeErr := os.UserHomeDir()
	if homeErr == nil {
		if v := coreExcludesFileFromFile(filepath.Join(home, ".gitconfig")); v != "" {
			return expandHome(v)
		}
	}

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" && homeErr == nil {
		xdgConfigHome = filepath.Join(home, ".config")
	}

	if xdgConfigHome == "" {
		return ""
	}

	if v := coreExcludesFileFromFile(filepath.Join(xdgConfigHome, "git", "config")); v != "" {
		return expandHome(v)
	}

	return filepath.Join(xdgConfigHome, "git", "ignore")
}

func coreExcludesFileFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	cfg, err := ini.Load(string(data))
	if err != nil {
		return ""
	}

	return cfg.Section("core").Key("excludesfile").String()
}

func expandHome(p string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}

	if p == "~" {
		return home
	}

	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(home, rest)
	}

	return p
}

// splitDomain turns a walk-relative directory path into the domain slice
// gitignore.ParsePattern expects, so patterns from a nested .gitignore only
// match beneath that directory.
func splitDomain(rel string) []string {
	if rel == "" {
		return nil
	}
	return strings.Split(rel, "/")
}

// readIgnoreFile parses a gitignore-format file, returning nil when the file
// doesn't exist (a common, non-error case for every source in the chain).
func readIgnoreFile(path string, domain []string) []gitignore.Pattern {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var out []gitignore.Pattern
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, gitignore.ParsePattern(line, domain))
	}

	return out
}

// dirHasVisibleFile reports whether dir contains at least one non-ignored
// file anywhere beneath it, stopping at the first match. Used by the
// "normal" untracked mode to decide whether an untracked directory counts
// as a single untracked entry.
func dirHasVisibleFile(dir, rel string, ignored func(string, bool) bool) bool {
	found := false

	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return fs.SkipAll
		}

		r, _ := filepath.Rel(dir, p)
		if r == "." {
			return nil
		}

		childRel := rel + "/" + filepath.ToSlash(r)
		if d.IsDir() {
			if ignored(childRel, true) {
				return fs.SkipDir
			}
			return nil
		}

		if !ignored(childRel, false) {
			found = true
			return fs.SkipAll
		}

		return nil
	})

	return found
}
