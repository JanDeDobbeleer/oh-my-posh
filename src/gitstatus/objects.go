package gitstatus

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

// commitInfo is the subset of a commit object the status engine needs.
type commitInfo struct {
	Tree          plumbing.Hash
	Parents       []plumbing.Hash
	CommitterWhen int64 // unix seconds
}

func readCommit(store *objectStore, h plumbing.Hash) (*commitInfo, error) {
	kind, data, err := store.object(h)
	if err != nil {
		return nil, err
	}
	if kind != kindCommit {
		return nil, fmt.Errorf("gitstatus: object %s is a %s, expected a commit", h, kind)
	}

	info := &commitInfo{}
	for line := range strings.SplitSeq(string(data), "\n") {
		if line == "" {
			break // end of headers, message follows
		}

		key, value, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}

		switch key {
		case "tree":
			if hash, ok := parseHash(value); ok {
				info.Tree = hash
			}
		case "parent":
			if hash, ok := parseHash(value); ok {
				info.Parents = append(info.Parents, hash)
			}
		case "committer":
			info.CommitterWhen = identTimestamp(value)
		}
	}

	if info.Tree.IsZero() {
		return nil, fmt.Errorf("gitstatus: commit %s has no tree header", h)
	}

	return info, nil
}

// identTimestamp extracts the unix timestamp from a "Name <email> 1234 +0100"
// ident line value. Returns 0 when the line is malformed.
func identTimestamp(ident string) int64 {
	fields := strings.Fields(ident)
	// timestamp is the second-to-last field, before the timezone offset
	if len(fields) < 2 {
		return 0
	}

	ts, err := strconv.ParseInt(fields[len(fields)-2], 10, 64)
	if err != nil {
		return 0
	}

	return ts
}

const treeModeDir = "40000"

// walkTree streams every blob in the tree rooted at h to visit, with its
// full slash-separated path. Non-blob, non-tree entries (gitlinks) are
// skipped.
func walkTree(store *objectStore, h plumbing.Hash, prefix string, visit func(path string, blob plumbing.Hash)) error {
	kind, data, err := store.object(h)
	if err != nil {
		return err
	}
	if kind != kindTree {
		return fmt.Errorf("gitstatus: object %s is a %s, expected a tree", h, kind)
	}

	// tree format: "<octal mode> <name>\x00" followed by a raw 20-byte hash
	for len(data) > 0 {
		nul := bytes.IndexByte(data, 0)
		if nul < 0 || len(data) < nul+1+20 {
			return fmt.Errorf("gitstatus: malformed tree object %s", h)
		}

		mode, name, ok := strings.Cut(string(data[:nul]), " ")
		if !ok {
			return fmt.Errorf("gitstatus: malformed tree entry in %s", h)
		}

		var entryHash plumbing.Hash
		copy(entryHash[:], data[nul+1:nul+1+20])
		data = data[nul+1+20:]

		path := name
		if prefix != "" {
			path = prefix + "/" + name
		}

		switch mode {
		case treeModeDir:
			if err := walkTree(store, entryHash, path, visit); err != nil {
				return err
			}
		case "160000": // gitlink (submodule): not a blob, skip
		default:
			visit(path, entryHash)
		}
	}

	return nil
}
