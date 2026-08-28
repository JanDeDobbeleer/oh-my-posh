package gitstatus

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

// ResolveRef resolves a full ref path (e.g. "refs/remotes/origin/main") to
// its commit hash within the repository whose common git dir is
// commonGitDir. found is false when the ref does not exist; that is not an
// error, callers just have nothing to report.
func ResolveRef(commonGitDir, refPath string) (hash string, found bool, err error) {
	h, ok, err := resolveRef(commonGitDir, refPath)
	if err != nil || !ok {
		return "", ok, err
	}

	return h.String(), true, nil
}

// listRefs returns every ref under prefix (e.g. "refs/tags/"), keyed by its
// full ref path ("refs/tags/v1.0"). Loose refs are read first and take
// precedence; packed-refs entries fill in the rest, matching git's own
// lookup order. A ref file that isn't a plain hash (a symref such as
// refs/remotes/origin/HEAD) is skipped rather than erroring the whole scan:
// one such ref is routine, not a sign the repo shape is unsupported.
func listRefs(commonGitDir, prefix string) (map[string]plumbing.Hash, error) {
	refs := map[string]plumbing.Hash{}

	root := filepath.Join(commonGitDir, filepath.FromSlash(prefix))
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hash, ok := parseHash(strings.TrimSpace(string(data)))
		if !ok {
			return nil
		}

		rel, err := filepath.Rel(commonGitDir, path)
		if err != nil {
			return nil
		}

		refs[filepath.ToSlash(rel)] = hash
		return nil
	})
	if err != nil {
		return nil, err
	}

	scanPackedRefsPrefix(commonGitDir, prefix, refs)

	return refs, nil
}

// scanPackedRefsPrefix adds every packed-refs entry under prefix to refs
// that isn't already present from a loose ref file.
func scanPackedRefsPrefix(commonGitDir, prefix string, refs map[string]plumbing.Hash) {
	f, err := os.Open(filepath.Join(commonGitDir, "packed-refs"))
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' || line[0] == '^' {
			continue
		}

		hashStr, name, ok := strings.Cut(line, " ")
		if !ok || !strings.HasPrefix(name, prefix) {
			continue
		}

		if _, exists := refs[name]; exists {
			continue // loose ref takes precedence
		}

		if hash, ok := parseHash(hashStr); ok {
			refs[name] = hash
		}
	}
}
