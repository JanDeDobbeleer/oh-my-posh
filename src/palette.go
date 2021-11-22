package main

import (
	"fmt"
	"sort"
	"strings"
)

type Palette map[string]string

const (
	paletteKeyPrefix         = "p:"
	paletteKeyError          = "palette: requested color %s does not exist in palette of colors %s"
	paletteRecursiveKeyError = "palette: resolution of color %s returned palette reference %s; recursive references are not supported"
)

// ResolveColor gets a color value from the palette using given colorName.
// If colorName is not a palette reference, it is returned as is.
func (p Palette) ResolveColor(colorName string) (string, error) {
	key, ok := p.asPaletteKey(colorName)
	// colorName is not a palette key, return it as is
	if !ok {
		return colorName, nil
	}

	color, ok := p[key]
	if !ok {
		return "", &PaletteKeyError{Key: key, palette: p}
	}

	if _, ok = p.asPaletteKey(color); ok {
		return "", &PaletteRecursiveKeyError{Key: key, Value: color}
	}

	return color, nil
}

func (p Palette) asPaletteKey(colorName string) (string, bool) {
	if !strings.HasPrefix(colorName, paletteKeyPrefix) {
		return "", false
	}

	key := strings.TrimPrefix(colorName, paletteKeyPrefix)

	return key, true
}

// PaletteKeyError records the missing Palette key.
type PaletteKeyError struct {
	Key     string
	palette Palette
}

func (p *PaletteKeyError) Error() string {
	keys := make([]string, 0, len(p.palette))
	for key := range p.palette {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	allColors := strings.Join(keys, ",")
	errorStr := fmt.Sprintf(paletteKeyError, p.Key, allColors)
	return errorStr
}

// PaletteRecursiveKeyError records the Palette key and resolved color value (which
// is also a Palette key)
type PaletteRecursiveKeyError struct {
	Key   string
	Value string
}

func (p *PaletteRecursiveKeyError) Error() string {
	errorStr := fmt.Sprintf(paletteRecursiveKeyError, p.Key, p.Value)
	return errorStr
}

// maybeResolveColor wraps resolveColor and silences possible errors, returning
// Transparent color by default, as a Block does not know how to handle color errors.
func (p Palette) MaybeResolveColor(colorName string) string {
	color, err := p.ResolveColor(colorName)
	if err != nil {
		return Transparent
	}

	return color
}
