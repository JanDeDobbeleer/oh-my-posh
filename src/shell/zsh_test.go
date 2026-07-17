package shell

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZshFeatures(t *testing.T) {
	got := allFeatures.Lines(ZSH).String("// these are the features")

	want := `// these are the features
enable_poshtooltips
_omp_create_widget zle-line-init _omp_zle-line-init
_omp_ftcs_marks=1
"$_omp_executable" upgrade --auto
"$_omp_executable" notice
_omp_cursor_positioning=1
_omp_enable_streaming=1
_omp_enable_vimode`

	assert.Equal(t, want, got)
}

// zshIsBufferComplete lifts _omp_is_buffer_complete out of the embedded init script,
// so the test drives the code we ship rather than a copy of it that can drift.
func zshIsBufferComplete(t *testing.T) string {
	t.Helper()

	const signature = "function _omp_is_buffer_complete() {"

	_, body, found := strings.Cut(zshInit, signature)
	require.True(t, found, "_omp_is_buffer_complete is missing from omp.zsh")

	body, _, found = strings.Cut(body, "\n}\n")
	require.True(t, found, "_omp_is_buffer_complete has no closing brace")

	return signature + body + "\n}\n"
}

// TestZshIsBufferComplete covers the check that keeps the primary prompt in place while
// a multi-line command is still being typed. Zsh reports an unterminated here-document
// as syntactically fine, so the here-document cases below are the interesting ones.
func TestZshIsBufferComplete(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh is not installed")
	}

	cases := []struct {
		Case     string
		Buffer   string
		Expected bool
	}{
		{Case: "Empty buffer", Buffer: "", Expected: true},
		{Case: "Single command", Buffer: "echo hi", Expected: true},
		{Case: "Unterminated quote", Buffer: `echo "abc`, Expected: false},
		{Case: "Trailing pipe", Buffer: "echo hi |", Expected: false},
		{Case: "Open for loop", Buffer: "for i in 1 2; do", Expected: false},
		{Case: "Closed for loop", Buffer: "for i in 1 2; do\necho $i\ndone", Expected: true},
		// A bare loop header parses as a complete empty loop under SHORT_LOOPS,
		// but the line editor still waits for the body - it must stay open.
		{Case: "Bare for header", Buffer: "for i in 1 2", Expected: false},
		{Case: "Bare while header", Buffer: "while true", Expected: false},
		{Case: "Bare until header", Buffer: "until false", Expected: false},
		{Case: "Bare select header", Buffer: "select x in a b", Expected: false},
		// The `<keyword> ... do;` form (a `do` immediately closed by `;`) is the
		// same SHORT_LOOPS leniency; it must stay open for every loop keyword.
		{Case: "for header with do;", Buffer: "for i in 1 2 do;", Expected: false},
		{Case: "while header with do;", Buffer: "while true do;", Expected: false},
		{Case: "until header with do;", Buffer: "until false do;", Expected: false},
		{Case: "select header with do;", Buffer: "select x in a b do;", Expected: false},
		// The keyword gate matches substrings ("before" contains "for"), so a
		// plain command that merely embeds a loop keyword must stay complete.
		{Case: "Keyword substring in a plain command", Buffer: "echo before", Expected: true},
		// Genuine zsh short-loop forms are complete and must not be held open.
		{Case: "for short-loop form", Buffer: "for i (1 2) echo $i", Expected: true},
		{Case: "C-style for short-loop form", Buffer: "for ((i = 0; i < 2; i++)) echo $i", Expected: true},
		{Case: "repeat short-loop form", Buffer: "repeat 2 echo hi", Expected: true},
		// foreach uses `end`, not `do`/`done`, so it needs no special handling -
		// but the keyword gate matches it (via "for"), so pin both ends.
		{Case: "Open foreach header", Buffer: "foreach i (1 2)", Expected: false},
		{Case: "Closed foreach loop", Buffer: "foreach i (1 2)\necho $i\nend", Expected: true},
		{Case: "Open case", Buffer: "case x in a) echo 1", Expected: false},
		{Case: "Closed case", Buffer: "case x in a) echo 1;; esac", Expected: true},
		{Case: "Odd trailing backslash", Buffer: `echo a \`, Expected: false},
		{Case: "Even trailing backslashes", Buffer: `echo a \\`, Expected: true},
		{Case: "Open here-document", Buffer: "cat <<EOF", Expected: false},
		{Case: "Open here-document with a body", Buffer: "cat <<EOF\nhello", Expected: false},
		{Case: "Closed here-document", Buffer: "cat <<EOF\nhello\nEOF", Expected: true},
		{Case: "Open here-document in a block", Buffer: "if true; then\ncat <<EOF\nhello\nEOF", Expected: false},
		{Case: "Closed here-document in a block", Buffer: "if true; then\ncat <<EOF\nhello\nEOF\nfi", Expected: true},
		{Case: "Open tab-stripped here-document", Buffer: "cat <<-EOF\n\thello", Expected: false},
		{Case: "Closed tab-stripped here-document", Buffer: "cat <<-EOF\n\thello\n\tEOF", Expected: true},
		{Case: "Two here-documents", Buffer: "cat <<EOF\nx\nEOF\ncat <<END\ny\nEND", Expected: true},
		{Case: "Here-string is not a here-document", Buffer: `cat <<<"hi"`, Expected: true},
		{Case: "Left shift is not a here-document", Buffer: "echo $((1 << 2))", Expected: true},
		{Case: "Delimiter word used as a command", Buffer: "echo one\nEOF\necho two", Expected: true},
		{Case: "Delimiter word before an open block", Buffer: "echo one\nEOF\nfor i in 1 2; do", Expected: false},
		{Case: "Here-document operator inside a body", Buffer: "cat <<EOF\ncat <<END\nEOF", Expected: true},
		{Case: "Case terminator in an open here-document", Buffer: "cat <<EOF\n;;", Expected: false},
		{Case: "Case terminator in a closed here-document", Buffer: "cat <<EOF\n;;\nEOF", Expected: true},
		{Case: "Case terminator as an open delimiter", Buffer: "cat <<';;'", Expected: false},
		{Case: "Case terminator as a closed delimiter", Buffer: "cat <<';;'\nbody\n;;", Expected: true},
	}

	script := zshIsBufferComplete(t) + `
PREBUFFER=''
BUFFER=$1
if _omp_is_buffer_complete; then print -r -- COMPLETE; else print -r -- INCOMPLETE; fi
`

	// -f plus an empty HOME keeps the startup files out of the parse: every zsh the
	// function shells out to reads ~/.zshenv, and an option set there can change how
	// a buffer parses.
	home := t.TempDir()

	for _, tc := range cases {
		cmd := exec.CommandContext(t.Context(), "zsh", "-f", "-c", script, "omp", tc.Buffer)
		cmd.Env = []string{"HOME=" + home, "PATH=" + os.Getenv("PATH")}

		out, err := cmd.Output()
		require.NoError(t, err, tc.Case)

		want := "INCOMPLETE"
		if tc.Expected {
			want = "COMPLETE"
		}

		assert.Equal(t, want, strings.TrimSpace(string(out)), tc.Case)
	}
}
