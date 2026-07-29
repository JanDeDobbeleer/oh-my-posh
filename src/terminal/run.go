package terminal

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
)

// CaptureRuns turns on the parallel Run stream Write/String build alongside the ANSI
// bytes they already emit, for a later encoder (an SVG renderer, eventually) to draw the
// prompt without parsing escape sequences. Off by default, mirroring Plain/Interactive:
// with it off, every capture site below is a single `if CaptureRuns` check and nothing
// else runs, so the ANSI hot path's allocation profile is unchanged (see the exact
// allocs/B-per-op gates in bench_test.go). With it on, capture is allowed to allocate.
var CaptureRuns bool

// RunMode discriminates the three ways writeSegmentColors (and writeAnchorOverride's
// matching branches) can render a segment or override. All three print bytes that read
// back as the same ANSI token "transparent" once decoded, so a Run needs this field to
// tell them apart; see writeSegmentColors' switch for the three branches it names.
type RunMode int

const (
	// RunNormal is a run rendered with its own foreground/background: neither of the
	// two transparent branches below applied.
	RunNormal RunMode = iota

	// RunBackgroundPainted is the branch taken when the foreground is transparent and a
	// terminal background color is known: the cell renders filled with the segment's own
	// background color, and the glyph itself renders in the terminal's default background
	// color, cutting its shape out of that fill — the SGR sequence sets the segment's
	// background as a foreground code, resets the background to default, then SGR 7
	// (reverse video) swaps the two, so what actually paints is the inverse of what the
	// codes name.
	RunBackgroundPainted

	// RunReverseVideo is the branch taken when the foreground is transparent and no
	// terminal background is known: reverse video (SGR 7) still swaps foreground/
	// background as described above, so the cell renders filled with the segment's own
	// background color; the glyph itself renders in whatever the real terminal's default
	// background turns out to be, which is unknown here (that's the "no terminal
	// background is known" half of this branch's condition).
	RunReverseVideo
)

// runAttributeSlots mirrors len(knownStyles): one nesting-depth counter per style
// anchor. knownStyles is declared as an array (`[...]*style{...}`), not a slice, so
// its length is a compile-time constant and can be used directly here instead of a
// literal pinned to it by a test.
const runAttributeSlots = len(knownStyles)

// Run is one contiguous span of text rendered under a single, unchanging style: the
// exact SGR pair and its pre-ToAnsi source form (a #RRGGBB hex, a colour name, accent,
// transparent, or a gradient definition — see color.History.Add), a render-mode
// discriminator, per-style-anchor nesting depth, and — for a cell stamped from the
// segment's own gradient — the true RGB behind the printed escape.
//
// Cells is the run's rendered width, a length delta matching what write() counts toward
// String()'s returned length; it can be zero (e.g. a hyperlink's URL text, which Text
// includes but write() never counts toward length) even when Text is non-empty.
type Run struct {
	BackgroundRGB    *color.RGB
	ForegroundRGB    *color.RGB
	Text             string
	Background       color.Ansi
	Foreground       color.Ansi
	BackgroundSource color.Ansi
	ForegroundSource color.Ansi
	Mode             RunMode
	Cells            int
	Attributes       [runAttributeSlots]uint8
}

// Runs returns the Run stream captured since the last String() call. Empty when
// CaptureRuns is false.
func Runs() []Run {
	return runsState.Runs
}

// runState bundles the run-capture bookkeeping threaded alongside colorState through the
// same style-anchor and color-change events (see colorState's own doc comment), only
// touched when CaptureRuns is true. text accumulates the pending run's content exactly as
// write() sends it to builder — see write()'s mirrored calls. depth is the running
// nesting counter per knownStyles slot, mutated at the three style-anchor emission sites
// (Write's leading-anchor prologue, writeAnchorOverride's Start/End cases) and reset
// alongside the writer's own end-of-segment SGR reset and in String()'s defer; attributes
// is depth's last snapshot, copied into a Run at flushRun. background/foreground/
// backgroundSource/foregroundSource/mode/backgroundRGB/foregroundRGB are the pending
// run's style, refreshed by syncPendingStyle after every event that can change it.
type runState struct {
	backgroundRGB    *color.RGB
	foregroundRGB    *color.RGB
	background       color.Ansi
	foreground       color.Ansi
	backgroundSource color.Ansi
	foregroundSource color.Ansi
	text             strings.Builder
	Runs             []Run
	mode             RunMode
	cellsAtFlush     int
	attributes       [runAttributeSlots]uint8
	depth            [runAttributeSlots]uint8
}

var runsState runState

// flushRun cuts the text accumulated since the last flush into a Run using the pending
// style snapshot (background/foreground/source/mode/attributes/RGB), then resets the
// text accumulator. A call with nothing accumulated — no text, no cell delta — is a
// no-op: several call sites flush unconditionally on entry (writeSegmentColors,
// writeAnchorOverride, endColorOverride all flush before touching any state), and most of
// those calls land here with nothing pending, since the previous flush already cut it.
func flushRun() {
	cells := length - runsState.cellsAtFlush
	text := runsState.text.String()

	if len(text) != 0 || cells != 0 {
		runsState.Runs = append(runsState.Runs, Run{
			Text:             text,
			Background:       runsState.background,
			Foreground:       runsState.foreground,
			BackgroundSource: runsState.backgroundSource,
			ForegroundSource: runsState.foregroundSource,
			BackgroundRGB:    runsState.backgroundRGB,
			ForegroundRGB:    runsState.foregroundRGB,
			Attributes:       runsState.attributes,
			Mode:             runsState.mode,
			Cells:            cells,
		})
	}

	runsState.text.Reset()
	runsState.cellsAtFlush = length
}

// syncPendingStyle snapshots cs's current active style — background/foreground/source,
// the render-mode discriminator, and the attribute nesting depth — as what governs
// whatever text comes next, and clears the gradient RGB (stampGradient sets it again,
// per cell, only for a channel it actually stamps). Called after every event that can
// change the active style: writeSegmentColors, writeAnchorOverride (including its
// style-anchor early returns, where only the attributes changed), and endColorOverride.
func syncPendingStyle(cs *colorState) {
	fg := activeForeground(cs)

	runsState.background = activeBackground(cs)
	runsState.foreground = fg
	runsState.backgroundSource = activeBackgroundSource(cs)
	runsState.foregroundSource = activeForegroundSource(cs)
	runsState.mode = deriveRunMode(fg)
	runsState.attributes = runsState.depth
	runsState.backgroundRGB = nil
	runsState.foregroundRGB = nil
}

// activeBackgroundSource/activeForegroundSource are activeBackground/activeForeground's
// source-form counterparts: the top of the override history's source, or the segment
// base's source when no override is active.
func activeBackgroundSource(cs *colorState) color.Ansi {
	if bg := cs.currentColor.Background(); !bg.IsEmpty() {
		return cs.currentColor.BackgroundSource()
	}

	return cs.backgroundColorSource
}

func activeForegroundSource(cs *colorState) color.Ansi {
	if fg := cs.currentColor.Foreground(); !fg.IsEmpty() {
		return cs.currentColor.ForegroundSource()
	}

	return cs.foregroundColorSource
}

// deriveRunMode mirrors writeSegmentColors' own switch (and writeAnchorOverride's
// matching branches) on the now-current active foreground (fg, resolved once by the
// caller — see syncPendingStyle, which needs the same value for runsState.foreground)
// and the package-level terminal background, so a single derivation covers all three
// call sites instead of each branch setting it inline (which would have to be threaded
// through every early return).
func deriveRunMode(fg color.Ansi) RunMode {
	switch {
	case fg.IsTransparent() && len(BackgroundColor) != 0:
		return RunBackgroundPainted
	case fg.IsTransparent():
		return RunReverseVideo
	default:
		return RunNormal
	}
}
