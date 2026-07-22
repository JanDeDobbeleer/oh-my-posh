package color

import (
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	gradientPrefix = "linear-gradient("
	gradientSuffix = ")"
)

// IsGradient reports whether c is a gradient definition, e.g. `linear-gradient(#FF0000, #0000FF)`.
func (c Ansi) IsGradient() bool {
	return strings.HasPrefix(c.String(), gradientPrefix)
}

// GradientStops performs syntax parsing only: it splits the comma-separated stop list inside
// `linear-gradient(...)` and trims whitespace around each stop. It does not resolve palette
// references or validate that a stop is a color. Returns nil when c is not a gradient, the
// closing paren is missing, the body contains a nested paren (angle/direction syntax is
// reserved but not implemented), or any stop is empty.
func (c Ansi) GradientStops() []Ansi {
	if !c.IsGradient() {
		return nil
	}

	value := c.String()
	if !strings.HasSuffix(value, gradientSuffix) {
		return nil
	}

	body := value[len(gradientPrefix) : len(value)-len(gradientSuffix)]
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
// a gradient, or when the gradient syntax is invalid.
func (c Ansi) GradientLast() Ansi {
	stops := c.GradientStops()
	if len(stops) == 0 {
		return c
	}

	return stops[len(stops)-1]
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
		log.Errorf("gradient %s: invalid syntax, expected linear-gradient(stop, stop, ...)", c)
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
