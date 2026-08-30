package gitstatus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HeadInfo is the resolved identity of a repository's HEAD.
type HeadInfo struct {
	// Hash is the full HEAD commit hash.
	Hash string
	// Ref is the branch name HEAD points at, or Detached when HEAD is not
	// on a branch.
	Ref      string
	Detached bool
}

// LoadHead resolves worktreeGitDir/HEAD to a commit hash, following a
// branch ref through loose or packed refs in commonGitDir. It is a
// lightweight sibling of Load: callers that only need the current hash and
// branch name (not the full status) can use this instead. Any error means
// the caller must fall back to exec git: a reftables HEAD, an unborn
// branch, or a corrupt/unsupported ref.
func LoadHead(worktreeGitDir, commonGitDir string) (*HeadInfo, error) {
	data, err := os.ReadFile(filepath.Join(worktreeGitDir, "HEAD"))
	if err != nil {
		return nil, err
	}

	head := strings.TrimSpace(string(data))
	if head == reftablesHead {
		return nil, errors.New("gitstatus: reftables HEAD requires exec fallback")
	}

	branchName, isBranch := strings.CutPrefix(head, branchRefPrefix)
	if !isBranch {
		hash, ok := parseHash(head)
		if !ok {
			return nil, fmt.Errorf("gitstatus: unrecognized HEAD content %q", head)
		}

		return &HeadInfo{Hash: hash.String(), Ref: Detached, Detached: true}, nil
	}

	hash, ok, err := resolveRef(commonGitDir, "refs/heads/"+branchName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("gitstatus: unborn branch")
	}

	return &HeadInfo{Hash: hash.String(), Ref: branchName}, nil
}
