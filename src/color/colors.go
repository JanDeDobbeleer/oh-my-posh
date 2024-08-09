package color

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

var TrueColor = true

// String is the interface that wraps ToColor method.
//
// ToColor gets the ANSI color code for a given color string.
// This can include a valid hex color in the format `#FFFFFF`,
// but also a name of one of the first 16 ANSI colors like `lightBlue`.
type String interface {
	ToAnsi(colorString Ansi, isBackground bool) Ansi
}

type Set struct {
	Background Ansi `json:"background" toml:"background"`
	Foreground Ansi `json:"foreground" toml:"foreground"`
}

type History []*Set

func (c *History) Len() int {
	return len(*c)
}

func (c *History) Add(background, foreground Ansi) {
	colors := &Set{
		Foreground: foreground,
		Background: background,
	}

	if c.Len() == 0 {
		*c = append(*c, colors)
		return
	}

	last := (*c)[c.Len()-1]
	// never add the same colors twice
	if last.Foreground == colors.Foreground && last.Background == colors.Background {
		return
	}

	*c = append(*c, colors)
}

func (c *History) Pop() {
	if c.Len() == 0 {
		return
	}

	*c = (*c)[:c.Len()-1]
}

func (c *History) Background() Ansi {
	if c.Len() == 0 {
		return emptyColor
	}

	return (*c)[c.Len()-1].Background
}

func (c *History) Foreground() Ansi {
	if c.Len() == 0 {
		return emptyColor
	}

	return (*c)[c.Len()-1].Foreground
}

// Ansi is an ANSI color code ready to be printed to the console.
// Example: "38;2;255;255;255", "48;2;255;255;255", "31", "95".
type Ansi string

const (
	emptyColor = Ansi("")
)

func (c Ansi) IsEmpty() bool {
	return c == emptyColor
}

func (c Ansi) IsTransparent() bool {
	return c == Transparent
}

func (c Ansi) IsClear() bool {
	return c == Transparent || c == emptyColor
}

func (c Ansi) ToForeground() Ansi {
	colorString := c.String()
	if strings.HasPrefix(colorString, "38;") {
		return Ansi(strings.Replace(colorString, "38;", "48;", 1))
	}
	return c
}

func (c Ansi) ResolveTemplate() Ansi {
	if c.IsEmpty() {
		return c
	}

	if c.IsTransparent() {
		return emptyColor
	}

	tmpl := &template.Text{
		Template: string(c),
		Context:  nil,
	}

	text, err := tmpl.Render()
	if err != nil {
		return Transparent
	}

	return Ansi(text)
}

func (c Ansi) String() string {
	return string(c)
}

func MakeColors(palette Palette, cacheEnabled bool, accentColor Ansi, env runtime.Environment) (colors String) {
	defaultColors := &Defaults{}
	defaultColors.SetAccentColor(env, accentColor)
	colors = defaultColors
	if palette != nil {
		colors = &PaletteColors{ansiColors: colors, palette: palette}
	}
	if cacheEnabled {
		colors = &Cached{ansiColors: colors}
	}
	return
}

type RGB struct {
	R, G, B uint8
}

// Defaults is the default AnsiColors implementation.
type Defaults struct {
	accent *Set
}

var (
	// Map for color names and their respective foreground [0] or background [1] color codes
	ansiColorCodes = map[Ansi][2]Ansi{
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

func (d *Defaults) ToAnsi(ansiColor Ansi, isBackground bool) Ansi {
	if len(ansiColor) == 0 {
		return emptyColor
	}

	if ansiColor.IsTransparent() {
		return ansiColor
	}

	if ansiColor == Accent {
		if d.accent == nil {
			return emptyColor
		}

		if isBackground {
			return d.accent.Background
		}

		return d.accent.Foreground
	}

	colorFromName, err := getAnsiColorFromName(ansiColor, isBackground)
	if err == nil {
		return colorFromName
	}

	colorString := ansiColor.String()

	if !strings.HasPrefix(colorString, "#") {
		val, err := strconv.ParseUint(colorString, 10, 64)
		if err != nil || val > 255 {
			return emptyColor
		}

		c256 := color.C256(uint8(val), isBackground)
		return Ansi(c256.String())
	}

	style := color.HEX(colorString, isBackground)
	if !style.IsEmpty() {
		if TrueColor {
			return Ansi(style.String())
		}

		return Ansi(style.C256().String())
	}

	if colorInt, err := strconv.ParseInt(colorString, 10, 8); err == nil {
		c := color.C256(uint8(colorInt), isBackground)

		return Ansi(c.String())
	}

	return emptyColor
}

// getAnsiColorFromName returns the color code for a given color name if the name is
// known ANSI color name.
func getAnsiColorFromName(colorValue Ansi, isBackground bool) (Ansi, error) {
	if colorCodes, found := ansiColorCodes[colorValue]; found {
		if isBackground {
			return colorCodes[backgroundIndex], nil
		}

		return colorCodes[foregroundIndex], nil
	}

	return "", fmt.Errorf("color name %s does not exist", colorValue)
}

func IsAnsiColorName(colorValue Ansi) bool {
	_, ok := ansiColorCodes[colorValue]
	return ok
}

// PaletteColors is the AnsiColors Decorator that uses the Palette to do named color
// lookups before ANSI color code generation.
type PaletteColors struct {
	ansiColors String
	palette    Palette
}

func (p *PaletteColors) ToAnsi(colorString Ansi, isBackground bool) Ansi {
	paletteColor, err := p.palette.ResolveColor(colorString)
	if err != nil {
		return emptyColor
	}

	ansiColor := p.ansiColors.ToAnsi(paletteColor, isBackground)

	return ansiColor
}

// Cached is the AnsiColors Decorator that does simple color lookup caching.
// ToColor calls are cheap, but not free, and having a simple cache in
// has measurable positive effect on performance.
type Cached struct {
	ansiColors String
	colorCache map[cachedColorKey]Ansi
}

type cachedColorKey struct {
	colorString  Ansi
	isBackground bool
}

func (c *Cached) ToAnsi(colorString Ansi, isBackground bool) Ansi {
	if c.colorCache == nil {
		c.colorCache = make(map[cachedColorKey]Ansi)
	}
	key := cachedColorKey{colorString, isBackground}
	if ansiColor, hit := c.colorCache[key]; hit {
		return ansiColor
	}
	ansiColor := c.ansiColors.ToAnsi(colorString, isBackground)
	c.colorCache[key] = ansiColor
	return ansiColor
}
