package color

import (
	"fmt"
	"sort"
	"strings"
)

type Palette map[Ansi]Ansi

const (
	paletteKeyPrefix         = "p:"
	paletteKeyError          = "palette: requested color %s does not exist in palette of colors %s"
	paletteMaxRecursionDepth = 3 // allows 3 or less recusive resolutions
	paletteRecursiveKeyError = "palette: recursive resolution of color %s returned palette reference %s and reached recursion depth %d"
)

// ResolveColor gets a color value from the palette using given colorName.
// If colorName is not a palette reference, it is returned as is.
func (p Palette) ResolveColor(colorName Ansi) (Ansi, error) {
	return p.resolveColor(colorName, 1, &colorName)
}

// originalColorName is a pointer to save allocations
func (p Palette) resolveColor(colorName Ansi, depth int, originalColorName *Ansi) (Ansi, error) {
	key, ok := asPaletteKey(colorName)
	// colorName is not a palette key, return it as is
	if !ok {
		return colorName, nil
	}

	color, ok := p[key]
	if !ok {
		return "", &PaletteKeyError{Key: key, palette: p}
	}

	if _, isKey := isPaletteKey(color); isKey {
		if depth > paletteMaxRecursionDepth {
			return "", &PaletteRecursiveKeyError{Key: *originalColorName, Value: color, depth: depth}
		}

		return p.resolveColor(color, depth+1, originalColorName)
	}

	return color, nil
}

func asPaletteKey(colorName Ansi) (Ansi, bool) {
	prefix, isKey := isPaletteKey(colorName)
	if !isKey {
		return "", false
	}

	key := strings.TrimPrefix(colorName.String(), prefix.String())

	return Ansi(key), true
}

func isPaletteKey(colorName Ansi) (Ansi, bool) {
	return paletteKeyPrefix, strings.HasPrefix(colorName.String(), paletteKeyPrefix)
}

// PaletteKeyError records the missing Palette key.
type PaletteKeyError struct {
	palette Palette
	Key     Ansi
}

func (p *PaletteKeyError) Error() string {
	keys := make([]string, 0, len(p.palette))
	for key := range p.palette {
		keys = append(keys, key.String())
	}
	sort.Strings(keys)
	allColors := strings.Join(keys, ",")
	errorStr := fmt.Sprintf(paletteKeyError, p.Key, allColors)
	return errorStr
}

// PaletteRecursiveKeyError records the Palette key and resolved color value (which
// is also a Palette key)
type PaletteRecursiveKeyError struct {
	Key   Ansi
	Value Ansi
	depth int
}

func (p *PaletteRecursiveKeyError) Error() string {
	errorStr := fmt.Sprintf(paletteRecursiveKeyError, p.Key, p.Value, p.depth)
	return errorStr
}

// maybeResolveColor wraps resolveColor and silences possible errors, returning
// Transparent color by default, as a Block does not know how to handle color errors.
func (p Palette) MaybeResolveColor(colorName Ansi) Ansi {
	color, err := p.ResolveColor(colorName)
	if err != nil {
		return ""
	}

	return color
}
