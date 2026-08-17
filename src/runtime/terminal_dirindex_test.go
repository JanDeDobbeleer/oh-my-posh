package runtime

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDirEntry is a minimal fs.DirEntry that never touches disk, so the
// dirIndex/linearMatch differential tests can construct arbitrary listings -
// including names no real filesystem would accept - without a temp dir.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string { return f.name }
func (f fakeDirEntry) IsDir() bool  { return f.isDir }

func (f fakeDirEntry) Type() fs.FileMode {
	if f.isDir {
		return fs.ModeDir
	}

	return 0
}

func (f fakeDirEntry) Info() (fs.FileInfo, error) {
	return nil, errors.New("fakeDirEntry: Info not implemented")
}

func file(name string) fakeDirEntry   { return fakeDirEntry{name: name} }
func folder(name string) fakeDirEntry { return fakeDirEntry{name: name, isDir: true} }

// matchWithIndex mirrors exactly what HasFilesInDir does once it has a
// listing and a built index: consult the index, fall back to linearMatch
// when it declines. pattern must already be lower-cased, the same contract
// HasFilesInDir upholds before calling either.
func matchWithIndex(idx *dirIndex, dirEntries []fs.DirEntry, pattern string) bool {
	if matched, ok := idx.match(pattern); ok {
		return matched
	}

	return linearMatch(dirEntries, pattern)
}

// TestDirIndexMatchesLinearScan is the differential property test: for a
// wide variety of (directory listing, pattern) combinations, the index-backed
// answer must equal linearMatch's - the unchanged, always-correct
// filepath.Match scan - for every pattern, not only the ones classified as
// fast-path eligible. Divergence anywhere means the classification in
// suffixPattern/hasGlobMeta is unsound.
func TestDirIndexMatchesLinearScan(t *testing.T) {
	corpus := []fakeDirEntry{
		file("main.go"), file("go.mod"), file("go.sum"),
		file("README.md"), file("LICENSE"), file(".gitignore"), file(".env"),
		file("package.json"), file("package-lock.json"), file("tsconfig.json"),
		file("webpack.config.js"), file("build.gradle.kts"), file("settings.gradle.kts"),
		file("Makefile"), file("Dockerfile"), file("docker-compose.yml"),
		file("archive.tar.gz"), file("backup.tar.bz2"), file("notes.txt"),
		file("a.b.c"), file("x.y"), file("trailing."), file("..hidden"),
		file("UPPER.TXT"), file("MixedCase.Go"), file("café.txt"), file("日本語.txt"),
		file("sitecore.json"), file("one.two.three.four"), file("noext"),
		file("weird name.txt"), file("Foo[Bar].txt"), file("q?.txt"),
		// directories must never contribute to a match, however the pattern
		// is shaped - these names deliberately collide with file names/exts
		// above so a bug that forgets the IsDir() check shows up.
		folder("node_modules"), folder(".git"), folder("src"), folder("vendor.go"),
	}

	dirEntries := make([]fs.DirEntry, 0, len(corpus))
	for _, entry := range corpus {
		dirEntries = append(dirEntries, entry)
	}

	idx := newDirIndex(dirEntries)

	// Curated patterns covering every classification path: literals, exact
	// dotted suffixes (single and multi-dot), case variations, unicode,
	// dotfiles, and every metacharacter shape that must fall back
	// (?, [...], embedded *, escapes, bare *, empty pattern).
	patterns := []string{
		"main.go", "MAIN.GO", "Main.Go",
		"go.mod", "GO.MOD",
		"package.json", "PACKAGE.JSON",
		"sitecore.json", "Makefile", "makefile", "MAKEFILE",
		"dockerfile", "docker-compose.yml",
		".gitignore", ".GITIGNORE", ".env", ".ENV",
		"a.b.c", "x.y", "trailing.", "..hidden", "noext",
		"café.txt", "CAFÉ.TXT", "日本語.txt",
		"weird name.txt", "node_modules", ".git", "vendor.go",
		"*.go", "*.GO", "*.md", "*.txt", "*.TXT", "*.json", "*.yml",
		"*.kts", "*.gradle.kts", "*.GRADLE.KTS", "*.tar.gz", "*.gz",
		"*.bz2", "*.tar.bz2", "*.c", "*.b.c", "*.y", "*.four",
		"*.three.four", "*.two.three.four", "*.hidden", "*.",
		"*.rb", "*.missing", "*",
		"", "?", "*?", "?.go", "?ain.go", "m?in.go",
		"[mM]ain.go", "[a-z]*.go", "main.[gG]o",
		"*x*", "*a*", "no*", "*t.go",
		`\*.go`, `main\.go`,
		"foo[bar].txt", "Foo[Bar].txt", "q?.txt",
	}

	for _, pattern := range patterns {
		t.Run(fmt.Sprintf("pattern=%q", pattern), func(t *testing.T) {
			// HasFilesInDir lower-cases the pattern once, before consulting
			// either the index or linearMatch - mirror that here so both
			// sides of the comparison see the exact input production code
			// would give them.
			lowered := strings.ToLower(pattern)
			got := matchWithIndex(idx, dirEntries, lowered)
			want := linearMatch(dirEntries, lowered)
			assert.Equal(t, want, got, "pattern %q", pattern)
		})
	}

	// Randomized layer, fixed seed for reproducibility: combine random
	// fragments of real entry names with random glob metacharacters to
	// stress the classifier with shapes the curated list would not think of.
	fragments := []string{"go", "GO", "json", "kts", "tar.gz", "b.c", "txt", "y", "z", "café", "日本語"}
	metaBits := []string{"*", "?", "[a-z]", `\`, ".", "-", "_", ""}

	rng := rand.New(rand.NewPCG(1, 2))

	generated := make([]string, 0, 300)
	for range 300 {
		var b []byte

		segments := 1 + rng.IntN(3)
		for range segments {
			if rng.IntN(2) == 0 {
				b = append(b, metaBits[rng.IntN(len(metaBits))]...)
				continue
			}

			b = append(b, fragments[rng.IntN(len(fragments))]...)
		}

		generated = append(generated, string(b))
	}

	for i, pattern := range generated {
		t.Run(fmt.Sprintf("generated_%d/pattern=%q", i, pattern), func(t *testing.T) {
			// HasFilesInDir lower-cases the pattern once, before consulting
			// either the index or linearMatch - mirror that here so both
			// sides of the comparison see the exact input production code
			// would give them.
			lowered := strings.ToLower(pattern)
			got := matchWithIndex(idx, dirEntries, lowered)
			want := linearMatch(dirEntries, lowered)
			assert.Equal(t, want, got, "pattern %q", pattern)
		})
	}
}

func TestNewDirIndex(t *testing.T) {
	cases := []struct {
		Case             string
		Entries          []fakeDirEntry
		ExpectedNames    []string
		ExpectedSuffixes []string
		ExcludedNames    []string
	}{
		{
			Case: "files only, directories excluded",
			Entries: []fakeDirEntry{
				file("main.go"),
				folder("main.go.d"), // a directory that even looks like a match
				folder("pkg"),
			},
			ExpectedNames: []string{"main.go"},
			ExcludedNames: []string{"main.go.d", "pkg"},
		},
		{
			Case: "multi-dot suffixes are all indexed",
			Entries: []fakeDirEntry{
				file("build.gradle.kts"),
			},
			ExpectedNames:    []string{"build.gradle.kts"},
			ExpectedSuffixes: []string{".kts", ".gradle.kts"},
		},
		{
			Case: "three dots index every dotted tail",
			Entries: []fakeDirEntry{
				file("a.b.c"),
			},
			ExpectedSuffixes: []string{".c", ".b.c"},
		},
		{
			Case: "names are lower-cased",
			Entries: []fakeDirEntry{
				file("README.MD"),
			},
			ExpectedNames:    []string{"readme.md"},
			ExpectedSuffixes: []string{".md"},
			ExcludedNames:    []string{"README.MD"},
		},
		{
			Case: "no extension indexes no suffix",
			Entries: []fakeDirEntry{
				file("Makefile"),
			},
			ExpectedNames: []string{"makefile"},
		},
		{
			Case: "trailing dot indexes the bare dot suffix",
			Entries: []fakeDirEntry{
				file("trailing."),
			},
			ExpectedSuffixes: []string{"."},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			dirEntries := make([]fs.DirEntry, 0, len(tc.Entries))
			for _, entry := range tc.Entries {
				dirEntries = append(dirEntries, entry)
			}

			idx := newDirIndex(dirEntries)

			for _, name := range tc.ExpectedNames {
				_, ok := idx.names[name]
				assert.True(t, ok, "expected name %q to be indexed", name)
			}

			for _, suffix := range tc.ExpectedSuffixes {
				_, ok := idx.suffixes[suffix]
				assert.True(t, ok, "expected suffix %q to be indexed", suffix)
			}

			for _, name := range tc.ExcludedNames {
				_, ok := idx.names[name]
				assert.False(t, ok, "did not expect name %q to be indexed", name)
			}
		})
	}
}

func TestDirIndexMatch(t *testing.T) {
	dirEntries := []fs.DirEntry{
		file("main.go"), file("go.mod"), file("build.gradle.kts"), folder("src"),
	}
	idx := newDirIndex(dirEntries)

	cases := []struct {
		Case     string
		Pattern  string
		Expected bool
		ExpectOK bool
	}{
		{Case: "literal hit", Pattern: "go.mod", Expected: true, ExpectOK: true},
		{Case: "literal miss", Pattern: "go.sum", Expected: false, ExpectOK: true},
		{Case: "literal against a directory name misses", Pattern: "src", Expected: false, ExpectOK: true},
		{Case: "suffix hit", Pattern: "*.go", Expected: true, ExpectOK: true},
		{Case: "multi-dot suffix hit", Pattern: "*.gradle.kts", Expected: true, ExpectOK: true},
		{Case: "suffix miss", Pattern: "*.rb", Expected: false, ExpectOK: true},
		{Case: "metacharacter pattern declines", Pattern: "*.g?", ExpectOK: false},
		{Case: "character class pattern declines", Pattern: "[mM]ain.go", ExpectOK: false},
		{Case: "bare star declines", Pattern: "*", ExpectOK: false},
		{Case: "non-dot suffix declines", Pattern: "*go", ExpectOK: false},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			matched, ok := idx.match(tc.Pattern)
			require.Equal(t, tc.ExpectOK, ok)

			if !tc.ExpectOK {
				return
			}

			assert.Equal(t, tc.Expected, matched)
		})
	}
}

// TestTerminalDirIndexReusedAcrossQueries proves the integration end to end
// through the public API against a real directory: the second and third
// HasFilesInDir calls for the same dir must answer from the cached index
// rather than re-reading the filesystem, and adding a file to disk after the
// first query must NOT change the answer - the same "read once per prompt"
// contract lsDirMap already gives HasFilesInDir.
func TestTerminalDirIndexReusedAcrossQueries(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), nil, 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "pkg.go"), 0o750))

	term := &Terminal{CmdFlags: &Flags{}}
	term.Init(term.CmdFlags)

	assert.True(t, term.HasFilesInDir(dir, "*.go"), "main.go should match")
	assert.True(t, term.HasFilesInDir(dir, "go.mod"), "literal go.mod should match")
	assert.False(t, term.HasFilesInDir(dir, "*.rb"), "no .rb file present")

	// the directory named "pkg.go" must never satisfy a *.go query
	dirEntries, err := term.readDir(dir)
	require.NoError(t, err)

	var sawPkgGoAsDir bool
	for _, entry := range dirEntries {
		if entry.Name() == "pkg.go" && entry.IsDir() {
			sawPkgGoAsDir = true
		}
	}
	require.True(t, sawPkgGoAsDir, "test fixture sanity check")

	_, cachedBefore := term.dirIndexMap.Get(dir)
	require.True(t, cachedBefore, "index should be cached after the first query")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "new.rb"), nil, 0o600))

	assert.False(t, term.HasFilesInDir(dir, "*.rb"), "cached index must not see files created after the first read")
}
