package svg

import (
	"strconv"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extractWidth parses the root <svg> tag's own width="..." attribute back
// out of a rendered document as a float, the one thing both
// TestEncodeCanvasIsFixedRegardlessOfContent and
// TestEncodeColumnsScalesCanvasWidth need from a document — collapsed here
// instead of each test carrying its own copy (one returning the raw
// substring, the other parsing it to float64).
func extractWidth(t *testing.T, doc string) float64 {
	t.Helper()

	tag := doc[:strings.Index(doc, ">")+1]
	start := strings.Index(tag, `width="`) + len(`width="`)
	end := start + strings.Index(tag[start:], `"`)

	v, err := strconv.ParseFloat(tag[start:end], 64)
	require.NoError(t, err)

	return v
}

// TestEncodeCanvasIsFixedRegardlessOfContent pins the whole point of
// restoring the terminal-window presentation (see the task brief this
// package was rebuilt from): every reference PNG render was a fixed
// 120-column window no matter what the theme rendered, and the SVG canvas
// must be exactly as content-independent. Two renders with wildly
// different content, same Columns, must produce the same canvas size.
func TestEncodeCanvasIsFixedRegardlessOfContent(t *testing.T) {
	opts := testOptions()

	short := [][]terminal.Run{{contentRun("hi")}}
	long := [][]terminal.Run{{contentRun(strings.Repeat("x", 30))}}

	docShort := Encode(short, opts)
	docLong := Encode(long, opts)

	assert.InDelta(t, extractWidth(t, docShort), extractWidth(t, docLong), 0.01)
}

// TestEncodeColumnsScalesCanvasWidth pins that Options.Columns, not row
// content, is what drives the canvas' content width: doubling Columns
// (same content) must add exactly the extra columns' worth of CellWidth to
// the canvas.
func TestEncodeColumnsScalesCanvasWidth(t *testing.T) {
	narrow := testOptions()
	narrow.Columns = 10

	wide := testOptions()
	wide.Columns = 20

	rows := [][]terminal.Run{{contentRun("hi")}}

	docNarrow := Encode(rows, narrow)
	docWide := Encode(rows, wide)

	gotDelta := extractWidth(t, docWide) - extractWidth(t, docNarrow)
	wantDelta := float64(wide.Columns-narrow.Columns) * wide.CellWidth
	assert.InDelta(t, wantDelta, gotDelta, 0.01)
}

// TestEncodeWindowChromeElements pins every distinct piece of window chrome:
// the hard offset shadow silhouette that matches the website's framed code
// blocks (a filled rect, not a filter — see shadowOffset's doc comment), the
// header bar path, the content pane path, the border rect (stroked,
// unfilled), and the "−"/"▢"/"×" window-control glyphs in fixed
// left-to-right order. No title or "+"/"⌄" tab-control glyphs are present
// (see writeWindowChrome's doc comment on why those were dropped).
func TestEncodeWindowChromeElements(t *testing.T) {
	rows := [][]terminal.Run{{contentRun("hi")}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, `class="omp-window-shadow"`)
	assert.Contains(t, doc, `class="omp-window-header"`)
	assert.Contains(t, doc, `class="omp-window-content"`)
	assert.Contains(t, doc, `class="omp-window-divider"`)
	assert.Contains(t, doc, `class="omp-window" x="`)
	assert.Contains(t, doc, `rx="`)
	assert.Contains(t, doc, `stroke="#ffffff"`)
	assert.Contains(t, doc, `stroke-width="1"`)
	assert.Contains(t, doc, `stroke-opacity="0.8"`)
	assert.Contains(t, doc, `fill="none"`)
	assert.NotContains(t, doc, "<feDropShadow")
	assert.NotContains(t, doc, "filter=")

	// The shadow must be painted before the window it sits behind.
	shadowIdx := strings.Index(doc, `class="omp-window-shadow"`)
	windowIdx := strings.Index(doc, `class="omp-window" x="`)
	assert.Less(t, shadowIdx, windowIdx, "the shadow silhouette must be painted before the window covering it")

	assert.NotContains(t, doc, `class="omp-window-title"`)
	assert.NotContains(t, doc, `class="omp-tab-control"`)

	minimizeIdx := strings.Index(doc, "\u2212")
	maximizeIdx := strings.Index(doc, "\u25a2")
	closeIdx := strings.Index(doc, ">\u00d7<")
	require.NotEqual(t, -1, minimizeIdx)
	require.NotEqual(t, -1, maximizeIdx)
	require.NotEqual(t, -1, closeIdx)
	assert.Less(t, minimizeIdx, maximizeIdx, "window controls must appear minimize, maximize, close left to right")
	assert.Less(t, maximizeIdx, closeIdx, "window controls must appear minimize, maximize, close left to right")

	assert.Equal(t, 3, strings.Count(doc, `class="omp-window-control"`))
}

// TestEncodeWindowHeaderDiffersFromContent pins the whole point of the
// two-tone chrome: the header bar's own fill must not equal the content
// pane's fill (opts.CanvasBackground), or the two panes would read as one
// continuous surface again.
func TestEncodeWindowHeaderDiffersFromContent(t *testing.T) {
	opts := testOptions()
	custom := color.RGB{R: 10, G: 20, B: 30}
	opts.CanvasBackground = &custom

	doc := Encode([][]terminal.Run{{contentRun("hi")}}, opts)
	decodeXML(t, doc)

	headerStart := strings.Index(doc, `class="omp-window-header" fill="`) + len(`class="omp-window-header" fill="`)
	headerFill := doc[headerStart : headerStart+7]

	assert.NotEqual(t, hexString(custom), headerFill)
}

// TestEncodeWindowControlsStayLegibleOnLightHeader pins the same "contrast
// with what's actually behind it" fix controlColor makes: the window
// controls sit on the header bar, not the content pane, so they must
// contrast with headerColor's own output rather than a fixed light gray -
// a light theme's header (see headerColor: darker than a light content
// fill, but still light overall) used to leave them nearly invisible.
func TestEncodeWindowControlsStayLegibleOnLightHeader(t *testing.T) {
	opts := testOptions()
	light := color.RGB{R: 0xf5, G: 0xf5, B: 0xf0}
	opts.CanvasBackground = &light

	doc := Encode([][]terminal.Run{{contentRun("hi")}}, opts)
	decodeXML(t, doc)

	assert.Contains(t, doc, `fill="`+hexString(defaultForegroundNearBlack)+`" text-anchor="middle">`+"\u2212</text>")
	assert.NotContains(t, doc, `fill="#cfcfcf"`)
}

func TestEncodeWindowBorderUsesThemeAwareStrokeAndShadow(t *testing.T) {
	lightOpts := testOptions()
	light := color.RGB{R: 0xf5, G: 0xf5, B: 0xf0}
	lightOpts.CanvasBackground = &light
	lightDoc := Encode([][]terminal.Run{{contentRun("hi")}}, lightOpts)
	assert.Contains(t, lightDoc, `stroke="#dadde1"`)
	assert.Contains(t, lightDoc, `class="omp-window-shadow"`)
	assert.Contains(t, lightDoc, `fill="#0f1722"`)

	darkOpts := testOptions()
	dark := color.RGB{R: 0x0f, G: 0x17, B: 0x22}
	darkOpts.CanvasBackground = &dark
	darkDoc := Encode([][]terminal.Run{{contentRun("hi")}}, darkOpts)
	assert.Contains(t, darkDoc, `stroke="#ffffff"`)
	assert.Contains(t, darkDoc, `fill="#ffffff"`)
}

// TestEncodeWindowFillMatchesCanvasBackground pins Options.CanvasBackground's
// dual role (see its doc comment): with no known terminal background, the
// window's own fill must be the exact same color a reverse-video cutout
// glyph paints, not an independently chosen one.
func TestEncodeWindowFillMatchesCanvasBackground(t *testing.T) {
	opts := testOptions()
	custom := color.RGB{R: 10, G: 20, B: 30}
	opts.CanvasBackground = &custom

	doc := Encode([][]terminal.Run{{contentRun("hi")}}, opts)

	assert.Contains(t, doc, `class="omp-window-content" fill="`+hexString(custom)+`"`)
}

// TestEncodeWindowFillPrefersKnownTerminalBackground pins the override:
// when the real terminal background is actually known, it wins over
// CanvasBackground for the window's own fill (see Encode's doc comment on
// windowFill).
func TestEncodeWindowFillPrefersKnownTerminalBackground(t *testing.T) {
	opts := testOptions()
	termBg := color.RGB{R: 40, G: 50, B: 60}
	opts.TerminalBackground = &termBg

	doc := Encode([][]terminal.Run{{contentRun("hi")}}, opts)

	assert.Contains(t, doc, `class="omp-window-content" fill="`+hexString(termBg)+`"`)
	assert.NotContains(t, doc, `fill="`+hexString(defaultCanvasBackground)+`"`)
}

// TestEncodeAppendsCursorAndWatermark pins full parity with the retired PNG
// renderer (image.go's cleanContent): a cursor glued onto the end of the
// last row of actual content, then a blank row, then the bold watermark on
// its own row - and the cursor must carry no background of its own,
// regardless of what background the row it's glued onto last used.
func TestEncodeAppendsCursorAndWatermark(t *testing.T) {
	rows := [][]terminal.Run{{contentRun("hi")}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, ">hi<")
	assert.Contains(t, doc, ">_<")
	assert.Contains(t, doc, ">ohmyposh.dev<")
	assert.Contains(t, doc, `>ohmyposh.dev</text>`)

	// The watermark is bold.
	watermarkIdx := strings.Index(doc, "ohmyposh.dev")
	textStart := strings.LastIndex(doc[:watermarkIdx], "<text")
	assert.Contains(t, doc[textStart:watermarkIdx], `font-weight="bold"`)

	// The cursor paints no rect of its own: exactly one <rect> for "hi"'s
	// own background, plus the window's border rect and the shadow
	// silhouette (see writeWindowChrome), none for the cursor.
	assert.Equal(t, 3, strings.Count(doc, "<rect"))
}

// TestEncodeEmptyRowsStillGetsChromeAndDecoration pins that even a
// completely empty capture still renders the full window, a cursor, and
// the watermark - a valid "empty prompt" render rather than a degenerate
// one.
func TestEncodeEmptyRowsStillGetsChromeAndDecoration(t *testing.T) {
	doc := Encode(nil, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, `class="omp-window"`)
	assert.Contains(t, doc, ">_<")
	assert.Contains(t, doc, ">ohmyposh.dev<")
}

// TestNewWindowGeometryScalesWithFontSize pins the FontSize/48 scaling the
// task brief specifies: at FontSize 48, the window retains the reference
// padding/titleOffset/corner ratios while the border stays at the website
// code blocks' 1px (see minStrokeWidth) rather than the retired PNG
// renderer's thicker 2px line.
func TestNewWindowGeometryScalesWithFontSize(t *testing.T) {
	opts := Options{FontSize: 48}
	geo := newWindowGeometry(&opts)

	assert.InDelta(t, 48, geo.padding, 0.001)
	assert.InDelta(t, 80, geo.titleOffset, 0.001)
	assert.InDelta(t, 12, geo.corner, 0.001)
	assert.InDelta(t, 1, geo.strokeWidth, 0.001)
	assert.InDelta(t, 80*0.55, geo.controlSize, 0.001)
	assert.InDelta(t, 80*1.15, geo.controlGap, 0.001)

	// Half the reference FontSize halves every one of those constants too.
	half := Options{FontSize: 24}
	geoHalf := newWindowGeometry(&half)
	assert.InDelta(t, geo.padding/2, geoHalf.padding, 0.001)
	assert.InDelta(t, geo.titleOffset/2, geoHalf.titleOffset, 0.001)
	assert.InDelta(t, geo.controlSize/2, geoHalf.controlSize, 0.001)
	assert.InDelta(t, geo.controlGap/2, geoHalf.controlGap, 0.001)
}

// TestNewCanvasSizeHasNoOuterMargin pins that the window sits flush at the
// canvas origin with no transparent gutter on the top or left, and that the
// canvas grows by exactly shadowOffset on the right and bottom so the hard
// offset silhouette (see shadowOffset) is never clipped by the viewBox.
func TestNewCanvasSizeHasNoOuterMargin(t *testing.T) {
	opts := Options{FontSize: 48}
	geo := newWindowGeometry(&opts)
	size := newCanvasSize(&geo, 10, 28, 2, 60)

	assert.Zero(t, size.windowX)
	assert.Zero(t, size.windowY)
	assert.InDelta(t, size.windowWidth+shadowOffset, size.width, 0.001)
	assert.InDelta(t, size.windowHeight+shadowOffset, size.height, 0.001)
}
