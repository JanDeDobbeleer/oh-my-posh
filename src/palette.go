package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Palette map[string]string

const (
	paletteColorMissingErrorTemplate = "palette: requested color %s does not exist in palette of colors %s"
)

var (
	paletteColorPrefixes = [...]string{"palette:", "p:"}
)

func (p Palette) resolveColor(colorName string) (string, error) {
	var selectedColorPrefix string

	for _, paletteColorPrefix := range paletteColorPrefixes {
		if strings.HasPrefix(colorName, paletteColorPrefix) {
			selectedColorPrefix = paletteColorPrefix
			break
		}
	}

	// colorName is not a palette reference, return it as is
	if selectedColorPrefix == "" {
		return colorName, nil
	}

	paletteName := strings.ReplaceAll(colorName, selectedColorPrefix, "")

	if paletteColor, ok := p[paletteName]; ok {
		return paletteColor, nil
	}

	return "", p.reportColorMissing(paletteName)
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
