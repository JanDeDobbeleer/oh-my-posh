package gitstatus

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// mergedStage is the on-disk stage value of a fully merged index entry.
// go-git's index.Merged constant is defined as 1, but the decoder actually
// extracts the raw on-disk stage bits (0-3) into Entry.Stage, where 0 means
// "no conflict". We compare against that zero value directly rather than
// the (seemingly mislabeled) named constant.
const mergedStage index.Stage = 0

// diffStaging compares the index against the HEAD tree to compute the
// staging-side counts: cache-tree fast path first, then a full tree diff.
func diffStaging(store storer.EncodedObjectStorer, idx *index.Index, headHash plumbing.Hash, headOK bool, result *Result) error {
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
	for _, e := range idx.Entries {
		if e.Stage == mergedStage {
			continue
		}
		unmerged[e.Name] = true
		delete(headFiles, e.Name)
	}

	var addedHashes, deletedHashes []plumbing.Hash

	for _, e := range idx.Entries {
		if e.Stage != mergedStage || e.IntentToAdd || unmerged[e.Name] {
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

func cacheTreeMatches(idx *index.Index, headTree plumbing.Hash) bool {
	if idx.Cache == nil || len(idx.Cache.Entries) == 0 {
		return false
	}

	root := idx.Cache.Entries[0]
	return root.Path == "" && root.Entries >= 0 && root.Hash == headTree
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
