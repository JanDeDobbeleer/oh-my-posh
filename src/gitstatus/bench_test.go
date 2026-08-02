package gitstatus

import (
	"fmt"
	"path/filepath"
	"testing"
)

// BenchmarkLoad runs Load against a repo with ~1k committed files, a small
// mix of unstaged/staged/untracked changes, and reports allocations.
func BenchmarkLoad(b *testing.B) {
	skipIfNoGit(b)

	dir := b.TempDir()

	runGit(b, dir, "init", "-q", "-b", "main", ".")
	runGit(b, dir, "config", "user.email", "bench@example.com")
	runGit(b, dir, "config", "user.name", "Bench")

	const fileCount = 1000
	for i := range fileCount {
		rel := filepath.Join("pkg", fmt.Sprintf("dir%d", i%20), fmt.Sprintf("file%d.txt", i))
		writeFile(b, dir, rel, fmt.Sprintf("content %d\n", i))
	}
	runGit(b, dir, "add", ".")
	runGit(b, dir, "commit", "-q", "-m", "seed")

	// a small, realistic mix of changes on top of the committed tree
	writeFile(b, dir, "pkg/dir0/file0.txt", "modified\n")
	writeFile(b, dir, "untracked.txt", "u\n")
	writeFile(b, dir, "pkg/dir1/staged-add.txt", "new\n")
	runGit(b, dir, "add", "pkg/dir1/staged-add.txt")

	opts := Options{
		WorktreeGitDir: gitPath(b, dir, "--git-dir"),
		CommonGitDir:   gitPath(b, dir, "--git-common-dir"),
		RepoRoot:       gitPath(b, dir, "--show-toplevel"),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		if _, err := Load(opts); err != nil {
			b.Fatal(err)
		}
	}
}
