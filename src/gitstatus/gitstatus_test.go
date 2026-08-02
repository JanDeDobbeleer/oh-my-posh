package gitstatus

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadParity builds throwaway repos with the real git CLI and asserts
// that Load's Result matches what parsePorcelainV2 extracts from
// `git status --porcelain=2 --branch` for the same repo state.
func TestLoadParity(t *testing.T) {
	skipIfNoGit(t)
	hermeticHome(t)

	cases := []struct {
		Setup         func(t *testing.T, dir string)
		Name          string
		UntrackedMode string
	}{
		{Name: "clean", Setup: setupClean},
		{Name: "dirty mix", Setup: setupDirtyMix},
		{Name: "dirty mix, untracked mode all", Setup: setupDirtyMix, UntrackedMode: "all"},
		{Name: "dirty mix, untracked mode no", Setup: setupDirtyMix, UntrackedMode: "no"},
		{Name: "merge conflict both modified (UU)", Setup: setupConflictUU},
		{Name: "merge conflict both added (AA)", Setup: setupConflictAA},
		{Name: "detached HEAD", Setup: setupDetached},
		{Name: "ahead and behind", Setup: setupAheadBehind},
		{Name: "upstream gone", Setup: setupUpstreamGone},
		{Name: "unborn branch", Setup: setupUnborn},
		{Name: "intent to add", Setup: setupIntentToAdd},
		{Name: "tracked file replaced by directory", Setup: setupFileToDir},
		{Name: "tracked directory replaced by file", Setup: setupDirToFile},
		{Name: "index.skipHash zero trailer", Setup: setupSkipHash},
		{Name: "index version 4", Setup: setupIndexV4},
		{Name: "fully packed objects", Setup: setupPacked},
		{Name: "packed ahead and behind", Setup: setupPackedAheadBehind},
	}

	// Every case runs under both stat strategies: the in-walk comparison
	// (Windows default) and the flat lstat pool (everywhere else), so each
	// platform's CI exercises the other's code path too.
	for _, inWalk := range []bool{true, false} {
		name := "flat-stat"
		if inWalk {
			name = "stat-in-walk"
		}

		t.Run(name, func(t *testing.T) {
			previous := statInWalk
			statInWalk = inWalk
			t.Cleanup(func() { statInWalk = previous })

			for _, tc := range cases {
				t.Run(tc.Name, func(t *testing.T) {
					dir := t.TempDir()
					initGitRepo(t, dir)
					tc.Setup(t, dir)
					assertParity(t, dir, tc.UntrackedMode)
				})
			}
		})
	}
}

// TestLoadParityLinkedWorktree exercises a `git worktree add` checkout,
// where WorktreeGitDir (HEAD/index) and CommonGitDir (objects/refs) live in
// different places on disk.
func TestLoadParityLinkedWorktree(t *testing.T) {
	skipIfNoGit(t)
	hermeticHome(t)

	main := t.TempDir()
	initGitRepo(t, main)
	writeFile(t, main, "a.txt", "a\n")
	runGit(t, main, "add", ".")
	runGit(t, main, "commit", "-q", "-m", "base")

	worktree := filepath.Join(t.TempDir(), "wt")
	runGit(t, main, "worktree", "add", "-q", "-b", "wt", worktree)

	writeFile(t, worktree, "b.txt", "b\n")
	runGit(t, worktree, "add", "b.txt")
	writeFile(t, worktree, "untracked.txt", "u\n")

	assertParity(t, worktree, "")
}

// assertParity loads the native status for dir and compares it against what
// parsePorcelainV2 derives from a real `git status` invocation in the same
// directory.
func assertParity(t *testing.T, dir, untrackedMode string) {
	t.Helper()

	opts := Options{
		WorktreeGitDir: gitPath(t, dir, "--git-dir"),
		CommonGitDir:   gitPath(t, dir, "--git-common-dir"),
		RepoRoot:       gitPath(t, dir, "--show-toplevel"),
		UntrackedMode:  untrackedMode,
	}

	got, err := Load(opts)
	require.NoError(t, err)

	mode := untrackedMode
	if mode == "" {
		mode = "normal"
	}

	output := runGit(t, dir, "status", "-u"+mode, "--branch", "--porcelain=2")
	want := parsePorcelainV2(output)

	assert.Equal(t, want, got)
}

func setupClean(t *testing.T, dir string) {
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "init")
}

func setupDirtyMix(t *testing.T, dir string) {
	writeFile(t, dir, "same-size.txt", "hello\n")
	writeFile(t, dir, "diff-size.txt", "short\n")
	writeFile(t, dir, "to-delete.txt", "bye\n")
	writeFile(t, dir, "staged-delete.txt", "gone\n")
	writeFile(t, dir, "staged-modify.txt", "orig\n")
	writeFile(t, dir, "rename-src.txt", "rename me please, needs enough content to be detected as a rename by gits similarity heuristic\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	// unstaged modify, same byte size
	writeFile(t, dir, "same-size.txt", "world\n")
	// unstaged modify, different byte size
	writeFile(t, dir, "diff-size.txt", "a much longer replacement body than before\n")
	// unstaged delete
	require.NoError(t, os.Remove(filepath.Join(dir, "to-delete.txt")))

	// staged add
	writeFile(t, dir, "staged-add.txt", "new\n")
	runGit(t, dir, "add", "staged-add.txt")
	// staged delete
	runGit(t, dir, "rm", "-q", "staged-delete.txt")
	// staged modify
	writeFile(t, dir, "staged-modify.txt", "changed\n")
	runGit(t, dir, "add", "staged-modify.txt")
	// staged rename (content untouched, so exact-hash pairing applies)
	runGit(t, dir, "mv", "rename-src.txt", "rename-dst.txt")

	// untracked file
	writeFile(t, dir, "untracked.txt", "u\n")
	// untracked dir
	writeFile(t, dir, "untracked-dir/inner.txt", "u\n")

	// ignored dir with files
	writeFile(t, dir, ".gitignore", "ignored-dir/\n")
	writeFile(t, dir, "ignored-dir/inner.txt", "ignored\n")

	// nested .gitignore with negation
	writeFile(t, dir, "nested/.gitignore", "*.log\n!keep.log\n")
	writeFile(t, dir, "nested/debug.log", "log\n")
	writeFile(t, dir, "nested/keep.log", "keep\n")
}

func setupConflictUU(t *testing.T, dir string) {
	writeFile(t, dir, "conflict.txt", "base\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	runGit(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "conflict.txt", "feature\n")
	runGit(t, dir, "commit", "-q", "-am", "feature change")

	runGit(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, "conflict.txt", "main\n")
	runGit(t, dir, "commit", "-q", "-am", "main change")

	runGitAllowFail(t, dir, "merge", "-q", "feature")
}

func setupConflictAA(t *testing.T, dir string) {
	writeFile(t, dir, "base.txt", "base\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	runGit(t, dir, "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "newfile.txt", "from feature\n")
	runGit(t, dir, "add", "newfile.txt")
	runGit(t, dir, "commit", "-q", "-m", "feature adds newfile")

	runGit(t, dir, "checkout", "-q", "main")
	writeFile(t, dir, "newfile.txt", "from main\n")
	runGit(t, dir, "add", "newfile.txt")
	runGit(t, dir, "commit", "-q", "-m", "main adds newfile")

	runGitAllowFail(t, dir, "merge", "-q", "feature")
}

func setupFileToDir(t *testing.T, dir string) {
	writeFile(t, dir, "swap", "file\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	require.NoError(t, os.Remove(filepath.Join(dir, "swap")))
	writeFile(t, dir, "swap/inner.txt", "now a directory\n")
}

func setupDirToFile(t *testing.T, dir string) {
	writeFile(t, dir, "swap/one.txt", "1\n")
	writeFile(t, dir, "swap/two.txt", "2\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "swap")))
	writeFile(t, dir, "swap", "now a file\n")
}

// setupSkipHash writes the index with index.skipHash (git >= 2.40), which
// replaces the trailing checksum with 20 zero bytes.
func setupSkipHash(t *testing.T, dir string) {
	runGit(t, dir, "config", "index.skipHash", "true")

	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	writeFile(t, dir, "a.txt", "changed\n")
	writeFile(t, dir, "untracked.txt", "u\n")
}

// setupIndexV4 forces the prefix-compressed index format (git writes v2 by
// default) with a mix of staged, dirty, and untracked files whose paths
// share long prefixes.
func setupIndexV4(t *testing.T, dir string) {
	runGit(t, dir, "config", "index.version", "4")

	writeFile(t, dir, "deeply/nested/directory/one.txt", "1\n")
	writeFile(t, dir, "deeply/nested/directory/two.txt", "2\n")
	writeFile(t, dir, "deeply/nested/other/three.txt", "3\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	writeFile(t, dir, "deeply/nested/directory/two.txt", "changed\n")
	writeFile(t, dir, "deeply/nested/staged.txt", "staged\n")
	runGit(t, dir, "add", "deeply/nested/staged.txt")
	writeFile(t, dir, "deeply/nested/untracked.txt", "u\n")
}

// setupPacked repacks every object into a single pack (with deltas) and
// stages a change afterwards, so the staging tree diff and commit reads all
// go through the pack reader instead of loose objects.
func setupPacked(t *testing.T, dir string) {
	for i := range 5 {
		writeFile(t, dir, "file.txt", strings.Repeat("line\n", 50+i))
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-q", "-m", "rev")
	}
	writeFile(t, dir, "dir/nested.txt", "nested\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "nested")

	runGit(t, dir, "repack", "-a", "-d", "-q")
	runGit(t, dir, "prune-packed")

	// staged change invalidates the cache-tree fast path: the tree diff must
	// read HEAD's trees from the pack
	writeFile(t, dir, "staged.txt", "staged\n")
	runGit(t, dir, "add", "staged.txt")
	writeFile(t, dir, "file.txt", "dirty\n")
}

// setupPackedAheadBehind diverges from a packed upstream so the
// ahead/behind walk reads its commits from the pack.
func setupPackedAheadBehind(t *testing.T, dir string) {
	setupAheadBehind(t, dir)
	runGit(t, dir, "repack", "-a", "-d", "-q")
	runGit(t, dir, "prune-packed")
}

func setupDetached(t *testing.T, dir string) {
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "one")

	writeFile(t, dir, "b.txt", "b\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "two")

	runGit(t, dir, "checkout", "-q", "HEAD~1")
}

func setupAheadBehind(t *testing.T, dir string) {
	remote := t.TempDir()
	runGit(t, remote, "init", "-q", "--bare", "-b", "main")

	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "one")
	runGit(t, dir, "remote", "add", "origin", remote)
	runGit(t, dir, "push", "-q", "-u", "origin", "main")

	// a local commit not yet pushed: this repo is ahead by one
	writeFile(t, dir, "b.txt", "b\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "ahead")

	// a second clone pushes a commit this repo doesn't know about yet
	other := t.TempDir()
	runGit(t, other, "clone", "-q", remote, ".")
	runGit(t, other, "config", "user.email", "test@example.com")
	runGit(t, other, "config", "user.name", "Test")
	writeFile(t, other, "c.txt", "c\n")
	runGit(t, other, "add", ".")
	runGit(t, other, "commit", "-q", "-m", "behind")
	runGit(t, other, "push", "-q", "origin", "main")

	// fetch the remote's new tip into refs/remotes/origin/main without
	// touching local main: this repo is now both ahead and behind
	runGit(t, dir, "fetch", "-q", "origin")
}

func setupUpstreamGone(t *testing.T, dir string) {
	remote := t.TempDir()
	runGit(t, remote, "init", "-q", "--bare", "-b", "main")

	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "one")
	runGit(t, dir, "remote", "add", "origin", remote)
	runGit(t, dir, "push", "-q", "-u", "origin", "main")

	// simulate a deleted remote branch: the tracking config survives, but
	// the remote-tracking ref it points at no longer resolves
	runGit(t, dir, "update-ref", "-d", "refs/remotes/origin/main")
}

func setupUnborn(t *testing.T, dir string) {
	writeFile(t, dir, "staged.txt", "s\n")
	runGit(t, dir, "add", "staged.txt")
	writeFile(t, dir, "untracked.txt", "u\n")
}

func setupIntentToAdd(t *testing.T, dir string) {
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")

	writeFile(t, dir, "intent.txt", "new content\n")
	runGit(t, dir, "add", "-N", "intent.txt")
}

// --- test infrastructure -----------------------------------------------

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
}

// hermeticHome points HOME/USERPROFILE/XDG_CONFIG_HOME at a fresh, empty
// directory for the duration of the test, so neither the native engine nor
// the real git CLI pick up the developer machine's actual global gitconfig
// or excludes file.
func hermeticHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init", "-q", "-b", "main", ".")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "config", "core.autocrlf", "false")
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := gitCommand(dir, args...)
	require.NoErrorf(t, err, "git %s failed: %s", strings.Join(args, " "), out)
	return out
}

// runGitAllowFail runs git and returns its output even on a non-zero exit,
// for commands like `git merge` that legitimately fail on conflict.
func runGitAllowFail(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, _ := gitCommand(dir, args...)
	return out
}

func gitCommand(dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func gitPath(t *testing.T, dir, arg string) string {
	t.Helper()
	out := runGit(t, dir, "rev-parse", "--path-format=absolute", arg)
	return filepath.FromSlash(strings.TrimSpace(out))
}
