package gitstatus

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

// peelTag resolves an object hash to the commit it ultimately points to,
// following a chain of annotated tag objects. A hash that is already a
// commit is returned unchanged. ok is false for anything that doesn't
// bottom out at a commit (a tag of a tree or blob, a broken object, or a
// chain deep enough to smell like a cycle).
func peelTag(store *objectStore, h plumbing.Hash) (plumbing.Hash, bool) {
	const maxDepth = 10 // no realistic tag chain nests anywhere close to this

	for range maxDepth {
		kind, data, err := store.object(h)
		if err != nil {
			return plumbing.ZeroHash, false
		}

		if kind == kindCommit {
			return h, true
		}
		if kind != kindTag {
			return plumbing.ZeroHash, false
		}

		next, ok := parseTagObjectTarget(data)
		if !ok {
			return plumbing.ZeroHash, false
		}
		h = next
	}

	return plumbing.ZeroHash, false
}

// parseTagObjectTarget reads the "object <hash>" header line of an
// annotated tag object.
func parseTagObjectTarget(data []byte) (plumbing.Hash, bool) {
	line, _, found := bytes.Cut(data, []byte("\n"))
	if !found {
		return plumbing.ZeroHash, false
	}

	key, value, ok := strings.Cut(string(line), " ")
	if !ok || key != "object" {
		return plumbing.ZeroHash, false
	}

	return parseHash(value)
}

// ExactTag returns the name of the tag pointing exactly at the commit
// hashHex, mirroring `git describe --tags --exact-match`. found is false
// when no tag matches, which is a normal, confident answer. err is
// non-nil only when the match is ambiguous (multiple tags on the same
// commit) or hashHex/the repo shape can't be read, both of which mean the
// caller should fall back to exec git rather than trust a guess.
func ExactTag(commonGitDir, hashHex string) (tag string, found bool, err error) {
	target, ok := parseHash(hashHex)
	if !ok {
		return "", false, fmt.Errorf("gitstatus: invalid hash %q", hashHex)
	}

	refs, err := listRefs(commonGitDir, "refs/tags/")
	if err != nil {
		return "", false, err
	}

	store := newObjectStore(commonGitDir)
	defer store.close()

	var matches []string
	for name, h := range refs {
		commit, ok := peelTag(store, h)
		if !ok {
			// One unresolvable tag object shouldn't block matching against
			// every other tag in the repo.
			continue
		}
		if commit == target {
			matches = append(matches, strings.TrimPrefix(name, "refs/tags/"))
		}
	}

	switch len(matches) {
	case 0:
		return "", false, nil
	case 1:
		return matches[0], true, nil
	default:
		return "", false, fmt.Errorf("gitstatus: ambiguous exact tag match for %s", hashHex)
	}
}
