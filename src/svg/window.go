package svg

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
)

// windowGeometry is every coordinate the terminal-window chrome needs,
// derived once per Encode call from opts. All of it traces back to the
// retired PNG renderer's own constants (image.go's initDefaults/SavePNG, at
// its factor = 2.0): padding 48, titleOffset 80, corner radius 12,
// traffic-light radius 18 spaced 50 apart, and a 2px window border. Those
// numbers are what the PNG renderer drew at its own reference em size — a
// 48px FontSize, backed out from the reference renders' fixed 3648×632
// canvas (120 columns × ~28px advance + 2×96 margin + 2×48 padding = 3648) —
// so every one of them is scaled here by FontSize/48 to keep the same
// proportions at this package's actual FontSize (16px by default) instead of
// reproducing the PNG renderer's pixel values verbatim.
//
// The PNG renderer's own margin 96 and its stackblur drop shadow are gone.
// The margin existed only to give that shadow somewhere to fall, and drawing
// both meant every export carried a transparent gutter — 32px a side at the
// default font size — that a caller then had to crop or lay out around. The
// canvas is the window now, edge to edge.
type windowGeometry struct {
	padding      float64
	titleOffset  float64
	corner       float64
	trafficR     float64
	trafficGap   float64
	trafficNudge float64
	strokeWidth  float64
}

func newWindowGeometry(opts *Options) windowGeometry {
	scale := opts.FontSize / 48.0

	return windowGeometry{
		padding:      48 * scale,
		titleOffset:  80 * scale,
		corner:       12 * scale,
		trafficR:     18 * scale,
		trafficGap:   50 * scale,
		trafficNudge: 4 * scale,
		strokeWidth:  2 * scale,
	}
}

// canvasSize is every box the window chrome and content grid need, computed
// once from a row count and geometry: the window itself (the rounded rect
// with the traffic lights and content inside it) and the content grid's own
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

// trafficLightColors are the three window-control dots, left to right, in
// the retired PNG renderer's own fixed colors (image.go's red/yellow/green
// constants) — never resolved against a theme, since a real terminal
// emulator's own window chrome isn't themed by the shell prompt running
// inside it either.
var trafficLightColors = [3]string{"#ED655A", "#E1C04C", "#71BD47"}

// writeWindowChrome draws the terminal window itself: one rounded <rect>
// filling the canvas, plus three traffic-light <circle>s to complete the
// impression of a window with controls, matching the reference PNG renders'
// layout. No drop shadow — see windowGeometry's doc comment.
//
// fill is the window's own background: always opts.CanvasBackground, which
// withDefaults has already set to the real terminal background whenever
// that's known — see Options.CanvasBackground's doc comment on why a
// cutout glyph's color and the window's fill must always be the same
// value.
// Each element below is written via a single format string rather than
// interleaved WriteString calls — see writeRect's doc comment (svg.go) on
// why: this package renders once per prompt, so the allocation-averse style
// the interleaved calls used to be written in buys nothing here.
func writeWindowChrome(b *strings.Builder, size canvasSize, geo *windowGeometry, fill color.RGB) {
	fmt.Fprintf(b, `<rect class="omp-window" x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" `+
		`stroke="#404040" stroke-width="%s"/>`+"\n",
		formatFloat(size.windowX), formatFloat(size.windowY), formatFloat(size.windowWidth), formatFloat(size.windowHeight),
		formatFloat(geo.corner), hexString(fill), formatFloat(geo.strokeWidth))

	cy := size.windowY + geo.padding + geo.trafficNudge

	for i, hex := range trafficLightColors {
		cx := size.windowX + geo.padding + float64(i)*geo.trafficGap + geo.trafficNudge

		fmt.Fprintf(b, `<circle class="omp-traffic-light" cx="%s" cy="%s" r="%s" fill="%s"/>`+"\n",
			formatFloat(cx), formatFloat(cy), formatFloat(geo.trafficR), hex)
	}
}
