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
// the rounded window rect (filled and stroked) and the three traffic lights
// in their fixed colors and left-to-right order. No drop shadow, and so no
// filter to reference - see windowGeometry's doc comment.
func TestEncodeWindowChromeElements(t *testing.T) {
	rows := [][]terminal.Run{{contentRun("hi")}}

	doc := Encode(rows, testOptions())
	decodeXML(t, doc)

	assert.Contains(t, doc, `class="omp-window"`)
	assert.Contains(t, doc, `rx="`)
	assert.Contains(t, doc, `stroke="#404040"`)
	assert.NotContains(t, doc, "<feDropShadow")
	assert.NotContains(t, doc, "filter=")

	redIdx := strings.Index(doc, "#ED655A")
	yellowIdx := strings.Index(doc, "#E1C04C")
	greenIdx := strings.Index(doc, "#71BD47")
	require.NotEqual(t, -1, redIdx)
	require.NotEqual(t, -1, yellowIdx)
	require.NotEqual(t, -1, greenIdx)
	assert.Less(t, redIdx, yellowIdx, "traffic lights must appear red, yellow, green left to right")
	assert.Less(t, yellowIdx, greenIdx, "traffic lights must appear red, yellow, green left to right")

	assert.Equal(t, 3, strings.Count(doc, `class="omp-traffic-light"`))
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

	assert.Contains(t, doc, `class="omp-window" x="`)
	assert.Contains(t, doc, `fill="`+hexString(custom)+`"`)
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

	assert.Contains(t, doc, `class="omp-window" x="`)
	assert.Contains(t, doc, `fill="`+hexString(termBg)+`"`)
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
	// own background, plus the window's, none for the cursor.
	assert.Equal(t, 2, strings.Count(doc, "<rect"))
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
// task brief specifies: at FontSize 48 (the retired PNG renderer's own
// reference em), every geometry constant must equal the PNG renderer's own
// pixel value verbatim.
func TestNewWindowGeometryScalesWithFontSize(t *testing.T) {
	opts := Options{FontSize: 48}
	geo := newWindowGeometry(&opts)

	assert.InDelta(t, 48, geo.padding, 0.001)
	assert.InDelta(t, 80, geo.titleOffset, 0.001)
	assert.InDelta(t, 12, geo.corner, 0.001)
	assert.InDelta(t, 18, geo.trafficR, 0.001)
	assert.InDelta(t, 50, geo.trafficGap, 0.001)
	assert.InDelta(t, 2, geo.strokeWidth, 0.001)

	// Half the reference FontSize halves every one of those constants too.
	half := Options{FontSize: 24}
	geoHalf := newWindowGeometry(&half)
	assert.InDelta(t, geo.padding/2, geoHalf.padding, 0.001)
	assert.InDelta(t, geo.titleOffset/2, geoHalf.titleOffset, 0.001)
}

// TestNewCanvasSizeHasNoOuterMargin pins that the window fills the canvas
// exactly: no transparent gutter on any side, which is what a caller
// embedding the SVG gets to lay out against.
func TestNewCanvasSizeHasNoOuterMargin(t *testing.T) {
	opts := Options{FontSize: 48}
	geo := newWindowGeometry(&opts)
	size := newCanvasSize(&geo, 10, 28, 2, 60)

	assert.Zero(t, size.windowX)
	assert.Zero(t, size.windowY)
	assert.InDelta(t, size.width, size.windowWidth, 0.001)
	assert.InDelta(t, size.height, size.windowHeight, 0.001)
}
