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
	setColors(background, foreground string)
	setParentColors(background, foreground string)
}

// AnsiColor writes colorized strings
type AnsiColor struct {
	builder            strings.Builder
	ansi               *ansiUtils
	terminalBackground string
	Colors             *Color
	ParentColors       *Color
}

type Color struct {
	Background string
	Foreground string
}

const (
	// Transparent implies a transparent color
	Transparent = "transparent"
	// Inherit takes the previous segment's color
	Inherit = "inherit"
	// Background takes the current segment's background color
	Background = "background"
	// Foreground takes the current segment's foreground color
	Foreground = "foreground"
)

func (a *AnsiColor) setColors(background, foreground string) {
	a.Colors = &Color{
		Background: background,
		Foreground: foreground,
	}
}

func (a *AnsiColor) setParentColors(background, foreground string) {
	a.ParentColors = &Color{
		Background: background,
		Foreground: foreground,
	}
}

// Gets the ANSI color code for a given color string.
// This can include a valid hex color in the format `#FFFFFF`,
// but also a name of one of the first 16 ANSI colors like `lightBlue`.
func (a *AnsiColor) getAnsiFromColorString(colorString string, isBackground bool) string {
	if colorString == Transparent || len(colorString) == 0 {
		return colorString
	}
	colorFromName, err := getColorFromName(colorString, isBackground)
	if err == nil {
		return colorFromName
	}
	style := color.HEX(colorString, isBackground)
	if style.IsEmpty() {
		return ""
	}
	return style.String()
}

func (a *AnsiColor) writeColoredText(background, foreground, text string) {
	// Avoid emitting empty strings with color codes
	if text == "" || (foreground == Transparent && background == Transparent) {
		return
	}
	// default to white fg if empty, empty backgrond is supported
	if len(foreground) == 0 {
		foreground = a.getAnsiFromColorString("white", false)
	}
	if foreground == Transparent && len(background) != 0 && len(a.terminalBackground) != 0 {
		fgAnsiColor := a.getAnsiFromColorString(a.terminalBackground, false)
		coloredText := fmt.Sprintf(a.ansi.colorFull, background, fgAnsiColor, text)
		a.builder.WriteString(coloredText)
		return
	}
	if foreground == Transparent && len(background) != 0 {
		coloredText := fmt.Sprintf(a.ansi.colorTransparent, background, text)
		a.builder.WriteString(coloredText)
		return
	} else if len(background) == 0 || background == Transparent {
		coloredText := fmt.Sprintf(a.ansi.colorSingle, foreground, text)
		a.builder.WriteString(coloredText)
		return
	}
	coloredText := fmt.Sprintf(a.ansi.colorFull, background, foreground, text)
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

	getAnsiColors := func(background, foreground string) (string, string) {
		if background == Background {
			background = a.Colors.Background
		}
		if background == Foreground {
			background = a.Colors.Foreground
		}
		if foreground == Foreground {
			foreground = a.Colors.Foreground
		}
		if foreground == Background {
			foreground = a.Colors.Background
		}
		if background == Inherit && a.ParentColors != nil {
			background = a.ParentColors.Background
		}
		if background == Inherit && a.ParentColors == nil {
			background = Transparent
		}
		if foreground == Inherit && a.ParentColors != nil {
			foreground = a.ParentColors.Foreground
		}
		if foreground == Inherit && a.ParentColors == nil {
			foreground = Transparent
		}
		inverted := foreground == Transparent && len(background) != 0
		background = a.getAnsiFromColorString(background, !inverted)
		foreground = a.getAnsiFromColorString(foreground, false)
		return background, foreground
	}

	bgAnsi, fgAnsi := getAnsiColors(background, foreground)

	text = a.ansi.escapeText(text)
	text = a.ansi.formatText(text)
	text = a.ansi.generateHyperlink(text)

	// first we match for any potentially valid colors enclosed in <>
	match := findAllNamedRegexMatch(`<(?P<foreground>[^,>]+)?,?(?P<background>[^>]+)?>(?P<content>[^<]*)<\/>`, text)
	for i := range match {
		fg := match[i]["foreground"]
		bg := match[i]["background"]
		if fg == Transparent && len(bg) == 0 {
			bg = background
		}
		bg, fg = getAnsiColors(bg, fg)
		// set colors if they are empty
		if len(bg) == 0 {
			bg = bgAnsi
		}
		if len(fg) == 0 {
			fg = fgAnsi
		}
		escapedTextSegment := match[i]["text"]
		innerText := match[i]["content"]
		textBeforeColorOverride := strings.Split(text, escapedTextSegment)[0]
		text = a.writeAndRemoveText(bgAnsi, fgAnsi, textBeforeColorOverride, textBeforeColorOverride, text)
		text = a.writeAndRemoveText(bg, fg, innerText, escapedTextSegment, text)
	}
	// color the remaining part of text with background and foreground
	a.writeColoredText(bgAnsi, fgAnsi, text)
}

func (a *AnsiColor) string() string {
	return a.builder.String()
}

func (a *AnsiColor) reset() {
	a.builder.Reset()
}
