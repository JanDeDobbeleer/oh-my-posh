package ansi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/gookit/color"
)

// Colors is the interface that wraps AnsiColorFromString method.
//
// AnsiColorFromString gets the ANSI color code for a given color string.
// This can include a valid hex color in the format `#FFFFFF`,
// but also a name of one of the first 16 ANSI colors like `lightBlue`.
type Colors interface {
	AnsiColorFromString(colorString string, isBackground bool) Color
}

// Color is an ANSI color code ready to be printed to the console.
// Example: "38;2;255;255;255", "48;2;255;255;255", "31", "95".
type Color string

const (
	emptyColor       = Color("")
	transparentColor = Color(Transparent)
)

func (c Color) IsEmpty() bool {
	return c == emptyColor
}

func (c Color) IsTransparent() bool {
	return c == transparentColor
}

func (c Color) IsClear() bool {
	return c == transparentColor || c == emptyColor
}

func (c Color) ToForeground() Color {
	colorString := string(c)
	if strings.HasPrefix(colorString, "38;") {
		return Color(strings.Replace(colorString, "38;", "48;", 1))
	}
	return c
}

func MakeColors(palette Palette, cacheEnabled bool, accentColor string, env platform.Environment) (colors Colors) {
	defaultColors := &DefaultColors{}
	defaultColors.SetAccentColor(env, accentColor)
	colors = defaultColors
	if palette != nil {
		colors = &PaletteColors{ansiColors: colors, palette: palette}
	}
	if cacheEnabled {
		colors = &CachedColors{ansiColors: colors}
	}
	return
}

type RGB struct {
	R, G, B uint8
}

// DefaultColors is the default AnsiColors implementation.
type DefaultColors struct {
	accent *cachedColor
}

var (
	// Map for color names and their respective foreground [0] or background [1] color codes
	ansiColorCodes = map[string][2]Color{
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

func (d *DefaultColors) AnsiColorFromString(colorString string, isBackground bool) Color {
	if len(colorString) == 0 {
		return emptyColor
	}
	if colorString == Transparent {
		return transparentColor
	}
	if colorString == Accent {
		if d.accent == nil {
			return emptyColor
		}
		if isBackground {
			return Color(d.accent.Background)
		}
		return Color(d.accent.Foreground)
	}
	colorFromName, err := getAnsiColorFromName(colorString, isBackground)
	if err == nil {
		return colorFromName
	}
	if !strings.HasPrefix(colorString, "#") {
		val, err := strconv.ParseUint(colorString, 10, 64)
		if err != nil || val > 255 {
			return emptyColor
		}
		c256 := color.C256(uint8(val), isBackground)
		return Color(c256.RGBColor().String())
	}
	style := color.HEX(colorString, isBackground)
	if !style.IsEmpty() {
		return Color(style.String())
	}
	if colorInt, err := strconv.ParseInt(colorString, 10, 8); err == nil {
		c := color.C256(uint8(colorInt), isBackground)
		return Color(c.String())
	}
	return emptyColor
}

// getAnsiColorFromName returns the color code for a given color name if the name is
// known ANSI color name.
func getAnsiColorFromName(colorName string, isBackground bool) (Color, error) {
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
	ansiColors Colors
	palette    Palette
}

func (p *PaletteColors) AnsiColorFromString(colorString string, isBackground bool) Color {
	paletteColor, err := p.palette.ResolveColor(colorString)
	if err != nil {
		return emptyColor
	}
	ansiColor := p.ansiColors.AnsiColorFromString(paletteColor, isBackground)
	return ansiColor
}

// CachedColors is the AnsiColors Decorator that does simple color lookup caching.
// AnsiColorFromString calls are cheap, but not free, and having a simple cache in
// has measurable positive effect on performance.
type CachedColors struct {
	ansiColors Colors
	colorCache map[cachedColorKey]Color
}

type cachedColorKey struct {
	colorString  string
	isBackground bool
}

func (c *CachedColors) AnsiColorFromString(colorString string, isBackground bool) Color {
	if c.colorCache == nil {
		c.colorCache = make(map[cachedColorKey]Color)
	}
	key := cachedColorKey{colorString, isBackground}
	if ansiColor, hit := c.colorCache[key]; hit {
		return ansiColor
	}
	ansiColor := c.ansiColors.AnsiColorFromString(colorString, isBackground)
	c.colorCache[key] = ansiColor
	return ansiColor
}
