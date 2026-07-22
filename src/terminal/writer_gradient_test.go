package terminal

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

const gradientReset = "\x1b[0m"

// colorise mirrors writer.go's colorisePrefix/coloriseSuffix wrapping, so tests can
// build expected output from color.GradientCells results instead of hard-coding codes.
func colorise(c color.Ansi) string {
	return "\x1b[" + c.String() + "m"
}

// saveGradientTestGlobals snapshots the package-level writer state this file's tests
// mutate and restores it via t.Cleanup, so a gradient CurrentColors/BackgroundColor
// value never leaks into a test in another file that runs afterward (see the golang
// skill's "save the original value and restore it" rule for global test state).
func saveGradientTestGlobals(t *testing.T) {
	t.Helper()

	origCurrentColors := CurrentColors
	origParentColors := ParentColors
	origColors := Colors
	origBackgroundColor := BackgroundColor
	origPlain := Plain

	t.Cleanup(func() {
		CurrentColors = origCurrentColors
		ParentColors = origParentColors
		Colors = origColors
		BackgroundColor = origBackgroundColor
		Plain = origPlain
	})
}

func TestWriteGradientRendering(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	fgGradient := color.Ansi("linear-gradient(#00FF00, #FF00FF)")
	invalidGradient := color.Ansi("linear-gradient(#FF0000)")

	resolver := &color.Defaults{}

	bgCells3 := color.GradientCells(bgGradient, 3, resolver, true, nil, nil)
	fgCells3 := color.GradientCells(fgGradient, 3, resolver, false, nil, nil)
	bgCells2 := color.GradientCells(bgGradient, 2, resolver, true, nil, nil)
	fgCells2 := color.GradientCells(fgGradient, 2, resolver, false, nil, nil)
	bgCells4 := color.GradientCells(bgGradient, 4, resolver, true, nil, nil)

	cases := []struct {
		Colors   *color.Set
		Case     string
		Input    string
		Expected string
	}{
		{
			Case:     "background gradient, plain ASCII",
			Input:    "abc",
			Colors:   &color.Set{Foreground: "white", Background: bgGradient},
			Expected: colorise("37") + colorise(bgCells3[0]) + "a" + colorise(bgCells3[1]) + "b" + colorise(bgCells3[2]) + "c" + gradientReset,
		},
		{
			Case:     "foreground gradient, plain ASCII",
			Input:    "abc",
			Colors:   &color.Set{Foreground: fgGradient, Background: "black"},
			Expected: colorise("40") + colorise(fgCells3[0]) + "a" + colorise(fgCells3[1]) + "b" + colorise(fgCells3[2]) + "c" + gradientReset,
		},
		{
			Case:  "both channels gradient",
			Input: "ab",
			Colors: &color.Set{
				Foreground: fgGradient,
				Background: bgGradient,
			},
			Expected: colorise(bgCells2[0]) + colorise(fgCells2[0]) + "a" + colorise(bgCells2[1]) + colorise(fgCells2[1]) + "b" + gradientReset,
		},
		{
			Case:     "wide rune advances index by its width",
			Input:    "a漢b",
			Colors:   &color.Set{Foreground: "white", Background: bgGradient},
			Expected: colorise("37") + colorise(bgCells4[0]) + "a" + colorise(bgCells4[1]) + "漢" + colorise(bgCells4[3]) + "b" + gradientReset,
		},
		{
			Case:     "single-stop gradient falls back to its only stop",
			Input:    "ab",
			Colors:   &color.Set{Foreground: "white", Background: invalidGradient},
			Expected: colorise("48;2;255;0;0") + colorise("37") + "ab" + gradientReset,
		},
		{
			Case:     "syntactically invalid gradient renders no background escape",
			Input:    "ab",
			Colors:   &color.Set{Foreground: "white", Background: "linear-gradient(#FF0000"},
			Expected: colorise("37") + "ab" + gradientReset,
		},
	}

	for _, tc := range cases {
		Init(shell.GENERIC)
		ParentColors = nil
		CurrentColors = tc.Colors
		Colors = &color.Defaults{}

		Write(tc.Colors.Background, tc.Colors.Foreground, tc.Input)

		got, _ := String()

		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

// TestWriteGradientOverrideBothChannels covers an inline override that sets both
// channels mid-text: the override wins for its duration and the gradient resumes
// stamping at the correct cell offset afterward (spec bullet: gradient + inline
// override mid-text).
func TestWriteGradientOverrideBothChannels(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	resolver := &color.Defaults{}

	// "ab" + "x" (override) + "cd" = 5 visible cells total.
	bgCells := color.GradientCells(bgGradient, 5, resolver, true, nil, nil)

	Init(shell.GENERIC)
	ParentColors = nil
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	Colors = &color.Defaults{}

	Write(colors.Background, colors.Foreground, "ab<red,blue>x</>cd")

	got, _ := String()

	expected := colorise("37") +
		colorise(bgCells[0]) + "a" +
		colorise(bgCells[1]) + "b" +
		colorise("44") + colorise("31") + "x" + // override: bg=blue(44), fg=red(31)
		colorise("37") + // fg restored explicitly on </>
		colorise(bgCells[3]) + "c" + // bg resumes stamping at cell offset 3, not 2
		colorise(bgCells[4]) + "d" +
		gradientReset

	assert.Equal(t, expected, got)
	// the overridden bg color (blue, "44") must never appear stamped again, and the
	// raw gradient string must never be printed literally.
	assert.NotContains(t, got, "linear-gradient")
}

// TestWriteGradientOverrideOneChannelOnly covers an inline override that only sets
// the foreground: the background gradient must keep stamping uninterrupted, and the
// foreground gradient must resume at the correct cell offset once the override ends
// (spec bullet: override of one channel only leaves the other channel active).
func TestWriteGradientOverrideOneChannelOnly(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	fgGradient := color.Ansi("linear-gradient(#00FF00, #FF00FF)")
	resolver := &color.Defaults{}

	// "ab" + "x" (fg override only) + "cd" = 5 visible cells total.
	bgCells := color.GradientCells(bgGradient, 5, resolver, true, nil, nil)
	fgCells := color.GradientCells(fgGradient, 5, resolver, false, nil, nil)

	Init(shell.GENERIC)
	ParentColors = nil
	colors := &color.Set{Foreground: fgGradient, Background: bgGradient}
	CurrentColors = colors
	Colors = &color.Defaults{}

	Write(colors.Background, colors.Foreground, "ab<red,>x</>cd")

	got, _ := String()

	expected := colorise(bgCells[0]) + colorise(fgCells[0]) + "a" +
		colorise(bgCells[1]) + colorise(fgCells[1]) + "b" +
		colorise("31") + // fg override to red; bg is untouched, no bg escape here
		colorise(bgCells[2]) + "x" + // bg keeps stamping uninterrupted through the override
		colorise(bgCells[3]) + colorise(fgCells[3]) + "c" + // fg resumes at cell offset 3, not 2
		colorise(bgCells[4]) + colorise(fgCells[4]) + "d" +
		gradientReset

	assert.Equal(t, expected, got)
	assert.NotContains(t, got, "linear-gradient")
}

func TestWriteGradientLengthMatchesNonGradient(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	input := "hello <b>world</b>, this is a segment"

	Init(shell.GENERIC)
	ParentColors = nil
	gradientColors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = gradientColors
	Colors = &color.Defaults{}

	Write(gradientColors.Background, gradientColors.Foreground, input)
	_, gradientLength := String()

	Init(shell.GENERIC)
	ParentColors = nil
	solidColors := &color.Set{Foreground: "white", Background: "blue"}
	CurrentColors = solidColors
	Colors = &color.Defaults{}

	Write(solidColors.Background, solidColors.Foreground, input)
	_, solidLength := String()

	assert.Equal(t, solidLength, gradientLength)
}

func TestWriteGradientPlainMode(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")

	Init(shell.GENERIC)
	Plain = true

	ParentColors = nil
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	Colors = &color.Defaults{}

	Write(colors.Background, colors.Foreground, "abc")

	got, length := String()

	assert.Equal(t, "abc", got)
	assert.Equal(t, 3, length)
	assert.NotContains(t, got, "\x1b")
}

// TestVisibleCells covers the exported cell-counting entry point the prompt engine uses to
// decide whether a gradient is wide enough to render per cell (Amendment 3). It does not touch
// any package-level state, so it needs no save/restore of the writer globals.
func TestVisibleCells(t *testing.T) {
	cases := []struct {
		Case     string
		Input    string
		Expected int
	}{
		{Case: "plain ASCII", Input: "abc", Expected: 3},
		{Case: "empty string", Input: "", Expected: 0},
		{Case: "style anchor is zero width", Input: "<b>abc</b>", Expected: 3},
		{Case: "color override anchor is zero width", Input: "<red>abc</>", Expected: 3},
		{Case: "wide rune counts its full width", Input: "a漢b", Expected: 4},
		{Case: "leading hyperlink anchor is stripped and zero width", Input: "<LINK>https://example.com<TEXT>abc</TEXT></LINK>", Expected: 3},
		{Case: "hyperlink with no text falls back to the 4-cell word 'link'", Input: "<LINK>https://example.com<TEXT></TEXT></LINK>", Expected: 4},
	}

	for _, tc := range cases {
		got := VisibleCells(tc.Input)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

// TestWriteGradientTransparentForeground covers the powerline start symbol: a
// transparent foreground with a gradient background must render byte-identical
// to the same text with a solid first-stop background (the transparentStart
// format takes a foreground code), and the transparent state must not leak into
// the next Write's gradient stamping.
func TestWriteGradientTransparentForeground(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")

	render := func(background color.Ansi) string {
		Init(shell.GENERIC)
		ParentColors = nil
		colors := &color.Set{Foreground: "transparent", Background: background}
		CurrentColors = colors
		Colors = &color.Defaults{}

		Write(background, "transparent", "")

		got, _ := String()
		return got
	}

	assert.Equal(t, render("#FF0000"), render(bgGradient), "transparent foreground must collapse to the first stop")

	// a follow-up Write must stamp its gradient: the previous Write's transparent
	// rendering may not suppress it via stale isTransparent state.
	resolver := &color.Defaults{}
	bgCells := color.GradientCells(bgGradient, 2, resolver, true, nil, nil)

	Init(shell.GENERIC)
	ParentColors = nil
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	Colors = &color.Defaults{}

	Write(bgGradient, "transparent", "")
	_, _ = String()

	Write(colors.Background, colors.Foreground, "ab")

	got, _ := String()
	expected := colorise("37") + colorise(bgCells[0]) + "a" + colorise(bgCells[1]) + "b" + gradientReset

	assert.Equal(t, expected, got, "gradient must stamp after a transparent Write")
}

// TestWriteGradientTransparentKeywordFirstStop covers the powerline separator in front
// of a chained gradient: a transparent foreground with a background of
// linear-gradient(parentBackground, ...) must resolve the keyword stop against the
// parent context instead of rendering a colorless glyph (regression: the arrow before
// a chained segment inherited stale terminal colors).
func TestWriteGradientTransparentKeywordFirstStop(t *testing.T) {
	saveGradientTestGlobals(t)

	chained := color.Ansi("linear-gradient(parentBackground, #EA76CB)")

	render := func(background color.Ansi) string {
		Init(shell.GENERIC)
		CurrentColors = &color.Set{Foreground: "#4c4f69", Background: background}
		ParentColors = []*color.Set{{Background: "#DD7878", Foreground: "#4c4f69"}}
		Colors = &color.Defaults{}

		Write(background, "transparent", "")

		got, _ := String()
		return got
	}

	got := render(chained)

	assert.Equal(t, render("#DD7878"), got, "the arrow must render the resolved parent color")
	assert.Contains(t, got, "38;2;221;120;120", "expected the parent background as a foreground code")
}

// TestWriteGradientKeywordAnchorOverride covers a template anchor resolving to a
// gradient that is NOT the segment's stamped base (e.g. the transient prompt's
// <background,transparent> caps): it must collapse to the first stop instead of
// rendering colorless or leaking the raw gradient string (regression).
func TestWriteGradientKeywordAnchorOverride(t *testing.T) {
	saveGradientTestGlobals(t)

	gradient := color.Ansi("linear-gradient(#DC8A78, #DD7878)")

	Init(shell.GENERIC)
	ParentColors = nil
	CurrentColors = &color.Set{Foreground: "#4c4f69", Background: gradient}
	Colors = &color.Defaults{}

	Write("transparent", "#4c4f69", "<background,transparent></>")

	got, _ := String()

	assert.Contains(t, got, "38;2;220;138;120", "cap must render the gradient's first stop as foreground")
	assert.NotContains(t, got, "linear-gradient")
}

// TestWriteGradientAnchorFollowsCell verifies that a background/foreground keyword
// anchor inside a gradient segment resolves to the gradient color at its cell
// position: a trailing <background,transparent> cap lands on the LAST cell's color,
// a leading one on the first (the transient prompt cap pattern).
func TestWriteGradientAnchorFollowsCell(t *testing.T) {
	saveGradientTestGlobals(t)

	gradient := color.Ansi("linear-gradient(#1E66F5, #40A02B)")
	resolver := &color.Defaults{}

	// leading cap + 4 body cells + trailing cap = 6 visible cells
	cells := color.GradientCells(gradient, 6, resolver, true, nil, nil)

	Init(shell.GENERIC)
	ParentColors = nil
	CurrentColors = &color.Set{Foreground: "#eff1f5", Background: gradient}
	Colors = &color.Defaults{}

	Write(gradient, "#eff1f5", "<background,transparent>X</><,background>abcd</><background,transparent>Y</>")

	got, _ := String()

	first := cells[0].ToChannel(false)
	last := cells[len(cells)-1].ToChannel(false)

	assert.Contains(t, got, colorise(first)+"X", "leading cap must use the first cell color")
	assert.Contains(t, got, colorise(last)+"Y", "trailing cap must use the last cell color")
	assert.NotContains(t, got, "linear-gradient")
}

// TestWriteGradientTransparentOverrideResumes pins the review fix for the stacked
// transparent override: after `</>` ends a mid-text transparent override, the
// gradient background must resume stamping for the remaining runes instead of
// leaving them on the terminal default background.
func TestWriteGradientTransparentOverrideResumes(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	resolver := &color.Defaults{}

	// "ab" + "X" + "cd" = 5 visible cells
	bgCells := color.GradientCells(bgGradient, 5, resolver, true, nil, nil)

	Init(shell.GENERIC)
	ParentColors = nil
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	Colors = &color.Defaults{}

	Write(colors.Background, colors.Foreground, "ab<transparent,#112233>X</>cd")

	got, _ := String()

	assert.Contains(t, got, colorise(bgCells[3])+"c", "background stamping must resume after the transparent override ends")
	assert.Contains(t, got, colorise(bgCells[4])+"d")
}

// TestWriteGradientInvisibleSpanExcluded pins the review fix for invisible spans:
// runes hidden by a <transparent,transparent> override are excluded from the
// pre-pass count, so the visible text still ends exactly on the last stop.
func TestWriteGradientInvisibleSpanExcluded(t *testing.T) {
	saveGradientTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	resolver := &color.Defaults{}

	// only "ab" is visible: the override hides everything after it (isInvisible
	// persists past `</>` by long-standing writer behavior).
	bgCells := color.GradientCells(bgGradient, 2, resolver, true, nil, nil)

	Init(shell.GENERIC)
	ParentColors = nil
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	Colors = &color.Defaults{}

	Write(colors.Background, colors.Foreground, "ab<transparent,transparent>hidden</>")

	got, _ := String()

	assert.Contains(t, got, colorise(bgCells[0])+"a")
	assert.Contains(t, got, colorise(bgCells[1])+"b", "the last visible cell must carry the last stop")
	assert.NotContains(t, got, "hidden")
}

// TestWriteGradientInvalidFallsBackToLastStop pins the review fix unifying the
// invalid-gradient fallback on the LAST stop, matching the engine's width collapse
// and every edge consumer.
func TestWriteGradientInvalidFallsBackToLastStop(t *testing.T) {
	saveGradientTestGlobals(t)

	Init(shell.GENERIC)
	ParentColors = nil
	colors := &color.Set{Foreground: "white", Background: "linear-gradient(red, blue)"}
	CurrentColors = colors
	Colors = &color.Defaults{}

	Write(colors.Background, colors.Foreground, "abcdefgh")

	got, _ := String()

	assert.Contains(t, got, colorise("44"), "body must render the last stop (blue) as a solid background")
	assert.NotContains(t, got, colorise("41"), "the first stop (red) must not appear")
}
