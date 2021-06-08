package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gookit/color"
)

var (
	// Map for color names and their respective foreground [0] or background [1] color codes
	colorMap = map[string][2]string{
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

// Returns the color code for a given color name
func getColorFromName(colorName string, isBackground bool) (string, error) {
	colorMapOffset := 0
	if isBackground {
		colorMapOffset = 1
	}
	if colorCodes, found := colorMap[colorName]; found {
		return colorCodes[colorMapOffset], nil
	}
	return "", errors.New("color name does not exist")
}

type colorWriter interface {
	write(background, foreground, text string)
	string() string
	reset()
}

// AnsiColor writes colorized strings
type AnsiColor struct {
	builder            strings.Builder
	ansi               *ansiUtils
	terminalBackground string
}

const (
	// Transparent implies a transparent color
	Transparent = "transparent"
)

// Gets the ANSI color code for a given color string.
// This can include a valid hex color in the format `#FFFFFF`,
// but also a name of one of the first 16 ANSI colors like `lightBlue`.
func (a *AnsiColor) getAnsiFromColorString(colorString string, isBackground bool) string {
	colorFromName, err := getColorFromName(colorString, isBackground)
	if err == nil {
		return colorFromName
	}
	style := color.HEX(colorString, isBackground)
	return style.Code()
}

func (a *AnsiColor) writeColoredText(background, foreground, text string) {
	// Avoid emitting empty strings with color codes
	if text == "" {
		return
	}
	if foreground == Transparent && background != "" && a.terminalBackground != "" {
		bgAnsiColor := a.getAnsiFromColorString(background, true)
		fgAnsiColor := a.getAnsiFromColorString(a.terminalBackground, false)
		coloredText := fmt.Sprintf(a.ansi.colorFull, bgAnsiColor, fgAnsiColor, text)
		a.builder.WriteString(coloredText)
		return
	}
	if foreground == Transparent && background != "" {
		ansiColor := a.getAnsiFromColorString(background, false)
		coloredText := fmt.Sprintf(a.ansi.colorTransparent, ansiColor, text)
		a.builder.WriteString(coloredText)
		return
	} else if background == "" || background == Transparent {
		ansiColor := a.getAnsiFromColorString(foreground, false)
		coloredText := fmt.Sprintf(a.ansi.colorSingle, ansiColor, text)
		a.builder.WriteString(coloredText)
		return
	}
	bgAnsiColor := a.getAnsiFromColorString(background, true)
	fgAnsiColor := a.getAnsiFromColorString(foreground, false)
	coloredText := fmt.Sprintf(a.ansi.colorFull, bgAnsiColor, fgAnsiColor, text)
	a.builder.WriteString(coloredText)
}

func (a *AnsiColor) writeAndRemoveText(background, foreground, text, textToRemove, parentText string) string {
	a.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (a *AnsiColor) write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}
	text = a.ansi.escapeText(text)
	text = a.ansi.formatText(text)
	text = a.ansi.generateHyperlink(text)

	// first we match for any potentially valid colors enclosed in <>
	match := findAllNamedRegexMatch(`<(?P<foreground>[^,>]+)?,?(?P<background>[^>]+)?>(?P<content>[^<]*)<\/>`, text)
	for i := range match {
		extractedForegroundColor := match[i]["foreground"]
		extractedBackgroundColor := match[i]["background"]
		if col := a.getAnsiFromColorString(extractedForegroundColor, false); col == "" && extractedForegroundColor != Transparent && len(extractedBackgroundColor) == 0 {
			continue // we skip invalid colors
		}
		if col := a.getAnsiFromColorString(extractedBackgroundColor, false); col == "" && extractedBackgroundColor != Transparent && len(extractedForegroundColor) == 0 {
			continue // we skip invalid colors
		}
		// reuse function colors if only one was specified
		if len(extractedBackgroundColor) == 0 {
			extractedBackgroundColor = background
		}
		if len(extractedForegroundColor) == 0 {
			extractedForegroundColor = foreground
		}
		escapedTextSegment := match[i]["text"]
		innerText := match[i]["content"]
		textBeforeColorOverride := strings.Split(text, escapedTextSegment)[0]
		text = a.writeAndRemoveText(background, foreground, textBeforeColorOverride, textBeforeColorOverride, text)
		text = a.writeAndRemoveText(extractedBackgroundColor, extractedForegroundColor, innerText, escapedTextSegment, text)
	}
	// color the remaining part of text with background and foreground
	a.writeColoredText(background, foreground, text)
}

func (a *AnsiColor) string() string {
	return a.builder.String()
}

func (a *AnsiColor) reset() {
	a.builder.Reset()
}
