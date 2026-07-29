package prompt

import (
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// CapturedRuns returns the Run stream captured for the whole rendered
// prompt, one []terminal.Run per output row (see startRunCapture/
// newCapturedRow), when terminal.CaptureRuns was true for the render that
// produced it (see Primary). Empty otherwise: every capture call below is a
// single `if terminal.CaptureRuns` check and nothing else runs, mirroring
// terminal.Runs' own off-by-default contract.
func (e *Engine) CapturedRuns() [][]terminal.Run {
	return e.capturedRows
}

// startRunCapture (re)initializes the row accumulator at the start of a
// render. A no-op unless terminal.CaptureRuns is set, so repeated renders off
// the same Engine with capture disabled never pay for the reset.
func (e *Engine) startRunCapture() {
	e.cursorRow = -1
	e.cursorRun = 0

	if !terminal.CaptureRuns {
		e.capturedRows = nil
		return
	}

	e.capturedRows = [][]terminal.Run{{}}
}

// markCursorAnchor records the current end of the captured stream as the
// place a shell would leave the cursor. primaryInternal calls it once the
// primary prompt itself is fully written (final space included) and before
// any right-aligned block is appended — writePrimaryRightPrompt writes the
// rprompt into that same row between a SaveCursorPosition/
// RestoreCursorPosition pair, precisely so the terminal's own cursor ends up
// back here rather than past the rprompt.
func (e *Engine) markCursorAnchor() {
	if !terminal.CaptureRuns || len(e.capturedRows) == 0 {
		return
	}

	e.cursorRow = len(e.capturedRows) - 1
	e.cursorRun = len(e.capturedRows[e.cursorRow])
}

// CursorAnchor returns the row index, and the index within that row's runs,
// that markCursorAnchor recorded. ok is false when nothing was captured (see
// CapturedRuns' own off-by-default contract), leaving a caller to fall back
// to the end of the last row.
func (e *Engine) CursorAnchor() (row, run int, ok bool) {
	if e.cursorRow < 0 || e.cursorRow >= len(e.capturedRows) {
		return 0, 0, false
	}

	return e.cursorRow, e.cursorRun, true
}

// captureBlockRuns returns a snapshot of terminal.Runs() for the block/rprompt
// that just wrote through terminal.Write, for writeBlock (via
// renderBlockSegments/renderBlockFromCache) to place once it knows the
// block's column (left-aligned, or right-aligned behind padding/a filler).
// This must run BEFORE the matching terminal.String() call, not after:
// String's own defer resets the run stream (runsState.Runs[:0]) before
// String returns to its caller, so capturing terminal.Runs() after String
// has already returned sees it truncated to nothing.
func (e *Engine) captureBlockRuns() []terminal.Run {
	if !terminal.CaptureRuns {
		return nil
	}

	return slices.Clone(terminal.Runs())
}

// expandFillerRuns rebuilds the Run-stream equivalent of shouldFill's own
// strings.Repeat(filler, repeat) + strings.Repeat(" ", unfilled) — from
// pattern, the single filler pattern shouldFill captured before its matching
// terminal.String() call (same reasoning as captureBlockRuns) — using the
// same padLength shouldFill already has in hand, so a filler-padded row shows
// the same repeated pattern the ANSI text does.
func expandFillerRuns(pattern []terminal.Run, padLength int) []terminal.Run {
	if !terminal.CaptureRuns || len(pattern) == 0 || padLength <= 0 {
		return nil
	}

	lenFiller := 0
	for _, run := range pattern {
		lenFiller += run.Cells
	}

	if lenFiller == 0 {
		return nil
	}

	repeat := padLength / lenFiller
	unfilled := padLength % lenFiller

	runs := make([]terminal.Run, 0, repeat*len(pattern)+1)
	for range repeat {
		runs = append(runs, pattern...)
	}

	return append(runs, gapRun(unfilled)...)
}

// appendCapturedRuns places gap (padding/filler cells that never passed
// through terminal.Write) followed by runs (a block's own captured content)
// at the end of the current row.
func (e *Engine) appendCapturedRuns(gap, runs []terminal.Run) {
	if !terminal.CaptureRuns || len(e.capturedRows) == 0 {
		return
	}

	idx := len(e.capturedRows) - 1
	e.capturedRows[idx] = append(e.capturedRows[idx], gap...)
	e.capturedRows[idx] = append(e.capturedRows[idx], runs...)
}

// newCapturedRow starts a new output row, mirroring writeNewline's own reset
// of currentLineLength: engine-appended newlines never pass through
// terminal.Write, so the Run stream has no way to see them on its own.
func (e *Engine) newCapturedRow() {
	if !terminal.CaptureRuns || len(e.capturedRows) == 0 {
		return
	}

	e.capturedRows = append(e.capturedRows, []terminal.Run{})
}

// gapRun builds a blank, colorless run of n cells: the Run-stream equivalent
// of the literal spaces engine.go writes directly for right-block padding,
// which never passes through terminal.Write and so never produces a Run of
// its own.
func gapRun(n int) []terminal.Run {
	if n <= 0 {
		return nil
	}

	return []terminal.Run{{Text: strings.Repeat(" ", n), Cells: n}}
}
