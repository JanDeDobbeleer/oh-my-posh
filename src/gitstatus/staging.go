package gitstatus

import (
	"github.com/go-git/go-git/v5/plumbing"
)

// diffStaging compares the index against the HEAD tree to compute the
// staging-side counts: cache-tree fast path first, then a full tree diff.
func diffStaging(store *objectStore, idx *gitIndex, headHash plumbing.Hash, headOK bool, result *Result) error {
	headFiles := map[string]plumbing.Hash{}

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

		err = walkTree(store, commit.Tree, "", func(path string, blob plumbing.Hash) {
			headFiles[path] = blob
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

		headEntryHash, inHead := headFiles[e.Name]
		if !inHead {
			result.Staging.Added++
			addedHashes = append(addedHashes, e.Hash)
			continue
		}

		if headEntryHash != e.Hash {
			result.Staging.Modified++
		}
		delete(headFiles, e.Name)
	}

	for _, h := range headFiles {
		result.Staging.Deleted++
		deletedHashes = append(deletedHashes, h)
	}

	pairRenames(&result.Staging, addedHashes, deletedHashes)
	return nil
}

func cacheTreeMatches(idx *gitIndex, headTree plumbing.Hash) bool {
	root := idx.CacheTreeRoot
	return root != nil && root.Path == "" && root.Entries >= 0 && root.Hash == headTree
}

// pairRenames mirrors git's exact-content rename detection: an Added path
// and a Deleted path with an identical blob hash are really one rename,
// which git.go's add() counts as a single Modified.
func pairRenames(counts *Counts, added, deleted []plumbing.Hash) {
	remaining := make(map[plumbing.Hash]int, len(deleted))
	for _, h := range deleted {
		remaining[h]++
	}

	for _, h := range added {
		if remaining[h] <= 0 {
			continue
		}

		remaining[h]--
		counts.Added--
		counts.Deleted--
		counts.Modified++
	}
}
