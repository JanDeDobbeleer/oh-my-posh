package gitstatus

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"

	"github.com/jandedobbeleer/oh-my-posh/src/ini"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConflictXY(t *testing.T) {
	cases := []struct {
		Case            string
		Stages          int
		ExpectedStaging byte
		ExpectedWorking byte
	}{
		{Case: "base only -> DD", Stages: 1 << 1, ExpectedStaging: 'D', ExpectedWorking: 'D'},
		{Case: "ours only -> AU", Stages: 1 << 2, ExpectedStaging: 'A', ExpectedWorking: 'U'},
		{Case: "theirs only -> UA", Stages: 1 << 3, ExpectedStaging: 'U', ExpectedWorking: 'A'},
		{Case: "base+ours -> UD", Stages: 1<<1 | 1<<2, ExpectedStaging: 'U', ExpectedWorking: 'D'},
		{Case: "base+theirs -> DU", Stages: 1<<1 | 1<<3, ExpectedStaging: 'D', ExpectedWorking: 'U'},
		{Case: "ours+theirs -> AA", Stages: 1<<2 | 1<<3, ExpectedStaging: 'A', ExpectedWorking: 'A'},
		{Case: "all three -> UU", Stages: 1<<1 | 1<<2 | 1<<3, ExpectedStaging: 'U', ExpectedWorking: 'U'},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			staging, working := conflictXY(tc.Stages)
			assert.Equal(t, tc.ExpectedStaging, staging, "staging code")
			assert.Equal(t, tc.ExpectedWorking, working, "working code")
		})
	}
}

func TestAddCode(t *testing.T) {
	cases := []struct {
		Case     string
		Code     byte
		Expected Counts
	}{
		{Case: "deleted", Code: 'D', Expected: Counts{Deleted: 1}},
		{Case: "added", Code: 'A', Expected: Counts{Added: 1}},
		{Case: "unmerged", Code: 'U', Expected: Counts{Unmerged: 1}},
		{Case: "modified", Code: 'M', Expected: Counts{Modified: 1}},
		{Case: "unknown code is a no-op", Code: '?', Expected: Counts{}},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			var counts Counts
			addCode(&counts, tc.Code)
			assert.Equal(t, tc.Expected, counts)
		})
	}
}

func TestScanPackedRefs(t *testing.T) {
	dir := t.TempDir()
	content := "" +
		"# pack-refs with: peeled fully-peeled sorted\n" +
		"1111111111111111111111111111111111111111 refs/heads/main\n" +
		"2222222222222222222222222222222222222222 refs/heads/feature\n" +
		"^3333333333333333333333333333333333333333\n" +
		"4444444444444444444444444444444444444444 refs/remotes/origin/main\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "packed-refs"), []byte(content), 0o644))

	cases := []struct {
		Case     string
		RefPath  string
		Expected string
		Found    bool
	}{
		{Case: "found branch", RefPath: "refs/heads/main", Expected: "1111111111111111111111111111111111111111", Found: true},
		{Case: "peeled line does not shadow the following ref", RefPath: "refs/heads/feature", Expected: "2222222222222222222222222222222222222222", Found: true},
		{Case: "remote-tracking ref", RefPath: "refs/remotes/origin/main", Expected: "4444444444444444444444444444444444444444", Found: true},
		{Case: "missing ref", RefPath: "refs/heads/missing", Found: false},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			hash, ok := scanPackedRefs(dir, tc.RefPath)
			assert.Equal(t, tc.Found, ok)
			if tc.Found {
				assert.Equal(t, tc.Expected, hash.String())
			}
		})
	}
}

func TestResolveRefPrefersLooseOverPacked(t *testing.T) {
	dir := t.TempDir()
	looseHash := "1111111111111111111111111111111111111111"
	packedHash := "2222222222222222222222222222222222222222"

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "refs", "heads"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "refs", "heads", "main"), []byte(looseHash+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "packed-refs"), []byte(packedHash+" refs/heads/main\n"), 0o644))

	hash, ok := resolveRef(dir, "refs/heads/main")
	require.True(t, ok)
	assert.Equal(t, looseHash, hash.String())

	hash, ok = resolveRef(dir, "refs/heads/only-packed")
	assert.False(t, ok)
	assert.True(t, hash.IsZero())
}

func TestResolveUpstream(t *testing.T) {
	cases := []struct {
		Case                 string
		Config               string
		ExpectedUpstream     string
		ExpectedUpstreamGone bool
	}{
		{
			Case: "named remote",
			Config: `[branch "main"]
	remote = origin
	merge = refs/heads/main
[remote "origin"]
	url = https://example.com/repo.git
`,
			ExpectedUpstream:     "origin/main",
			ExpectedUpstreamGone: true, // configured, but the ref doesn't exist on disk
		},
		{
			Case: "remote value is a bare URL with no matching section",
			Config: `[branch "main"]
	remote = https://example.com/other.git
	merge = refs/heads/main
`,
			ExpectedUpstream: "",
			// porcelain prints no upstream line for a URL remote, so the
			// exec path leaves its initial UpstreamGone = true untouched
			ExpectedUpstreamGone: true,
		},
		{
			Case: "dot remote (local branch upstream)",
			Config: `[branch "main"]
	remote = .
	merge = refs/heads/other
`,
			ExpectedUpstream:     "other",
			ExpectedUpstreamGone: true, // configured, but refs/heads/other doesn't exist
		},
		{
			Case:                 "no branch section configured",
			Config:               "",
			ExpectedUpstream:     "",
			ExpectedUpstreamGone: true, // unchanged default: matches "no upstream at all"
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			commonGitDir := t.TempDir()

			var cfg *ini.File
			if tc.Config != "" {
				var err error
				cfg, err = ini.Load(tc.Config)
				require.NoError(t, err)
			}

			opts := Options{CommonGitDir: commonGitDir}
			result := &Result{UpstreamGone: true}

			err := resolveUpstream(opts, cfg, nil, "main", plumbing.ZeroHash, result)
			require.NoError(t, err)

			assert.Equal(t, tc.ExpectedUpstream, result.Upstream)
			assert.Equal(t, tc.ExpectedUpstreamGone, result.UpstreamGone)
			assert.Zero(t, result.Ahead)
			assert.Zero(t, result.Behind)
		})
	}
}

func TestLoadFallsBack(t *testing.T) {
	cases := []struct {
		Setup func(t *testing.T, dir string)
		Case  string
	}{
		{
			Case: "gitlink entry",
			Setup: func(t *testing.T, dir string) {
				t.Helper()
				idx := &index.Index{
					Version: 2,
					Entries: []*index.Entry{
						{Name: "submodule", Mode: filemode.Submodule, Hash: plumbing.NewHash("1111111111111111111111111111111111111111")},
					},
				}
				encodeIndex(t, filepath.Join(dir, "index"), idx)
			},
		},
		{
			Case: "mandatory index extension",
			Setup: func(t *testing.T, dir string) {
				t.Helper()
				// DIRC header, version 2, zero entries, then a mandatory
				// ("sdir", lowercase first byte) extension header.
				// Sparse-index and split-index (signatures "sdir"/"link")
				// both hit this same decoder path.
				var buf bytes.Buffer
				buf.WriteString("DIRC")
				require.NoError(t, binary.Write(&buf, binary.BigEndian, uint32(2)))
				require.NoError(t, binary.Write(&buf, binary.BigEndian, uint32(0)))
				buf.WriteString("sdir")
				require.NoError(t, binary.Write(&buf, binary.BigEndian, uint32(0)))
				buf.Write(make([]byte, 20)) // trailer so the extension scan sees a complete block

				require.NoError(t, os.WriteFile(filepath.Join(dir, "index"), buf.Bytes(), 0o644))
			},
		},
		{
			Case: "reftables HEAD",
			Setup: func(t *testing.T, dir string) {
				t.Helper()
				encodeIndex(t, filepath.Join(dir, "index"), &index.Index{Version: 2})
				require.NoError(t, os.WriteFile(filepath.Join(dir, "HEAD"), []byte("ref: refs/heads/.invalid\n"), 0o644))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			dir := t.TempDir()
			tc.Setup(t, dir)

			_, err := Load(Options{
				WorktreeGitDir: dir,
				CommonGitDir:   dir,
				RepoRoot:       dir,
			})
			assert.Error(t, err)
		})
	}
}

func encodeIndex(t *testing.T, path string, idx *index.Index) {
	t.Helper()

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	require.NoError(t, index.NewEncoder(f).Encode(idx))
}
