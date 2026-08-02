// Package gitstatus computes the same counts as `git status --porcelain=2`
// in-process, using go-git plumbing for object/index access and a custom
// worktree scan for stat-cache comparison. It never spawns a process.
//
// Load returns an error for every repository shape it does not support
// (sparse/split index, reftables, submodules, unreadable objects, ...). The
// caller must treat any error as "fall back to exec git" — Load never
// panics.
package gitstatus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// Detached is the branch name reported when HEAD is not on a branch. It must
// stay equal to segments.DETACHED; gitstatus cannot import the segments
// package (segments imports gitstatus), so the value is duplicated here.
const Detached = "(detached)"

// Counts mirrors what the git segment extracts from porcelain v2 for one
// side (working tree or staging area) of the status.
type Counts struct {
	Added     int
	Deleted   int
	Modified  int
	Untracked int
	Unmerged  int
}

// Result is the outcome of a successful Load.
type Result struct {
	// Hash is the full HEAD hash, or "(initial)" on an unborn branch.
	Hash string
	// Ref is the branch name, or Detached.
	Ref string
	// Upstream is "origin/main"-style, empty when not configured.
	Upstream     string
	Working      Counts
	Staging      Counts
	Ahead        int
	Behind       int
	UpstreamGone bool
}

// Options describes the on-disk locations Load needs. All fields are
// required except UntrackedMode.
type Options struct {
	WorktreeGitDir string // per-worktree git dir: HEAD, index, rebase state
	CommonGitDir   string // shared git dir: objects, refs, packed-refs, config
	RepoRoot       string // worktree root on disk
	UntrackedMode  string // "normal" (default when empty), "all", "no"
}

// Load computes the working tree and staging area status for the repository
// described by opts. Any error means the caller must fall back to exec git.
func Load(opts Options) (*Result, error) {
	defer log.Trace(time.Now(), opts.RepoRoot)

	untrackedMode := opts.UntrackedMode
	if untrackedMode == "" {
		untrackedMode = "normal"
	}
	if untrackedMode != "normal" && untrackedMode != "all" && untrackedMode != "no" {
		return nil, fmt.Errorf("gitstatus: unsupported untracked mode %q", untrackedMode)
	}

	idx, indexModTime, err := readIndex(opts.WorktreeGitDir)
	if err != nil {
		return nil, err
	}

	for i := range idx.Entries {
		if idx.Entries[i].Mode == modeGitlink {
			return nil, errors.New("gitstatus: submodule entries require exec fallback")
		}
	}

	result := &Result{UpstreamGone: true}

	// A missing or malformed config is not fatal on its own: it just means
	// there is no upstream to resolve and no repo-level excludesfile.
	cfg, _ := loadRepoConfig(opts.CommonGitDir)

	if err := checkRepoFormat(cfg); err != nil {
		return nil, err
	}

	store := newObjectStore(opts.CommonGitDir)
	defer store.close()

	headHash, headOK, err := resolveBranch(opts, cfg, store, result)
	if err != nil {
		return nil, err
	}

	basePatterns := loadBasePatterns(opts, cfg)
	scanWorktree(opts, idx, indexModTime, untrackedMode, trustExecutableBit(cfg), basePatterns, result)

	if err := diffStaging(store, idx, headHash, headOK, result); err != nil {
		return nil, err
	}

	return result, nil
}

// readIndex decodes the index file, recording its mtime beforehand so the
// worktree scan can detect racily-clean entries (files modified in the same
// timestamp tick the index was written).
func readIndex(worktreeGitDir string) (*gitIndex, time.Time, error) {
	indexPath := filepath.Join(worktreeGitDir, "index")

	fi, err := os.Stat(indexPath)
	if err != nil {
		return nil, time.Time{}, err
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, time.Time{}, err
	}

	idx, err := decodeIndex(data)
	if err != nil {
		return nil, time.Time{}, err
	}

	return idx, fi.ModTime(), nil
}
