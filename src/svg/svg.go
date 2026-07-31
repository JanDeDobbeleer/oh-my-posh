// Package svg renders a captured terminal.Run stream to a self-contained SVG
// document. It is a sibling encoder to the ANSI writer (see
// terminal.CaptureRuns): both start from the same Run stream, one emitting
// escape sequences, the other emitting markup, so neither has to parse the
// other's output.
package svg

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// Options configures the SVG canvas: font metrics, plus the two colors a Run
// stream cannot carry itself. BackgroundSource/ForegroundSource can hold
// "accent" or the ansiColorCodes name "default", both of which resolve
// against live package state in the ANSI writer (color.Defaults' accent,
// the real terminal background) rather than anything in the Run itself.
// Taking them as explicit fields — instead of Encode reaching for that
// package state — is what makes color resolution unit-testable.
type Options struct {
	TerminalBackground *color.RGB
	AccentForeground   *color.RGB
	AccentBackground   *color.RGB
	// CanvasBackground serves two roles that must stay the same color for a
	// reverse-video cutout to read as a hole punched through to the window
	// behind it, rather than a mismatched patch: it is both the glyph color
	// painted for a transparent-foreground run (terminal.RunBackgroundPainted
	// and terminal.RunReverseVideo alike — see paintRun) and, since 6e2377a9,
	// the fill of the terminal window Encode draws behind every row (see
	// window.go). A caller only ever has reason to set this explicitly when
	// TerminalBackground is unknown: withDefaults enforces the invariant
	// both roles depend on — the real terminal background where known,
	// this field's own value otherwise — once, by overwriting
	// CanvasBackground with TerminalBackground whenever the latter is
	// non-nil, so every reader downstream (windowFill in Encode, paintRun's
	// transparent arms) can use CanvasBackground unconditionally instead of
	// re-deriving the same choice. Absent a known terminal background,
	// withDefaults falls back to the retired PNG renderer's own fixed
	// #151515.
	CanvasBackground *color.RGB
	// Cursor is where the cursor indicator is glued onto the prompt (see
	// Cursor's own doc comment and decorate). Nil puts it at the end of the
	// last row, which is only correct for a prompt with no right-aligned
	// block.
	Cursor     *Cursor
	FontFamily string
	FontSize   float64
	// CellWidth is the horizontal advance of one monospace cell, in the same
	// unit FontSize is (an absolute value, not a ratio of it — a caller sets
	// both explicitly, e.g. testOptions' FontSize:10/CellWidth:10 pairing).
	// Zero falls back to FontSize*defaultCellWidth (~0.6021em), Hack Nerd
	// Font's own advance ratio read straight from its hmtx table (ASCII 'M':
	// 1233 font units / 2048 units-per-em — see defaultCellWidth's doc
	// comment). That default matches the bundled font stack (see
	// defaultFontFamily), but the website renders these in Victor Mono,
	// whose advance ratio is measurably different (0.5455em) — a caller
	// with a different font should set CellWidth explicitly rather than
	// assume this default (see cli/config_export_svg.go's svgOptions for
	// how the CLI's --cell-width, quoted as a font-size ratio, converts to
	// this field's absolute unit).
	CellWidth float64
	// LineHeight is the vertical advance of one row, in the same absolute
	// unit as CellWidth. Zero falls back to FontSize*defaultLineHeight
	// (~1.397em), Hack Nerd Font's own hhea-derived line height times the
	// retired PNG renderer's lineSpacing constant (see defaultLineHeight's
	// doc comment). Like CellWidth, this is font-specific — Victor Mono's
	// is ~1.691em — and a caller rendering in a different font should set
	// it explicitly.
	LineHeight float64
	// FillAscent/FillDescent bound a segment's background rect vertically,
	// as distances above and below the text baseline in the same absolute
	// unit as CellWidth/LineHeight. Together they are the box a powerline
	// separator glyph actually inks, which is not the row box: a row is
	// LineHeight tall, but U+E0B0 and the rounded U+E0B4/E0B6 caps ink only
	// about 84% of that, and asymmetrically about the baseline. Filling the
	// whole row instead left every segment's background standing taller than
	// the caps that are supposed to close it, so a rounded theme rendered as
	// a tall block with short caps stuck on its ends.
	//
	// Zero falls back to Hack Nerd Font's own U+E0B0 ink box (see
	// defaultFillAscent/defaultFillDescent). Like CellWidth and LineHeight
	// these are font-specific — Victor Mono's are 1.0982/0.3255 em — so a
	// caller rendering in a different font should set them explicitly.
	FillAscent  float64
	FillDescent float64
	// Columns fixes the canvas' content width at Columns*CellWidth,
	// independent of any row's actual content: the terminal-window
	// presentation this package restores needs a stable frame size
	// regardless of what a theme renders, exactly like every reference PNG
	// the old renderer produced was a fixed-width window rather than one
	// that shrink-wrapped to content. Zero falls back to defaultColumns
	// (120), matching both the retired renderer's own default and the
	// image-export CLI command's --terminal-width flag default, so a
	// caller that doesn't set this still gets the terminal-shaped canvas
	// the reference PNGs always had.
	Columns int
}

const (
	// baselineRatio is how far down a row its glyphs sit, as a share of
	// LineHeight, backed out from the retired PNG renderer's own
	// background-rect placement (bgY = y - fontLineHeight*0.75).
	baselineRatio = 0.75

	// defaultFontFamily prefers the "Mono" patched Nerd Font variants, and
	// only falls back to the plain ones. Both patch the same codepoints at
	// the same 1-cell advance, so the choice does not move a single glyph on
	// this package's grid — but the plain variants leave an icon's *ink* at
	// its natural width while advancing one cell, and Nerd Font icons are
	// far wider than a cell: measured in HackNerdFont-Regular.ttf, U+F07B
	// (folder) inks 1.661 cells, U+F00C (check) 1.453, U+F120 (terminal)
	// 1.869. That ink overruns the following cell, which for nearly every
	// theme is the space its own template writes after the icon — so the
	// prompt reads as though the padding were missing. The Mono variants
	// scale each icon down to exactly one cell of ink, which is also what a
	// terminal shows, since a Mono patch is the variant Nerd Fonts ships for
	// terminal use.
	defaultFontFamily = `"CaskaydiaCove Nerd Font Mono", "FiraCode Nerd Font Mono", "Hack Nerd Font Mono", ` +
		`"CaskaydiaCove Nerd Font", "FiraCode Nerd Font", "Hack Nerd Font", ui-monospace, monospace`
	// DefaultFontSize is the font-size withDefaults falls back to when a caller's
	// Options leaves FontSize unset. It is exported so a caller building its own
	// FontSize-relative values (e.g. the image CLI command's --cell-width/
	// --line-height/--fill-ascent/--fill-descent flags, each a multiple of
	// font-size) can scale against the same default this package renders with,
	// instead of hardcoding a second copy of it.
	DefaultFontSize = 16.0
	// defaultColumns matches the retired PNG renderer's own defaultColumns
	// and the image-export CLI command's --terminal-width flag default: a
	// caller that never sets Options.Columns still gets a 120-column window.
	defaultColumns = 120

	// defaultCellWidth is Hack Nerd Font's true advance ratio for an ASCII
	// glyph, read directly from the font's own hmtx table rather than
	// reverse-engineered from a rendered reference: HackNerdFont-Regular.ttf
	// reports unitsPerEm 2048 and an 'M' advance of 1233 units, 1233/2048 =
	// 0.60205...em. Every Nerd Font Private Use Area icon measured
	// alongside it (E0B0, F408, F044, E62A, E725, F489) also advances
	// exactly 1.000 cell, in both Hack and Victor Mono — the icons overflow
	// their own cell's ink, sometimes past 1.8 cells wide, but that is by
	// design (a terminal renders the next glyph a full cell over
	// regardless, letting the ink spill) and not something to budget extra
	// cell width for; see the now-removed doublewidth.go this package used
	// to carry for exactly that mistake.
	//
	// A prior version of this default used 7.0/12.0 (0.58333em), backed out
	// from the retired PNG renderer's own reference canvas arithmetic. That
	// number encodes a bug, not a font property: the PNG renderer computed
	// its advance via `font.Drawer.MeasureString(" ") >> 6`, a 26.6
	// fixed-point shift that truncates the fractional pixel rather than
	// rounding it (image.go's fontHeight/spaceAdvance). At its 48px
	// reference em, Hack's true advance is 48*0.60205 = 28.9px, which the
	// PNG renderer truncated down to 28px — 0.9px lost per column, 3.2% of
	// the true advance. Reproducing that truncation here would bake a
	// rounding defect specific to one em size into a vector format that has
	// no rounding to do at all, and it would be flatly wrong for any other
	// font or FontSize. See defaultLineHeight for the matching line-height
	// derivation and the residual drift this choice leaves against the old
	// PNG references.
	defaultCellWidth = 1233.0 / 2048.0

	// defaultLineHeight is Hack Nerd Font's own line height — hhea's
	// Ascender - Descender + LineGap, 2384 units at unitsPerEm 2048 =
	// 1.1641em — times 1.2, the retired PNG renderer's own lineSpacing
	// constant (image.go: `ir.lineSpacing = 1.2`, applied as
	// `regular.Metrics().Height * lineSpacing`, image.go ~:310-312,
	// ~:727-761). 1.1641 * 1.2 = 1.397em.
	//
	// A prior version of this default was a flat 1.4, backed out from the
	// PNG reference renders' pixel dimensions (48px em, 66px content rows,
	// 66/48 ≈ 1.375) rather than font metrics. That number is close to
	// Victor Mono's own ratio (1.4091 * 1.2 = 1.691em) but every reference
	// PNG was actually rendered in Hack (1.397em) — a coincidence of
	// fitting the wrong font, not a derivation. See defaultCellWidth's doc
	// comment for why a metric derived from the font, not from the old
	// renderer's rounded pixels, is the correct default here too.
	defaultLineHeight = (2384.0 / 2048.0) * 1.2

	// defaultFillAscent/defaultFillDescent are Hack Nerd Font's own U+E0B0
	// ink box, read from its glyf entry: y -495..1913 at unitsPerEm 2048.
	// Every powerline separator the bundled themes draw shares that exact
	// box — U+E0B0, U+E0B2, U+E0B4, U+E0B6 and U+E0B8 all report the same
	// -495..1913 — so one pair covers triangles, rounded caps and slants
	// alike rather than needing a per-glyph table.
	//
	// The box is deliberately a little taller than the font's own hhea line
	// box (1.1758 em against 1.1641 em): Nerd Fonts gives these glyphs a
	// small overshoot so adjacent segments meet with no seam. Deriving the
	// fill from hhea instead would leave a hairline of canvas showing
	// through between a segment and its own separator.
	defaultFillAscent  = 1913.0 / 2048.0
	defaultFillDescent = 495.0 / 2048.0
)

// defaultCanvasBackground is the retired PNG renderer's own fixed fallback
// (image.go's defaultBackgroundColor, #151515) for the glyph color a
// transparent-foreground run paints when no real terminal background is
// known; see Options.CanvasBackground.
var defaultCanvasBackground = color.RGB{R: 0x15, G: 0x15, B: 0x15}

// withDefaults fills every zero-value metric with a sensible default, so a
// caller only has to set the fields it cares about — typically just the two
// colors — and still gets a fully specified canvas.
func (o *Options) withDefaults() Options {
	out := *o

	if out.FontFamily == "" {
		out.FontFamily = defaultFontFamily
	}

	if out.FontSize == 0 {
		out.FontSize = DefaultFontSize
	}

	if out.CellWidth == 0 {
		out.CellWidth = out.FontSize * defaultCellWidth
	}

	if out.LineHeight == 0 {
		out.LineHeight = out.FontSize * defaultLineHeight
	}

	if out.FillAscent == 0 {
		out.FillAscent = out.FontSize * defaultFillAscent
	}

	if out.FillDescent == 0 {
		out.FillDescent = out.FontSize * defaultFillDescent
	}

	// The real terminal background wins over any explicit CanvasBackground
	// whenever it's known — see Options.CanvasBackground's doc comment on
	// why this is the one place that invariant is decided, rather than each
	// reader re-deriving "the real terminal background where known, else
	// CanvasBackground" for itself.
	if out.TerminalBackground != nil {
		out.CanvasBackground = out.TerminalBackground
	}

	if out.CanvasBackground == nil {
		out.CanvasBackground = &defaultCanvasBackground
	}

	if out.Columns <= 0 {
		out.Columns = defaultColumns
	}

	return out
}

// Encode renders rows of runs — one []terminal.Run per output row, as
// prompt.Engine.CapturedRuns returns — into a self-contained SVG document.
//
// Layout is a monospace cell grid: every run advances the running column by
// its own Cells, never by measuring glyphs. Each <text> element also carries
// textLength/lengthAdjust pinning it to exactly Cells*CellWidth, so it lands
// on the grid regardless of the browser's own font metrics. A Nerd Font
// icon's ink commonly overflows that one-cell box (see defaultCellWidth's
// doc comment on measured ink-vs-advance ratios) — that overflow is left
// alone, exactly like a real terminal lets a wide glyph draw over the
// following cell's left edge, rather than compensated for with extra grid
// width.
//
// A run with Cells == 0 (e.g. a hyperlink's URL text — see terminal.Run's
// doc comment) is skipped entirely: no rect, no text, no column advance.
//
// font-family/font-size are written as presentation attributes on the root
// <svg> element rather than a <style> rule: both are ordinary inherited CSS
// properties, so every descendant <text> picks them up for free, and a
// document embedding many of these SVGs inline never gets a document-scoped
// <style> block fighting its own page styles. Whitespace is the exception —
// it is declared per <text> instead, see writeText. Each <text>
// element carries no dominant-baseline/alignment-baseline attribute at all,
// so it renders on SVG's default alphabetic baseline; encodeRow computes
// that baseline's y itself (rowTop + 0.75*LineHeight, see writeText) rather
// than anchoring to a rect edge via "hanging", because this package has no
// server-side font metrics and "hanging" baseline placement is implemented
// inconsistently across browsers — a computed alphabetic-baseline offset is
// deterministic everywhere instead.
//
// A <style> block is emitted only when a run actually blinks (attrBlink >
// 0): no bundled theme does, so a gallery of these SVGs carries no <style>
// at all. When one does blink, its class name and @keyframes name are
// suffixed with a hash of the row content, so two different blinking SVGs
// inlined on the same page never share (and can't clash on) the same
// document-scoped class/keyframes name.
//
// Encode takes opts by value, not by pointer, even though Options is 80
// bytes (over gocritic's hugeParam threshold): Encode runs once per prompt
// render, so the extra copy is immaterial next to the string-building it
// does, and the by-value signature keeps the public API's ergonomic call
// shape (Encode(rows, Options{...})) rather than forcing every caller through
// a throwaway local just to take its address.
//
//nolint:gocritic
func Encode(rows [][]terminal.Run, opts Options) string {
	opts = opts.withDefaults()
	// Decorate first, fit second: Options.Cursor indexes the captured rows,
	// and fitRow is free to collapse gaps and wrap a row into several, which
	// would invalidate that index. Doing it in this order also means the
	// cursor is fitted like any other content — though insertCursor takes its
	// cell out of the following alignment gap where there is one, so in
	// practice a row that fitted before still fits.
	rows = fitRows(decorate(rows, opts.Cursor), opts.Columns)

	blinks := false
	hasher := fnv.New64a()

	for _, row := range rows {
		for i := range row {
			if row[i].Attributes[attrBlink] > 0 {
				blinks = true
			}

			hasher.Write([]byte(row[i].Text))
			hasher.Write(row[i].Attributes[:])
		}
	}

	var blinkClass string
	if blinks {
		blinkClass = "omp-blink-" + strconv.FormatUint(hasher.Sum64(), 36)
	}

	geo := newWindowGeometry(&opts)
	size := newCanvasSize(&geo, opts.Columns, opts.CellWidth, len(rows), opts.LineHeight)

	var b strings.Builder

	// A single format string in place of the interleaved WriteString calls
	// this root tag used to be built from — see writeRect's doc comment on
	// why: Encode runs once per prompt, so there is no allocation budget
	// this would need to respect.
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" font-family="%s" font-size="%spx">`+"\n",
		formatFloat(size.width), formatFloat(size.height), formatFloat(size.width), formatFloat(size.height),
		escapeAttr(opts.FontFamily), formatFloat(opts.FontSize))

	if blinkClass != "" {
		writeBlinkStyle(&b, blinkClass)
	}

	// The window's own fill is always opts.CanvasBackground: withDefaults
	// has already folded a known TerminalBackground into it (see Options.
	// CanvasBackground's doc comment), so there is no separate preference to
	// apply here.
	writeWindowChrome(&b, size, &geo, *opts.CanvasBackground)

	for rowIndex, row := range rows {
		encodeRow(&b, row, rowIndex, size, opts, blinkClass)
	}

	b.WriteString("</svg>")

	return b.String()
}

// fitRows applies fitRow (see fit.go) to every captured row, flattening any
// row that had to be collapsed/wrapped into several output rows. columns
// must be positive — Encode's own call site always passes a post-
// withDefaults Columns of at least 1 (see its doc comment) — a caller
// bypassing withDefaults must enforce that itself.
func fitRows(rows [][]terminal.Run, columns int) [][]terminal.Run {
	fitted := make([][]terminal.Run, 0, len(rows))

	for _, row := range rows {
		fitted = append(fitted, fitRow(row, columns)...)
	}

	return fitted
}

func writeBlinkStyle(b *strings.Builder, blinkClass string) {
	b.WriteString("<style>\n.")
	b.WriteString(blinkClass)
	b.WriteString(" { animation: ")
	b.WriteString(blinkClass)
	b.WriteString(" 1s steps(1, end) infinite; }\n@keyframes ")
	b.WriteString(blinkClass)
	b.WriteString(" { 50% { opacity: 0; } }\n</style>\n")
}

// paintState carries the last resolved color per channel across runs in the
// same row. Foreground keeps the prior color on an unresolved source so a run
// with no explicit foreground still inherits the terminal state; background
// instead clears on an unresolved or transparent source so a separator glyph
// does not inherit the previous segment's rect and render as a solid block.
type paintState struct {
	fg *color.RGB
	bg *color.RGB
}

// encodeRow takes opts by value for the same reason Encode does (see its
// comment): it runs once per row, not once per Run, so the 80-byte copy
// gocritic flags is immaterial. size.contentX/contentY (see canvasSize) is
// the content grid's top-left corner inside the window, replacing what used
// to be a flat opts.Padding offset from the canvas edge before the window
// chrome existed.
//
//nolint:gocritic
func encodeRow(b *strings.Builder, runs []terminal.Run, rowIndex int, size canvasSize, opts Options, blinkClass string) {
	state := paintState{}
	cell := 0
	rowTop := size.contentY + float64(rowIndex)*opts.LineHeight

	// baseline places the glyphs roughly centered in the row instead of
	// hugging its top edge (see the "hanging" removal note on Encode).
	baseline := rowTop + opts.LineHeight*baselineRatio

	// A segment's background fills the box its own separator glyphs ink, not
	// the row box — see Options.FillAscent/FillDescent.
	fillTop := baseline - opts.FillAscent
	fillHeight := opts.FillAscent + opts.FillDescent

	for i := range runs {
		run := &runs[i]
		textRGB, rectRGB := paintRun(run, &state, &opts)

		if run.Cells == 0 {
			continue
		}

		x := size.contentX + float64(cell)*opts.CellWidth
		runWidth := float64(run.Cells) * opts.CellWidth

		if rectRGB != nil {
			writeRect(b, x, fillTop, runWidth, fillHeight, *rectRGB)
		}

		if run.Text != "" {
			writeText(b, run.Text, x, baseline, runWidth, textRGB, run.Attributes, blinkClass)
		}

		cell += run.Cells
	}
}

// writeRect writes a self-closing <rect> element via a single format string
// rather than interleaved WriteString calls: this package renders once per
// prompt (see Encode's doc comment on why it takes Options by value), so the
// extra formatting cost buys nothing over the allocation-averse style the
// interleaved calls used to be written in.
func writeRect(b *strings.Builder, x, y, width, height float64, rgb color.RGB) {
	fmt.Fprintf(b, `<rect x="%s" y="%s" width="%s" height="%s" fill="%s"/>`+"\n",
		formatFloat(x), formatFloat(y), formatFloat(width), formatFloat(height), hexString(rgb))
}

// writeText writes a <text> element whose y is already the baseline
// coordinate (see encodeRow's baseline), not a rect's top edge: the element
// deliberately carries no dominant-baseline attribute, so it renders on
// SVG's default alphabetic baseline (see Encode's doc comment).
//
// xml:space="preserve" has to sit on every <text> element, not once on the
// root <svg>: a segment template's own padding (" {{ .Path }} ", the leading
// and trailing spaces nearly every theme writes) is ordinary XML whitespace,
// which the default xml:space="default" strips at the ends of an element and
// collapses in the middle. Measured in headless Chromium, the string "  ab  "
// renders as 2 cells rather than 6 under all of: no declaration, the root
// carrying white-space="pre" (what this package wrote before), the root
// carrying style="white-space:pre", and even the root carrying
// xml:space="preserve" — only the per-element form preserves it. The damage
// compounds with textLength/lengthAdjust below: the surviving glyphs are
// stretched across the full run width, so a collapsed 3-cell " ✓ " renders as
// one glyph smeared 3 cells wide rather than a centered check.
func writeText(b *strings.Builder, text string, x, y, width float64, rgb *color.RGB, attrs [8]uint8, blinkClass string) {
	b.WriteString(`<text xml:space="preserve" x="`)
	b.WriteString(formatFloat(x))
	b.WriteString(`" y="`)
	b.WriteString(formatFloat(y))
	b.WriteString(`" textLength="`)
	b.WriteString(formatFloat(width))
	b.WriteString(`" lengthAdjust="spacingAndGlyphs"`)

	if rgb != nil {
		b.WriteString(` fill="`)
		b.WriteString(hexString(*rgb))
		b.WriteString(`"`)
	}

	if attrs[attrBold] > 0 {
		b.WriteString(` font-weight="bold"`)
	}

	if attrs[attrItalic] > 0 {
		b.WriteString(` font-style="italic"`)
	}

	if decoration := textDecoration(attrs); decoration != "" {
		b.WriteString(` text-decoration="`)
		b.WriteString(decoration)
		b.WriteString(`"`)
	}

	if attrs[attrDim] > 0 {
		b.WriteString(` opacity="0.6"`)
	}

	if attrs[attrBlink] > 0 {
		b.WriteString(` class="`)
		b.WriteString(blinkClass)
		b.WriteString(`"`)
	}

	b.WriteString(">")
	b.WriteString(escapeXML(text))
	b.WriteString("</text>\n")
}

// attribute slot indices mirror knownStyles' order in terminal/writer.go
// (bold, underline, overline, italic, strikethrough, dim/faint, blink,
// reverse), which Run.Attributes' index space is keyed on; see run.go's
// runAttributeSlots.
const (
	attrBold = iota
	attrUnderline
	attrOverline
	attrItalic
	attrStrikethrough
	attrDim
	attrBlink
	attrReverse
)

func textDecoration(attrs [8]uint8) string {
	var parts []string

	if attrs[attrUnderline] > 0 {
		parts = append(parts, "underline")
	}

	if attrs[attrOverline] > 0 {
		parts = append(parts, "overline")
	}

	if attrs[attrStrikethrough] > 0 {
		parts = append(parts, "line-through")
	}

	return strings.Join(parts, " ")
}

// paintRun resolves the RGB a run's rect and text should paint, applying its
// RunMode (see terminal.RunMode's doc comment) and its reverse-video
// attribute. bold/italic/underline/overline/strikethrough/dim/blink are
// applied later, directly from Run.Attributes, since they don't affect which
// color goes where. The caller decides whether to draw a rect at all by
// testing rectRGB != nil itself; a run with no resolved background paints no
// rect, so a separate bool would only ever restate that nil check.
func paintRun(run *terminal.Run, state *paintState, opts *Options) (textRGB, rectRGB *color.RGB) {
	fg, fgResolved := resolveChannel(run.ForegroundSource, run.ForegroundRGB, false, opts)
	if fgResolved {
		state.fg = fg
	}

	bg, bgResolved := resolveChannel(run.BackgroundSource, run.BackgroundRGB, true, opts)
	state.bg = nil
	if bgResolved {
		state.bg = bg
	}

	switch run.Mode {
	case terminal.RunBackgroundPainted, terminal.RunReverseVideo:
		// the foreground is transparent: the cell fills with the segment's
		// own background, and the glyph is cut out of that fill in
		// opts.CanvasBackground — matching the real terminal rendering (see
		// terminal.RunMode's doc comment): SGR 7 swaps a foreground code
		// carrying the segment's background with the just-reset default
		// background, so the segment background ends up as the visible
		// fill, not the glyph color. Both RunModes resolve to the exact same
		// textRGB here: withDefaults has already folded a known
		// TerminalBackground into CanvasBackground (see Options.
		// CanvasBackground's doc comment), so there is no longer a
		// mode-specific "known vs unknown terminal background" choice to
		// make at this call site.
		rectRGB = state.bg
		textRGB = opts.CanvasBackground
	default:
		textRGB = state.fg
		rectRGB = state.bg
	}

	// SGR 7 (reverse video) is a real terminal's on/off flag, not a counter:
	// sending it again while already in one of the writer's own
	// transparent-driven reverse renderings above changes nothing, so the
	// explicit <r> attribute only swaps anything in the RunNormal case.
	if run.Attributes[attrReverse] > 0 && run.Mode == terminal.RunNormal {
		textRGB, rectRGB = rectRGB, textRGB
	}

	return textRGB, rectRGB
}

// formatFloat renders v rounded to at most 2 decimal places, using the
// shortest representation that round-trips at that precision — an integer
// value still prints as "10", not "10.00" — since Encode's own arithmetic
// (cell/row counts times a float metric) routinely produces values like
// 140.79999999999998 that carry no visual meaning past the hundredths place.
func formatFloat(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

func hexString(rgb color.RGB) string {
	const hexDigits = "0123456789abcdef"

	buf := [7]byte{'#'}

	for i, c := range [3]uint8{rgb.R, rgb.G, rgb.B} {
		buf[1+i*2] = hexDigits[c>>4]
		buf[2+i*2] = hexDigits[c&0x0f]
	}

	return string(buf[:])
}

func escapeXML(s string) string {
	return xmlReplacer.Replace(s)
}

var xmlReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// escapeAttr escapes a value destined for a double-quoted XML attribute.
// This is stricter than escapeXML: an attribute value like Options.FontFamily
// ordinarily contains literal double quotes (e.g. `"CaskaydiaCove Nerd
// Font", ui-monospace`), which was harmless as plain text content inside the
// old <style> block but would prematurely close a font-family="..." attribute
// if left unescaped.
func escapeAttr(s string) string {
	return attrReplacer.Replace(s)
}

var attrReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
