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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"github.com/go-git/go-billy/v5/osfs"
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
	Working      Counts
	Staging      Counts
	Hash         string // full HEAD hash, "(initial)" when unborn
	Ref          string // branch name, or Detached
	Upstream     string // "origin/main"-style, "" when not configured
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

	for _, e := range idx.Entries {
		if e.Mode == filemode.Submodule {
			return nil, errors.New("gitstatus: submodule entries require exec fallback")
		}
	}

	result := &Result{UpstreamGone: true}

	// A missing or malformed config is not fatal on its own: it just means
	// there is no upstream to resolve and no repo-level excludesfile.
	cfg, _ := loadRepoConfig(opts.CommonGitDir)

	fs := osfs.New(opts.CommonGitDir)
	store := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())

	headHash, headOK, err := resolveBranch(opts, cfg, store, result)
	if err != nil {
		return nil, err
	}

	basePatterns := loadBasePatterns(opts, cfg)
	scanWorktree(opts, idx, indexModTime, untrackedMode, basePatterns, result)

	if err := diffStaging(store, idx, headHash, headOK, result); err != nil {
		return nil, err
	}

	return result, nil
}

// readIndex decodes the index file, recording its mtime beforehand so the
// worktree scan can detect racily-clean entries (files modified in the same
// timestamp tick the index was written).
func readIndex(worktreeGitDir string) (*index.Index, time.Time, error) {
	indexPath := filepath.Join(worktreeGitDir, "index")

	fi, err := os.Stat(indexPath)
	if err != nil {
		return nil, time.Time{}, err
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, time.Time{}, err
	}

	idx := &index.Index{}
	err = index.NewDecoder(bytes.NewReader(data)).Decode(idx)
	// index.skipHash (git >= 2.40, default with feature.manyFiles) writes an
	// all-zero trailer instead of a checksum; the decoder has fully populated
	// idx by the time it rejects that, so accept it
	if errors.Is(err, index.ErrInvalidChecksum) && isZeroTrailer(data) {
		err = nil
	}

	if err != nil {
		return nil, time.Time{}, fmt.Errorf("gitstatus: decode index: %w", err)
	}

	return idx, fi.ModTime(), nil
}

func isZeroTrailer(data []byte) bool {
	const trailerLen = 20
	if len(data) < trailerLen {
		return false
	}

	for _, b := range data[len(data)-trailerLen:] {
		if b != 0 {
			return false
		}
	}

	return true
}
