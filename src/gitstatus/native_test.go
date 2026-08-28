package gitstatus

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadHeadParity covers both a branch checkout and a detached HEAD
// against the real git CLI's own idea of the current commit.
func TestLoadHeadParity(t *testing.T) {
	skipIfNoGit(t)
	hermeticHome(t)

	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "one")

	worktreeGitDir := gitPath(t, dir, "--git-dir")
	commonGitDir := gitPath(t, dir, "--git-common-dir")
	wantHash := runGit(t, dir, "rev-parse", "HEAD")

	head, err := LoadHead(worktreeGitDir, commonGitDir)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(wantHash), head.Hash)
	assert.Equal(t, "main", head.Ref)
	assert.False(t, head.Detached)

	runGit(t, dir, "checkout", "-q", "--detach", "HEAD")
	head, err = LoadHead(worktreeGitDir, commonGitDir)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(wantHash), head.Hash)
	assert.Equal(t, Detached, head.Ref)
	assert.True(t, head.Detached)
}

func TestLoadHeadUnbornBranchFallsBack(t *testing.T) {
	skipIfNoGit(t)
	hermeticHome(t)

	dir := t.TempDir()
	initGitRepo(t, dir)

	_, err := LoadHead(gitPath(t, dir, "--git-dir"), gitPath(t, dir, "--git-common-dir"))
	assert.Error(t, err)
}

// TestExactTagParity covers a lightweight tag, an annotated tag (which
// requires peeling), a non-tagged commit, and an ambiguous multi-tag commit
// against the real `git describe --tags --exact-match`.
func TestExactTagParity(t *testing.T) {
	skipIfNoGit(t)
	hermeticHome(t)

	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "one")
	runGit(t, dir, "tag", "lightweight")

	writeFile(t, dir, "b.txt", "b\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "two")
	runGit(t, dir, "tag", "-a", "annotated", "-m", "msg")

	writeFile(t, dir, "c.txt", "c\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "three")

	commonGitDir := gitPath(t, dir, "--git-common-dir")

	lightweightHash := strings.TrimSpace(runGit(t, dir, "rev-parse", "lightweight"))
	tag, found, err := ExactTag(commonGitDir, lightweightHash)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "lightweight", tag)

	annotatedHash := strings.TrimSpace(runGit(t, dir, "rev-parse", "annotated^{commit}"))
	tag, found, err = ExactTag(commonGitDir, annotatedHash)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "annotated", tag)

	headHash := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
	_, found, err = ExactTag(commonGitDir, headHash)
	require.NoError(t, err)
	assert.False(t, found)

	// two tags on the same commit: the engine must refuse to guess
	runGit(t, dir, "tag", "second-tag", "lightweight")
	_, _, err = ExactTag(commonGitDir, lightweightHash)
	assert.Error(t, err)
}

// TestLoadCommitParity covers a commit decorated with a local branch, two
// tags, and a remote-tracking branch against `git log -1 --decorate=full`.
func TestLoadCommitParity(t *testing.T) {
	skipIfNoGit(t)
	hermeticHome(t)

	remote := t.TempDir()
	runGit(t, remote, "init", "-q", "--bare", "-b", "main")

	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "feat: decorated commit")
	runGit(t, dir, "tag", "v1.0")
	runGit(t, dir, "tag", "-a", "v1.1", "-m", "annotated")
	runGit(t, dir, "remote", "add", "origin", remote)
	runGit(t, dir, "push", "-q", "-u", "origin", "main")

	commonGitDir := gitPath(t, dir, "--git-common-dir")
	hash := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	info, err := LoadCommit(commonGitDir, hash)
	require.NoError(t, err)

	assert.Equal(t, "Test", info.Author.Name)
	assert.Equal(t, "test@example.com", info.Author.Email)
	assert.Equal(t, "Test", info.Committer.Name)
	assert.Equal(t, "test@example.com", info.Committer.Email)
	assert.Equal(t, "feat: decorated commit", info.Subject)
	assert.Equal(t, hash, info.Hash)
	assert.ElementsMatch(t, []string{"main"}, info.Refs.Heads)
	assert.ElementsMatch(t, []string{"v1.0", "v1.1"}, info.Refs.Tags)
	assert.ElementsMatch(t, []string{"origin/main"}, info.Refs.Remotes)
}

// TestAheadBehindAndResolveRefParity mirrors setupAheadBehind's scenario but
// drives it through the public AheadBehind/ResolveRef wrappers instead of
// Load, matching `git rev-list --count` for the same repo.
func TestAheadBehindAndResolveRefParity(t *testing.T) {
	skipIfNoGit(t)
	hermeticHome(t)

	remote := t.TempDir()
	runGit(t, remote, "init", "-q", "--bare", "-b", "main")

	dir := t.TempDir()
	initGitRepo(t, dir)
	writeFile(t, dir, "a.txt", "a\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "base")
	runGit(t, dir, "remote", "add", "origin", remote)
	runGit(t, dir, "push", "-q", "-u", "origin", "main")

	writeFile(t, dir, "b.txt", "b\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-q", "-m", "local only")

	commonGitDir := gitPath(t, dir, "--git-common-dir")
	ours := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	theirs, found, err := ResolveRef(commonGitDir, "refs/remotes/origin/main")
	require.NoError(t, err)
	require.True(t, found)

	ahead, behind, err := AheadBehind(commonGitDir, ours, theirs)
	require.NoError(t, err)
	assert.Equal(t, 1, ahead)
	assert.Equal(t, 0, behind)

	_, found, err = ResolveRef(commonGitDir, "refs/remotes/origin/does-not-exist")
	require.NoError(t, err)
	assert.False(t, found)
}
