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

// TestGetPowerlineColorGradient covers engine.go's getPowerlineColor: the
// powerline separator symbol sits at the previous segment's right edge, so a
// gradient background must collapse to its last stop, never the first.
func TestGetPowerlineColorGradient(t *testing.T) {
	previous := &config.Segment{Type: "text", Style: config.Powerline, Background: gradientStops}
	active := &config.Segment{Type: "text", Style: config.Powerline, Background: "green"}

	engine := &Engine{previousActiveSegment: previous, activeSegment: active}

	got := engine.getPowerlineColor()

	assert.Equal(t, gradientStops.GradientLast(), got, "powerline separator color must be the gradient's last stop")
	assert.NotEqual(t, gradientStops.GradientFirst(), got, "powerline separator color must not be the gradient's first stop")
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
