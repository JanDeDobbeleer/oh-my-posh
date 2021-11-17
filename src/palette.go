package main

import (
	"strings"
)

type Palette map[string]string

const (
	paletteColorPrefix = "palette:"
)

func (p Palette) resolveColor(colorName string) (string, error) {
	paletteName := strings.ReplaceAll(colorName, paletteColorPrefix, "")

	if paletteColor, ok := p[paletteName]; ok {
		return paletteColor, nil
	}

	return "", nil
}
