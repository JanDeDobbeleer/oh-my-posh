package main

import (
	"fmt"

	"github.com/gookit/color"
)

// MakeColors creates instance of AnsiColors to use in AnsiWriter according to
// environment and configuration.
func MakeColors(env Environment, cfg *Config) AnsiColors {
	cacheDisabled := env.getenv("OMP_CACHE_DISABLED") == "1"
	return makeColors(cfg.Palette, !cacheDisabled)
}

func makeColors(palette Palette, cacheEnabled bool) (colors AnsiColors) {
	colors = &DefaultColors{}
	if palette != nil {
		colors = &PaletteColors{ansiColors: colors, palette: palette}
	}
	if cacheEnabled {
		colors = &CachedColors{ansiColors: colors}
	}
	return
}

// DefaultColors is the default AnsiColors implementation.
type DefaultColors struct{}

var (
	// Map for color names and their respective foreground [0] or background [1] color codes
	ansiColorCodes = map[string][2]AnsiColor{
		"black":        {"30", "40"},
		"red":          {"31", "41"},
		"green":        {"32", "42"},
		"yellow":       {"33", "43"},
		"blue":         {"34", "44"},
		"magenta":      {"35", "45"},
		"cyan":         {"36", "46"},
		"white":        {"37", "47"},
		"default":      {"39", "49"},
		"darkGray":     {"90", "100"},
		"lightRed":     {"91", "101"},
		"lightGreen":   {"92", "102"},
		"lightYellow":  {"93", "103"},
		"lightBlue":    {"94", "104"},
		"lightMagenta": {"95", "105"},
		"lightCyan":    {"96", "106"},
		"lightWhite":   {"97", "107"},
	}
)

const (
	foregroundIndex = 0
	backgroundIndex = 1
)

func (*DefaultColors) AnsiColorFromString(colorString string, isBackground bool) AnsiColor {
	if len(colorString) == 0 {
		return emptyAnsiColor
	}
	if colorString == Transparent {
		return transparentAnsiColor
	}
	colorFromName, err := getAnsiColorFromName(colorString, isBackground)
	if err == nil {
		return colorFromName
	}
	style := color.HEX(colorString, isBackground)
	if style.IsEmpty() {
		return emptyAnsiColor
	}
	return AnsiColor(style.String())
}

// getAnsiColorFromName returns the color code for a given color name if the name is
// known ANSI color name.
func getAnsiColorFromName(colorName string, isBackground bool) (AnsiColor, error) {
	if colorCodes, found := ansiColorCodes[colorName]; found {
		if isBackground {
			return colorCodes[backgroundIndex], nil
		}
		return colorCodes[foregroundIndex], nil
	}
	return "", fmt.Errorf("color name %s does not exist", colorName)
}

func IsAnsiColorName(colorString string) bool {
	_, ok := ansiColorCodes[colorString]
	return ok
}

// PaletteColors is the AnsiColors Decorator that uses the Palette to do named color
// lookups before ANSI color code generation.
type PaletteColors struct {
	ansiColors AnsiColors
	palette    Palette
}

func (p *PaletteColors) AnsiColorFromString(colorString string, isBackground bool) AnsiColor {
	paletteColor, err := p.palette.ResolveColor(colorString)
	if err != nil {
		return emptyAnsiColor
	}
	ansiColor := p.ansiColors.AnsiColorFromString(paletteColor, isBackground)
	return ansiColor
}

// CachedColors is the AnsiColors Decorator that does simple color lookup caching.
// AnsiColorFromString calls are cheap, but not free, and having a simple cache in
// has measurable positive effect on performance.
type CachedColors struct {
	ansiColors AnsiColors
	colorCache map[cachedColorKey]AnsiColor
}

type cachedColorKey struct {
	colorString  string
	isBackground bool
}

func (c *CachedColors) AnsiColorFromString(colorString string, isBackground bool) AnsiColor {
	if c.colorCache == nil {
		c.colorCache = make(map[cachedColorKey]AnsiColor)
	}
	key := cachedColorKey{colorString, isBackground}
	if ansiColor, hit := c.colorCache[key]; hit {
		return ansiColor
	}
	ansiColor := c.ansiColors.AnsiColorFromString(colorString, isBackground)
	c.colorCache[key] = ansiColor
	return ansiColor
}
