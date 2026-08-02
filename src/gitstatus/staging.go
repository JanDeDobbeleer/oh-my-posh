package gitstatus

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing"
)

// headEntry is one blob in the HEAD tree: its hash plus the normalized file
// mode, so staged mode-only changes (chmod + git add) are visible.
type headEntry struct {
	hash plumbing.Hash
	mode uint32
}

// diffStaging compares the index against the HEAD tree to compute the
// staging-side counts: cache-tree fast path first, then a full tree diff.
func diffStaging(store *objectStore, idx *gitIndex, headHash plumbing.Hash, headOK bool, result *Result) error {
	headFiles := map[string]headEntry{}

	if headOK {
		commit, err := readCommit(store, headHash)
		if err != nil {
			return err
		}

		// Fast path: a cache-tree root that matches the HEAD tree means the
		// index and HEAD are identical, so nothing is staged.
		if cacheTreeMatches(idx, commit.Tree) {
			return nil
		}

		err = walkTree(store, commit.Tree, "", func(path string, blob plumbing.Hash, mode uint32) {
			headFiles[path] = headEntry{hash: blob, mode: normalizeMode(mode)}
		})
		if err != nil {
			return err
		}
	}

	unmerged := map[string]bool{}
	for i := range idx.Entries {
		e := &idx.Entries[i]
		if e.Stage == 0 {
			continue
		}
		unmerged[e.Name] = true
		delete(headFiles, e.Name)
	}

	var addedHashes, deletedHashes []plumbing.Hash

	for i := range idx.Entries {
		e := &idx.Entries[i]
		if e.Stage != 0 || e.IntentToAdd || unmerged[e.Name] {
			continue
		}

		head, inHead := headFiles[e.Name]
		if !inHead {
			result.Staging.Added++
			addedHashes = append(addedHashes, e.Hash)
			continue
		}

		// mode participates unconditionally: this compares two recorded
		// values (index vs tree), no filesystem stat involved, so
		// core.filemode does not apply — matching git diff --cached
		if head.hash != e.Hash || head.mode != normalizeMode(e.Mode) {
			result.Staging.Modified++
		}
		delete(headFiles, e.Name)
	}

	for _, h := range headFiles {
		result.Staging.Deleted++
		deletedHashes = append(deletedHashes, h.hash)
	}

	return pairRenames(&result.Staging, addedHashes, deletedHashes)
}

// normalizeMode reduces a recorded git mode to the canonical values status
// distinguishes: regular files collapse to 100644 or 100755 (historic modes
// like 100664 count as non-executable), symlinks and gitlinks pass through.
func normalizeMode(mode uint32) uint32 {
	if mode&0o170000 != 0o100000 {
		return mode
	}

	if mode&0o111 != 0 {
		return 0o100755
	}

	return 0o100644
}

func cacheTreeMatches(idx *gitIndex, headTree plumbing.Hash) bool {
	root := idx.CacheTreeRoot
	return root != nil && root.Path == "" && root.Entries >= 0 && root.Hash == headTree
}

// pairRenames mirrors git's exact-content rename detection: an Added path
// and a Deleted path with an identical blob hash are really one rename,
// which git.go's add() counts as a single Modified. When adds and deletes
// remain unpaired on both sides, git may still pair them through similarity
// detection this engine deliberately does not implement — computing counts
// here could silently diverge, so that shape falls back to exec git.
func pairRenames(counts *Counts, added, deleted []plumbing.Hash) error {
	remaining := make(map[plumbing.Hash]int, len(deleted))
	for _, h := range deleted {
		remaining[h]++
	}

	unpairedAdds := 0
	for _, h := range added {
		if remaining[h] <= 0 {
			unpairedAdds++
			continue
		}

		remaining[h]--
		counts.Added--
		counts.Deleted--
		counts.Modified++
	}

	unpairedDeletes := 0
	for _, n := range remaining {
		unpairedDeletes += n
	}

	if unpairedAdds > 0 && unpairedDeletes > 0 {
		return errors.New("gitstatus: staged adds and deletes may be renames, deferring to git")
	}

	return nil
}
