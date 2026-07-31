package color

import (
	"math"
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

const (
	darkenPrefix  = "darken("
	lightenPrefix = "lighten("
)

// IsShade reports whether c is a darken(...)/lighten(...) single-color shade call.
func (c Ansi) IsShade() bool {
	_, ok := c.shadeCallDirection()
	return ok
}

func (c Ansi) shadeCallDirection() (shadeDirection, bool) {
	s := c.String()

	switch {
	case strings.HasPrefix(s, darkenPrefix):
		return shadeDark, true
	case strings.HasPrefix(s, lightenPrefix):
		return shadeLight, true
	default:
		return shadeNone, false
	}
}

// ShadeArgs parses darken(<color>, <percent>)/lighten(<color>, <percent>) into its color
// argument and a percent in [0, 100]. ok is false when c isn't a shade call, its closing
// paren is missing, the body isn't exactly two comma-separated arguments, or the percent
// argument doesn't parse as a number in [0, 100].
func (c Ansi) ShadeArgs() (dir shadeDirection, inner Ansi, percent float64, ok bool) {
	dir, isShade := c.shadeCallDirection()
	if !isShade {
		return shadeNone, "", 0, false
	}

	s := c.String()
	if !strings.HasSuffix(s, gradientSuffix) {
		return shadeNone, "", 0, false
	}

	prefix := darkenPrefix
	if dir == shadeLight {
		prefix = lightenPrefix
	}

	body := s[len(prefix) : len(s)-len(gradientSuffix)]
	if strings.ContainsAny(body, "()") {
		return shadeNone, "", 0, false
	}

	parts := strings.Split(body, ",")
	if len(parts) != 2 {
		return shadeNone, "", 0, false
	}

	innerPart := strings.TrimSpace(parts[0])
	percentPart := strings.TrimSpace(parts[1])

	if innerPart == "" || percentPart == "" {
		return shadeNone, "", 0, false
	}

	pct, err := strconv.ParseFloat(percentPart, 64)
	if err != nil || pct < 0 || pct > 100 {
		return shadeNone, "", 0, false
	}

	return dir, Ansi(innerPart), pct, true
}

// withShadeCall rebuilds a darken/lighten call from its parts, so a palette reference in
// the color argument can be resolved to a concrete color before Defaults.ToAnsi sees it.
func withShadeCall(dir shadeDirection, inner Ansi, percent float64) Ansi {
	prefix := darkenPrefix
	if dir == shadeLight {
		prefix = lightenPrefix
	}

	return Ansi(prefix + inner.String() + ", " + strconv.FormatFloat(percent, 'g', -1, 64) + gradientSuffix)
}

// shadeHex shifts inner's HCL lightness toward black (shadeDark) or white (shadeLight)
// by percent - the same technique as autoShade, but driven by a caller-chosen
// percentage instead of the auto-tuned curve.
//
// ok is false when inner isn't a hex color: darken()/lighten() only operate on colors
// oh-my-posh knows the actual RGB of, which rules out named ANSI/terminal colors, 256
// palette indices, and keywords - their real color is decided by the terminal's own
// scheme (or, for keywords, by segment context this single-color call has no access
// to), not by oh-my-posh. Define the color in a palette entry with a literal hex value
// and reference it with darken(p:name, n)/lighten(p:name, n) instead.
func shadeHex(inner Ansi, dir shadeDirection, percent float64) (string, bool) {
	innerStr := inner.String()
	if !strings.HasPrefix(innerStr, "#") {
		return "", false
	}

	base, err := colorful.Hex(innerStr)
	if err != nil {
		return "", false
	}

	h, c, l := base.Hcl()
	pct := percent / 100

	if dir == shadeLight {
		l += (1 - l) * pct
	}

	if dir != shadeLight {
		l -= l * pct
	}

	l = math.Max(0, math.Min(1, l))

	shade := colorful.Hcl(h, c, l)
	for !shade.IsValid() && c > 0 {
		c -= autoShadeChromaStep
		shade = colorful.Hcl(h, c, l)
	}

	return shade.Clamped().Hex(), true
}
