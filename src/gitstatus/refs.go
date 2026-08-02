package gitstatus

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/jandedobbeleer/oh-my-posh/src/ini"
)

const (
	branchRefPrefix = "ref: refs/heads/"
	reftablesHead   = "ref: refs/heads/.invalid"
)

// resolveBranch reads HEAD and resolves it to a commit hash. For a branch
// checkout it also resolves the configured upstream and, when one exists,
// the ahead/behind counts. The returned bool is false only for an unborn
// branch (HEAD points at a branch ref that has never been committed to).
func resolveBranch(opts Options, cfg *ini.File, store *objectStore, result *Result) (plumbing.Hash, bool, error) {
	data, err := os.ReadFile(filepath.Join(opts.WorktreeGitDir, "HEAD"))
	if err != nil {
		return plumbing.ZeroHash, false, err
	}

	head := strings.TrimSpace(string(data))
	if head == reftablesHead {
		return plumbing.ZeroHash, false, errors.New("gitstatus: reftables HEAD requires exec fallback")
	}

	branchName, isBranch := strings.CutPrefix(head, branchRefPrefix)
	if !isBranch {
		return resolveDetached(head, result)
	}

	result.Ref = branchName

	hash, ok, err := resolveRef(opts.CommonGitDir, "refs/heads/"+branchName)
	if err != nil {
		return plumbing.ZeroHash, false, err
	}

	if !ok {
		// Unborn branch: no commits yet. Matches porcelain's
		// `# branch.oid (initial)`. Upstream/ahead-behind are left at their
		// zero values, same as the exec-git parity path for this case.
		result.Hash = "(initial)"
		return plumbing.ZeroHash, false, nil
	}
	result.Hash = hash.String()

	if err := resolveUpstream(opts, cfg, store, branchName, hash, result); err != nil {
		return plumbing.ZeroHash, false, err
	}

	return hash, true, nil
}

func resolveDetached(head string, result *Result) (plumbing.Hash, bool, error) {
	result.Ref = Detached

	hash, ok := parseHash(head)
	if !ok {
		return plumbing.ZeroHash, false, fmt.Errorf("gitstatus: unrecognized HEAD content %q", head)
	}

	result.Hash = hash.String()
	return hash, true, nil
}

// resolveUpstream fills in Upstream/UpstreamGone/Ahead/Behind from the
// branch's remote-tracking configuration, if any.
func resolveUpstream(opts Options, cfg *ini.File, store *objectStore, branchName string, headHash plumbing.Hash, result *Result) error {
	if cfg == nil {
		return nil
	}

	section, err := cfg.GetSection(fmt.Sprintf(`branch "%s"`, branchName))
	if err != nil {
		return nil
	}

	remote := section.Key("remote").String()
	merge := section.Key("merge").String()
	if remote == "" || merge == "" {
		return nil
	}

	mergeBranch := strings.TrimPrefix(merge, "refs/heads/")

	var upstreamRefPath string
	switch remote {
	case ".":
		result.Upstream = mergeBranch
		upstreamRefPath = "refs/heads/" + mergeBranch
	default:
		if _, err := cfg.GetSection(fmt.Sprintf(`remote "%s"`, remote)); err != nil {
			// branch.<name>.remote isn't a configured remote name (e.g. a
			// bare URL): porcelain prints no upstream at all, which leaves
			// UpstreamGone at its initial true — mirror that exactly
			return nil
		}
		result.Upstream = remote + "/" + mergeBranch
		upstreamRefPath = "refs/remotes/" + remote + "/" + mergeBranch
	}

	upstreamHash, ok, err := resolveRef(opts.CommonGitDir, upstreamRefPath)
	if err != nil {
		return err
	}

	if !ok {
		result.UpstreamGone = true
		return nil
	}

	result.UpstreamGone = false

	ahead, behind, err := aheadBehind(store, headHash, upstreamHash)
	if err != nil {
		return err
	}
	result.Ahead, result.Behind = ahead, behind

	return nil
}

// resolveRef resolves a ref path (e.g. "refs/heads/main") to its commit
// hash, checking the loose ref file first and falling back to packed-refs.
// A loose ref file that exists but does not hold a plain hash (a symref,
// or corruption) is an error: guessing here would silently misreport the
// branch as unborn, so the caller must fall back to exec git instead.
func resolveRef(commonGitDir, refPath string) (plumbing.Hash, bool, error) {
	data, err := os.ReadFile(filepath.Join(commonGitDir, filepath.FromSlash(refPath)))
	if err == nil {
		hash, ok := parseHash(strings.TrimSpace(string(data)))
		if !ok {
			return plumbing.ZeroHash, false, fmt.Errorf("gitstatus: unsupported loose ref content in %s", refPath)
		}
		return hash, true, nil
	}

	hash, found := scanPackedRefs(commonGitDir, refPath)
	return hash, found, nil
}

// scanPackedRefs looks up refPath in CommonGitDir/packed-refs. Lines are
// "<hash> <refname>"; "#"-prefixed comment lines and "^"-prefixed peeled-tag
// lines are ignored.
func scanPackedRefs(commonGitDir, refPath string) (plumbing.Hash, bool) {
	f, err := os.Open(filepath.Join(commonGitDir, "packed-refs"))
	if err != nil {
		return plumbing.ZeroHash, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' || line[0] == '^' {
			continue
		}

		hashStr, name, ok := strings.Cut(line, " ")
		if !ok || name != refPath {
			continue
		}

		if hash, ok := parseHash(hashStr); ok {
			return hash, true
		}
	}

	return plumbing.ZeroHash, false
}

// parseHash accepts only well-formed 40-character hex SHA-1 strings, so a
// stray or corrupt ref file can't silently resolve to the zero hash.
func parseHash(s string) (plumbing.Hash, bool) {
	if len(s) != 40 {
		return plumbing.ZeroHash, false
	}

	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return plumbing.ZeroHash, false
		}
	}

	return plumbing.NewHash(s), true
}

func loadRepoConfig(commonGitDir string) (*ini.File, error) {
	data, err := os.ReadFile(filepath.Join(commonGitDir, "config"))
	if err != nil {
		return nil, err
	}

	return ini.Load(string(data))
}

// checkRepoFormat rejects repository formats the engine cannot read
// correctly, deterministically instead of by parse luck: SHA-256 object
// format, reftables ref storage, and future format versions.
func checkRepoFormat(cfg *ini.File) error {
	if cfg == nil {
		return nil
	}

	core := cfg.Section("core")
	if v := core.Key("repositoryformatversion").String(); v != "" && v != "0" && v != "1" {
		return fmt.Errorf("gitstatus: unsupported repositoryformatversion %s", v)
	}

	extensions := cfg.Section("extensions")

	if v := extensions.Key("objectformat").String(); v != "" && !strings.EqualFold(v, "sha1") {
		return fmt.Errorf("gitstatus: unsupported object format %s", v)
	}

	if v := extensions.Key("refstorage").String(); v != "" && !strings.EqualFold(v, "files") {
		return fmt.Errorf("gitstatus: unsupported ref storage %s", v)
	}

	return nil
}

// trustExecutableBit mirrors git's core.filemode: whether the filesystem
// records the executable bit reliably. Platform default, overridden by the
// repo config (git init probes the filesystem and records the verdict).
func trustExecutableBit(cfg *ini.File) bool {
	trust := runtime.GOOS != goosWindows
	if cfg == nil {
		return trust
	}

	v := cfg.Section("core").Key("filemode").String()
	if v == "" {
		return trust
	}

	return strings.EqualFold(v, "true")
}
