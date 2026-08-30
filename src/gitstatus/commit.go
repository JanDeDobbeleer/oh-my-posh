package gitstatus

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
)

// Ident is a name/email pair as recorded in a commit's author or committer
// line.
type Ident struct {
	Name  string
	Email string
}

// CommitRefs mirrors what `git log --decorate=full` reports for the refs
// pointing at a commit.
type CommitRefs struct {
	Heads   []string
	Tags    []string
	Remotes []string
}

// CommitInfo is the subset of `git log -1`'s output the git segment
// displays.
type CommitInfo struct {
	Timestamp time.Time
	Author    Ident
	Committer Ident
	Subject   string
	Hash      string
	Refs      CommitRefs
}

// LoadCommit reads the commit hashHex points at, plus every local branch,
// tag, and remote-tracking branch that also points at it. Any error means
// the caller must fall back to exec git.
func LoadCommit(commonGitDir, hashHex string) (*CommitInfo, error) {
	hash, ok := parseHash(hashHex)
	if !ok {
		return nil, fmt.Errorf("gitstatus: invalid hash %q", hashHex)
	}

	store := newObjectStore(commonGitDir)
	defer store.close()

	kind, data, err := store.object(hash)
	if err != nil {
		return nil, err
	}
	if kind != kindCommit {
		return nil, fmt.Errorf("gitstatus: object %s is a %s, expected a commit", hashHex, kind)
	}

	info := &CommitInfo{Hash: hash.String()}

	body := data
	for {
		line, rest, found := bytes.Cut(body, []byte("\n"))
		if !found || len(line) == 0 {
			body = rest
			break
		}
		body = rest

		key, value, ok := strings.Cut(string(line), " ")
		if !ok {
			continue
		}

		switch key {
		case "author":
			name, email, ts, ok := parseIdent(value)
			if ok {
				info.Author = Ident{Name: name, Email: email}
				info.Timestamp = time.Unix(ts, 0)
			}
		case "committer":
			name, email, _, ok := parseIdent(value)
			if ok {
				info.Committer = Ident{Name: name, Email: email}
			}
		}
	}

	subject, _, _ := bytes.Cut(body, []byte("\n"))
	info.Subject = string(subject)

	refs, err := decorate(store, commonGitDir, hash)
	if err != nil {
		return nil, err
	}
	info.Refs = *refs

	return info, nil
}

// parseIdent parses a commit's "author"/"committer" header value:
// "Name <email> <unix-timestamp> <tz-offset>".
func parseIdent(value string) (name, email string, timestamp int64, ok bool) {
	lt := strings.IndexByte(value, '<')
	gt := strings.IndexByte(value, '>')
	if lt < 0 || gt < 0 || gt < lt {
		return "", "", 0, false
	}

	name = strings.TrimSpace(value[:lt])
	email = value[lt+1 : gt]

	fields := strings.Fields(strings.TrimSpace(value[gt+1:]))
	if len(fields) > 0 {
		timestamp, _ = strconv.ParseInt(fields[0], 10, 64)
	}

	return name, email, timestamp, true
}

// decorate finds every local branch, tag, and remote-tracking branch that
// points at target, the same set `git log --decorate=full` prints.
func decorate(store *objectStore, commonGitDir string, target plumbing.Hash) (*CommitRefs, error) {
	refs := &CommitRefs{}

	heads, err := listRefs(commonGitDir, "refs/heads/")
	if err != nil {
		return nil, err
	}
	for name, h := range heads {
		if h == target {
			refs.Heads = append(refs.Heads, strings.TrimPrefix(name, "refs/heads/"))
		}
	}
	sort.Strings(refs.Heads)

	remotes, err := listRefs(commonGitDir, "refs/remotes/")
	if err != nil {
		return nil, err
	}
	for name, h := range remotes {
		if h == target {
			refs.Remotes = append(refs.Remotes, strings.TrimPrefix(name, "refs/remotes/"))
		}
	}
	sort.Strings(refs.Remotes)

	tags, err := listRefs(commonGitDir, "refs/tags/")
	if err != nil {
		return nil, err
	}
	for name, h := range tags {
		commit, ok := peelTag(store, h)
		if !ok {
			continue
		}
		if commit == target {
			refs.Tags = append(refs.Tags, strings.TrimPrefix(name, "refs/tags/"))
		}
	}
	sort.Strings(refs.Tags)

	return refs, nil
}
