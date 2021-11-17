package main

import (
	"strings"
)

type Palette map[string]string

const (
	paletteColorPrefix = "palette:"
)

func (p Palette) resolveColor(colorName string) (string, error) {
	if !strings.HasPrefix(colorName, paletteColorPrefix) {
		return colorName, nil
	}

	paletteName := strings.ReplaceAll(colorName, paletteColorPrefix, "")

	if paletteColor, ok := p[paletteName]; ok {
		return paletteColor, nil
	}

	return "", nil
}
