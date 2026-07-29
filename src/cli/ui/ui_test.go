package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelectReadsAChoice pins the picker's contract: it prints a numbered list and answers with a
// zero-based index, so a caller indexes its own slice with what comes back.
func TestSelectReadsAChoice(t *testing.T) {
	var out bytes.Buffer

	index, err := Select(strings.NewReader("2\n"), &out, "Pick one", []string{"alpha", "beta", "gamma"})

	require.NoError(t, err)
	assert.Equal(t, 1, index, "the answer is zero-based; the prompt is not")
	assert.Contains(t, out.String(), "1  alpha")
	assert.Contains(t, out.String(), "3  gamma")
}

// TestSelectCancels covers the two ways a user declines: an empty line, and a closed stdin. Both
// are intent, not failure, so both answer ErrCancelled rather than something a caller would print
// as an error.
func TestSelectCancels(t *testing.T) {
	for name, input := range map[string]string{
		"empty line": "\n",
		"eof":        "",
	} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer

			_, err := Select(strings.NewReader(input), &out, "Pick one", []string{"alpha"})

			assert.ErrorIs(t, err, ErrCancelled)
		})
	}
}

// TestSelectRejectsOutOfRange pins that a number outside the list is an error rather than a
// silent clamp - installing the wrong font because 99 became 12 is worse than being told to
// try again.
func TestSelectRejectsOutOfRange(t *testing.T) {
	for name, input := range map[string]string{
		"too high":     "4\n",
		"zero":         "0\n",
		"not a number": "beta\n",
	} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer

			_, err := Select(strings.NewReader(input), &out, "Pick one", []string{"a", "b", "c"})

			require.Error(t, err)
			assert.NotErrorIs(t, err, ErrCancelled, "a bad answer is not the same as declining to answer")
		})
	}
}

// TestSelectAlignsIndices pins the column width: a two-digit list indents the single-digit rows
// so the labels line up.
func TestSelectAlignsIndices(t *testing.T) {
	items := make([]string, 12)
	for i := range items {
		items[i] = "font"
	}

	var out bytes.Buffer

	_, _ = Select(strings.NewReader("1\n"), &out, "Pick one", items)

	assert.Contains(t, out.String(), "   1  font")
	assert.Contains(t, out.String(), "  12  font")
}

// TestStatusRepaintsInPlace pins that the status line erases before it writes. Without that a
// shorter message leaves the tail of a longer one behind, which is the whole reason this repaints
// rather than just printing.
func TestStatusRepaintsInPlace(t *testing.T) {
	var out bytes.Buffer

	status := NewStatus(&out)
	status.Start("downloading something with a long name")
	status.Set("done")
	status.Stop("installed")

	written := out.String()

	assert.Contains(t, written, hideCursor)
	assert.Contains(t, written, showCursor, "the cursor must come back, or the terminal is left broken")
	assert.Contains(t, written, eraseLine+"installed")
	assert.True(t, strings.HasSuffix(written, "\n"), "the final line must be terminated")
}

// TestStatusStopClearsWhenSilent covers the other ending: no final message means the line goes
// away entirely, leaving the terminal as it was found.
func TestStatusStopClearsWhenSilent(t *testing.T) {
	var out bytes.Buffer

	status := NewStatus(&out)
	status.Start("working")
	status.Stop("")

	assert.Contains(t, out.String(), eraseLine+showCursor)
}

// TestStatusStopIsSafeWithoutStart pins that stopping something never started is a no-op rather
// than a panic on a nil channel - a deferred Stop runs even when the path that would have
// started it returned early.
func TestStatusStopIsSafeWithoutStart(t *testing.T) {
	var out bytes.Buffer

	assert.NotPanics(t, func() { NewStatus(&out).Stop("done") })
	assert.Empty(t, out.String())
}

// TestProgressClampsAndFills pins both ends of the bar. A server that understates content-length
// drives the fraction past 1, which must not print a negative-count repeat - that panics.
func TestProgressClampsAndFills(t *testing.T) {
	for name, test := range map[string]struct {
		expect   string
		fraction float64
	}{
		"empty":     {"  0%", 0},
		"half":      {" 50%", 0.5},
		"full":      {"100%", 1},
		"overshoot": {"100%", 1.7},
		"negative":  {"  0%", -0.5},
	} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer

			assert.NotPanics(t, func() { NewProgress(&out, "downloading").Set(test.fraction) })
			assert.Contains(t, out.String(), test.expect)
		})
	}
}
