package main

import (
	"errors"
	"fmt"

	"github.com/gookit/color"
)

const (
	foregroundIndex = 0
	backgroundIndex = 1
)

var (
	// Map for color names and their respective foreground [0] or background [1] color codes
	ansiColors = map[string][2]AnsiColor{
		"black":        {"30", "40"},
		"red":          {"31", "41"},
		"green":        {"32", "42"},
		"yellow":       {"33", "43"},
		"blue":         {"34", "44"},
		"magenta":      {"35", "45"},
		"cyan":         {"36", "46"},
		"white":        {"37", "47"},
		"default":      {"39", "49"},
		"darkGray":     {"90", "100"},
		"lightRed":     {"91", "101"},
		"lightGreen":   {"92", "102"},
		"lightYellow":  {"93", "103"},
		"lightBlue":    {"94", "104"},
		"lightMagenta": {"95", "105"},
		"lightCyan":    {"96", "106"},
		"lightWhite":   {"97", "107"},
	}
)

type DefaultAnsiColors struct{}

func (_ *DefaultAnsiColors) AnsiColorFromString(colorString string, isBackground bool) AnsiColor {
	if len(colorString) == 0 {
		return emptyAnsiColor
	}
	if colorString == Transparent {
		return transparentAnsiColor
	}
	colorFromName, err := getAnsiColorFromName(colorString, isBackground)
	if err == nil {
		return colorFromName
	}
	style := color.HEX(colorString, isBackground)
	if style.IsEmpty() {
		return emptyAnsiColor
	}
	return AnsiColor(style.String())
}

// getAnsiColorFromName returns the color code for a given color name if the name is
// knows ANSI color name.
func getAnsiColorFromName(colorName string, isBackground bool) (AnsiColor, error) {
	if colorCodes, found := ansiColors[colorName]; found {
		if isBackground {
			return colorCodes[backgroundIndex], nil
		}
		return colorCodes[foregroundIndex], nil
	}
	return "", errors.New(fmt.Sprintf("color name %s does not exist", colorName))
}

func isAnsiColorName(colorString string) bool {
	_, ok := ansiColors[colorString]
	return ok
}
