package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Palette map[string]string

const (
	paletteColorMissingErrorTemplate        = "palette: requested color %s does not exist in palette of colors %s"
	paletteRecursiveResolutionErrorTemplate = "palette: resolution of color %s returned palette reference %s; recursive resolution is not supported"
)

var (
	paletteColorPrefixes = [...]string{"palette:", "p:"}
)

func (p Palette) resolveColor(colorName string) (string, error) {
	palettePrefix := p.checkPalettePrefix(colorName)

	// colorName is not a palette reference, return it as is
	if palettePrefix == "" {
		return colorName, nil
	}

	paletteName := strings.ReplaceAll(colorName, palettePrefix, "")

	if paletteColor, ok := p[paletteName]; ok {
		palettePrefix = p.checkPalettePrefix(paletteColor)

		if palettePrefix != "" {
			return "", p.reportRecursiveResolution(paletteName, paletteColor)
		}

		return paletteColor, nil
	}

	return "", p.reportColorMissing(paletteName)
}

func (p Palette) checkPalettePrefix(colorName string) (selectedPalettePrefix string) {
	for _, palettePrefix := range paletteColorPrefixes {
		if strings.HasPrefix(colorName, palettePrefix) {
			selectedPalettePrefix = palettePrefix
			break
		}
	}

	return
}

func (p Palette) reportRecursiveResolution(colorName, colorValue string) error {
	errorStr := fmt.Sprintf(paletteRecursiveResolutionErrorTemplate, colorName, colorValue)
	return errors.New(errorStr)
}

func (p Palette) reportColorMissing(colorName string) error {
	colorNames := make([]string, 0, len(p))
	for k := range p {
		colorNames = append(colorNames, k)
	}
	sort.Strings(colorNames)
	allColors := strings.Join(colorNames, ",")
	errorStr := fmt.Sprintf(paletteColorMissingErrorTemplate, colorName, allColors)
	return errors.New(errorStr)
}

// maybeResolveColor wraps resolveColor and silences possible errors, returning
// Transparent color by default, as a Block does not know how to handle color errors.
func (p Palette) maybeResolveColor(colorName string) string {
	color, err := p.resolveColor(colorName)
	if err != nil {
		return Transparent
	}

	return color
}
