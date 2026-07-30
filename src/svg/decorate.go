package svg

import (
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// cursorText/watermarkText are the retired PNG renderer's own fixed cursor
// glyph and watermark string (image.go's cleanContent: ir.Cursor's default
// "_", and the literal "ohmyposh.dev" it always appended). The maintainer
// chose full parity with the old renderer here: both are kept, not just one
// of the two — see the task brief this package was rebuilt from.
const (
	cursorText    = "_"
	watermarkText = "ohmyposh.dev"
)

// Cursor locates where the cursor indicator belongs in a captured Run
// stream: a row index, and an index within that row's runs. It is
// prompt.Engine.CursorAnchor's return value, passed through Options rather
// than rediscovered here, because nothing in the Run stream itself
// distinguishes a left-aligned block's runs from a right-aligned one's — the
// engine appends both to the same row (see prompt.appendCapturedRuns).
type Cursor struct {
	Row int
	Run int
}

// decorate appends the retired PNG renderer's own trailing content — a
// cursor indicator glued onto the end of the prompt's own output, then one
// blank line, then the bold "ohmyposh.dev" watermark on its own line — to
// rows, returning a new slice rather than mutating the caller's own capture.
//
// cursor says where "the end of the prompt's own output" is. Appending to
// the last row's tail — which is what this did before — puts the cursor
// after a right-aligned block, since the engine writes the rprompt into the
// same row; a real terminal leaves the cursor back at the primary prompt's
// end, which is exactly the position writePrimaryRightPrompt saves and
// restores around the rprompt. A nil cursor falls back to the tail, for a
// caller with no anchor to give.
//
// Both additions use the ansiColorCodes name "default" as their foreground
// source rather than leaving it empty: an empty source means "keep
// whatever color was already active" (see resolveChannel's doc comment),
// which for the cursor would mean inheriting the last segment's own
// foreground — image.go's cursor and watermark render in the fixed default
// foreground color regardless of what precedes them, because cleanContent
// appends them after the prompt's own trailing SGR reset, and "default" is
// this package's equivalent of that reset state.
func decorate(rows [][]terminal.Run, cursor *Cursor) [][]terminal.Run {
	if len(rows) == 0 {
		rows = [][]terminal.Run{{}}
	}

	rows = append([][]terminal.Run{}, rows...)

	row := len(rows) - 1
	at := len(rows[row])

	if cursor != nil && cursor.Row >= 0 && cursor.Row < len(rows) && cursor.Run >= 0 && cursor.Run <= len(rows[cursor.Row]) {
		row, at = cursor.Row, cursor.Run
	}

	rows[row] = insertCursor(rows[row], at)

	rows = append(rows, nil, []terminal.Run{{
		Text:             watermarkText,
		Cells:            len(watermarkText),
		ForegroundSource: color.Ansi("default"),
		Attributes:       [8]uint8{attrBold: 1},
	}})

	return rows
}

// insertCursor splices the cursor run into row at index at. When an
// engine-inserted alignment gap follows (see isGapRun) the cursor's one cell
// is taken out of that gap rather than added to the row, so a row carrying a
// right-aligned block keeps its exact width and the rprompt stays flush
// against the right edge — the same thing a real terminal does, where the
// cursor occupies a cell that is already there rather than pushing the line
// wider.
func insertCursor(row []terminal.Run, at int) []terminal.Run {
	out := make([]terminal.Run, 0, len(row)+1)
	out = append(out, row[:at]...)

	out = append(out, terminal.Run{
		Text:             cursorText,
		Cells:            len(cursorText),
		ForegroundSource: color.Ansi("default"),
		// Explicit, not empty: an empty BackgroundSource means "keep
		// whatever was already active" (see resolveChannel's doc comment),
		// which here would paint the cursor with whatever segment's
		// background happened to precede it. The cursor itself carries no
		// background of its own, exactly like image.go's reference renders
		// show it.
		BackgroundSource: color.Transparent,
	})

	tail := row[at:]

	if len(tail) > 0 && isGapRun(&tail[0]) {
		gap := tail[0]
		gap.Text = gap.Text[1:]
		gap.Cells--

		if gap.Cells > 0 {
			out = append(out, gap)
		}

		tail = tail[1:]
	}

	return append(out, tail...)
}
