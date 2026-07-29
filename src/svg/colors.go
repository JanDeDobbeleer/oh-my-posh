package svg

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"

	gookit "github.com/gookit/color"
)

// resolveChannel resolves one channel (background or foreground) of a Run to
// an RGB value, given the run's own captured RGB (set only for a gradient
// cell — see terminal.Run's doc comment, populated from color.GradientCellsRGB
// rather than inverted from an SGR escape) and the run's pre-ToAnsi source
// form (a #RRGGBB hex, an ansiColorCodes name, "accent", a 0-255 palette
// index, "transparent", or empty).
//
// The bool return reports whether the source actually resolved to something.
// False means color.Defaults.ToAnsi's failure path: nothing was emitted (an
// unresolved palette key, a malformed hex, ...), so the caller must keep
// whatever color was already active on that channel, exactly like a real
// terminal that never received an SGR code for it. True with a nil RGB means
// the source resolved to "paint nothing here" (transparent, or an unset
// accent/terminal background) — also a real resolution, just not a color,
// and it must override whatever was active before it.
func resolveChannel(source color.Ansi, rgb *color.RGB, isBackground bool, opts *Options) (*color.RGB, bool) {
	if rgb != nil {
		return rgb, true
	}

	if source.IsEmpty() {
		return nil, false
	}

	if source.IsTransparent() {
		return nil, true
	}

	if source == color.Accent {
		return resolveAccent(isBackground, opts)
	}

	value := source.String()

	if strings.HasPrefix(value, "#") {
		return hexToRGB(value)
	}

	if rgb, ok := namedRGB(value, isBackground, opts); ok {
		return rgb, true
	}

	if index, err := strconv.ParseUint(value, 10, 8); err == nil {
		rgb := palette256(uint8(index))
		return &rgb, true
	}

	return nil, false
}

// resolveAccent resolves the "accent" keyword against Options' pre-resolved
// accent RGB. An unset accent (opts.Accent* nil) mirrors color.Defaults.ToAnsi
// returning its own empty-string failure value for an unresolved accent: no
// SGR code would be emitted, so the previous color must stay active.
func resolveAccent(isBackground bool, opts *Options) (*color.RGB, bool) {
	accent := opts.AccentForeground
	if isBackground {
		accent = opts.AccentBackground
	}

	if accent == nil {
		return nil, false
	}

	return accent, true
}

// ResolveStaticRGB resolves a single, non-gradient color source — such as
// terminal.BackgroundColor, resolved once up front rather than carried per
// Run — to RGB. "accent"/"default" resolve against opts exactly like a Run's
// source would through Encode.
func ResolveStaticRGB(source color.Ansi, isBackground bool, opts *Options) (*color.RGB, bool) {
	return resolveChannel(source, nil, isBackground, opts)
}

// ParseTrueColorSGR parses a 24-bit SGR payload ("38;2;r;g;b" or
// "48;2;r;g;b") into RGB, via color.ParseTrueColorRGB — the same byte-level
// parse color/gradient.go's own colorful.Color-returning parseTrueColor
// builds on, so the two packages can't drift out of parity with each other.
// It exists for the one color a Run's source form can't carry as a hex
// string: "accent" is a keyword whose actual RGB only exists behind
// color.Defaults' private accent field, reachable through
// color.String.ToAnsi(color.Accent, ...) — which always answers in full
// truecolor for an accent regardless of color.TrueColor (see
// color.Defaults.SetAccentColor building it via gookit's color.RGB) — so a
// caller resolves it once through that and hands the result here, rather
// than Options resolving "accent" itself from package state.
func ParseTrueColorSGR(ansi color.Ansi) (*color.RGB, bool) {
	rgb, ok := color.ParseTrueColorRGB(ansi)
	if !ok {
		return nil, false
	}

	return &rgb, true
}

// hexToRGB parses the same `#`-prefixed color color.Defaults.ToAnsi ultimately
// hands to gookit (see color/colors.go's `color.HEX(colorString,
// isBackground)` call) by calling gookit's own color.HexToRgb directly,
// rather than a transcription of its rules kept in sync by hand: commit
// 86de20dd exists because that transcription had drifted — accepting only
// the canonical #RRGGBB — and a color this parser rejects is not "no color"
// but one the terminal actually painted and the SVG silently dropped:
// resolveChannel reads a false here as "keep whatever was active", which
// after a leading diamond is the segment's own background — text painted in
// its own background color, invisible. Six bundled themes write 3-digit
// shorthand (#fff/#FFF/#000) and rendered exactly that way before that fix:
// cert, iterm2, montys, poshmon, uew, unicorn. Calling gookit's own parser
// makes that parity structural instead of something a future edit here can
// drift away from again; it also trims surrounding whitespace, which the
// hand-rolled version never did, so a whitespace-padded hex the ANSI writer
// already painted now renders here too instead of being dropped.
func hexToRGB(value string) (*color.RGB, bool) {
	rgb := gookit.HexToRgb(value)
	if len(rgb) != 3 {
		return nil, false
	}

	return &color.RGB{R: uint8(rgb[0]), G: uint8(rgb[1]), B: uint8(rgb[2])}, true
}

// namedRGB resolves an ansiColorCodes name (see color/colors.go's
// ansiColorCodes map) to RGB. "default" is context-dependent rather than a
// fixed color: its background channel is the real terminal background
// (Options.TerminalBackground, possibly unknown), and its foreground channel
// has no equivalent explicit input (see Options' doc comment), so it falls
// back to a fixed approximation.
func namedRGB(name string, isBackground bool, opts *Options) (*color.RGB, bool) {
	if name == "default" {
		if isBackground {
			return opts.TerminalBackground, true
		}

		return &defaultForegroundRGB, true
	}

	rgb, ok := ansiNamedRGB[name]
	if !ok {
		return nil, false
	}

	return &rgb, true
}

// defaultForegroundRGB is a fixed approximation of a terminal's default
// foreground color, used only for the ansiColorCodes name "default" on a
// foreground channel; see namedRGB's doc comment for why this can't be
// resolved any more precisely from Options.
var defaultForegroundRGB = color.RGB{R: 255, G: 255, B: 255}

// ansiNamedRGB gives each of ansiColorCodes' 16 names (color/colors.go) an
// actual RGB, since ansiColorCodes itself only carries the SGR number, not a
// color a browser can render.
var ansiNamedRGB = map[string]color.RGB{
	"black":        {R: 1, G: 1, B: 1},
	"red":          {R: 222, G: 56, B: 43},
	"green":        {R: 57, G: 181, B: 74},
	"yellow":       {R: 255, G: 199, B: 6},
	"blue":         {R: 0, G: 111, B: 184},
	"magenta":      {R: 118, G: 38, B: 113},
	"cyan":         {R: 44, G: 181, B: 233},
	"white":        {R: 204, G: 204, B: 204},
	"darkGray":     {R: 128, G: 128, B: 128},
	"lightRed":     {R: 255, G: 0, B: 0},
	"lightGreen":   {R: 0, G: 255, B: 0},
	"lightYellow":  {R: 255, G: 255, B: 0},
	"lightBlue":    {R: 0, G: 0, B: 255},
	"lightMagenta": {R: 255, G: 0, B: 255},
	"lightCyan":    {R: 101, G: 194, B: 205},
	"lightWhite":   {R: 255, G: 255, B: 255},
}

// ansi16RGB is ansiNamedRGB in SGR code order (30-37, then 90-97), for
// palette256's first 16 indices.
var ansi16RGB = [16]color.RGB{
	ansiNamedRGB["black"], ansiNamedRGB["red"], ansiNamedRGB["green"], ansiNamedRGB["yellow"],
	ansiNamedRGB["blue"], ansiNamedRGB["magenta"], ansiNamedRGB["cyan"], ansiNamedRGB["white"],
	ansiNamedRGB["darkGray"], ansiNamedRGB["lightRed"], ansiNamedRGB["lightGreen"], ansiNamedRGB["lightYellow"],
	ansiNamedRGB["lightBlue"], ansiNamedRGB["lightMagenta"], ansiNamedRGB["lightCyan"], ansiNamedRGB["lightWhite"],
}

// cube6 is the 6-step intensity ramp the 256-color palette's 6x6x6 color
// cube (indices 16-231) is built from; the standard xterm values.
var cube6 = [6]uint8{0, 95, 135, 175, 215, 255}

// palette256 converts a 256-color palette index to RGB using the standard
// xterm layout: 0-15 are the 16 named colors above, 16-231 are a 6x6x6 color
// cube, and 232-255 are a 24-step grayscale ramp.
func palette256(index uint8) color.RGB {
	if index < 16 {
		return ansi16RGB[index]
	}

	if index >= 232 {
		level := uint8(8 + (int(index)-232)*10)
		return color.RGB{R: level, G: level, B: level}
	}

	i := index - 16
	return color.RGB{R: cube6[i/36], G: cube6[(i/6)%6], B: cube6[i%6]}
}
