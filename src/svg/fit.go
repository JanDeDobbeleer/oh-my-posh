package svg

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// minGapRun mirrors the retired PNG renderer's own minPaddingRun
// (image.go): the minimum length, in cells, of a run of literal space
// characters treated as prompt-engine alignment padding (a right-aligned
// block or rprompt's own filler — see prompt.gapRun/fillerCaptureRuns)
// rather than incidental whitespace that happens to be part of a segment's
// own content.
const minGapRun = 5

// fitRow fits a single captured row to at most maxCells cells, porting the
// retired PNG renderer's own fitRow policy (image.go, ~:539-562) from pixel
// measurements to cell counts: this package already lays runs out on a
// fixed cell grid (see Encode's doc comment), so there is no font metric to
// measure against here, only Run.Cells arithmetic.
//
//   - a row already within budget is returned unchanged
//   - a row that overflows because it contains a long (>= minGapRun) run of
//     colorless literal-space cells (an engine-inserted alignment gap — see
//     isGapRun) has cells removed from that gap first, from its front,
//     keeping right-aligned content flush against the right edge
//   - anything still over budget is hard-wrapped onto as many additional
//     rows as it takes, never splitting a Run's text mid-rune (see wrapRow)
//
// The engine renders at exactly Columns columns, so in practice almost
// every row already fits and this is a no-op; it exists for the rows a
// later change (double-width glyph cell allocation) will push over by a
// few cells despite that.
func fitRow(row []terminal.Run, maxCells int) [][]terminal.Run {
	cells := rowCells(row)
	if cells <= maxCells {
		return [][]terminal.Run{row}
	}

	if start, count, gapCells, ok := longestGapRegion(row); ok {
		remove := min(cells-maxCells, gapCells)
		row = collapseGap(row, start, count, remove)
		cells -= remove

		if cells <= maxCells {
			return [][]terminal.Run{row}
		}
	}

	return wrapRow(row, maxCells)
}

// rowCells sums a row's rendered width in cells, the same quantity Encode's
// own column-advance loop tracks per run (see encodeRow).
func rowCells(row []terminal.Run) int {
	cells := 0

	for i := range row {
		cells += row[i].Cells
	}

	return cells
}

// isGapRun reports whether run is the Run-stream shape prompt.gapRun and
// fillerCaptureRuns' own leftover cells produce for engine-inserted
// alignment padding: a colorless span of literal space characters, as
// opposed to genuine content that happens to be blank, which always
// carries a resolved foreground/background source, an RGB, a non-normal
// render mode, or a style attribute.
func isGapRun(run *terminal.Run) bool {
	if run.Cells == 0 {
		return false
	}

	if run.ForegroundRGB != nil || run.BackgroundRGB != nil {
		return false
	}

	if !run.ForegroundSource.IsEmpty() || !run.BackgroundSource.IsEmpty() {
		return false
	}

	if run.Mode != terminal.RunNormal {
		return false
	}

	if run.Attributes != ([8]uint8{}) {
		return false
	}

	return len(run.Text) == run.Cells && strings.Trim(run.Text, " ") == ""
}

// longestGapRegion finds the longest maximal run of consecutive gap runs
// (see isGapRun) whose combined width is at least minGapRun, mirroring the
// retired PNG renderer's own longestPaddingRun but over Runs rather than
// individual glyphs: a filler-repeated pattern followed by its own leftover
// cells (see fillerCaptureRuns) can appear as several adjacent gap Runs, and
// collapsing must be able to remove cells across all of them, not just the
// last one.
func longestGapRegion(row []terminal.Run) (start, count, cells int, ok bool) {
	bestStart, bestCount, bestCells := -1, 0, 0
	curStart, curCount, curCells := -1, 0, 0

	flush := func() {
		if curCells >= minGapRun && curCells > bestCells {
			bestStart, bestCount, bestCells = curStart, curCount, curCells
		}
	}

	for i := range row {
		if isGapRun(&row[i]) {
			if curCount == 0 {
				curStart = i
			}

			curCount++
			curCells += row[i].Cells

			continue
		}

		flush()
		curCount, curCells = 0, 0
	}

	flush()

	if bestCount == 0 {
		return 0, 0, 0, false
	}

	return bestStart, bestCount, bestCells, true
}

// collapseGap removes remove cells from the front of the gap region
// row[start:start+count], mirroring the retired PNG renderer's own
// row.glyphs[:start] + row.glyphs[start+remove:] splice (image.go's
// fitRow) at Run granularity: runs entirely inside the removed span are
// dropped, and the run remove finally lands in has its leading cells (and
// matching leading spaces of its Text) trimmed rather than dropped whole.
func collapseGap(row []terminal.Run, start, count, remove int) []terminal.Run {
	out := make([]terminal.Run, 0, len(row))
	out = append(out, row[:start]...)

	left := remove

	for i := start; i < start+count; i++ {
		run := row[i]

		if left >= run.Cells {
			left -= run.Cells
			continue
		}

		if left > 0 {
			run.Cells -= left
			run.Text = strings.Repeat(" ", run.Cells)
			left = 0
		}

		out = append(out, run)
	}

	out = append(out, row[start+count:]...)

	return out
}

// wrapRow hard-wraps row into as many rows as it takes to keep each one at
// most maxCells cells wide, exactly like a real terminal would, mirroring
// the retired PNG renderer's own wrapRow (image.go) at Run granularity
// instead of per-glyph: a Run's Text is only ever split on a rune boundary
// (never mid-rune), and every split half keeps the parent Run's full style.
//
// Splitting assumes one cell per rune, true for every Run this package's
// caller currently produces (see terminal.Run's doc comment) — a run whose
// Cells already differs from its rune count only does so via zero-width
// hyperlink text, handled separately below. A future double-width-glyph
// allocation (the same change this whole fitRow port exists to absorb the
// impact of — see fitRow's doc comment) would need this split taught actual
// per-glyph cell widths instead of one-cell-per-rune.
//
// maxCells must be positive — fitRow only ever calls this after failing to
// bring a row into budget by collapsing a gap run, which requires a positive
// budget to have overflowed in the first place; a direct caller must enforce
// that itself.
func wrapRow(row []terminal.Run, maxCells int) [][]terminal.Run {
	var rows [][]terminal.Run

	var current []terminal.Run

	used := 0

	flush := func() {
		rows = append(rows, current)
		current = nil
		used = 0
	}

	for _, run := range row {
		if run.Cells == 0 {
			current = append(current, run)
			continue
		}

		remaining := run

		for {
			budget := maxCells - used
			if budget <= 0 {
				flush()
				budget = maxCells
			}

			if remaining.Cells <= budget {
				current = append(current, remaining)
				used += remaining.Cells

				break
			}

			head, tail := splitRun(&remaining, budget)
			current = append(current, head)
			flush()

			remaining = tail
		}
	}

	if len(current) > 0 || len(rows) == 0 {
		rows = append(rows, current)
	}

	return rows
}

// splitRun splits run into a leading head of exactly n cells and a
// trailing tail with the remainder, both carrying run's full style; see
// wrapRow's doc comment for the one-cell-per-rune assumption this relies
// on. Takes run by pointer (it's 120 bytes, over gocritic's hugeParam
// threshold) purely to avoid the copy; run itself is never mutated.
func splitRun(run *terminal.Run, n int) (head, tail terminal.Run) {
	runes := []rune(run.Text)
	if n > len(runes) {
		n = len(runes)
	}

	head = *run
	head.Text = string(runes[:n])
	head.Cells = n

	tail = *run
	tail.Text = string(runes[n:])
	tail.Cells = run.Cells - n

	return head, tail
}
