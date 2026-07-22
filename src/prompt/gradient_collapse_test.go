package prompt

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
)

// threeStopGradient is used by the width tests below: a 3-stop gradient needs
// 2 * 3 = 6 visible cells before it renders per cell (see minGradientCellsPerStop
// in engine.go / Amendment 3 in the gradient spec).
const threeStopGradient = color.Ansi("linear-gradient(#FF0000, #00FF00, #0000FF)")

// renderPowerlineSegment builds a single Powerline-style segment with the given
// background and template text, runs it through the same setActiveSegment/
// renderActiveSegment path the engine uses, and returns the rendered output.
func renderPowerlineSegment(t *testing.T, background color.Ansi, template string) string {
	t.Helper()

	engine := New(&runtime.Flags{IsPrimary: true})
	terminal.String() // drain any state left over from a previous case
	terminal.ParentColors = nil

	segment := &config.Segment{
		Type:       "text",
		Template:   template,
		Style:      config.Powerline,
		Foreground: "white",
		Background: background,
	}
	assert.NoError(t, segment.MapSegmentWithWriter(engine.Env))
	segment.Render(0, true)

	engine.setActiveSegment(segment)
	engine.renderActiveSegment()

	out, _ := terminal.String()
	return out
}

// TestSetActiveSegmentCollapsesNarrowGradient covers Amendment 3: a segment
// narrower than minGradientCellsPerStop*stops must render byte-identical to the
// same segment configured with a plain background equal to the gradient's last
// stop - no per-cell escapes anywhere, including caps/separators.
func TestSetActiveSegmentCollapsesNarrowGradient(t *testing.T) {
	// "abc" is 3 visible cells; a 2-stop gradient needs >= 4 to render per cell.
	gradientOut := renderPowerlineSegment(t, gradientStops, "abc")
	solidOut := renderPowerlineSegment(t, gradientStops.GradientLast(), "abc")

	assert.Equal(t, solidOut, gradientOut, "a narrow gradient segment must render byte-identical to its last stop configured as a plain background")
	assert.NotContains(t, gradientOut, "linear-gradient", "the raw gradient string must never reach the output")

	// exactly one background truecolor escape: the collapsed solid color, not one per cell.
	assert.Equal(t, 1, strings.Count(gradientOut, "\x1b[48;2;"))
}

// TestSetActiveSegmentKeepsWideGradientPerCell is TestSetActiveSegmentCollapsesNarrowGradient's
// counterpart: a segment at or above the minimum width still renders one interpolated
// color per visible cell.
func TestSetActiveSegmentKeepsWideGradientPerCell(t *testing.T) {
	// "abcd" is 4 visible cells, exactly meeting a 2-stop gradient's minimum (2 * 2).
	gradientOut := renderPowerlineSegment(t, gradientStops, "abcd")
	solidOut := renderPowerlineSegment(t, gradientStops.GradientLast(), "abcd")

	assert.NotEqual(t, solidOut, gradientOut, "a wide-enough gradient segment must not collapse to a single solid color")

	resolver := &color.Defaults{}
	cells := color.GradientCells(gradientStops, 4, resolver, true, nil, nil)

	for _, cell := range cells {
		assert.Contains(t, gradientOut, "\x1b["+cell.String()+"m")
	}

	assert.Equal(t, 4, strings.Count(gradientOut, "\x1b[48;2;"), "each of the 4 cells must stamp its own background escape")
}

// TestSetActiveSegmentThreeStopGradientNeedsSixCells covers the per-stop scaling of
// the threshold: a 3-stop gradient collapses below 6 cells and renders per cell at
// or above it.
func TestSetActiveSegmentThreeStopGradientNeedsSixCells(t *testing.T) {
	// 5 visible cells, one short of 2 * 3 stops: must collapse.
	narrowGradientOut := renderPowerlineSegment(t, threeStopGradient, "abcde")
	narrowSolidOut := renderPowerlineSegment(t, threeStopGradient.GradientLast(), "abcde")
	assert.Equal(t, narrowSolidOut, narrowGradientOut, "a 3-stop gradient under 6 cells must collapse to its last stop")
	assert.Equal(t, 1, strings.Count(narrowGradientOut, "\x1b[48;2;"))

	// 6 visible cells, exactly meeting 2 * 3 stops: must render per cell.
	wideGradientOut := renderPowerlineSegment(t, threeStopGradient, "abcdef")
	wideSolidOut := renderPowerlineSegment(t, threeStopGradient.GradientLast(), "abcdef")
	assert.NotEqual(t, wideSolidOut, wideGradientOut, "a 3-stop gradient at 6 cells must not collapse")
	assert.Equal(t, 6, strings.Count(wideGradientOut, "\x1b[48;2;"))
}

// TestPendingSegmentKeepsGradientPlaceholder covers streaming placeholders: a
// pending segment's "..." text is narrower than the collapse threshold, but the
// placeholder must still render the gradient per cell - it previews the segment's
// final look instead of flashing a collapsed solid block mid-stream.
func TestPendingSegmentKeepsGradientPlaceholder(t *testing.T) {
	engine := New(&runtime.Flags{IsPrimary: true})
	terminal.String()

	pending := &config.Segment{
		Type:       "text",
		Template:   " never used ",
		Style:      config.Powerline,
		Foreground: "white",
		Background: gradientStops,
	}
	assert.NoError(t, pending.MapSegmentWithWriter(engine.Env))
	pending.Pending = true
	pending.Render(0, false)

	engine.setActiveSegment(pending)
	engine.renderActiveSegment()

	out, _ := terminal.String()

	// "..." is 3 cells; without the pending exemption this would collapse to a
	// single solid background escape.
	assert.Equal(t, 3, strings.Count(out, "\x1b[48;2;"), "the placeholder must stamp one background escape per cell")
	assert.NotContains(t, out, "linear-gradient")
}

// TestPaletteReferencedGradientCollapses pins the review fix for palette-indirected
// gradients: a segment background of "p:grad" must be palette-resolved BEFORE the
// engine's gradient handling, so the width collapse and edge consumers behave
// identically to the same gradient written inline.
func TestPaletteReferencedGradientCollapses(t *testing.T) {
	engine := New(&runtime.Flags{IsPrimary: true})
	terminal.String()
	terminal.ParentColors = nil

	palette := color.Palette{"grad": "linear-gradient(#FF0000, #0000FF)"}
	origColors := terminal.Colors
	terminal.Colors = color.MakeColors(palette, false, "", engine.Env)
	t.Cleanup(func() { terminal.Colors = origColors })

	segment := &config.Segment{
		Type:       "text",
		Template:   "abc", // 3 cells: below the 2-stop threshold, must collapse
		Style:      config.Powerline,
		Foreground: "white",
		Background: "p:grad",
	}
	assert.NoError(t, segment.MapSegmentWithWriter(engine.Env))
	segment.Render(0, true)

	engine.setActiveSegment(segment)
	engine.renderActiveSegment()

	out, _ := terminal.String()

	assert.Equal(t, 1, strings.Count(out, "\x1b[48;2;"), "a narrow palette-referenced gradient must collapse to one solid background")
	assert.Contains(t, out, "48;2;0;0;255", "the collapsed color must be the gradient's last stop")
	assert.NotContains(t, out, "linear-gradient")
}
