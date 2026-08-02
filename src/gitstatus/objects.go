package gitstatus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// commitInfo is the subset of a commit object the status engine needs.
// Parsing it directly from the encoded object keeps go-git's plumbing/object
// package — and its openpgp/diff dependency tree — out of the binary.
type commitInfo struct {
	Tree          plumbing.Hash
	Parents       []plumbing.Hash
	CommitterWhen int64 // unix seconds
}

func readCommit(store storer.EncodedObjectStorer, h plumbing.Hash) (*commitInfo, error) {
	obj, err := store.EncodedObject(plumbing.CommitObject, h)
	if err != nil {
		return nil, err
	}

	r, err := obj.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	info := &commitInfo{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
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

	if err := scanner.Err(); err != nil {
		return nil, err
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
// skipped, matching the previous TreeWalker-based behavior.
func walkTree(store storer.EncodedObjectStorer, h plumbing.Hash, prefix string, visit func(path string, blob plumbing.Hash)) error {
	obj, err := store.EncodedObject(plumbing.TreeObject, h)
	if err != nil {
		return err
	}

	r, err := obj.Reader()
	if err != nil {
		return err
	}

	data, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		return err
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
