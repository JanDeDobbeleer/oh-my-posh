package prompt

// The golden harness renders every bundled theme against one committed,
// sanitized --data fixture shared by all of them (fixturePath below) and
// checks the result against two artifacts under src/prompt/testdata/goldens:
//
//   - manifest.txt: one "<theme> <sha256-of-rendered-bytes>" line per theme,
//     sorted. A theme change touches exactly one line, so a PR diff stays
//     readable even though every theme is covered.
//   - <theme>.golden: the full rendered bytes, committed only for a small,
//     deliberately chosen set of representative themes (representativeThemes
//     below) so a refactor has at least a few real ANSI blobs to diff by eye.
//
// A single shared fixture (rather than one per theme) is possible because
// every theme's fixture data is small and mostly overlapping: for every
// segment key any bundled theme uses, the fixture records one representative
// value (see cli/config_export_data.go's --themes merge mode), and every
// theme reads whichever of those keys its own config references.
//
// Regenerate both the fixture and the goldens/manifest after an intentional
// rendering change with:
//
//	go run . config export data --sanitize --themes ../themes --output prompt/testdata/fixtures/prompt.data.json
//	go test ./prompt/... -run TestGoldenThemes -update
//
// run from src/. Those two commands are named in every failure message below.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/require"
)

const regenerateCommand = "go test ./prompt/... -run TestGoldenThemes -update (from src/)"

const regenerateFixtureCommand = "go run . config export data --sanitize --themes ../themes " +
	"--output prompt/testdata/fixtures/prompt.data.json (from src/)"

// fixturePath is the single sanitized --data fixture every bundled theme
// renders against - see the package comment above for why one shared fixture
// replaces what used to be 125 per-theme files.
var fixturePath = filepath.Join("testdata", "fixtures", "prompt.data.json")

// sanitizedPWD is the PWD value cli/sanitize.go bakes into fixturePath (see
// package comment above) - kept in sync manually since the two live in
// different packages that must not import each other (see applyFixtureData).
const sanitizedPWD = "~/dev/oh-my-posh"

var update = flag.Bool("update", false, "regenerate the golden manifest and representative golden files")

// TestMain neutralises every ambient input identified while building this
// harness that would otherwise make a theme's rendered bytes depend on the
// machine or session running the test, rather than only on the theme and its
// fixture. code.yml runs `go test ./...` on ubuntu, macOS and windows, so all
// three matter, not just this machine.
func TestMain(m *testing.M) {
	// 65 of 125 themes render a time segment through dateInZone(fmt, date,
	// "Local") (template/date.go). The segment's own state is baked into the
	// fixture at record time (restoreData never re-probes it), but formatting
	// that recorded instant still happens at render time in the process's
	// Local zone - so without this, the same fixture renders different text
	// on a UTC CI runner than on a contributor's own machine. Setting the TZ
	// env var does not work on Windows; time.Local must be assigned directly.
	time.Local = time.UTC

	// color.TrueColor = Program != AppleTerminal (terminal/writer.go) gates
	// truecolor vs 256-colour emission based on these two env vars. A
	// contributor's real terminal (or macOS Terminal.app in CI) must not leak
	// into what gets rendered here.
	os.Unsetenv("TERM_PROGRAM")
	os.Unsetenv("WT_SESSION")

	cacheDir, err := os.MkdirTemp("", "omp-golden-cache")
	if err != nil {
		fmt.Fprintln(os.Stderr, "golden harness: failed to create isolated cache dir:", err)
		os.Exit(1)
	}

	// OMP_CACHE_DIR must be set before the first call to cache.Path() (it
	// memoizes into a package-level var), which cache.Init below triggers.
	// Isolates every segment's device/session cache from whatever a real omp
	// session on this machine has ever written - segment.isToggled()
	// (config/segment.go) reads the session cache and returns before
	// restoreData ever runs, so a developer who has run `oh-my-posh toggle`
	// would otherwise get different renders than CI.
	os.Setenv("OMP_CACHE_DIR", cacheDir)

	// cache.NewSession mints a fresh session ID unconditionally, ignoring
	// POSH_SESSION_ID, so a real, live session cache can never be resolved
	// and read here even if the harness happens to run inside one.
	cache.Init("golden-harness", cache.NewSession)

	flag.Parse()

	code := m.Run()

	// Deliberately not cache.Close(): nothing was persisted (no Persist
	// option above), and closing is unnecessary work on a cache directory
	// that's about to be removed.
	os.RemoveAll(cacheDir)

	os.Exit(code)
}

// representativeThemes get a full byte-for-byte golden file in addition to
// their manifest line. Selected to cover:
//   - both bundled .omp.yaml themes (devious-diamonds, glowsticks) - they go
//     through Segment.UnmarshalYAML (config/segment.go), a different
//     unmarshal path than the 122 .omp.json themes.
//   - a theme pure in each of the three styles bundled themes actually use
//     (plain: avit, powerline: agnoster, diamond: bubbles); no bundled theme
//     uses the "accordion" style.
//   - a theme with an rprompt (right-aligned) block: atomic, chips.
//   - chips specifically, since it makes a live wakatime HTTP call, and this
//     harness's hermeticity (rendering from the recorded fixture, never
//     touching the network) was verified against it.
//   - the project's own default theme (jandedobbeleer) and two more
//     real-world themes explicitly analyzed while scoping this harness
//     (tokyo, powerlevel10k_rainbow).
//   - the busiest theme by segment count (night-owl, 43) and the
//     second-busiest (tiwahu, 40), to keep at least one dense config honest.
//
// No bundled theme uses a gradient color (linear-gradient(/dark-gradient(/
// light-gradient() - confirmed by grep across themes/*.omp.* - so no theme
// could be selected for that criterion; gradient rendering is covered
// separately by prompt/gradient_test.go and prompt/gradient_collapse_test.go.
// Likewise no bundled theme uses <LINK> hyperlink markup - confirmed the same
// way - so that slot has no candidate either.
var representativeThemes = map[string]bool{
	"devious-diamonds":      true,
	"glowsticks":            true,
	"avit":                  true,
	"agnoster":              true,
	"bubbles":               true,
	"jandedobbeleer":        true,
	"atomic":                true,
	"chips":                 true,
	"tokyo":                 true,
	"night-owl":             true,
	"tiwahu":                true,
	"powerlevel10k_rainbow": true,
}

// themeFiles enumerates the bundled themes via config.ThemeFiles, from the
// repo's themes/ directory. go test's working directory is the package
// directory (src/prompt), so themes/ sits two levels up.
func themeFiles(t *testing.T) []string {
	t.Helper()

	const themesDir = "../../themes"

	files, err := config.ThemeFiles(themesDir)
	require.NoError(t, err)

	return files
}

// themeStem strips the theme directory and one of config.ThemeFileExtensions,
// leaving the name shared by a theme file, its fixture and its golden:
// "agnoster.omp.json" -> "agnoster".
func themeStem(themePath string) string {
	name := filepath.Base(themePath)

	for _, ext := range config.ThemeFileExtensions {
		if stem, ok := strings.CutSuffix(name, ext); ok {
			return stem
		}
	}

	return name
}

// applyFixtureData loads a sanitized --data fixture and copies it onto flags via
// config.Data.ApplyFlags, the same routing cli.applyDataFile uses for the
// print/image commands (cli/data.go) - cli itself imports this package, so it
// cannot be shared the other way around. A nil changed means "never
// overridden," since this harness has no CLI flags of its own to check.
func applyFixtureData(t *testing.T, flags *runtime.Flags, fixturePath string) {
	t.Helper()

	data, err := config.LoadData(fixturePath)
	require.NoErrorf(t, err, "failed to load fixture %s", fixturePath)

	require.GreaterOrEqualf(t, data.Version, config.DataVersion,
		"fixture %s has no recorder version marker; regenerate it with: %s", fixturePath, regenerateFixtureCommand)

	require.NoErrorf(t, data.ApplyFlags(flags, nil), "failed to apply fixture %s", fixturePath)
}

// renderTheme renders themePath's primary prompt against fixturePath, using
// newEngine (engine.go) rather than New: New resolves its config through
// config.Get, which returns whatever config is session-cached (config/gob.go)
// and ignores the path entirely inside a real omp shell. newEngine is the
// post-config.Get half of New, extracted so this harness gets the same
// terminal.Plain/forceRender/prompt.Grow/rectifyTerminalWidth handling New
// performs, instead of silently skipping it the way hand-rolling an Engine
// (as cli/config_export_data.go does for recording) would.
func renderTheme(t *testing.T, themePath, fixturePath string) []byte {
	t.Helper()

	// template.Init only rebuilds the package-level template.Cache singleton
	// when it is nil (template/init.go) - by design, for the serve daemon,
	// which stays alive across many renders and calls ResetCache itself
	// between them (cli/serve.go). This harness renders 125 different themes
	// (each with its own env identity) in one process, and other tests in
	// this package also populate Cache directly, so without forcing a rebuild
	// here every render after the first would silently reuse whichever env
	// data happened to be built last, rather than this theme's fixture.
	template.ResetCache()

	flags := &runtime.Flags{
		ConfigPath:    themePath,
		Shell:         shell.GENERIC,
		TerminalWidth: 120,
		IsPrimary:     true,
		Escape:        true,
	}

	applyFixtureData(t, flags, fixturePath)

	env := &runtime.Terminal{}
	env.Init(flags)

	cfg := config.Load(themePath)
	require.NotEmptyf(t, cfg.Source, "config.Load(%s) fell back to the default theme - it failed to parse", themePath)

	eng := newEngine(cfg, env)

	out := []byte(eng.Primary())

	// Terminal.setPwd (runtime/terminal.go) feeds the fixture's PWD through
	// runtime/path.Clean, which - only on a Windows test runner - turns
	// sanitizedPWD's forward slashes into backslashes before any OSC 9;9 or
	// title escape that embeds the cwd (engine.go's e.Env.Pwd()); correct for
	// a real Windows shell, but it splits this one manifest into three, one
	// per test-runner OS. Rewriting back only this exact literal (not every
	// backslash in the render) restores cross-platform parity without
	// touching an escape sequence's own backslash bytes, such as an OSC
	// string terminator (ESC \).
	out = bytes.ReplaceAll(out, []byte(filepath.FromSlash(sanitizedPWD)), []byte(sanitizedPWD))

	return out
}

// escapeWindow renders b so escape sequences and other control bytes are
// visible on a terminal or in a CI log instead of silently mutating it - a raw
// ANSI blob diffed as-is is unreadable, which is the entire reason this
// harness reports a decoded window instead of the raw bytes.
func escapeWindow(b []byte) string {
	var sb strings.Builder

	for _, r := range string(b) {
		switch r {
		case '\x1b':
			sb.WriteString("<ESC>")
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&sb, "<%02X>", r)
				continue
			}

			sb.WriteRune(r)
		}
	}

	return sb.String()
}

// firstDiffOffset returns the index of the first byte at which want and got
// disagree, or -1 if one is a prefix of the other only by trailing length (in
// which case the shorter length is returned) or they are identical.
func firstDiffOffset(want, got []byte) int {
	n := min(len(got), len(want))

	for i := range n {
		if want[i] != got[i] {
			return i
		}
	}

	if len(want) != len(got) {
		return n
	}

	return -1
}

const contextRadius = 40

func contextWindow(b []byte, offset int) []byte {
	start := max(offset-contextRadius, 0)
	end := min(offset+contextRadius, len(b))

	return b[start:end]
}

// TestGoldenThemes renders every bundled theme from its committed fixture and
// checks it two ways: a sha256 line in the manifest (all 125 themes) and,
// additionally for representativeThemes, a byte-for-byte comparison against a
// committed golden file. Run with -update to regenerate both instead of
// comparing.
func TestGoldenThemes(t *testing.T) {
	themePaths := themeFiles(t)
	require.Lenf(t, themePaths, 125, "expected 125 bundled themes (123 .omp.json + 2 .omp.yaml); "+
		"if this changed intentionally, update the fixture/golden set for the new/removed theme(s)")

	type manifestEntry struct {
		stem string
		hash string
	}

	manifest := make([]manifestEntry, 0, len(themePaths))
	seenRepresentative := make(map[string]bool, len(representativeThemes))

	for _, themePath := range themePaths {
		stem := themeStem(themePath)

		out := renderTheme(t, themePath, fixturePath)

		sum := sha256.Sum256(out)
		manifest = append(manifest, manifestEntry{stem: stem, hash: hex.EncodeToString(sum[:])})

		if !representativeThemes[stem] {
			continue
		}

		seenRepresentative[stem] = true
		goldenPath := filepath.Join("testdata", "goldens", stem+".golden")

		if *update {
			require.NoErrorf(t, os.WriteFile(goldenPath, out, 0o644), "failed to write golden for %s", stem)
			continue
		}

		want, err := os.ReadFile(goldenPath)
		require.NoErrorf(t, err, "failed to read golden %s; regenerate with: %s", goldenPath, regenerateCommand)

		if !bytes.Equal(want, out) {
			offset := firstDiffOffset(want, out)
			t.Errorf(
				"golden mismatch for theme %q at byte offset %d\n  want: %q\n  got:  %q\nregenerate with: %s",
				stem, offset, escapeWindow(contextWindow(want, offset)), escapeWindow(contextWindow(out, offset)),
				regenerateCommand,
			)
		}
	}

	for stem := range representativeThemes {
		require.Truef(t, seenRepresentative[stem], "representative theme %q was not found among the bundled themes", stem)
	}

	sort.Slice(manifest, func(i, j int) bool { return manifest[i].stem < manifest[j].stem })

	manifestPath := filepath.Join("testdata", "goldens", "manifest.txt")

	if *update {
		var sb strings.Builder

		for _, entry := range manifest {
			fmt.Fprintf(&sb, "%s %s\n", entry.stem, entry.hash)
		}

		require.NoError(t, os.WriteFile(manifestPath, []byte(sb.String()), 0o644))

		return
	}

	raw, err := os.ReadFile(manifestPath)
	require.NoErrorf(t, err, "failed to read manifest %s; regenerate with: %s", manifestPath, regenerateCommand)

	wantHashes := make(map[string]string, len(manifest))

	for lineNo, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		require.Lenf(t, fields, 2, "manifest.txt line %d is malformed: %q", lineNo+1, line)

		wantHashes[fields[0]] = fields[1]
	}

	for _, entry := range manifest {
		wantHash, ok := wantHashes[entry.stem]
		if !ok {
			t.Errorf("theme %q is missing from manifest.txt; regenerate with: %s", entry.stem, regenerateCommand)
			continue
		}

		if wantHash != entry.hash {
			t.Errorf("theme %q hash mismatch: manifest has %s, rendered %s; regenerate with: %s",
				entry.stem, wantHash, entry.hash, regenerateCommand)
		}
	}

	for stem := range wantHashes {
		found := false

		for _, entry := range manifest {
			if entry.stem == stem {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("manifest.txt has a stale entry for %q, which no longer matches a bundled theme; regenerate with: %s",
				stem, regenerateCommand)
		}
	}
}
