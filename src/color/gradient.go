package color

import (
	"math"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	linearGradientPrefix = "linear-gradient("
	darkGradientPrefix   = "dark-gradient("
	lightGradientPrefix  = "light-gradient("
	gradientSuffix       = ")"
)

// gradientPrefixes lists every recognized gradient prefix, checked in order.
var gradientPrefixes = [...]string{linearGradientPrefix, darkGradientPrefix, lightGradientPrefix}

// IsGradient reports whether c is a gradient definition: a multi-stop
// `linear-gradient(#FF0000, #0000FF)`, or a single-color auto-shade
// `dark-gradient(#3465a4)` / `light-gradient(#3465a4)`.
func (c Ansi) IsGradient() bool {
	_, ok := c.gradientPrefix()
	return ok
}

// gradientPrefix returns the gradient prefix c starts with, and whether it matched any.
func (c Ansi) gradientPrefix() (string, bool) {
	s := c.String()

	for _, prefix := range gradientPrefixes {
		if strings.HasPrefix(s, prefix) {
			return prefix, true
		}
	}

	return "", false
}

// shadeDirection identifies an auto-shade gradient's single stop as darkening
// (dark-gradient) or lightening (light-gradient); shadeNone means c is a plain
// linear-gradient, or not a gradient at all.
type shadeDirection int

const (
	shadeNone shadeDirection = iota
	shadeDark
	shadeLight
)

func (c Ansi) shadeDirection() shadeDirection {
	s := c.String()

	switch {
	case strings.HasPrefix(s, darkGradientPrefix):
		return shadeDark
	case strings.HasPrefix(s, lightGradientPrefix):
		return shadeLight
	default:
		return shadeNone
	}
}

// GradientStops performs syntax parsing only: it splits the comma-separated stop list inside
// `linear-gradient(...)`/`dark-gradient(...)`/`light-gradient(...)` and trims whitespace around
// each stop. It does not resolve palette references or validate that a stop is a color. Returns
// nil when c is not a gradient, the closing paren is missing, the body contains a nested paren
// (angle/direction syntax is reserved but not implemented), or any stop is empty.
func (c Ansi) GradientStops() []Ansi {
	prefix, ok := c.gradientPrefix()
	if !ok {
		return nil
	}

	value := c.String()
	if !strings.HasSuffix(value, gradientSuffix) {
		return nil
	}

	body := value[len(prefix) : len(value)-len(gradientSuffix)]
	if strings.ContainsAny(body, "()") {
		return nil
	}

	parts := strings.Split(body, ",")
	stops := make([]Ansi, 0, len(parts))

	for _, part := range parts {
		stop := strings.TrimSpace(part)
		if stop == "" {
			return nil
		}

		stops = append(stops, Ansi(stop))
	}

	return stops
}

// WithGradientStops rebuilds c with the same gradient prefix (linear-gradient, dark-gradient,
// or light-gradient) but stops in place of the original ones. Returns c unchanged when c is
// not a gradient, so a caller can use it unconditionally on a value that might not be one.
func (c Ansi) WithGradientStops(stops []Ansi) Ansi {
	prefix, ok := c.gradientPrefix()
	if !ok {
		return c
	}

	parts := make([]string, len(stops))
	for i, stop := range stops {
		parts[i] = stop.String()
	}

	return Ansi(prefix + strings.Join(parts, ", ") + gradientSuffix)
}

// GradientFirst returns the first stop of the gradient. It returns c unchanged when c is not
// a gradient, or when the gradient syntax is invalid.
func (c Ansi) GradientFirst() Ansi {
	stops := c.GradientStops()
	if len(stops) == 0 {
		return c
	}

	return stops[0]
}

// GradientLast returns the last stop of the gradient. It returns c unchanged when c is not
// a gradient, or when the gradient syntax is invalid. Equivalent to GradientLastForCells(0):
// a dark-gradient/light-gradient shades using the narrowest (gentlest) auto-shade step, since
// the segment's actual width isn't known here. Callers that know it — the segment's own
// separators, diamond caps, and inline overrides — should call GradientLastForCells instead,
// so the edge matches the actual last cell GradientCells renders rather than the fallback.
func (c Ansi) GradientLast() Ansi {
	return c.GradientLastForCells(0)
}

// GradientLastForCells is GradientLast, but for a dark-gradient/light-gradient it shades by
// exactly as much as GradientCells(c, cells, ...) would shade the segment's actual last cell
// by, so a separator/cap/parent color reference for a segment of this width matches the body
// precisely instead of jumping to a different shade right after it. cells <= 0 (width unknown)
// uses the same gentle single-step shade as a 2-cell segment.
func (c Ansi) GradientLastForCells(cells int) Ansi {
	stops := c.GradientStops()
	if len(stops) == 0 {
		return c
	}

	last := stops[len(stops)-1]

	if dir := c.shadeDirection(); dir != shadeNone && len(stops) == 1 {
		if clr, err := colorful.Hex(last.String()); err == nil {
			return Ansi(autoShade(clr, dir, cells).Hex())
		}
	}

	return last
}

// GradientCells resolves each stop of the gradient c — keywords like parentBackground against
// current/parents, palette references through resolver — parses the result as a hex color, and
// interpolates across cells steps in HCL space. It returns one ready-to-print ANSI color code
// per cell, honoring the package-level TrueColor flag: a truecolor escape when true, a gookit
// C256 downgrade when false. cells == 1 returns only the first stop. Returns nil when fewer
// than two stops resolve to a valid color; the caller falls back to a single collapsed color
// per the gradient-invalid rule.
func GradientCells(c Ansi, cells int, resolver String, isBackground bool, current *Set, parents []*Set) []Ansi {
	if cells <= 0 {
		return nil
	}

	stops := c.GradientStops()
	if len(stops) == 0 {
		log.Errorf("gradient %s: invalid syntax, expected linear-gradient(stop, stop, ...), dark-gradient(color), or light-gradient(color)", c)
		return nil
	}

	colors := make([]colorful.Color, 0, len(stops))

	for _, stop := range stops {
		// a keyword stop (parentBackground, foreground, ...) resolves against the
		// segment context first; a parent gradient collapses to its last stop there.
		resolved := stop.Resolve(current, parents)

		resolved, err := resolver.Resolve(resolved)
		if err != nil {
			log.Errorf("gradient %s: unable to resolve stop %s: %s", c, stop, err)
			continue
		}

		// the OS accent color only resolves in ToAnsi, to a truecolor payload
		// rather than hex; parseTrueColor recovers the RGB triplet from it.
		if resolved == Accent {
			resolved = resolver.ToAnsi(Accent, false)
		}

		clr, err := colorful.Hex(resolved.String())
		if err != nil {
			var ok bool
			if clr, ok = parseTrueColor(resolved); !ok {
				log.Errorf("gradient %s: stop %s does not resolve to a color, only hex colors, palette references, and keywords resolving to a color can be interpolated", c, stop)
				continue
			}
		}

		colors = append(colors, clr)
	}

	// dark-gradient(color)/light-gradient(color) is an auto-shade request: turn the one
	// resolved color into a real two-stop gradient running from that exact color to a
	// darker/lighter shade of it, sized to cells, reusing the interpolation below
	// unchanged. See GradientLastForCells for the matching edge (separators, diamond
	// caps, parent color refs).
	if dir := c.shadeDirection(); dir != shadeNone {
		if len(stops) != 1 {
			log.Errorf("gradient %s: expects exactly one color stop, e.g. dark-gradient(#3465a4)", c)
			return nil
		}

		if len(colors) == 0 {
			log.Errorf("gradient %s: stop does not resolve to a color", c)
			return nil
		}

		colors = []colorful.Color{colors[0], autoShade(colors[0], dir, cells)}
	}

	if len(colors) < 2 {
		log.Errorf("gradient %s: needs at least two valid stops, rendering the last stop as a solid color", c)
		return nil
	}

	if cached, ok := gradientCellCache[gradientKey(colors, cells, isBackground)]; ok {
		return cached
	}

	if cells == 1 {
		return cacheGradientCells(colors, cells, isBackground, []Ansi{ansiFromColorful(colors[0], isBackground)})
	}

	segments := len(colors) - 1
	result := make([]Ansi, cells)

	for i := range cells {
		position := float64(i) / float64(cells-1) * float64(segments)

		segment := int(position)
		if segment >= segments {
			segment = segments - 1
		}

		blended := colors[segment].BlendHcl(colors[segment+1], position-float64(segment)).Clamped()
		result[i] = ansiFromColorful(blended, isBackground)
	}

	return cacheGradientCells(colors, cells, isBackground, result)
}

// autoShadeBaseSlope/Floor/Ceiling shape the lightness shift autoShade targets AT THE
// REFERENCE WIDTH (2 steps, i.e. a 3-cell segment - the tuned-by-feel narrow case dark-
// gradient/light-gradient exists for), as a fraction of the base color's own headroom
// toward black or white (so a color already close to the target end still shifts
// visibly, and one far from it doesn't overshoot), clamped to a floor and a ceiling.
//
// autoShadeWidthMultiplier grows that reference delta for a wider segment, so a wide
// gradient still reads as a clear effect instead of fading into an imperceptibly fine
// ramp - but as a SATURATING curve (autoShadeMaxMultiplier, approached over roughly
// autoShadeWidthSteepness steps), not linear growth with a hard clamp: linear growth
// hit its clamp by ~15-20 cells and stayed there, crushing every segment wider than
// that to the exact same near-white/near-black color regardless of how much wider it
// got. The curve instead keeps easing toward the cap, so a 90-cell segment shades
// further than a 20-cell one rather than looking identical to it, while a very wide
// segment still tops out at a moderate, recognizably-still-the-same-hue shade instead
// of washing out to white or crushing to black.
const (
	autoShadeBaseSlope      = 0.08
	autoShadeBaseFloor      = 0.025
	autoShadeBaseCeiling    = 0.05
	autoShadeMaxMultiplier  = 4.0
	autoShadeWidthSteepness = 15.0
	autoShadeMinLightness   = 0.02
	autoShadeMaxLightness   = 0.98
	autoShadeChromaStep     = 0.005
)

// autoShade derives a dark-gradient/light-gradient's second stop: same hue, same chroma
// other than what the sRGB gamut forces away, lightness shifted toward black (shadeDark)
// or white (shadeLight) by the reference delta (see autoShadeBaseSlope) times a width
// multiplier that saturates as cells grows (see autoShadeMaxMultiplier), so
// GradientCells(c, cells, ...) always lands its actual last cell here regardless of
// width. Blending toward black/white directly (an earlier approach) pulls chroma along
// with it, which reads as the color going muddy rather than deepening/brightening;
// walking chroma down only as far as IsValid demands keeps the hue exact and preserves as
// much saturation as the gamut allows at the new lightness. GradientLastForCells mirrors
// this on the raw (unresolved) stop text so a segment's trailing separator/cap matches
// the color GradientCells renders the last cell as, for the SAME cells.
func autoShade(base colorful.Color, dir shadeDirection, cells int) colorful.Color {
	h, c, l := base.Hcl()

	// steps == 2 (a 3-cell segment) is the reference width: widthMultiplier == 1 there,
	// so the delta below equals exactly what a 3-cell segment was tuned to look like.
	steps := float64(max(1, cells-1))
	widthMultiplier := 1 + (autoShadeMaxMultiplier-1)*(1-math.Exp(-(steps-2)/autoShadeWidthSteepness))

	if dir == shadeLight {
		reference := math.Max(autoShadeBaseFloor, math.Min(autoShadeBaseCeiling, autoShadeBaseSlope*(1-l)))
		l = math.Min(autoShadeMaxLightness, l+reference*widthMultiplier)
	} else {
		reference := math.Max(autoShadeBaseFloor, math.Min(autoShadeBaseCeiling, autoShadeBaseSlope*l))
		l = math.Max(autoShadeMinLightness, l-reference*widthMultiplier)
	}

	shade := colorful.Hcl(h, c, l)
	for !shade.IsValid() && c > 0 {
		c -= autoShadeChromaStep
		shade = colorful.Hcl(h, c, l)
	}

	return shade.Clamped()
}

// gradientCellCache memoizes interpolation results keyed on the RESOLVED stop colors
// (keyword and palette stops resolve before the key is built, so context changes miss
// the cache correctly), the cell count, the channel, and the TrueColor mode. Prompt
// rendering is single-threaded (see the terminal writer's package state), so a plain
// map suffices. Bounded to keep long-lived daemons from growing it unchecked.
var gradientCellCache = make(map[string][]Ansi)

const gradientCellCacheLimit = 128

func gradientKey(colors []colorful.Color, cells int, isBackground bool) string {
	buf := make([]byte, 0, 8+len(colors)*12)

	for _, clr := range colors {
		r, g, b := clr.RGB255()
		buf = strconv.AppendUint(buf, uint64(r), 10)
		buf = append(buf, ';')
		buf = strconv.AppendUint(buf, uint64(g), 10)
		buf = append(buf, ';')
		buf = strconv.AppendUint(buf, uint64(b), 10)
		buf = append(buf, ',')
	}

	buf = strconv.AppendInt(buf, int64(cells), 10)

	if isBackground {
		buf = append(buf, 'b')
	}

	if TrueColor {
		buf = append(buf, 't')
	}

	return string(buf)
}

func cacheGradientCells(colors []colorful.Color, cells int, isBackground bool, result []Ansi) []Ansi {
	if len(gradientCellCache) >= gradientCellCacheLimit {
		gradientCellCache = make(map[string][]Ansi)
	}

	gradientCellCache[gradientKey(colors, cells, isBackground)] = result
	return result
}

// parseTrueColor parses a truecolor ANSI payload ("38;2;r;g;b" or "48;2;r;g;b") back into
// a colorful.Color. The OS accent color resolves to this form instead of hex.
func parseTrueColor(c Ansi) (colorful.Color, bool) {
	parts := strings.Split(c.String(), ";")
	if len(parts) != 5 || (parts[0] != "38" && parts[0] != "48") || parts[1] != "2" {
		return colorful.Color{}, false
	}

	rgb := make([]uint8, 3)
	for i, part := range parts[2:] {
		val, err := strconv.ParseUint(part, 10, 8)
		if err != nil {
			return colorful.Color{}, false
		}

		rgb[i] = uint8(val)
	}

	return colorful.Color{R: float64(rgb[0]) / 255.0, G: float64(rgb[1]) / 255.0, B: float64(rgb[2]) / 255.0}, true
}

// ansiFromColorful converts an interpolated HCL color to a ready-to-print ANSI code, honoring
// the package-level TrueColor flag. The truecolor payload is built with strconv appends
// rather than gookit's fmt.Sprintf path: one allocation per cell instead of four, on what
// is the gradient hot path's dominant allocation site.
func ansiFromColorful(c colorful.Color, isBackground bool) Ansi {
	r, g, b := c.RGB255()

	if !TrueColor {
		return Ansi(color.RGB(r, g, b, isBackground).C256().String())
	}

	buf := make([]byte, 0, 16)

	if isBackground {
		buf = append(buf, "48;2;"...)
	} else {
		buf = append(buf, "38;2;"...)
	}

	buf = strconv.AppendUint(buf, uint64(r), 10)
	buf = append(buf, ';')
	buf = strconv.AppendUint(buf, uint64(g), 10)
	buf = append(buf, ';')
	buf = strconv.AppendUint(buf, uint64(b), 10)

	return Ansi(buf)
}
