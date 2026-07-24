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

// gradientStops is the two-stop gradient used across these tests; red -> blue
// keeps the first/last stops trivially distinguishable in assertions.
const gradientStops = color.Ansi("linear-gradient(#FF0000, #0000FF)")

// newPowerlineSegment builds a fully-initialized Powerline-style segment (writer
// mapped, text rendered) so backgroundEdge's VisibleCells(segment.Text()) call has
// something real to measure instead of panicking on a nil writer.
func newPowerlineSegment(t *testing.T, engine *Engine, template string, background color.Ansi) *config.Segment {
	t.Helper()

	segment := &config.Segment{Type: "text", Template: template, Style: config.Powerline, Background: background}
	assert.NoError(t, segment.MapSegmentWithWriter(engine.Env))
	segment.Render(0, true)

	return segment
}

// TestGetPowerlineColorGradient covers engine.go's getPowerlineColor: the
// powerline separator symbol sits at the previous segment's right edge, so a
// gradient background must collapse to its last stop, never the first.
func TestGetPowerlineColorGradient(t *testing.T) {
	engine := New(&runtime.Flags{IsPrimary: true})

	previous := newPowerlineSegment(t, engine, "abc", gradientStops)
	active := newPowerlineSegment(t, engine, "def", "green")

	engine.previousActiveSegment = previous
	engine.activeSegment = active

	got := engine.getPowerlineColor()

	assert.Equal(t, gradientStops.GradientLast(), got, "powerline separator color must be the gradient's last stop")
	assert.NotEqual(t, gradientStops.GradientFirst(), got, "powerline separator color must not be the gradient's first stop")
}

// TestGetPowerlineColorSingleStopAutoShade covers the bug report behind the auto-shade
// fix: a single-stop gradient's powerline separator must render the same shaded color
// GradientCells renders the segment's last cell as, not the unshaded configured color -
// otherwise the separator visibly jumps back to a different shade right after the body.
func TestGetPowerlineColorSingleStopAutoShade(t *testing.T) {
	engine := New(&runtime.Flags{IsPrimary: true})

	singleStop := color.Ansi("dark-gradient(#3465A4)")
	previous := newPowerlineSegment(t, engine, "abc", singleStop)
	active := newPowerlineSegment(t, engine, "def", "green")

	engine.previousActiveSegment = previous
	engine.activeSegment = active

	got := engine.getPowerlineColor()

	assert.Equal(t, singleStop.GradientLastForCells(3), got, "powerline separator must be the auto-shaded last stop, sized to the segment's 3 visible cells")
	assert.NotEqual(t, singleStop.GradientFirst(), got, "powerline separator must not be the unshaded configured color")
}

// TestGetPowerlineColorPaletteReferencedStop pins the fix for a palette reference used
// as an individual gradient STOP rather than the whole gradient value:
// dark-gradient(p:teal) must resolve p:teal before GradientLast shades it, so the
// powerline separator matches the auto-shaded color GradientCells renders the segment's
// last cell as, instead of the raw "p:teal" text GradientLast previously couldn't shade
// without a resolver (it fell through unshaded to the bright, unmodified base color).
func TestGetPowerlineColorPaletteReferencedStop(t *testing.T) {
	engine := New(&runtime.Flags{IsPrimary: true})
	terminal.String()

	palette := color.Palette{"teal": "#179299"}
	origColors := terminal.Colors
	terminal.Colors = color.MakeColors(palette, false, "", engine.Env)
	t.Cleanup(func() { terminal.Colors = origColors })

	previous := newPowerlineSegment(t, engine, "abc", "dark-gradient(p:teal)")
	active := newPowerlineSegment(t, engine, "def", "green")

	engine.previousActiveSegment = previous
	engine.activeSegment = active

	got := engine.getPowerlineColor()

	assert.Equal(t, color.Ansi("#179299").GradientFirst(), color.Ansi("#179299"), "sanity: base color unchanged by GradientFirst")
	assert.NotEqual(t, color.Ansi("#179299"), got, "the separator must not render the raw, unshaded base color")

	unresolvedLast := color.Ansi("dark-gradient(p:teal)").GradientLastForCells(3)
	assert.NotEqual(t, unresolvedLast, got, "GradientLastForCells on the raw p:teal stop can't shade without a resolver; the engine must resolve the palette reference first")

	resolvedLast := color.Ansi("dark-gradient(#179299)").GradientLastForCells(3)
	assert.Equal(t, resolvedLast, got, "the separator must match the auto-shaded color GradientCells renders the last body cell as, sized to the segment's 3 visible cells")

	// the resolved gradient must keep its dark-gradient prefix, not silently become a
	// plain linear-gradient: unlike dark-gradient, a single-stop linear-gradient is not
	// auto-shaded, so its GradientLast is the raw, unshaded base color.
	assert.Equal(t, color.Ansi("#179299"), color.Ansi("linear-gradient(#179299)").GradientLast(), "sanity: a single-stop linear-gradient is not auto-shaded")
	assert.NotEqual(t, color.Ansi("#179299"), got, "the resolved gradient must have kept dark-gradient's auto-shade semantics")
}

// TestWriteSeparatorTrailingDiamondGradient covers writeSeparator's final
// trailing-diamond branch: the glyph sits at the segment's right edge, so a
// gradient background must render as its last stop rather than the writer's
// default cells==1 (first stop) behavior for a single-cell glyph.
func TestWriteSeparatorTrailingDiamondGradient(t *testing.T) {
	render := func(background color.Ansi) string {
		engine := New(&runtime.Flags{IsPrimary: true})
		terminal.String() // drain any state left over from a previous case
		terminal.ParentColors = nil

		segment := &config.Segment{
			Type:            "text",
			Template:        "X",
			Style:           config.Diamond,
			TrailingDiamond: "",
			Foreground:      "white",
			Background:      background,
		}

		assert.NoError(t, segment.MapSegmentWithWriter(engine.Env))
		segment.Render(0, true)

		engine.setActiveSegment(segment)
		engine.writeSeparator(true)

		out, _ := terminal.String()
		return out
	}

	gradientOut := render(gradientStops)
	lastStopOut := render(gradientStops.GradientLast())
	firstStopOut := render(gradientStops.GradientFirst())

	assert.Equal(t, lastStopOut, gradientOut, "trailing diamond must render the gradient's last stop")
	assert.NotEqual(t, firstStopOut, gradientOut, "trailing diamond must not render the gradient's first stop")
}

// TestRenderActiveSegmentDiamondPreviousGradient covers renderActiveSegment's
// diamond branch: when the previous segment has no trailing diamond, the next
// segment's leading diamond is drawn against the previous segment's
// background - its right edge - so a gradient there must collapse to its
// last stop.
func TestRenderActiveSegmentDiamondPreviousGradient(t *testing.T) {
	const leadingGlyph = ""

	leadingDiamondOutput := func(previousBackground color.Ansi) string {
		engine := New(&runtime.Flags{IsPrimary: true})
		terminal.String()

		// TrailingDiamond intentionally left empty so HasEmptyDiamondAtEnd() is true.
		previous := &config.Segment{
			Type:       "text",
			Template:   "P",
			Style:      config.Diamond,
			Foreground: "white",
			Background: previousBackground,
		}
		assert.NoError(t, previous.MapSegmentWithWriter(engine.Env))
		previous.Render(0, true)

		engine.setActiveSegment(previous)
		engine.renderActiveSegment()
		terminal.String() // drain the previous segment's own output

		next := &config.Segment{
			Type:           "text",
			Template:       "N",
			Style:          config.Diamond,
			LeadingDiamond: leadingGlyph,
			Foreground:     "white",
			Background:     "green",
		}
		assert.NoError(t, next.MapSegmentWithWriter(engine.Env))
		next.Render(1, true)

		engine.setActiveSegment(next)
		engine.renderActiveSegment()

		out, _ := terminal.String()

		idx := strings.Index(out, leadingGlyph)
		if !assert.GreaterOrEqual(t, idx, 0, "leading diamond glyph not found in output") {
			return out
		}

		return out[:idx+len(leadingGlyph)]
	}

	gradientOut := leadingDiamondOutput(gradientStops)
	lastStopOut := leadingDiamondOutput(gradientStops.GradientLast())
	firstStopOut := leadingDiamondOutput(gradientStops.GradientFirst())

	assert.Equal(t, lastStopOut, gradientOut, "next segment's leading diamond must pick up the previous gradient's last stop")
	assert.NotEqual(t, firstStopOut, gradientOut, "next segment's leading diamond must not pick up the previous gradient's first stop")
}

// TestParentBackgroundKeywordCollapsesGradientLastStop is an integration-level
// check that a following segment's `parentBackground` keyword resolves a
// gradient previous segment to its last stop. The collapse itself lives in
// color/keywords.go (T1); this exercises it through the engine's actual
// SetParentColors wiring (engine.go's renderActiveSegment) rather than at the
// color package unit level.
func TestParentBackgroundKeywordCollapsesGradientLastStop(t *testing.T) {
	render := func(previousBackground color.Ansi) string {
		engine := New(&runtime.Flags{IsPrimary: true})
		terminal.String()

		previous := &config.Segment{
			Type:       "text",
			Template:   "P",
			Style:      config.Powerline,
			Foreground: "white",
			Background: previousBackground,
		}
		assert.NoError(t, previous.MapSegmentWithWriter(engine.Env))
		previous.Render(0, true)

		engine.setActiveSegment(previous)
		engine.renderActiveSegment()
		terminal.String()

		next := &config.Segment{
			Type:       "text",
			Template:   "N",
			Style:      config.Powerline,
			Foreground: "white",
			Background: "parentBackground",
		}
		assert.NoError(t, next.MapSegmentWithWriter(engine.Env))
		next.Render(1, true)

		engine.setActiveSegment(next)
		engine.renderActiveSegment()

		out, _ := terminal.String()
		return out
	}

	gradientOut := render(gradientStops)
	lastStopOut := render(gradientStops.GradientLast())

	assert.Equal(t, lastStopOut, gradientOut, "parentBackground must collapse a gradient background to its last stop")
}
