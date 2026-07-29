package svg

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contentRun is a small helper building a Run that is definitely NOT a gap
// run (a resolved foreground/background), so tests reliably distinguish
// genuine content from engine-inserted alignment padding.
func contentRun(text string) terminal.Run {
	return terminal.Run{
		Text: text, Cells: len([]rune(text)),
		ForegroundSource: "red", BackgroundSource: "blue", Mode: terminal.RunNormal,
	}
}

// TestIsGapRun pins the exact shape prompt.gapRun/fillerCaptureRuns produce
// for engine-inserted alignment padding (see runs.go), as opposed to every
// other way a Run can end up carrying only space characters.
func TestIsGapRun(t *testing.T) {
	cases := []struct {
		Case     string
		Run      terminal.Run
		Expected bool
	}{
		{Case: "gapRun shape", Run: terminal.Run{Text: "     ", Cells: 5}, Expected: true},
		{Case: "zero cells", Run: terminal.Run{Text: "", Cells: 0}, Expected: false},
		{
			Case: "colored space run is genuine content, not padding",
			Run:  terminal.Run{Text: "   ", Cells: 3, BackgroundSource: "blue"}, Expected: false,
		},
		{
			Case: "non-space text", Run: terminal.Run{Text: "ab ", Cells: 3}, Expected: false,
		},
		{
			Case: "reverse-video space is a real cutout glyph, not padding",
			Run:  terminal.Run{Text: " ", Cells: 1, Mode: terminal.RunReverseVideo}, Expected: false,
		},
		{
			Case: "attributed space (e.g. underlined) is not padding",
			Run:  terminal.Run{Text: " ", Cells: 1, Attributes: [8]uint8{attrUnderline: 1}}, Expected: false,
		},
		{
			Case: "gradient-stamped space is not padding",
			Run:  terminal.Run{Text: " ", Cells: 1, ForegroundRGB: &color.RGB{R: 1, G: 2, B: 3}}, Expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			assert.Equal(t, tc.Expected, isGapRun(&tc.Run))
		})
	}
}

// TestLongestGapRegion pins that a filler pattern's own leftover cells can
// appear as several adjacent gap Runs (see fillerCaptureRuns), and the
// region search must span all of them, not stop at the first.
func TestLongestGapRegion(t *testing.T) {
	row := []terminal.Run{
		contentRun("left"),
		{Text: "  ", Cells: 2},  // too short alone to count as padding
		{Text: "   ", Cells: 3}, // adjacent gap run: combined with the above, 5 cells
		contentRun("right"),
		{Text: "          ", Cells: 10}, // the longer, single gap region
	}

	start, count, cells, ok := longestGapRegion(row)
	require.True(t, ok)
	assert.Equal(t, 4, start)
	assert.Equal(t, 1, count)
	assert.Equal(t, 10, cells)

	// Dropping the standalone long run makes the two merged short ones win.
	row = row[:4]
	start, count, cells, ok = longestGapRegion(row)
	require.True(t, ok)
	assert.Equal(t, 1, start)
	assert.Equal(t, 2, count)
	assert.Equal(t, 5, cells)
}

// TestLongestGapRegionNoneQualifies pins the false-ok case: every gap run
// present is shorter than minGapRun.
func TestLongestGapRegionNoneQualifies(t *testing.T) {
	row := []terminal.Run{contentRun("a"), {Text: "  ", Cells: 2}, contentRun("b")}

	_, _, _, ok := longestGapRegion(row)
	assert.False(t, ok)
}

// TestFitRowWithinBudgetIsUnchanged pins fitRow's no-op path: nothing about
// a row within budget is touched, not even a defensive copy.
func TestFitRowWithinBudgetIsUnchanged(t *testing.T) {
	row := []terminal.Run{contentRun("hello")}

	got := fitRow(row, 10)

	require.Len(t, got, 1)
	assert.Equal(t, row, got[0])
}

// TestFitRowCollapsesGapRunFirst pins the retired PNG renderer's own policy
// (image.go's fitRow): an overflowing row with a long padding run gets cells
// removed from the gap before anything is hard-wrapped, keeping the
// right-aligned content flush against the row's own right edge.
func TestFitRowCollapsesGapRunFirst(t *testing.T) {
	row := []terminal.Run{
		contentRun("left"),              // 4 cells
		{Text: "          ", Cells: 10}, // 10 cells of alignment padding
		contentRun("right"),             // 5 cells
	}
	// total 19 cells; budget 15 -> must remove 4 padding cells, nothing else.

	got := fitRow(row, 15)

	require.Len(t, got, 1)
	assert.Equal(t, 15, rowCells(got[0]))

	// The gap run shrank from the front (6 cells left), content on both
	// sides is untouched.
	require.Len(t, got[0], 3)
	assert.Equal(t, "left", got[0][0].Text)
	assert.Equal(t, "      ", got[0][1].Text)
	assert.Equal(t, 6, got[0][1].Cells)
	assert.Equal(t, "right", got[0][2].Text)
}

// TestFitRowCollapsesAcrossAdjacentGapRuns pins collapseGap's own multi-run
// case: a filler-repeated pattern followed by its own leftover cells (see
// fillerCaptureRuns) is several adjacent gap Runs, and removal must be able
// to fully drop an early one and partially trim the next.
func TestFitRowCollapsesAcrossAdjacentGapRuns(t *testing.T) {
	row := []terminal.Run{
		contentRun("l"),           // 1 cell
		{Text: "abc", Cells: 3},   // a colorless "filler" run - not spaces, so NOT a gap run
		{Text: "   ", Cells: 3},   // gap run 1
		{Text: "     ", Cells: 5}, // gap run 2, adjacent - combined region is 8 cells
		contentRun("r"),           // 1 cell
	}
	// total 1+3+3+5+1 = 13 cells; budget 7 -> remove 6 from the 8-cell region.

	got := fitRow(row, 7)

	require.Len(t, got, 1)
	assert.Equal(t, 7, rowCells(got[0]))

	// First gap run (3 cells) fully consumed, second trimmed from 5 to 2.
	var texts []string
	for _, r := range got[0] {
		texts = append(texts, r.Text)
	}
	assert.Equal(t, []string{"l", "abc", "  ", "r"}, texts)
}

// TestFitRowWrapsWhenGapAloneIsNotEnough pins the fallback path: an
// overflowing row whose padding run alone can't bring it under budget is
// hard-wrapped with whatever's left of that padding.
func TestFitRowWrapsWhenGapAloneIsNotEnough(t *testing.T) {
	row := []terminal.Run{
		contentRun("0123456789"),   // 10 cells
		{Text: "      ", Cells: 6}, // 6 cells of padding - not enough alone
		contentRun("abcde"),        // 5 cells
	}
	// total 21 cells; budget 10. The whole 6-cell gap collapses (removing
	// all of it still leaves 15 > 10), so it must also hard-wrap.

	got := fitRow(row, 10)

	require.Len(t, got, 2)
	assert.LessOrEqual(t, rowCells(got[0]), 10)
	assert.LessOrEqual(t, rowCells(got[1]), 10)

	// No run's text lost any runes across the whole wrap: reassembling every
	// row's text must reproduce the original content minus the padding that
	// was removed to make room.
	var rebuilt strings.Builder
	for _, r := range got {
		for i := range r {
			rebuilt.WriteString(r[i].Text)
		}
	}
	assert.Equal(t, "0123456789abcde", rebuilt.String())
}

// TestWrapRowNeverSplitsARune pins wrapRow's own core guarantee: a run's
// text is only ever cut on a rune boundary, even for multi-byte runes, and
// every half keeps the parent run's full style.
func TestWrapRowNeverSplitsARune(t *testing.T) {
	row := []terminal.Run{{
		Text: "café", Cells: 4, // é is multi-byte in UTF-8
		ForegroundSource: "red", BackgroundSource: "blue", Attributes: [8]uint8{attrBold: 1},
	}}

	got := wrapRow(row, 3)

	require.Len(t, got, 2)
	require.Len(t, got[0], 1)
	require.Len(t, got[1], 1)

	assert.Equal(t, "caf", got[0][0].Text)
	assert.Equal(t, 3, got[0][0].Cells)
	assert.Equal(t, "é", got[1][0].Text)
	assert.Equal(t, 1, got[1][0].Cells)

	// Style survives the split on both halves.
	for _, half := range []terminal.Run{got[0][0], got[1][0]} {
		assert.Equal(t, color.Ansi("red"), half.ForegroundSource)
		assert.Equal(t, color.Ansi("blue"), half.BackgroundSource)
		assert.Equal(t, uint8(1), half.Attributes[attrBold])
	}
}

// TestWrapRowZeroCellRunNeverSplits pins that a zero-cell run (e.g. a
// hyperlink's URL text - see terminal.Run's doc comment) rides along with
// whichever row it lands next to instead of ever being split or dropped.
func TestWrapRowZeroCellRunNeverSplits(t *testing.T) {
	row := []terminal.Run{
		contentRun("0123456789"),
		{Text: "https://example.com", Cells: 0},
	}

	got := wrapRow(row, 10)

	require.Len(t, got, 1)
	require.Len(t, got[0], 2)
	assert.Equal(t, "https://example.com", got[0][1].Text)
	assert.Equal(t, 0, got[0][1].Cells)
}
