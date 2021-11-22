package main

import (
	"fmt"
	"strings"
)

const (
	colorRegex = `<(?P<foreground>[^,>]+)?,?(?P<background>[^>]+)?>(?P<content>[^<]*)<\/>`
)

type promptWriter interface {
	write(background, foreground, text string)
	string() string
	reset()
	setColors(background, foreground string)
	setParentColors(background, foreground string)
	clearParentColors()
}

// AnsiWriter writes colorized strings
type AnsiWriter struct {
	builder            strings.Builder
	ansi               *ansiUtils
	terminalBackground string
	Colors             *Color
	ParentColors       *Color
	ansiColors         AnsiColors
}

type Color struct {
	Background string
	Foreground string
}

// AnsiColors is the interface that wraps AnsiColorFromString method.
//
// AnsiColorFromString gets the ANSI color code for a given color string.
// This can include a valid hex color in the format `#FFFFFF`,
// but also a name of one of the first 16 ANSI colors like `lightBlue`.
type AnsiColors interface {
	AnsiColorFromString(colorString string, isBackground bool) AnsiColor
}

// AnsiColor is an ANSI color code ready to be printed to the console.
// Example: "38;2;255;255;255", "48;2;255;255;255", "31", "95".
type AnsiColor string

const (
	emptyAnsiColor       = AnsiColor("")
	transparentAnsiColor = AnsiColor(Transparent)
)

func (c AnsiColor) IsEmpty() bool {
	return c == emptyAnsiColor
}

func (c AnsiColor) IsTransparent() bool {
	return c == transparentAnsiColor
}

const (
	// Transparent implies a transparent color
	Transparent = "transparent"
	// ParentBackground takes the previous segment's background color
	ParentBackground = "parentBackground"
	// ParentForeground takes the previous segment's color
	ParentForeground = "parentForeground"
	// Background takes the current segment's background color
	Background = "background"
	// Foreground takes the current segment's foreground color
	Foreground = "foreground"
)

func (a *AnsiWriter) setColors(background, foreground string) {
	a.Colors = &Color{
		Background: background,
		Foreground: foreground,
	}
}

func (a *AnsiWriter) setParentColors(background, foreground string) {
	a.ParentColors = &Color{
		Background: background,
		Foreground: foreground,
	}
}

func (a *AnsiWriter) clearParentColors() {
	a.ParentColors = nil
}

func (a *AnsiWriter) getAnsiFromColorString(colorString string, isBackground bool) AnsiColor {
	return a.ansiColors.AnsiColorFromString(colorString, isBackground)
}

func (a *AnsiWriter) writeColoredText(background, foreground AnsiColor, text string) {
	// Avoid emitting empty strings with color codes
	if text == "" || (foreground.IsTransparent() && background.IsTransparent()) {
		return
	}
	// default to white fg if empty, empty backgrond is supported
	if foreground.IsEmpty() {
		foreground = a.getAnsiFromColorString("white", false)
	}
	if foreground.IsTransparent() && !background.IsEmpty() && len(a.terminalBackground) != 0 {
		fgAnsiColor := a.getAnsiFromColorString(a.terminalBackground, false)
		coloredText := fmt.Sprintf(a.ansi.colorFull, background, fgAnsiColor, text)
		a.builder.WriteString(coloredText)
		return
	}
	if foreground.IsTransparent() && !background.IsEmpty() {
		coloredText := fmt.Sprintf(a.ansi.colorTransparent, background, text)
		a.builder.WriteString(coloredText)
		return
	} else if background.IsEmpty() || background.IsTransparent() {
		coloredText := fmt.Sprintf(a.ansi.colorSingle, foreground, text)
		a.builder.WriteString(coloredText)
		return
	}
	coloredText := fmt.Sprintf(a.ansi.colorFull, background, foreground, text)
	a.builder.WriteString(coloredText)
}

func (a *AnsiWriter) writeAndRemoveText(background, foreground AnsiColor, text, textToRemove, parentText string) string {
	a.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (a *AnsiWriter) write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}

	bgAnsi, fgAnsi := a.asAnsiColors(background, foreground)
	text = a.ansi.escapeText(text)
	text = a.ansi.formatText(text)
	text = a.ansi.generateHyperlink(text)

	// first we match for any potentially valid colors enclosed in <>
	// i.e., find color overrides
	overrides := findAllNamedRegexMatch(colorRegex, text)
	for _, override := range overrides {
		fgOverride := override["foreground"]
		bgOverride := override["background"]
		if fgOverride == Transparent && len(bgOverride) == 0 {
			bgOverride = background
		}
		bgOverrideAnsi, fgOverrideAnsi := a.asAnsiColors(bgOverride, fgOverride)
		// set colors if they are empty
		if bgOverrideAnsi.IsEmpty() {
			bgOverrideAnsi = bgAnsi
		}
		if fgOverrideAnsi.IsEmpty() {
			fgOverrideAnsi = fgAnsi
		}
		escapedTextSegment := override["text"]
		innerText := override["content"]
		textBeforeColorOverride := strings.Split(text, escapedTextSegment)[0]
		text = a.writeAndRemoveText(bgAnsi, fgAnsi, textBeforeColorOverride, textBeforeColorOverride, text)
		text = a.writeAndRemoveText(bgOverrideAnsi, fgOverrideAnsi, innerText, escapedTextSegment, text)
	}
	// color the remaining part of text with background and foreground
	a.writeColoredText(bgAnsi, fgAnsi, text)
}

func (a *AnsiWriter) asAnsiColors(background, foreground string) (AnsiColor, AnsiColor) {
	if backgroundValue, ok := a.isKeyword(background); ok {
		background = backgroundValue
	}
	if foregroundValue, ok := a.isKeyword(foreground); ok {
		foreground = foregroundValue
	}
	inverted := foreground == Transparent && len(background) != 0
	backgroundAnsi := a.getAnsiFromColorString(background, !inverted)
	foregroundAnsi := a.getAnsiFromColorString(foreground, false)
	return backgroundAnsi, foregroundAnsi
}

func (a *AnsiWriter) isKeyword(color string) (string, bool) {
	switch {
	case color == Background:
		return a.Colors.Background, true
	case color == Foreground:
		return a.Colors.Foreground, true
	case color == ParentBackground && a.ParentColors != nil:
		return a.ParentColors.Background, true
	case color == ParentForeground && a.ParentColors != nil:
		return a.ParentColors.Foreground, true
	case (color == ParentBackground || color == ParentForeground) && a.ParentColors == nil:
		return Transparent, true
	default:
		return "", false
	}
}

func (a *AnsiWriter) string() string {
	return a.builder.String()
}

func (a *AnsiWriter) reset() {
	a.builder.Reset()
}
