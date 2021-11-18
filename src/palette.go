package main

import (
	"fmt"
	"sort"
	"strings"
)

type Palette map[string]string

const (
	paletteColorMissingErrorTemplate       = "palette: requested color %s does not exist in palette of colors %s"
	paletteRecursiveReferenceErrorTemplate = "palette: resolution of color %s returned palette reference %s; recursive references are not supported"
)

var (
	paletteColorPrefixes = [...]string{"palette:", "p:"}
)

type PaletteColorMissingError struct {
	Reference string
	palette   Palette
}

func (p *PaletteColorMissingError) Error() string {
	colorNames := make([]string, 0, len(p.palette))
	for k := range p.palette {
		colorNames = append(colorNames, k)
	}
	sort.Strings(colorNames)
	allColors := strings.Join(colorNames, ",")
	errorStr := fmt.Sprintf(paletteColorMissingErrorTemplate, p.Reference, allColors)
	return errorStr
}

type PaletteRecursiveReferenceError struct {
	Reference string
	Value     string
}

func (p *PaletteRecursiveReferenceError) Error() string {
	errorStr := fmt.Sprintf(paletteRecursiveReferenceErrorTemplate, p.Reference, p.Value)
	return errorStr
}

func (p Palette) resolveColor(colorName string) (string, error) {
	palettePrefix := p.checkPalettePrefix(colorName)

	// colorName is not a palette reference, return it as is
	if palettePrefix == "" {
		return colorName, nil
	}

	paletteName := strings.ReplaceAll(colorName, palettePrefix, "")

	if paletteColor, ok := p[paletteName]; ok {
		if palettePrefix = p.checkPalettePrefix(paletteColor); palettePrefix != "" {
			return "", &PaletteRecursiveReferenceError{Reference: paletteName, Value: paletteColor}
		}

		return paletteColor, nil
	}

	return "", &PaletteColorMissingError{Reference: paletteName, palette: p}
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

// maybeResolveColor wraps resolveColor and silences possible errors, returning
// Transparent color by default, as a Block does not know how to handle color errors.
func (p Palette) maybeResolveColor(colorName string) string {
	color, err := p.resolveColor(colorName)
	if err != nil {
		return Transparent
	}

	return color
}
