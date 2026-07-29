package color

import (
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/generics"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

func init() {
	gob.Register(&Set{})
	gob.Register((*Ansi)(nil))
	gob.Register(&Palette{})
	gob.Register(&Palettes{})
	gob.Register(&Cycle{})
}

const (
	accentColor = "accent_color"
)

var TrueColor = true

// String converts a color string — a hex color like `#FFFFFF`, or one of the
// first 16 ANSI color names like `lightBlue` — to an ANSI code.
type String interface {
	ToAnsi(colorString Ansi, isBackground bool) Ansi
	Resolve(colorString Ansi) (Ansi, error)
}

// Set holds one background/foreground color pair. The two unexported source-tracking
// fields below add +96 B/op (BenchmarkWriteAnchors) / +32 B/op (BenchmarkWritePlainASCII)
// versus a Set without them, at identical allocation counts, measured against
// origin/main with terminal.CaptureRuns off: History.Add allocates a Set per push, and
// every push now carries two extra Ansi (string header) fields regardless of whether
// anything reads them. This is accepted, not a bug: the fields are unconditional rather
// than threading CaptureRuns into this package, which would risk the run-capture and
// ANSI-only paths drifting apart between builds that enable it and builds that don't.
type Set struct {
	Background Ansi `json:"background" toml:"background" yaml:"background"`
	Foreground Ansi `json:"foreground" toml:"foreground" yaml:"foreground"`

	// backgroundSource/foregroundSource carry each channel's resolved SOURCE
	// form (a #RRGGBB hex, a colour name, accent, transparent, ...) alongside
	// Background/Foreground's SGR payload. Unexported: they never round-trip
	// through config (un)marshalling or the gob-encoded device cache, which
	// only see exported fields, so their presence changes neither. See
	// History.Add for why they ride on the same entry instead of a second stack.
	backgroundSource Ansi
	foregroundSource Ansi
}

func (c *Set) String() string {
	return fmt.Sprintf("%s|%s", c.Foreground, c.Background)
}

func (c *Set) ParseString(colors string) {
	parts := strings.SplitN(colors, "|", 3)
	if len(parts) != 2 {
		return
	}

	c.Foreground = Ansi(parts[0])
	c.Background = Ansi(parts[1])
}

type History []*Set

func (c *History) Len() int {
	return len(*c)
}

// Add pushes background/foreground (SGR) onto the stack, together with
// backgroundSource/foregroundSource — each channel's source form from before
// ToAnsi's SGR conversion, which a later encoder needs because SGR cannot be
// inverted back to it reliably (the 256-colour downgrade and base-16 codes
// are lossy).
//
// The dedupe decision below compares only background/foreground (SGR), same
// as before this field existed: under color.TrueColor == false, two distinct
// source forms can collapse to the same SGR pair, so the decision must not
// be made on source forms, or it would diverge from this one.
func (c *History) Add(background, foreground, backgroundSource, foregroundSource Ansi) {
	colors := &Set{
		Foreground:       foreground,
		Background:       background,
		foregroundSource: foregroundSource,
		backgroundSource: backgroundSource,
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

// BackgroundSource returns the top entry's background source form (see Add);
// emptyColor when the stack is empty, matching Background.
func (c *History) BackgroundSource() Ansi {
	if c.Len() == 0 {
		return emptyColor
	}

	return (*c)[c.Len()-1].backgroundSource
}

// ForegroundSource is BackgroundSource's foreground counterpart.
func (c *History) ForegroundSource() Ansi {
	if c.Len() == 0 {
		return emptyColor
	}

	return (*c)[c.Len()-1].foregroundSource
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

// ToChannel returns the color code adjusted for the requested channel, converting
// between the foreground (38;...) and background (48;...) escape payload prefixes.
func (c Ansi) ToChannel(isBackground bool) Ansi {
	colorString := c.String()

	switch {
	case isBackground && strings.HasPrefix(colorString, "38;"):
		return Ansi("48;" + colorString[3:])
	case !isBackground && strings.HasPrefix(colorString, "48;"):
		return Ansi("38;" + colorString[3:])
	default:
		return c
	}
}

func (c Ansi) ResolveTemplate() Ansi {
	if c.IsEmpty() {
		return c
	}

	if c.IsTransparent() {
		return emptyColor
	}

	text, err := template.RenderTrusted(string(c), nil)
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

// unresolvedAccent is the negative marker cached when the OS accent color
// cannot be resolved, so we don't keep retrying the (expensive) OS query
// on every prompt render.
var unresolvedAccent = &Set{}

func (d *Defaults) SetAccentColor(env runtime.Environment, defaultColor Ansi) {
	defer log.Trace(time.Now())

	// a gradient accent_color cannot serve as the single accent value; collapse to
	// its first stop so the cached accent Set never holds a raw gradient string.
	if defaultColor.IsGradient() {
		defaultColor = defaultColor.GradientFirst()
	}

	// get the resolved OS accent color from the device cache first, regardless
	// of whether a default was configured, so we never repeat the underlying
	// OS query (e.g. a DWM registry read) once we know the answer.
	accent, OK := cache.Get[*Set](cache.Device, accentColor)
	if !OK {
		rgb, err := GetAccentColor(env)
		if err != nil {
			accent = unresolvedAccent
		} else {
			foreground := color.RGB(rgb.R, rgb.G, rgb.B, false)
			background := color.RGB(rgb.R, rgb.G, rgb.B, true)

			accent = &Set{
				Foreground: Ansi(foreground.String()),
				Background: Ansi(background.String()),
			}
		}

		cache.Set(cache.Device, accentColor, accent, cache.INFINITE)
	}

	if accent.Foreground.IsEmpty() && accent.Background.IsEmpty() {
		d.accent = &Set{
			Foreground: d.ToAnsi(defaultColor, false),
			Background: d.ToAnsi(defaultColor, true),
		}

		return
	}

	d.accent = accent
}

type RGB struct {
	R, G, B uint8
}

// ParseTrueColorRGB parses a 24-bit SGR payload ("38;2;r;g;b" or
// "48;2;r;g;b") into RGB. It exists for the one color that can't be carried
// as a hex string: the OS accent color resolves through ToAnsi to this form
// rather than hex (see Defaults.SetAccentColor building it via gookit's
// color.RGB), so both a gradient stop resolving "accent" (gradient.go's
// parseTrueColor, a thin colorful.Color-returning wrapper around this) and
// the svg package's own Options resolving "accent" (svg/colors.go's
// ParseTrueColorSGR, an RGB-pointer-returning wrapper) need the same parse.
// Kept in this package rather than duplicated in each caller so the two
// never drift out of parity with each other again.
func ParseTrueColorRGB(c Ansi) (RGB, bool) {
	parts := strings.Split(c.String(), ";")
	if len(parts) != 5 || (parts[0] != "38" && parts[0] != "48") || parts[1] != "2" {
		return RGB{}, false
	}

	r, err := strconv.ParseUint(parts[2], 10, 8)
	if err != nil {
		return RGB{}, false
	}

	g, err := strconv.ParseUint(parts[3], 10, 8)
	if err != nil {
		return RGB{}, false
	}

	b, err := strconv.ParseUint(parts[4], 10, 8)
	if err != nil {
		return RGB{}, false
	}

	return RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, true
}

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

func (d *Defaults) ToAnsi(ansiColor Ansi, isBackground bool) Ansi {
	if ansiColor == "" {
		return emptyColor
	}

	if ansiColor.IsTransparent() {
		return ansiColor
	}

	// a gradient rides the ANSI plumbing as a plain string; the terminal writer detects
	// and renders it per cell, so it must never be mangled into hex/256 parsing here.
	if ansiColor.IsGradient() {
		return ansiColor
	}

	if dir, inner, percent, ok := ansiColor.ShadeArgs(); ok {
		hex, ok := shadeHex(inner, dir, percent)
		if !ok {
			log.Errorf("%s: darken()/lighten() only support hex colors and palette references resolving to one; "+
				"named ANSI/terminal colors have no fixed RGB oh-my-posh can shift, define the color in your palette with a hex value instead", ansiColor)
			return emptyColor
		}

		ansiColor = Ansi(hex)
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

func (d *Defaults) Resolve(colorString Ansi) (Ansi, error) {
	return colorString, nil
}

func getAnsiColorFromName(colorValue Ansi, isBackground bool) (Ansi, error) {
	if colorCodes, found := ansiColorCodes[colorValue]; found {
		return colorCodes[generics.ToInt[int](isBackground)], nil
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
	// a gradient string is not a palette key; guard it explicitly so it never round-trips
	// through palette resolution and reaches the next decorator untouched.
	if colorString.IsGradient() {
		return colorString
	}

	colorString, err := p.palette.resolveShade(colorString)
	if err != nil {
		return emptyColor
	}

	paletteColor, err := p.palette.ResolveColor(colorString)
	if err != nil {
		return emptyColor
	}

	ansiColor := p.ansiColors.ToAnsi(paletteColor, isBackground)

	return ansiColor
}

func (p *PaletteColors) Resolve(colorString Ansi) (Ansi, error) {
	colorString, err := p.palette.resolveShade(colorString)
	if err != nil {
		return "", err
	}

	return p.palette.ResolveColor(colorString)
}

// Cached is the AnsiColors Decorator that does simple color lookup caching.
// ToAnsi calls are cheap but not free, and caching has a measurable positive effect on performance.
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

func (c *Cached) Resolve(colorString Ansi) (Ansi, error) {
	return c.ansiColors.Resolve(colorString)
}
