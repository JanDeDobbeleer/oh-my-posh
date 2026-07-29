package terminal

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveRunTestGlobals snapshots the package-level writer state this file's tests mutate and
// restores it via t.Cleanup, mirroring writer_gradient_test.go's saveGradientTestGlobals
// (not reused directly: that helper lives in a different file and this one additionally
// needs CaptureRuns, which defaults to false and must never leak into another file's tests).
func saveRunTestGlobals(t *testing.T) {
	t.Helper()

	origCurrentColors := CurrentColors
	origParentColors := ParentColors
	origColors := Colors
	origBackgroundColor := BackgroundColor
	origPlain := Plain
	origCaptureRuns := CaptureRuns

	t.Cleanup(func() {
		CurrentColors = origCurrentColors
		ParentColors = origParentColors
		Colors = origColors
		BackgroundColor = origBackgroundColor
		Plain = origPlain
		CaptureRuns = origCaptureRuns
	})
}

// TestCaptureRunsOffProducesNoRuns pins the off-by-default contract: with CaptureRuns
// false, nothing must ever land in the run stream, regardless of what the segment does.
func TestCaptureRunsOffProducesNoRuns(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.GENERIC)
	Colors = &color.Defaults{}
	CaptureRuns = false
	CurrentColors = &color.Set{Foreground: "white", Background: "blue"}
	ParentColors = nil

	Write("blue", "white", "hello <b>world</b>")

	runs := Runs()
	_, _ = String()

	assert.Empty(t, runs)
}

// TestRunNestedStyleAnchorsUseNestingDepth pins the counter-not-boolean requirement:
// <b><b>x</b></b>y emits Start twice and End twice, so a boolean would clear bold at the
// FIRST </b> (while x is still open at depth 2) and diverge from the actual bytes. The
// depth counter must instead still read 2 for x and drop to 0 only once both </b> have
// fired, by y.
func TestRunNestedStyleAnchorsUseNestingDepth(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.GENERIC)
	Colors = &color.Defaults{}
	CaptureRuns = true
	CurrentColors = &color.Set{Foreground: "white", Background: "blue"}
	ParentColors = nil

	Write("blue", "white", "<b><b>x</b></b>y")

	runs := Runs()
	_, length := String()

	require.Len(t, runs, 2)

	assert.Equal(t, "x", runs[0].Text)
	assert.EqualValues(t, 2, runs[0].Attributes[0], "the second <b> nests bold to depth 2 before x")

	assert.Equal(t, "y", runs[1].Text)
	assert.EqualValues(t, 0, runs[1].Attributes[0], "both </b> have fired by the time y is written")

	assert.Equal(t, 2, length)
}

// TestRunAttributesIgnoreResetOverride pins that `</>` (resetStyle) must never clear
// attributes: it routes to endColorOverride, which only ever touches colours. A `<b>` left
// open across a `</>` must still read as bold afterward.
func TestRunAttributesIgnoreResetOverride(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.GENERIC)
	Colors = &color.Defaults{}
	CaptureRuns = true
	CurrentColors = &color.Set{Foreground: "white", Background: "blue"}
	ParentColors = nil

	Write("blue", "white", "<b>a<red>b</>c</b>")

	runs := Runs()
	_, _ = String()

	for _, r := range runs {
		assert.EqualValues(t, 1, r.Attributes[0], "run %q: bold must survive the </> that ends the inline color override", r.Text)
	}
}

// TestRunGradientCapturesPerCellRGB covers the gradient capture path end to end: every
// visible cell of a gradient channel is its own Run (stampGradient re-stamps every cell,
// never diffing against the previous one), and its RGB must be the true interpolated
// color.GradientCellsRGB value — taken where colorful.Color is already in hand — not
// something inverted back out of the printed escape.
func TestRunGradientCapturesPerCellRGB(t *testing.T) {
	saveRunTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	resolver := &color.Defaults{}

	Init(shell.GENERIC)
	Colors = resolver
	CaptureRuns = true
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	ParentColors = nil

	Write(colors.Background, colors.Foreground, "abc")

	runs := Runs()
	_, length := String()

	expectedRGB := color.GradientCellsRGB(bgGradient, 3, resolver, nil, nil)
	require.Len(t, runs, 3)
	require.Len(t, expectedRGB, 3)

	for i, r := range runs {
		assert.Equal(t, string(rune('a'+i)), r.Text)
		assert.Equal(t, 1, r.Cells)

		require.NotNil(t, r.BackgroundRGB, "cell %d", i)
		assert.Equal(t, expectedRGB[i], *r.BackgroundRGB, "cell %d RGB must be the true interpolated color", i)
		assert.Nil(t, r.ForegroundRGB, "foreground is solid white, not this segment's gradient")
	}

	assert.Equal(t, 3, length)
}

// TestRunBothChannelsGradientCapturePerCellRGB covers stampGradient's fgActive branch
// (TestRunGradientCapturesPerCellRGB only exercises bgActive): when both channels are the
// segment's own gradient, each cell's Run must carry BOTH a background and a foreground
// RGB, independently interpolated.
func TestRunBothChannelsGradientCapturePerCellRGB(t *testing.T) {
	saveRunTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")
	fgGradient := color.Ansi("linear-gradient(#00FF00, #FF00FF)")
	resolver := &color.Defaults{}

	Init(shell.GENERIC)
	Colors = resolver
	CaptureRuns = true
	colors := &color.Set{Foreground: fgGradient, Background: bgGradient}
	CurrentColors = colors
	ParentColors = nil

	Write(colors.Background, colors.Foreground, "ab")

	runs := Runs()
	_, _ = String()

	expectedBg := color.GradientCellsRGB(bgGradient, 2, resolver, nil, nil)
	expectedFg := color.GradientCellsRGB(fgGradient, 2, resolver, nil, nil)
	require.Len(t, runs, 2)

	for i, r := range runs {
		require.NotNil(t, r.BackgroundRGB, "cell %d", i)
		require.NotNil(t, r.ForegroundRGB, "cell %d", i)
		assert.Equal(t, expectedBg[i], *r.BackgroundRGB, "cell %d", i)
		assert.Equal(t, expectedFg[i], *r.ForegroundRGB, "cell %d", i)
	}
}

// TestRunGradientOverrideChannelIsNotRGB covers the inline-override case: while an anchor
// overrides the gradient channel, the active color is a plain solid, so the Run for that
// span must carry no RGB at all (BackgroundRGB nil), distinct from the gradient cells
// before and after it.
func TestRunGradientOverrideChannelIsNotRGB(t *testing.T) {
	saveRunTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")

	Init(shell.GENERIC)
	Colors = &color.Defaults{}
	CaptureRuns = true
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	ParentColors = nil

	// both channels overridden: bg no longer matches the gradient, so it stops stamping
	// for x's duration (a bg-only-empty override like <blue,> would instead inherit the
	// current background, which is still the gradient, and keep stamping through it —
	// see TestWriteGradientOverrideOneChannelOnly).
	Write(colors.Background, colors.Foreground, "a<red,blue>x</>b")

	runs := Runs()
	_, _ = String()

	var overrideRun *Run
	for i := range runs {
		if runs[i].Text == "x" {
			overrideRun = &runs[i]
		}
	}

	require.NotNil(t, overrideRun)
	assert.Nil(t, overrideRun.BackgroundRGB, "an overridden channel is a solid color, not this segment's gradient cell")
	assert.Equal(t, RunNormal, overrideRun.Mode)
}

// TestRunModeDiscriminatesTransparentRenderings covers all three RunMode values: the
// three renderings writeSegmentColors can produce all print bytes that read back as the
// same ANSI token "transparent", so Mode is the only way a consumer of the Run stream can
// tell them apart.
func TestRunModeDiscriminatesTransparentRenderings(t *testing.T) {
	saveRunTestGlobals(t)

	cases := []struct {
		Case               string
		Foreground         color.Ansi
		TerminalBackground color.Ansi
		Expected           RunMode
	}{
		{Case: "normal", Foreground: "white", Expected: RunNormal},
		{Case: "background painted", Foreground: color.Transparent, TerminalBackground: "#212F3C", Expected: RunBackgroundPainted},
		{Case: "reverse video", Foreground: color.Transparent, Expected: RunReverseVideo},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			Init(shell.GENERIC)
			Colors = &color.Defaults{}
			CaptureRuns = true
			BackgroundColor = tc.TerminalBackground
			colors := &color.Set{Foreground: tc.Foreground, Background: "#FF5733"}
			CurrentColors = colors
			ParentColors = nil

			Write(colors.Background, colors.Foreground, "test")

			runs := Runs()
			_, _ = String()

			require.Len(t, runs, 1)
			assert.Equal(t, tc.Expected, runs[0].Mode, tc.Case)
		})
	}
}

// TestRunHyperlinkFlush covers a hyperlink transition inside a captured segment: no color
// change happens across <LINK>/<TEXT>/</LINK>, so the anchor text and the surrounding text
// stay in ONE run — and the URL runes appear in neither Text nor Cells, because a terminal
// consumes them as part of the OSC 8 escape rather than painting them.
func TestRunHyperlinkFlush(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.PWSH)
	Colors = &color.Defaults{}
	CaptureRuns = true
	CurrentColors = &color.Set{Foreground: "black", Background: "white"}
	ParentColors = nil

	Write("white", "black", "a<LINK>http://x<TEXT>bc</TEXT></LINK>d")

	runs := Runs()
	_, length := String()

	require.Len(t, runs, 1, "no color change occurs across the hyperlink boundary")
	assert.Equal(t, "abcd", runs[0].Text, "Text is what gets painted: the URL is not")
	assert.Equal(t, 4, runs[0].Cells, "the URL must not inflate Cells: only a, b, c, d are counted")
	assert.Equal(t, 4, length)
}

// TestRunHyperlinkNoTextFallback covers writeBody's "link" no-text fallback under
// capture: per the design, this is an emit of 4 cells, not a flush, so it stays part of
// whatever run is already accumulating.
func TestRunHyperlinkNoTextFallback(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.PWSH)
	Colors = &color.Defaults{}
	CaptureRuns = true
	CurrentColors = &color.Set{Foreground: "black", Background: "white"}
	ParentColors = nil

	Write("white", "black", "ab<LINK>http://x<TEXT></TEXT></LINK>cd")

	runs := Runs()
	_, length := String()

	require.Len(t, runs, 1)
	assert.Equal(t, "ablinkcd", runs[0].Text, "the 4-cell \"link\" fallback is painted, the URL is not")
	assert.Equal(t, 8, runs[0].Cells, "ab (2) + the 4-cell 'link' fallback + cd (2) = 8")
	assert.Equal(t, 8, length)
}

// TestRunGradientHyperlinkNoTextFallback is TestRunHyperlinkNoTextFallback's gradient
// counterpart: writeBodyGradient's fallback calls stampGradient once for all 4 "link"
// cells (not once per character), so it must land in a single run carrying one stamped
// gradient cell, not four.
func TestRunGradientHyperlinkNoTextFallback(t *testing.T) {
	saveRunTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")

	Init(shell.GENERIC)
	Colors = &color.Defaults{}
	CaptureRuns = true
	colors := &color.Set{Foreground: "white", Background: bgGradient}
	CurrentColors = colors
	ParentColors = nil

	Write(colors.Background, colors.Foreground, "ab<LINK>http://x<TEXT></TEXT></LINK>cd")

	runs := Runs()
	_, length := String()

	var gotText strings.Builder
	totalCells := 0
	var linkRun *Run

	for i := range runs {
		gotText.WriteString(runs[i].Text)
		totalCells += runs[i].Cells

		if runs[i].Text == "link" {
			linkRun = &runs[i]
		}
	}

	assert.Equal(t, "ablinkcd", gotText.String(), "the 4-cell \"link\" fallback is painted, the URL is not")
	assert.Equal(t, 8, totalCells)
	assert.Equal(t, 8, length)

	require.NotNil(t, linkRun, "the 4-cell fallback must appear as its own run's text")
	assert.Equal(t, 4, linkRun.Cells)
	assert.NotNil(t, linkRun.BackgroundRGB, "stampGradient runs once for the fallback, stamping one cell for all 4 characters")
}

// TestRunInvisibleSpanIsSticky pins the isInvisible-transition case: a
// <transparent,transparent> override hides everything to the end of the SEGMENT (Write
// call), since endColorOverride never clears isInvisible — so text after the matching
// </> must be hidden too, not just the text directly inside the override.
func TestRunInvisibleSpanIsSticky(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.GENERIC)
	Colors = &color.Defaults{}
	CaptureRuns = true
	CurrentColors = &color.Set{Foreground: "black", Background: "white"}
	ParentColors = nil

	Write("white", "black", "ab<transparent,transparent>hidden</>cd")

	runs := Runs()
	_, length := String()

	var gotText strings.Builder
	totalCells := 0

	for _, r := range runs {
		gotText.WriteString(r.Text)
		totalCells += r.Cells
	}

	assert.Equal(t, "ab", gotText.String(), "hidden text, and everything after it in this segment, must never reach the run stream")
	assert.Equal(t, 2, totalCells)
	assert.Equal(t, 2, length)
}

// TestRunPlainMode pins Plain-mode capture: the OSC 8 wrappers and the hyperlink URL text
// must be excluded exactly like the ANSI builder excludes them, so the captured run's
// Text is exactly the visible link text, matching String()'s own Plain output byte for
// byte.
func TestRunPlainMode(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.PWSH)
	Plain = true
	Colors = &color.Defaults{}
	CaptureRuns = true
	CurrentColors = &color.Set{Foreground: "black", Background: "white"}
	ParentColors = nil

	Write("white", "black", "<LINK>http://www.google.be<TEXT>google</TEXT></LINK>")

	runs := Runs()
	got, length := String()

	require.Len(t, runs, 1)
	assert.Equal(t, "google", got)
	assert.Equal(t, "google", runs[0].Text)
	assert.Equal(t, 6, length)
	assert.Equal(t, 6, runs[0].Cells)
	assert.NotContains(t, runs[0].Text, "google.be")
}

// TestRunShellEscapingMirrorsBuilder covers the shell-rewrite paths write() applies
// (zsh doubles '%', bash doubles '\\'): Run.Text must carry the SAME rewritten form that
// reached the ANSI builder, not the original source rune, since neither slicing the
// source text nor slicing the builder can reproduce it otherwise.
func TestRunShellEscapingMirrorsBuilder(t *testing.T) {
	saveRunTestGlobals(t)

	cases := []struct {
		Case      string
		ShellName string
		Input     string
		Expected  string
	}{
		{Case: "zsh percent", ShellName: shell.ZSH, Input: "100% done", Expected: "100%% done"},
		{Case: "bash backslash", ShellName: shell.BASH, Input: `C:\path`, Expected: `C:\\path`},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			Init(tc.ShellName)
			Colors = &color.Defaults{}
			CaptureRuns = true
			CurrentColors = &color.Set{Foreground: "black", Background: "white"}
			ParentColors = nil

			Write("white", "black", tc.Input)

			runs := Runs()
			_, _ = String()

			require.Len(t, runs, 1)
			assert.Equal(t, tc.Expected, runs[0].Text)
		})
	}
}

// TestRunResetBetweenStringCalls pins String()'s defer resetting the run slice (and the
// pending style/depth) alongside length/builder: without it, runs would accumulate across
// every block, rprompt and transient render in the process.
func TestRunResetBetweenStringCalls(t *testing.T) {
	saveRunTestGlobals(t)

	Init(shell.GENERIC)
	Colors = &color.Defaults{}
	CaptureRuns = true
	CurrentColors = &color.Set{Foreground: "white", Background: "blue"}
	ParentColors = nil

	Write("blue", "white", "first")
	first := Runs()
	_, _ = String()
	require.Len(t, first, 1)
	assert.Equal(t, "first", first[0].Text)

	Write("blue", "white", "second")
	second := Runs()
	_, _ = String()

	require.Len(t, second, 1, "the previous String() call must have cleared the run stream")
	assert.Equal(t, "second", second[0].Text)
}

// TestRunCorrectnessBarPerBlock is the plan's correctness bar, scoped per String() call
// (one block), asserted under both shell.GENERIC (nil EscapeSequences, what the golden
// harness exercises) and shell.ZSH (populated EscapeSequences, which the goldens never
// exercise). Segments deliberately avoid '%' and '\\' — shell-escaping is covered
// separately by TestRunShellEscapingMirrorsBuilder — so a Plain-mode render of the exact
// same segment sequence is a safe stand-in for "escape-stripped block output": with no
// escapable runes, Plain's output is byte-identical to the styled render with every
// escape sequence removed, regardless of shell.
func TestRunCorrectnessBarPerBlock(t *testing.T) {
	saveRunTestGlobals(t)

	bgGradient := color.Ansi("linear-gradient(#FF0000, #0000FF)")

	type segment struct {
		Background color.Ansi
		Foreground color.Ansi
		Text       string
	}

	segments := []segment{
		{Background: "blue", Foreground: "white", Text: "hello <b>world</b>"},
		{Background: bgGradient, Foreground: "white", Text: "ab<red>cd</>ef"},
		{Background: "green", Foreground: "black", Text: "<blue,yellow>x</>y"},
	}

	render := func(shellName string, plain bool) (text string, length int, runs []Run) {
		Init(shellName)
		Colors = &color.Defaults{}
		CaptureRuns = !plain
		Plain = plain
		ParentColors = nil

		for _, seg := range segments {
			CurrentColors = &color.Set{Background: seg.Background, Foreground: seg.Foreground}
			Write(seg.Background, seg.Foreground, seg.Text)
		}

		runs = Runs()
		text, length = String()
		return text, length, runs
	}

	for _, shellName := range []string{shell.GENERIC, shell.ZSH} {
		t.Run(shellName, func(t *testing.T) {
			plainText, _, _ := render(shellName, true)
			_, length, runs := render(shellName, false)

			var gotText strings.Builder
			totalCells := 0
			for _, r := range runs {
				gotText.WriteString(r.Text)
				totalCells += r.Cells
			}

			assert.Equal(t, plainText, gotText.String(), "concatenated run text must equal the escape-stripped block output")
			assert.Equal(t, length, totalCells, "summed run cells must equal the length String() returns")
		})
	}
}
