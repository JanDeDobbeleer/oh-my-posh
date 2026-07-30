package svg

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
)

// windowGeometry is every coordinate the terminal-window chrome needs,
// derived once per Encode call from opts. padding, titleOffset, corner
// radius and the 2px window border trace back to the retired PNG
// renderer's own constants (image.go's initDefaults/SavePNG, at its factor
// = 2.0): padding 48, titleOffset 80, corner radius 12. Those numbers are
// what the PNG renderer drew at its own reference em size — a 48px
// FontSize, backed out from the reference renders' fixed 3648×632 canvas
// (120 columns × ~28px advance + 2×96 margin + 2×48 padding = 3648) — so
// every one of them is scaled here by FontSize/48 to keep the same
// proportions at this package's actual FontSize (16px by default) instead
// of reproducing the PNG renderer's pixel values verbatim.
//
// The PNG renderer's own margin 96 and its stackblur drop shadow are gone.
// The margin existed only to give that shadow somewhere to fall, and drawing
// both meant every export carried a transparent gutter — 32px a side at the
// default font size — that a caller then had to crop or lay out around. The
// canvas is the window now, edge to edge.
//
// The header's own controlSize/controlGap are new with the two-tone header
// (see writeWindowChrome): unlike the retired PNG renderer's fixed macOS
// traffic lights, they have no reference pixel value to preserve, so they
// are sized directly as fractions of titleOffset (the header's own height)
// instead of their own scale-derived constants.
type windowGeometry struct {
	padding     float64
	titleOffset float64 // the header bar's own height
	corner      float64
	strokeWidth float64

	controlSize float64 // font-size of the − ▢ × window-control glyphs
	controlGap  float64 // spacing between the three window controls
}

// minStrokeWidth is the floor writeWindowChrome's border stroke never
// scales below: below the default FontSize (16, scale ~0.33) the plain
// 2*scale formula drops under 1px - thin enough that anti-aliasing fades it
// to near-nothing, which is what left the window's rounded corners barely
// visible at every FontSize this package is actually used at (the website's
// theme gallery and Studio preview both render at the default). 1px is
// still crisp at FontSize 48's own scale=1 (2*1 > 1, so the clamp is a
// no-op there and TestNewWindowGeometryScalesWithFontSize's pinned value of
// 2 is untouched).
const minStrokeWidth = 1.0

func newWindowGeometry(opts *Options) windowGeometry {
	scale := opts.FontSize / 48.0

	geo := windowGeometry{
		padding:     48 * scale,
		titleOffset: 80 * scale,
		corner:      12 * scale,
		strokeWidth: max(2*scale, minStrokeWidth),
	}

	geo.controlSize = geo.titleOffset * 0.55
	geo.controlGap = geo.titleOffset * 1.15

	return geo
}

// canvasSize is every box the window chrome and content grid need, computed
// once from a row count and geometry: the window itself (the rounded rect
// with the title bar and content inside it) and the content grid's own
// top-left corner inside it. The window fills the canvas exactly, so the two
// share an origin; windowX/windowY are kept rather than folded away because
// every chrome coordinate is written relative to them.
type canvasSize struct {
	width, height             float64 // the full <svg> canvas
	windowX, windowY          float64
	windowWidth, windowHeight float64
	contentX, contentY        float64
}

func newCanvasSize(geo *windowGeometry, columns int, cellWidth float64, rows int, lineHeight float64) canvasSize {
	contentWidth := float64(columns) * cellWidth
	contentHeight := float64(rows) * lineHeight

	windowWidth := contentWidth + 2*geo.padding
	windowHeight := contentHeight + 2*geo.padding + geo.titleOffset

	return canvasSize{
		width:  windowWidth,
		height: windowHeight,

		windowX:      0,
		windowY:      0,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,

		contentX: geo.padding,
		contentY: geo.padding + geo.titleOffset,
	}
}

// writeWindowChrome draws the terminal window itself: a header bar across
// the top (the window's own rounded top corners) and the content pane below
// it (the rounded bottom corners), each its own fill, plus the border
// stroke and the "−"/"▢"/"×" minimize/maximize/close glyphs flush against
// the header's right edge. No title text and no "+"/"⌄" tab controls: this
// package tried a title/tab-strip row too (see the now-removed macOS
// traffic-light and Windows Terminal tab-control chrome this replaced), but
// a window this small never has real room for a title next to controls
// without either cramping into the edges or reading illegibly tiny, and a
// plain two-tone header - the header bar in one color, the content pane in
// another - is what every reference (Windows Terminal, Ghostty, macOS
// Terminal) has in common once a title is set aside.
//
// Two separate rounded-rect paths (rather than one rect plus a clip-path)
// draw the header/content split: this package has no <defs> or ids
// anywhere else, and a caller that inlines more than one Encode result
// directly into the same HTML document (rather than each behind its own
// <img>) would collide on a shared clipPath id. Two paths, each rounding
// only the two corners it owns (top corners for the header, bottom corners
// for the content pane) and square on the edge where they meet, need
// neither.
//
// Each element below is written via a single format string rather than
// interleaved WriteString calls — see writeRect's doc comment (svg.go) on
// why: this package renders once per prompt, so the allocation-averse style
// the interleaved calls used to be written in buys nothing here.
func writeWindowChrome(b *strings.Builder, size canvasSize, geo *windowGeometry, contentFill color.RGB) {
	x0, y0 := size.windowX, size.windowY
	w, h, r := size.windowWidth, size.windowHeight, geo.corner

	header := headerColor(contentFill)

	fmt.Fprintf(b, `<path class="omp-window-header" fill="%s" d="M %s,%s H %s A %s,%s 0 0 1 %s,%s `+
		`V %s H %s V %s A %s,%s 0 0 1 %s,%s Z"/>`+"\n",
		hexString(header),
		formatFloat(x0+r), formatFloat(y0), formatFloat(x0+w-r),
		formatFloat(r), formatFloat(r), formatFloat(x0+w), formatFloat(y0+r),
		formatFloat(y0+geo.titleOffset), formatFloat(x0), formatFloat(y0+r),
		formatFloat(r), formatFloat(r), formatFloat(x0+r), formatFloat(y0))

	fmt.Fprintf(b, `<path class="omp-window-content" fill="%s" d="M %s,%s H %s V %s A %s,%s 0 0 1 %s,%s `+
		`H %s A %s,%s 0 0 1 %s,%s Z"/>`+"\n",
		hexString(contentFill),
		formatFloat(x0), formatFloat(y0+geo.titleOffset), formatFloat(x0+w), formatFloat(y0+h-r),
		formatFloat(r), formatFloat(r), formatFloat(x0+w-r), formatFloat(y0+h),
		formatFloat(x0+r), formatFloat(r), formatFloat(r), formatFloat(x0), formatFloat(y0+h-r))

	// Inset by half the stroke width on every side: a rect stroked exactly on
	// the canvas edge (x=0,y=0,width=w,height=h) has its outer half clipped
	// by the SVG viewport (the root <svg>'s overflow is hidden by default),
	// leaving only a half-width, barely-visible line - most noticeable along
	// the bottom edge, where the content pane's own fill often matches the
	// page background behind it and there is nothing else to read as a
	// border. Insetting keeps the full stroke width inside the viewBox on
	// every edge instead of only the top/left ever reading as intended.
	half := geo.strokeWidth / 2

	fmt.Fprintf(b, `<rect class="omp-window" x="%s" y="%s" width="%s" height="%s" rx="%s" fill="none" `+
		`stroke="#404040" stroke-width="%s"/>`+"\n",
		formatFloat(x0+half), formatFloat(y0+half), formatFloat(w-geo.strokeWidth), formatFloat(h-geo.strokeWidth),
		formatFloat(r), formatFloat(geo.strokeWidth))

	barMidY := y0 + geo.titleOffset/2
	rightEdge := x0 + w

	writeWindowControls(b, geo, barMidY, rightEdge, controlColor(header))
}

// controlBaselineY centers a glyph of the given font-size on midY: a fixed
// fraction of font-size (rather than a per-caller ratio) below the visual
// center line, close enough to the actual cap-height midpoint for the thin
// glyphs this bar draws that every one of them reads as sitting on the same
// line.
func controlBaselineY(midY, fontSize float64) float64 {
	return midY + fontSize*0.35
}

// writeWindowControls draws the minimize/maximize/close glyphs left to
// right in both position and document order, so close always lands flush
// against the right edge regardless of how many controls precede it.
func writeWindowControls(b *strings.Builder, geo *windowGeometry, barMidY, rightEdge float64, fill color.RGB) {
	labels := [3]string{"−", "▢", "×"}
	fontSize := geo.controlSize

	for i, label := range labels {
		cx := rightEdge - geo.padding*1.0 - float64(len(labels)-1-i)*geo.controlGap
		fmt.Fprintf(b, `<text xml:space="preserve" class="omp-window-control" x="%s" y="%s" font-size="%spx" fill="%s" text-anchor="middle">%s</text>`+"\n",
			formatFloat(cx), formatFloat(controlBaselineY(barMidY, fontSize)), formatFloat(fontSize), hexString(fill), label)
	}
}

// controlColor picks the window-control glyph color that reads against the
// header bar itself (see luminance) rather than a fixed light gray: header
// is always a fixed +/-22 shade away from the content fill (see
// headerColor), so on a light theme's header a fixed light-gray glyph reads
// nearly invisible - the same contrast problem headerColor's own delta
// solves for the header/content split, applied here to the glyphs sitting
// on top of it.
func controlColor(header color.RGB) color.RGB {
	if luminance(header) < 128 {
		return color.RGB{R: 0xcf, G: 0xcf, B: 0xcf}
	}

	return defaultForegroundNearBlack
}

// headerColor gives the header bar a shade related to bg but visually
// distinct from it - lighter when bg is dark, darker when bg is light - so
// the two-tone split reads regardless of the caller's own terminal theme
// instead of assuming a dark background the way a single fixed header color
// would.
func headerColor(bg color.RGB) color.RGB {
	const delta = 22

	if luminance(bg) < 128 {
		return color.RGB{R: addClamped(bg.R, delta), G: addClamped(bg.G, delta), B: addClamped(bg.B, delta)}
	}

	return color.RGB{R: subClamped(bg.R, delta), G: subClamped(bg.G, delta), B: subClamped(bg.B, delta)}
}

func addClamped(c uint8, delta int) uint8 {
	v := int(c) + delta
	if v > 255 {
		return 255
	}

	return uint8(v)
}

func subClamped(c uint8, delta int) uint8 {
	v := int(c) - delta
	if v < 0 {
		return 0
	}

	return uint8(v)
}
