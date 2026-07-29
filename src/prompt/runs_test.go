package prompt

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveCaptureRunsGlobal snapshots terminal.CaptureRuns and restores it via
// t.Cleanup, so a test enabling capture never leaks it into another test in
// this package (which does not expect the run-stream bookkeeping to run).
func saveCaptureRunsGlobal(t *testing.T) {
	t.Helper()

	orig := terminal.CaptureRuns
	t.Cleanup(func() { terminal.CaptureRuns = orig })
}

// TestGapRun pins gapRun's contract: a positive n produces one blank,
// colorless run of exactly n cells (the Run-stream equivalent of the literal
// spaces engine.go writes for right-block padding); zero or negative
// produces nothing, so callers never need to guard the call site themselves.
func TestGapRun(t *testing.T) {
	cases := []struct {
		Case     string
		Expected []terminal.Run
		N        int
	}{
		{Case: "positive", N: 3, Expected: []terminal.Run{{Text: "   ", Cells: 3}}},
		{Case: "zero", N: 0, Expected: nil},
		{Case: "negative", N: -1, Expected: nil},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			assert.Equal(t, tc.Expected, gapRun(tc.N))
		})
	}
}

// TestEngineRunCaptureOffIsANoOp pins the off-by-default contract: with
// terminal.CaptureRuns false, none of the capture helpers touch Engine state,
// mirroring terminal.Runs' own off-by-default contract (see run.go).
func TestEngineRunCaptureOffIsANoOp(t *testing.T) {
	saveCaptureRunsGlobal(t)
	terminal.CaptureRuns = false

	e := &Engine{}

	e.startRunCapture()
	assert.Nil(t, e.capturedRows)

	runs := e.captureBlockRuns()
	assert.Nil(t, runs)

	e.appendCapturedRuns(gapRun(2), []terminal.Run{{Text: "x", Cells: 1}})
	assert.Nil(t, e.capturedRows)

	e.newCapturedRow()
	assert.Nil(t, e.capturedRows)
}

// TestEngineRunCaptureRowBookkeeping exercises the capture primitives
// directly (startRunCapture/appendCapturedRuns/newCapturedRow), independent
// of a full render, mirroring how writeBlock/writeNewline/writePrimaryRightPrompt
// drive them: a gap run followed by a block's own runs lands on the current
// row, and newCapturedRow starts a fresh one.
func TestEngineRunCaptureRowBookkeeping(t *testing.T) {
	saveCaptureRunsGlobal(t)
	terminal.CaptureRuns = true

	e := &Engine{}
	e.startRunCapture()
	require.Len(t, e.capturedRows, 1)

	e.appendCapturedRuns(nil, []terminal.Run{{Text: "AB", Cells: 2}})
	e.appendCapturedRuns(gapRun(3), []terminal.Run{{Text: "R", Cells: 1}})

	require.Len(t, e.capturedRows, 1)
	assert.Equal(t, []terminal.Run{
		{Text: "AB", Cells: 2},
		{Text: "   ", Cells: 3},
		{Text: "R", Cells: 1},
	}, e.capturedRows[0])

	e.newCapturedRow()
	e.appendCapturedRuns(nil, []terminal.Run{{Text: "CD", Cells: 2}})

	require.Len(t, e.capturedRows, 2)
	assert.Equal(t, []terminal.Run{{Text: "CD", Cells: 2}}, e.capturedRows[1])
}

// TestExpandFillerRuns pins expandFillerRuns' repeat math against
// shouldFill's own strings.Repeat(filler, repeat) + strings.Repeat(" ",
// unfilled) (engine.go): the captured filler pattern must repeat whole, with
// the remainder padded as blank cells, never truncating a repeat mid-run.
func TestExpandFillerRuns(t *testing.T) {
	saveCaptureRunsGlobal(t)
	terminal.CaptureRuns = true

	cases := []struct {
		Case          string
		Pattern       []terminal.Run
		PadLength     int
		ExpectedCells int
		ExpectedRuns  int
	}{
		{
			Case: "exact multiple leaves no remainder", PadLength: 5, ExpectedRuns: 5, ExpectedCells: 5,
			Pattern: []terminal.Run{{Text: "-", Cells: 1, ForegroundSource: "red"}},
		},
		{
			// a 2-cell filler repeated into 7 cells: 3 whole repeats (6
			// cells) plus a 1-cell blank remainder, never a truncated 4th
			// repeat of the filler pattern itself.
			Case: "remainder pads with a blank cell run", PadLength: 7, ExpectedRuns: 4, ExpectedCells: 7,
			Pattern: []terminal.Run{{Text: "--", Cells: 2, ForegroundSource: "red"}},
		},
		{
			Case: "zero pad produces nothing", PadLength: 0, ExpectedRuns: 0, ExpectedCells: 0,
			Pattern: []terminal.Run{{Text: "-", Cells: 1, ForegroundSource: "red"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			runs := expandFillerRuns(tc.Pattern, tc.PadLength)

			require.Len(t, runs, tc.ExpectedRuns)

			total := 0
			for _, run := range runs {
				total += run.Cells
			}

			assert.Equal(t, tc.ExpectedCells, total)
		})
	}
}

// TestCapturedRunsPrimaryLayout renders a small multi-block, multi-row
// config through the real Primary() pipeline (newEngine, matching
// golden_test.go's own pattern) with terminal.CaptureRuns on, and checks the
// resulting rows account for every cell the engine itself lays out: a left
// block, a right-aligned block sharing its row (as chips.omp.json exercises,
// via canWriteRightBlock's padding), and a second row started by
// block.Newline.
func TestCapturedRunsPrimaryLayout(t *testing.T) {
	saveCaptureRunsGlobal(t)

	const terminalWidth = 20

	flags := &runtime.Flags{
		Shell:         shell.GENERIC,
		TerminalWidth: terminalWidth,
		IsPrimary:     true,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	cfg := &config.Config{
		Blocks: []*config.Block{
			{
				Type:      config.Prompt,
				Alignment: config.Left,
				Segments: []*config.Segment{
					{Type: "text", Template: "AB", Foreground: "red", Background: "blue"},
				},
			},
			{
				Type:      config.Prompt,
				Alignment: config.Right,
				Segments: []*config.Segment{
					{Type: "text", Template: "R", Foreground: "green", Background: "black"},
				},
			},
			{
				Type:      config.Prompt,
				Alignment: config.Left,
				Newline:   true,
				Segments: []*config.Segment{
					{Type: "text", Template: "CD", Foreground: "yellow", Background: "black"},
				},
			},
		},
	}

	eng := newEngine(cfg, env)

	terminal.CaptureRuns = true

	primaryPrompt := eng.Primary()
	require.NotEmpty(t, primaryPrompt)

	rows := eng.CapturedRuns()
	require.Len(t, rows, 2)

	firstRowCells := 0
	for _, run := range rows[0] {
		firstRowCells += run.Cells
	}

	// AB (2) + the right block's own padding + R (1) fill the row out to the
	// full terminal width, exactly like the ANSI text canWriteRightBlock pads.
	assert.Equal(t, terminalWidth, firstRowCells)

	lastRun := rows[0][len(rows[0])-1]
	assert.Equal(t, "R", lastRun.Text)

	secondRowCells := 0
	for _, run := range rows[1] {
		secondRowCells += run.Cells
	}

	assert.Equal(t, 2, secondRowCells)
	assert.Equal(t, "CD", rows[1][len(rows[1])-1].Text)
}

// TestCapturedRunsRPromptBlock covers the other right-side mechanism
// (jandedobbeleer.omp.json's "rprompt" block type, distinct from an
// alignment: "right" block): rendered via writePrimaryRightPrompt's cursor
// save/restore rather than writeBlock's own padding, it must still land its
// runs on the same row the primary block already wrote.
func TestCapturedRunsRPromptBlock(t *testing.T) {
	saveCaptureRunsGlobal(t)

	// canWriteRightBlock demands 30 cells of breathing room for an actual
	// rprompt (vs. 5 for an alignment:"right" block — see its own doc
	// comment), so the width here must clear currentLineLength + rprompt
	// length + that margin, unlike TestCapturedRunsPrimaryLayout's 20.
	flags := &runtime.Flags{
		Shell:         shell.GENERIC,
		TerminalWidth: 40,
		IsPrimary:     true,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	cfg := &config.Config{
		Blocks: []*config.Block{
			{
				Type:      config.Prompt,
				Alignment: config.Left,
				Segments: []*config.Segment{
					{Type: "text", Template: "AB", Foreground: "red", Background: "blue"},
				},
			},
			{
				Type: config.RPrompt,
				Segments: []*config.Segment{
					{Type: "text", Template: "R", Foreground: "green", Background: "black"},
				},
			},
		},
	}

	eng := newEngine(cfg, env)

	terminal.CaptureRuns = true

	_ = eng.Primary()

	rows := eng.CapturedRuns()
	require.Len(t, rows, 1)

	lastRun := rows[0][len(rows[0])-1]
	assert.Equal(t, "R", lastRun.Text)

	// The cursor anchor must land between the primary prompt and the gap the
	// rprompt is padded out with, not at the row's end: writePrimaryRightPrompt
	// wraps the rprompt in SaveCursorPosition/RestoreCursorPosition precisely
	// so a real terminal leaves the cursor back here.
	row, run, ok := eng.CursorAnchor()
	require.True(t, ok)
	assert.Equal(t, 0, row)
	assert.Less(t, run, len(rows[0]), "anchor must precede the gap and the rprompt")
	assert.Equal(t, "AB", rows[row][run-1].Text)
	assert.Equal(t, strings.Repeat(" ", 37), rows[row][run].Text)
}

// TestCursorAnchorWithoutCapture pins CursorAnchor's contract when run
// capture is off: no anchor, mirroring CapturedRuns returning nothing.
func TestCursorAnchorWithoutCapture(t *testing.T) {
	saveCaptureRunsGlobal(t)

	terminal.CaptureRuns = false

	eng := &Engine{}
	eng.startRunCapture()
	eng.markCursorAnchor()

	_, _, ok := eng.CursorAnchor()
	assert.False(t, ok)
}

// TestCapturedRunsFinalSpace pins that final_space reaches the Run stream. The engine writes
// that space straight into its own builder rather than through terminal.Write, so nothing
// captures it on its own - and an SVG export then drew the cursor flush against the last
// segment while a real terminal left a space there. The cursor anchor must land after it too,
// since markCursorAnchor runs once the primary prompt (final space included) is written.
func TestCapturedRunsFinalSpace(t *testing.T) {
	saveCaptureRunsGlobal(t)

	flags := &runtime.Flags{
		Shell:         shell.GENERIC,
		TerminalWidth: 40,
		IsPrimary:     true,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	cfg := &config.Config{
		FinalSpace: true,
		Blocks: []*config.Block{
			{
				Type:      config.Prompt,
				Alignment: config.Left,
				Segments: []*config.Segment{
					{Type: "text", Template: "AB", Foreground: "red", Background: "blue"},
				},
			},
		},
	}

	eng := newEngine(cfg, env)

	terminal.CaptureRuns = true

	require.NotEmpty(t, eng.Primary())

	rows := eng.CapturedRuns()
	require.Len(t, rows, 1)

	cells := 0
	for _, run := range rows[0] {
		cells += run.Cells
	}

	assert.Equal(t, 3, cells, "AB (2) plus the final space (1)")

	last := rows[0][len(rows[0])-1]
	assert.Equal(t, " ", last.Text)
	assert.Equal(t, 1, last.Cells)

	row, run, ok := eng.CursorAnchor()
	require.True(t, ok)
	assert.Equal(t, 0, row)
	assert.Equal(t, len(rows[0]), run, "the cursor belongs after the final space, not before it")
}

// TestCapturedRunsWithoutFinalSpace is the control: with final_space off, nothing is appended.
func TestCapturedRunsWithoutFinalSpace(t *testing.T) {
	saveCaptureRunsGlobal(t)

	flags := &runtime.Flags{
		Shell:         shell.GENERIC,
		TerminalWidth: 40,
		IsPrimary:     true,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	cfg := &config.Config{
		Blocks: []*config.Block{
			{
				Type:      config.Prompt,
				Alignment: config.Left,
				Segments: []*config.Segment{
					{Type: "text", Template: "AB", Foreground: "red", Background: "blue"},
				},
			},
		},
	}

	eng := newEngine(cfg, env)

	terminal.CaptureRuns = true

	require.NotEmpty(t, eng.Primary())

	rows := eng.CapturedRuns()
	require.Len(t, rows, 1)

	cells := 0
	for _, run := range rows[0] {
		cells += run.Cells
	}

	assert.Equal(t, 2, cells, "AB only")
	assert.Equal(t, "AB", rows[0][len(rows[0])-1].Text)
}

// TestRPromptBreathingRoomOverride pins the render-only margin. canWriteRightBlock keeps
// DefaultRPromptBreathingRoom cells free between the prompt and an rprompt so a command being
// typed does not run into it - a margin nothing types into an exported image, which is why
// render.Config asks for a smaller one. The width here is chosen to sit in the gap: wide enough
// to clear 20 cells of room, too narrow to clear 30.
func TestRPromptBreathingRoomOverride(t *testing.T) {
	newEngineWithRoom := func(room int) *Engine {
		flags := &runtime.Flags{
			Shell:         shell.GENERIC,
			TerminalWidth: 30,
			IsPrimary:     true,
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		cfg := &config.Config{
			Blocks: []*config.Block{
				{
					Type:      config.Prompt,
					Alignment: config.Left,
					Segments: []*config.Segment{
						{Type: "text", Template: "AB", Foreground: "red", Background: "blue"},
					},
				},
				{
					Type: config.RPrompt,
					Segments: []*config.Segment{
						{Type: "text", Template: "RP", Foreground: "green", Background: "black"},
					},
				},
			},
		}

		eng := newEngine(cfg, env)
		eng.RPromptBreathingRoom = room

		return eng
	}

	saveCaptureRunsGlobal(t)
	terminal.CaptureRuns = true

	// 30 wide, a 2-cell prompt and a 2-cell rprompt leaves 26 cells: past 20, short of 30.
	assert.NotContains(t, newEngineWithRoom(0).Primary(), "RP", "the default 30 does not fit here")
	assert.Contains(t, newEngineWithRoom(20).Primary(), "RP", "20 does")
}
