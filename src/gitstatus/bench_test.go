package gitstatus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkLoad runs Load against a repo with ~1k committed files, a small
// mix of unstaged/staged/untracked changes, and reports allocations.
func BenchmarkLoad(b *testing.B) {
	if _, err := exec.LookPath("git"); err != nil {
		b.Skip("git not found on PATH")
	}

	dir := b.TempDir()

	runBenchGit(b, dir, "init", "-q", "-b", "main", ".")
	runBenchGit(b, dir, "config", "user.email", "bench@example.com")
	runBenchGit(b, dir, "config", "user.name", "Bench")

	const fileCount = 1000
	for i := range fileCount {
		rel := filepath.Join("pkg", fmt.Sprintf("dir%d", i%20), fmt.Sprintf("file%d.txt", i))
		writeBenchFile(b, dir, rel, fmt.Sprintf("content %d\n", i))
	}
	runBenchGit(b, dir, "add", ".")
	runBenchGit(b, dir, "commit", "-q", "-m", "seed")

	// a small, realistic mix of changes on top of the committed tree
	writeBenchFile(b, dir, "pkg/dir0/file0.txt", "modified\n")
	writeBenchFile(b, dir, "untracked.txt", "u\n")
	writeBenchFile(b, dir, "pkg/dir1/staged-add.txt", "new\n")
	runBenchGit(b, dir, "add", "pkg/dir1/staged-add.txt")

	opts := Options{
		WorktreeGitDir: gitBenchPath(b, dir, "--git-dir"),
		CommonGitDir:   gitBenchPath(b, dir, "--git-common-dir"),
		RepoRoot:       gitBenchPath(b, dir, "--show-toplevel"),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		if _, err := Load(opts); err != nil {
			b.Fatal(err)
		}
	}
}

func writeBenchFile(b *testing.B, dir, rel, content string) {
	b.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	require.NoError(b, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(b, os.WriteFile(full, []byte(content), 0o644))
}

func runBenchGit(b *testing.B, dir string, args ...string) {
	b.Helper()
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	require.NoErrorf(b, err, "git %s failed: %s", strings.Join(args, " "), out)
}

func gitBenchPath(b *testing.B, dir, arg string) string {
	b.Helper()
	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--path-format=absolute", arg)
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(b, err)
	return filepath.FromSlash(strings.TrimSpace(string(out)))
}
