package svg

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testOptions fixes every metric so pixel math in assertions is exact and
// readable: a 10px cell, a 20px row, and a fixed Columns wide enough that no
// existing test's row gets wrapped by fitRow (see fit.go) — tests that
// specifically exercise wrapping set their own, narrower Columns. The window
// padding these assertions are computed against isn't set here at all: it
// comes from windowGeometry, scaled from FontSize (see newWindowGeometry) —
// contentOriginX recomputes that real geometry rather than assuming it away.
//
// FillAscent/FillDescent are set to 15/5 rather than left to default so the
// fill box works out to exactly the row box (baseline sits at 0.75*20 = 15
// below the row top, so the rect spans rowTop..rowTop+20). Every position
// assertion here predates the fill box being a metric of its own and is
// written against the row, so pinning them this way keeps those assertions
// meaningful; TestEncodeFillBoxTracksGlyphInk covers the case where the two
// genuinely differ.
func testOptions() Options {
	return Options{
		FontSize:    10,
		CellWidth:   10,
		LineHeight:  20,
		FillAscent:  15,
		FillDescent: 5,
		Columns:     40,
	}
}

// contentOriginX computes the same content grid top-left X coordinate
// Encode itself derives (see newWindowGeometry/newCanvasSize), so position
// assertions stay correct if that geometry ever changes instead of
// hardcoding numbers that would silently drift out of sync. Every caller
// only ever needs the X coordinate (row-position assertions derive Y from
// LineHeight/rowIndex directly), so this returns just that rather than a
// (x, y) pair whose y half no caller uses.
//
//nolint:gocritic
func contentOriginX(opts Options) float64 {
	opts = opts.withDefaults()
	geo := newWindowGeometry(&opts)
	size := newCanvasSize(&geo, opts.Columns, opts.CellWidth, 1, opts.LineHeight)

	return size.contentX
}

// contentBaselineY is contentOriginX's counterpart for the other axis: the
// baseline Encode writes row n's glyphs on, derived from the same geometry
// rather than hardcoded. Row baselines used to be written into assertions as
// literals, which meant a change to the window chrome failed them all with a
// number rather than a reason.
//
//nolint:gocritic
func contentBaselineY(opts Options, row int) float64 {
	opts = opts.withDefaults()
	geo := newWindowGeometry(&opts)
	size := newCanvasSize(&geo, opts.Columns, opts.CellWidth, 1, opts.LineHeight)

	return size.contentY + float64(row)*opts.LineHeight + baselineRatio*opts.LineHeight
}

// decodeXML asserts doc is well-formed XML — the bar the task sets for
// "renders correctly when opened in a browser" — draining every token so a
// mismatched tag or a bad escape surfaces as a test failure rather than
// something only a browser would notice.
func decodeXML(t *testing.T, doc string) {
	t.Helper()

	decoder := xml.NewDecoder(strings.NewReader(doc))

	for {
		_, err := decoder.Token()
		if err == nil {
			continue
		}

		require.ErrorIs(t, err, io.EOF)

		return
	}
}

// TestEncodeSkipsZeroCellRuns pins the hyperlink case explicitly called out
// in the parent stage plan: Run.Cells can be 0 while Run.Text is non-empty
// (a hyperlink's URL text — see terminal.Run's doc comment), and such a run
// must contribute no glyphs and no column advance, so the run after it lands
// immediately next to the one before it, not shifted by the hidden text's
// own width.
func TestEncodeSkipsZeroCellRuns(t *testing.T) {
	rows := [][]terminal.Run{
		{
			contentRun("ab"),
			{Text: "https://example.com/hidden", Cells: 0, ForegroundSource: "red", BackgroundSource: "blue", Mode: terminal.RunNormal},
			contentRun("cd"),
		},
	}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.NotContains(t, doc, "hidden")
	assert.Contains(t, doc, ">ab<")
	assert.Contains(t, doc, ">cd<")

	// "cd" must start immediately after "ab" (2 cells further, at the
	// content origin + 2*cellWidth), not pushed out further by the
	// zero-cell hyperlink run's own text in between.
	opts := testOptions()
	contentX := contentOriginX(opts)
	wantX := contentX + 2*opts.CellWidth
	assert.Contains(t, doc, `x="`+formatFloat(wantX)+`"`)

	// One <rect> for each of the two visible runs, plus the window's own
	// border rect (see writeWindowChrome; the header/content fills are now
	// <path> elements, not <rect>s); one <text> for each of the two visible
	// runs, plus the cursor, the watermark decorate always appends (see
	// decorate), and the "−"/"▢"/"×" window-control glyphs the chrome draws
	// (no title/tab-control text since those were dropped — see
	// writeWindowChrome's doc comment).
	assert.Equal(t, 3, strings.Count(doc, "<rect"))
	assert.Equal(t, 7, strings.Count(doc, "<text"))
}

// TestEncodeRunModes covers all three RunMode values (see terminal.RunMode's
// doc comment): the ordinary case, and the two distinct transparency
// renderings. Both transparent modes must paint a rect filled with the
// segment's own background (the "cutout" a real terminal renders via SGR
// 7 — see terminal.RunMode's doc comment and paintRun's), with the glyph
// itself painted in Options.CanvasBackground — the real terminal background
// where known, its own fallback otherwise (see Options.CanvasBackground's
// doc comment) — regardless of which of the two transparent RunModes is
// active, since withDefaults folds a known terminal background into
// CanvasBackground once rather than each RunMode reading its own field.
//
// The third transparent case pins the one behavior change that unification
// brought: a RunBackgroundPainted run with no known terminal background
// used to read opts.TerminalBackground directly for its glyph color, a nil
// pointer that emitted no fill attribute at all (an SVG text element with no
// fill renders black, which is very likely wrong for the run's own
// background). It now falls back to CanvasBackground exactly like
// RunReverseVideo always did — the same "punch a hole to the window behind
// it" cutout this package already relies on CanvasBackground for
// everywhere else, so this is a fix, not just a side effect of sharing code.
func TestEncodeRunModes(t *testing.T) {
	red := hexString(color.RGB{R: 222, G: 56, B: 43})
	blue := hexString(color.RGB{R: 0, G: 111, B: 184})
	termBg := color.RGB{R: 5, G: 6, B: 7}
	canvasBg := hexString(defaultCanvasBackground)

	cases := []struct {
		Case         string
		WantTextFill string
		WantRectFill string
		Opts         Options
		Mode         terminal.RunMode
		WantRect     bool
	}{
		{
			Case: "normal paints its own foreground and background", Mode: terminal.RunNormal,
			WantRect: true, WantTextFill: red, WantRectFill: blue, Opts: testOptions(),
		},
		{
			Case:     "background-painted fills the segment's background, cutting the glyph out in the known terminal background",
			Mode:     terminal.RunBackgroundPainted,
			WantRect: true, WantTextFill: hexString(termBg), WantRectFill: blue,
			Opts: func() Options { o := testOptions(); o.TerminalBackground = &termBg; return o }(),
		},
		{
			Case:     "background-painted with no known terminal background falls back to the canvas background, not a missing fill",
			Mode:     terminal.RunBackgroundPainted,
			WantRect: true, WantTextFill: canvasBg, WantRectFill: blue, Opts: testOptions(),
		},
		{
			Case:     "reverse video fills the segment's background, cutting the glyph out in the canvas background fallback",
			Mode:     terminal.RunReverseVideo,
			WantRect: true, WantTextFill: canvasBg, WantRectFill: blue, Opts: testOptions(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			rows := [][]terminal.Run{
				{{Text: "hi", Cells: 2, ForegroundSource: "red", BackgroundSource: "blue", Mode: tc.Mode}},
			}

			doc := Encode(rows, tc.Opts)
			decodeXML(t, doc)

			assert.Equal(t, tc.WantRect, strings.Contains(doc, "<rect"))

			if tc.WantRect {
				assert.Contains(t, doc, `fill="`+tc.WantRectFill+`"`)
			}

			assert.Contains(t, doc, `fill="`+tc.WantTextFill+`"`)
		})
	}
}

// TestEncodeCarriesForwardOnEmptySource pins the "empty" color form: a
// failed color.Defaults.ToAnsi resolution emits nothing, so the glyph must
// inherit whatever color was already active on that channel, rather than
// falling back to no color at all.
func TestEncodeCarriesForwardOnEmptySourceLegacy(t *testing.T) {
	rows := [][]terminal.Run{
		{
			contentRun("ab"),
			{Text: "cd", Cells: 2, ForegroundSource: "", BackgroundSource: "", Mode: terminal.RunNormal},
		},
	}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	red := hexString(color.RGB{R: 222, G: 56, B: 43})

	assert.Equal(t, 2, strings.Count(doc, `fill="`+red+`"`))
}

// TestEncodeCarriesForwardOnEmptySource pins the "empty" color form: a
// failed color.Defaults.ToAnsi resolution emits nothing, so the glyph must
// inherit whatever foreground color was already active, while the background
// clears so a separator glyph does not inherit the previous segment's rect.
func TestEncodeCarriesForwardOnEmptySource(t *testing.T) {
	rows := [][]terminal.Run{
		{
			contentRun("ab"),
			{Text: "cd", Cells: 2, ForegroundSource: "", BackgroundSource: "", Mode: terminal.RunNormal},
		},
	}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	red := hexString(color.RGB{R: 222, G: 56, B: 43})

	assert.Equal(t, 2, strings.Count(doc, `fill="`+red+`"`))
	assert.Equal(t, 2, strings.Count(doc, "<rect"))
}

// TestEncodeClearsBackgroundOnEmptySource ensures a separator-like run with no
// explicit background paints no rect behind its glyph, so the glyph keeps its
// own shape instead of becoming a solid block over the previous segment's
// background.
func TestEncodeClearsBackgroundOnEmptySource(t *testing.T) {
	rows := [][]terminal.Run{{
		{Text: "A", Cells: 1, ForegroundSource: "red", BackgroundSource: "blue", Mode: terminal.RunNormal},
		{Text: "", Cells: 1, ForegroundSource: "green", BackgroundSource: "", Mode: terminal.RunNormal},
	}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.Equal(t, 2, strings.Count(doc, "<rect"))
}

// TestEncodeAttributes maps every Run.Attributes slot (see knownStyles'
// order in terminal/writer.go) to its CSS equivalent, plus the explicit <r>
// reverse attribute, which swaps text/rect fill in the RunNormal case since
// it is independent of the writer's own transparent-driven RunMode swap.
func TestEncodeAttributes(t *testing.T) {
	blue := hexString(color.RGB{R: 0, G: 111, B: 184})

	cases := []struct {
		Case       string
		WantSubstr []string
		Attrs      [8]uint8
	}{
		{Case: "bold", Attrs: [8]uint8{attrBold: 1}, WantSubstr: []string{`font-weight="bold"`}},
		{Case: "italic", Attrs: [8]uint8{attrItalic: 1}, WantSubstr: []string{`font-style="italic"`}},
		{Case: "underline", Attrs: [8]uint8{attrUnderline: 1}, WantSubstr: []string{`text-decoration="underline"`}},
		{Case: "overline", Attrs: [8]uint8{attrOverline: 1}, WantSubstr: []string{`text-decoration="overline"`}},
		{Case: "strikethrough", Attrs: [8]uint8{attrStrikethrough: 1}, WantSubstr: []string{`text-decoration="line-through"`}},
		{
			Case:       "underline+overline+strikethrough combine into one attribute",
			Attrs:      [8]uint8{attrUnderline: 1, attrOverline: 1, attrStrikethrough: 1},
			WantSubstr: []string{`text-decoration="underline overline line-through"`},
		},
		{Case: "dim reduces opacity", Attrs: [8]uint8{attrDim: 1}, WantSubstr: []string{`opacity="0.6"`}},
		{
			// The class is now content-hashed (see Encode's doc comment on
			// blinkClass), not the fixed "omp-blink" literal, so two different
			// blinking SVGs inlined on the same page can't collide. Assert the
			// shape (prefix + non-empty suffix) rather than the exact value.
			Case: "blink applies a per-render blink class", Attrs: [8]uint8{attrBlink: 1},
			WantSubstr: []string{`class="omp-blink-`},
		},
		{
			Case:  "nested bold (depth 2) still renders as bold, not a stronger weight",
			Attrs: [8]uint8{attrBold: 2}, WantSubstr: []string{`font-weight="bold"`},
		},
		{
			Case: "reverse swaps text/rect fill in RunNormal", Attrs: [8]uint8{attrReverse: 1},
			WantSubstr: []string{`fill="` + blue + `"`},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			rows := [][]terminal.Run{
				{{Text: "hi", Cells: 2, ForegroundSource: "red", BackgroundSource: "blue", Mode: terminal.RunNormal, Attributes: tc.Attrs}},
			}

			doc := Encode(rows, testOptions())
			decodeXML(t, doc)

			for _, substr := range tc.WantSubstr {
				assert.Contains(t, doc, substr)
			}
		})
	}
}

// TestEncodeEscapesXML pins the three characters that must not break the
// document: &, <, and > appearing in prompt text (e.g. a git branch name, or
// literal shell redirection text) must round-trip through the SVG as the
// same text, not as broken markup.
func TestEncodeEscapesXML(t *testing.T) {
	text := `a & b <not-a-tag> c > d`

	rows := [][]terminal.Run{{contentRun(text)}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, "a &amp; b &lt;not-a-tag&gt; c &gt; d")
	assert.NotContains(t, doc, "<not-a-tag>")

	// round-trip through an XML decoder to prove the escaped text decodes
	// back to the original, not just that the raw bytes look escaped.
	decoder := xml.NewDecoder(strings.NewReader(doc))

	var found string

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		charData, ok := token.(xml.CharData)
		if !ok {
			continue
		}

		if strings.Contains(string(charData), "not-a-tag") {
			found = string(charData)
		}
	}

	assert.Equal(t, text, found)
}

// TestEncodeGradientRGBOverridesSource pins that a gradient cell's captured
// RGB (see terminal.Run's doc comment) always wins over its source string,
// since the source can't carry a gradient stop's true color losslessly.
func TestEncodeGradientRGBOverridesSource(t *testing.T) {
	gradientRGB := &color.RGB{R: 1, G: 2, B: 3}

	rows := [][]terminal.Run{
		{{
			Text: "x", Cells: 1, Mode: terminal.RunNormal,
			ForegroundSource: "red", ForegroundRGB: gradientRGB,
			BackgroundSource: "blue",
		}},
	}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, `fill="`+hexString(*gradientRGB)+`"`)
}

// TestEncodeRootCarriesFontMetrics pins the presentation-attribute migration
// (see Encode's doc comment): font-family/font-size/white-space move to the
// root <svg> - ordinary inherited CSS properties, so every <text> descendant
// gets them for free - instead of a <style> rule, so an inlined SVG never
// carries a document-scoped <style> block that would restyle the rest of
// the page. No <text> element carries dominant-baseline at all (see Encode's
// doc comment): every glyph renders on SVG's default alphabetic baseline, at
// a y computed from the row's own geometry rather than anchored via
// "hanging" to a rect edge.
func TestEncodeRootCarriesFontMetrics(t *testing.T) {
	opts := testOptions()
	opts.FontFamily = `"Victor Mono", monospace`

	rows := [][]terminal.Run{{contentRun("hi")}}

	doc := Encode(rows, opts)
	decodeXML(t, doc)

	svgTag := doc[:strings.Index(doc, ">")+1]
	assert.Contains(t, svgTag, `font-family="&quot;Victor Mono&quot;, monospace"`)
	assert.Contains(t, svgTag, `font-size="10px"`)
	assert.NotContains(t, svgTag, "dominant-baseline")

	// Whitespace is declared per <text>, not on the root: neither
	// white-space="pre" nor xml:space="preserve" on the root survives a
	// browser's XML whitespace processing — see writeText's doc comment.
	assert.NotContains(t, svgTag, "white-space")
	assert.NotContains(t, svgTag, "xml:space")

	assert.NotContains(t, doc, "dominant-baseline")
}

// TestEncodeNoStyleWithoutBlink pins the inlining hazard the presentation-
// attribute migration exists to fix: a render with no blinking run must
// carry no <style> element at all, so 124 of these inlined on one page (the
// website's theme gallery) never fight over a document-scoped selector or
// duplicate an unused @keyframes rule.
func TestEncodeNoStyleWithoutBlink(t *testing.T) {
	rows := [][]terminal.Run{{contentRun("hi")}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.NotContains(t, doc, "<style")
	assert.NotContains(t, doc, "@keyframes")
}

// TestEncodeBlinkClassScopedPerRender pins that two Encode calls whose rows
// differ get different blink class/@keyframes names, so two such SVGs
// inlined on the same page can't collide (see Encode's doc comment).
// Identical rows are allowed to collide - the point pinned here is that
// different content can't.
func TestEncodeBlinkClassScopedPerRender(t *testing.T) {
	rowsA := [][]terminal.Run{
		{{Text: "aa", Cells: 2, ForegroundSource: "red", BackgroundSource: "blue", Mode: terminal.RunNormal, Attributes: [8]uint8{attrBlink: 1}}},
	}
	rowsB := [][]terminal.Run{
		{{Text: "bb", Cells: 2, ForegroundSource: "red", BackgroundSource: "blue", Mode: terminal.RunNormal, Attributes: [8]uint8{attrBlink: 1}}},
	}

	docA := Encode(rowsA, testOptions())

	docB := Encode(rowsB, testOptions())

	classPattern := regexp.MustCompile(`class="(omp-blink-[a-z0-9]+)"`)
	classA := classPattern.FindStringSubmatch(docA)
	classB := classPattern.FindStringSubmatch(docB)
	require.Len(t, classA, 2)
	require.Len(t, classB, 2)
	assert.NotEqual(t, classA[1], classB[1])

	assert.Contains(t, docA, "@keyframes "+classA[1])
	assert.Contains(t, docB, "@keyframes "+classB[1])
}

// TestEncodeEmptyRows pins Encode's zero-row behavior: a valid, if minimal,
// SVG document rather than an error.
func TestEncodeEmptyRows(t *testing.T) {
	doc := Encode(nil, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, "<svg")
	assert.Contains(t, doc, "</svg>")
}

// TestEncodeTextPreservesWhitespace pins the declaration that keeps a
// segment template's own padding — the leading/trailing spaces nearly every
// theme writes around its content — from being stripped. Measured in
// headless Chromium, only the per-element form survives; see writeText's doc
// comment for the four declarations that do not. Asserting per-element also
// catches a regression that moves it back onto the root, which decodes as
// valid XML and looks right in the markup while rendering wrong.
func TestEncodeTextPreservesWhitespace(t *testing.T) {
	rows := [][]terminal.Run{{contentRun(" padded ")}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, `<text xml:space="preserve"`)
	assert.Contains(t, doc, "> padded </text>")

	opened := strings.Count(doc, "<text ")
	preserved := strings.Count(doc, `<text xml:space="preserve"`)
	assert.Equal(t, opened, preserved, "every <text> element must declare xml:space")
}

// TestEncodeFillBoxTracksGlyphInk pins that a segment's background is bounded
// by FillAscent/FillDescent — the box a powerline separator glyph inks — and
// not by LineHeight. Filling the row instead left the background standing
// taller than the rounded caps meant to close it, so a bubbles-style theme
// rendered as a tall block with short caps stuck on its ends.
func TestEncodeFillBoxTracksGlyphInk(t *testing.T) {
	rows := [][]terminal.Run{{contentRun("ab")}}

	opts := testOptions()
	opts.FillAscent = 12
	opts.FillDescent = 3

	doc := Encode(rows, opts)
	decodeXML(t, doc)

	baseline := contentBaselineY(opts, 0)

	// The baseline is 0.75*LineHeight below the row top, so with a 15-unit
	// fill box the rect starts 3 units into the row and stops 2 short of its
	// bottom — deliberately not the row box.
	assert.Contains(t, doc, `<rect x="`+formatFloat(contentOriginX(opts))+`" y="`+
		formatFloat(baseline-opts.FillAscent)+`" width="20" height="15" fill="#006fb8"/>`)
	assert.NotContains(t, doc, `height="20" fill="#006fb8"`)

	// The text itself still sits on the row's own baseline: the fill box
	// moves the background, never the glyphs.
	assert.Contains(t, doc, `y="`+formatFloat(baseline)+`"`)
}

// TestEncodeFillBoxDefaultsToHackInkBox pins the fallback: a caller that
// specifies neither metric gets Hack Nerd Font's own U+E0B0 ink box, which is
// measurably shorter than a row.
func TestEncodeFillBoxDefaultsToHackInkBox(t *testing.T) {
	opts := testOptions()
	opts.FillAscent = 0
	opts.FillDescent = 0

	resolved := opts.withDefaults()

	assert.InDelta(t, 9.3408, resolved.FillAscent, 0.0001)
	assert.InDelta(t, 2.4170, resolved.FillDescent, 0.0001)
	assert.Less(t, resolved.FillAscent+resolved.FillDescent, resolved.LineHeight)
}

// TestEncodeCursorAtAnchor pins where the cursor indicator lands on a row
// that carries a right-aligned block: at the primary prompt's own end, not
// after the rprompt. The gap between them absorbs the cursor's cell, so the
// right-aligned content stays flush against the right edge.
func TestEncodeCursorAtAnchor(t *testing.T) {
	rows := [][]terminal.Run{
		{
			contentRun("prompt"),
			{Text: strings.Repeat(" ", 10), Cells: 10},
			{Text: "right", Cells: 5, ForegroundSource: "green", BackgroundSource: "blue", Mode: terminal.RunNormal},
		},
	}

	opts := testOptions()
	opts.Cursor = &Cursor{Row: 0, Run: 1}

	doc := Encode(rows, opts)
	decodeXML(t, doc)

	originX := contentOriginX(opts)
	y := formatFloat(contentBaselineY(opts, 0))

	// The cursor sits at column 6, right after "prompt", and takes the first
	// cell of the 10-cell gap — so the gap that follows is 9 cells and
	// "right" still starts at column 16, exactly where it was without a
	// cursor.
	assert.Contains(t, doc, `x="`+formatFloat(originX+6*opts.CellWidth)+`" y="`+y+`" textLength="10" lengthAdjust="spacingAndGlyphs" fill="#ffffff">_</text>`)
	assert.Contains(t, doc, `x="`+formatFloat(originX+7*opts.CellWidth)+`" y="`+y+`" textLength="90"`)
	assert.Contains(t, doc, `x="`+formatFloat(originX+16*opts.CellWidth)+`" y="`+y+`" textLength="50" lengthAdjust="spacingAndGlyphs" fill="#39b54a">right</text>`)
}

// TestEncodeCursorStaysLegibleOnLightBackground pins the fix for the "default"
// foreground bug: the cursor's ForegroundSource is "default" (see decorate.go's
// insertCursor), which used to resolve to a fixed white regardless of
// CanvasBackground - invisible against a light one. It must now switch to
// defaultForegroundNearBlack whenever CanvasBackground reads as light (see
// defaultForegroundColor).
func TestEncodeCursorStaysLegibleOnLightBackground(t *testing.T) {
	rows := [][]terminal.Run{{contentRun("prompt")}}

	opts := testOptions()
	opts.Cursor = &Cursor{Row: 0, Run: 0}
	light := color.RGB{R: 0xf5, G: 0xf5, B: 0xf0}
	opts.CanvasBackground = &light

	doc := Encode(rows, opts)
	decodeXML(t, doc)

	assert.Contains(t, doc, `fill="`+hexString(defaultForegroundNearBlack)+`">_</text>`)
	assert.NotContains(t, doc, `fill="#ffffff">_</text>`)
}

// TestEncodeCursorWithoutAnchor pins the fallback for a caller that has no
// anchor to give: the end of the last row, which is what a prompt with no
// right-aligned block wants anyway.
func TestEncodeCursorWithoutAnchor(t *testing.T) {
	rows := [][]terminal.Run{{contentRun("prompt")}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	originX := contentOriginX(testOptions())
	y := formatFloat(contentBaselineY(testOptions(), 0))

	assert.Contains(t, doc, `x="`+formatFloat(originX+6*testOptions().CellWidth)+`" y="`+y+`" textLength="10" lengthAdjust="spacingAndGlyphs" fill="#ffffff">_</text>`)
}
