package color

import (
	"fmt"
	"strconv"
	"strings"
)

// TmuxColors wraps a String implementation and converts ANSI color codes
// to tmux native format strings (e.g., "fg=#rrggbb", "bg=colour42").
// This allows the rendering engine to emit tmux #[fg=...] / #[bg=...] tokens
// instead of ANSI escape sequences when rendering for the tmux status bar.
type TmuxColors struct {
	inner String
}

// NewTmuxColors creates a TmuxColors that wraps the given String implementation.
func NewTmuxColors(inner String) *TmuxColors {
	return &TmuxColors{inner: inner}
}

func (t *TmuxColors) Resolve(colorString Ansi) (Ansi, error) {
	return t.inner.Resolve(colorString)
}

func (t *TmuxColors) ToAnsi(c Ansi, isBackground bool) Ansi {
	ansiCode := t.inner.ToAnsi(c, isBackground)
	return convertAnsiToTmux(ansiCode, isBackground)
}

// ansiCodeToColorName maps ANSI numeric codes (both fg and bg variants) to color names.
var ansiCodeToColorName = func() map[Ansi]string {
	m := make(map[Ansi]string, len(ansiColorCodes)*2)
	for name, codes := range ansiColorCodes {
		m[codes[0]] = string(name) // fg code → name
		m[codes[1]] = string(name) // bg code → name
	}
	return m
}()

// convertAnsiToTmux converts an ANSI color code string (as returned by Defaults.ToAnsi)
// into a tmux format token like "fg=#rrggbb" or "bg=colour42".
func convertAnsiToTmux(code Ansi, isBackground bool) Ansi {
	if code.IsEmpty() || code.IsTransparent() {
		return code
	}

	prefix := "fg"
	if isBackground {
		prefix = "bg"
	}

	s := string(code)

	// True color: "38;2;R;G;B" (fg) or "48;2;R;G;B" (bg)
	if (strings.HasPrefix(s, "38;2;") || strings.HasPrefix(s, "48;2;")) && strings.Count(s, ";") == 4 {
		rgb := s[5:]
		parts := strings.SplitN(rgb, ";", 3)
		if len(parts) == 3 {
			r, errR := strconv.ParseUint(parts[0], 10, 8)
			g, errG := strconv.ParseUint(parts[1], 10, 8)
			b, errB := strconv.ParseUint(parts[2], 10, 8)
			if errR == nil && errG == nil && errB == nil {
				return Ansi(fmt.Sprintf("%s=#%02x%02x%02x", prefix, r, g, b))
			}
		}
	}

	// 256-color: "38;5;N" (fg) or "48;5;N" (bg)
	if strings.HasPrefix(s, "38;5;") || strings.HasPrefix(s, "48;5;") {
		n := s[5:]
		return Ansi(fmt.Sprintf("%s=colour%s", prefix, n))
	}

	// Named color (e.g., "30"=black fg, "40"=black bg, "90"=darkGray fg, …)
	if name, ok := ansiCodeToColorName[code]; ok {
		return Ansi(fmt.Sprintf("%s=%s", prefix, name))
	}

	return emptyColor
}
