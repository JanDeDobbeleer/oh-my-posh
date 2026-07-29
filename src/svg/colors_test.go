package svg

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"

	"github.com/stretchr/testify/assert"
)

// TestResolveChannel covers every form Run.BackgroundSource/ForegroundSource
// can hold (see terminal.Run's doc comment): a #RRGGBB hex, an ansiColorCodes
// name, "accent" (both resolved and unresolved), a plain numeric 256-palette
// index (in range and out of range), "transparent", and empty — the one that
// most often gets missed, since it means "keep whatever was already active"
// rather than "no color", and must report ok=false to say so.
func TestResolveChannel(t *testing.T) {
	accentFg := &color.RGB{R: 10, G: 20, B: 30}
	accentBg := &color.RGB{R: 40, G: 50, B: 60}
	termBg := &color.RGB{R: 5, G: 5, B: 5}

	opts := Options{
		TerminalBackground: termBg,
		AccentForeground:   accentFg,
		AccentBackground:   accentBg,
	}

	cases := []struct {
		Expected     *color.RGB
		GradientRGB  *color.RGB
		Case         string
		Source       color.Ansi
		IsBackground bool
		ExpectedOK   bool
	}{
		{Case: "hex", Source: "#FF00AA", Expected: &color.RGB{R: 0xFF, G: 0x00, B: 0xAA}, ExpectedOK: true},
		{Case: "hex lowercase", Source: "#00ff00", Expected: &color.RGB{R: 0, G: 255, B: 0}, ExpectedOK: true},
		// gookit's color.HexToRgb — the parser the ANSI writer ultimately
		// reaches through color.HEX — accepts a 3-digit shorthand, an 8-digit
		// 0x-prefixed value, and is case-insensitive; hexToRGB has to accept
		// exactly the same set, or a color the terminal painted is dropped
		// from the SVG. Six bundled themes write #fff/#FFF/#000.
		{Case: "hex shorthand", Source: "#FFF", Expected: &color.RGB{R: 0xFF, G: 0xFF, B: 0xFF}, ExpectedOK: true},
		{Case: "hex shorthand lowercase", Source: "#0a8", Expected: &color.RGB{R: 0x00, G: 0xAA, B: 0x88}, ExpectedOK: true},
		{Case: "hex 0x prefixed", Source: "#0xad99c0", Expected: &color.RGB{R: 0xAD, G: 0x99, B: 0xC0}, ExpectedOK: true},
		{Case: "malformed hex non-numeric", Source: "#GGGGGG", ExpectedOK: false},
		{Case: "malformed hex four digits", Source: "#FFFF", ExpectedOK: false},
		{Case: "named color", Source: "red", Expected: &color.RGB{R: 222, G: 56, B: 43}, ExpectedOK: true},
		{Case: "named color unknown", Source: "chartreuse", ExpectedOK: false},
		{
			Case: "default background uses terminal background", Source: "default", IsBackground: true,
			Expected: termBg, ExpectedOK: true,
		},
		{
			Case: "default background with no terminal background known", Source: "default", IsBackground: true,
			Expected: nil, ExpectedOK: true,
		},
		{Case: "default foreground uses fixed fallback", Source: "default", Expected: &defaultForegroundRGB, ExpectedOK: true},
		{Case: "accent foreground resolved", Source: color.Accent, Expected: accentFg, ExpectedOK: true},
		{Case: "accent background resolved", Source: color.Accent, IsBackground: true, Expected: accentBg, ExpectedOK: true},
		{Case: "numeric palette index", Source: "196", Expected: &color.RGB{R: 255, G: 0, B: 0}, ExpectedOK: true},
		{Case: "numeric palette index 0", Source: "0", Expected: &color.RGB{R: 1, G: 1, B: 1}, ExpectedOK: true},
		{Case: "numeric palette index out of range", Source: "256", ExpectedOK: false},
		{Case: "transparent resolves to no fill", Source: color.Transparent, Expected: nil, ExpectedOK: true},
		{Case: "empty keeps previous", Source: "", Expected: nil, ExpectedOK: false},
		{
			Case: "gradient RGB wins outright", Source: "", GradientRGB: &color.RGB{R: 9, G: 9, B: 9},
			Expected: &color.RGB{R: 9, G: 9, B: 9}, ExpectedOK: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			runOpts := opts
			if tc.Case == "default background with no terminal background known" {
				runOpts.TerminalBackground = nil
			}

			rgb, ok := resolveChannel(tc.Source, tc.GradientRGB, tc.IsBackground, &runOpts)

			assert.Equal(t, tc.ExpectedOK, ok)
			assert.Equal(t, tc.Expected, rgb)
		})
	}
}

// TestResolveChannelAccentUnresolvedKeepsPrevious pins the one accent edge
// case a table entry can't express cleanly: a nil Options.Accent* — the
// real-world state when SetAccentColor never resolved one (see
// color.Defaults.SetAccentColor) — must behave exactly like an empty source,
// not like a resolved "no color".
func TestResolveChannelAccentUnresolvedKeepsPrevious(t *testing.T) {
	rgb, ok := resolveChannel(color.Accent, nil, false, &Options{})

	assert.False(t, ok)
	assert.Nil(t, rgb)
}

func TestPalette256(t *testing.T) {
	cases := []struct {
		Case     string
		Expected color.RGB
		Index    uint8
	}{
		{Case: "standard black", Index: 0, Expected: color.RGB{R: 1, G: 1, B: 1}},
		{Case: "bright white", Index: 15, Expected: color.RGB{R: 255, G: 255, B: 255}},
		{Case: "cube first entry", Index: 16, Expected: color.RGB{R: 0, G: 0, B: 0}},
		{Case: "cube red", Index: 196, Expected: color.RGB{R: 255, G: 0, B: 0}},
		{Case: "cube last entry", Index: 231, Expected: color.RGB{R: 255, G: 255, B: 255}},
		{Case: "grayscale first", Index: 232, Expected: color.RGB{R: 8, G: 8, B: 8}},
		{Case: "grayscale last", Index: 255, Expected: color.RGB{R: 238, G: 238, B: 238}},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			assert.Equal(t, tc.Expected, palette256(tc.Index))
		})
	}
}

func TestParseTrueColorSGR(t *testing.T) {
	cases := []struct {
		Expected   *color.RGB
		Case       string
		Input      color.Ansi
		ExpectedOK bool
	}{
		{Case: "foreground truecolor", Input: "38;2;18;52;86", Expected: &color.RGB{R: 18, G: 52, B: 86}, ExpectedOK: true},
		{Case: "background truecolor", Input: "48;2;1;2;3", Expected: &color.RGB{R: 1, G: 2, B: 3}, ExpectedOK: true},
		{Case: "256 color code is not truecolor", Input: "38;5;196", ExpectedOK: false},
		{Case: "16 color code is not truecolor", Input: "31", ExpectedOK: false},
		{Case: "empty", Input: "", ExpectedOK: false},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			rgb, ok := ParseTrueColorSGR(tc.Input)

			assert.Equal(t, tc.ExpectedOK, ok)
			assert.Equal(t, tc.Expected, rgb)
		})
	}
}
