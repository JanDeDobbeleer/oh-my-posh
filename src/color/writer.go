package color

import (
	"fmt"
	"oh-my-posh/regex"
	"strings"
)

const (
	colorRegex = `<(?P<foreground>[^,>]+)?,?(?P<background>[^>]+)?>(?P<content>[^<]*)<\/>`
)

type Writer interface {
	Write(background, foreground, text string)
	String() string
	Reset()
	SetColors(background, foreground string)
	SetParentColors(background, foreground string)
	ClearParentColors()
}

// AnsiWriter writes colorized ANSI strings
type AnsiWriter struct {
	Ansi               *Ansi
	TerminalBackground string
	Colors             *Color
	ParentColors       []*Color
	AnsiColors         AnsiColors

	builder strings.Builder
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

func (a *AnsiWriter) SetColors(background, foreground string) {
	a.Colors = &Color{
		Background: background,
		Foreground: foreground,
	}
}

func (a *AnsiWriter) SetParentColors(background, foreground string) {
	if a.ParentColors == nil {
		a.ParentColors = make([]*Color, 0)
	}
	a.ParentColors = append([]*Color{{
		Background: background,
		Foreground: foreground,
	}}, a.ParentColors...)
}

func (a *AnsiWriter) ClearParentColors() {
	a.ParentColors = nil
}

func (a *AnsiWriter) getAnsiFromColorString(colorString string, isBackground bool) AnsiColor {
	return a.AnsiColors.AnsiColorFromString(colorString, isBackground)
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
	if foreground.IsTransparent() && !background.IsEmpty() && len(a.TerminalBackground) != 0 {
		fgAnsiColor := a.getAnsiFromColorString(a.TerminalBackground, false)
		coloredText := fmt.Sprintf(a.Ansi.colorFull, background, fgAnsiColor, text)
		a.builder.WriteString(coloredText)
		return
	}
	if foreground.IsTransparent() && !background.IsEmpty() {
		coloredText := fmt.Sprintf(a.Ansi.colorTransparent, background, text)
		a.builder.WriteString(coloredText)
		return
	} else if background.IsEmpty() || background.IsTransparent() {
		coloredText := fmt.Sprintf(a.Ansi.colorSingle, foreground, text)
		a.builder.WriteString(coloredText)
		return
	}
	coloredText := fmt.Sprintf(a.Ansi.colorFull, background, foreground, text)
	a.builder.WriteString(coloredText)
}

func (a *AnsiWriter) writeAndRemoveText(background, foreground AnsiColor, text, textToRemove, parentText string) string {
	a.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (a *AnsiWriter) Write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}

	bgAnsi, fgAnsi := a.asAnsiColors(background, foreground)
	text = a.Ansi.EscapeText(text)
	text = a.Ansi.formatText(text)
	text = a.Ansi.generateHyperlink(text)

	// first we match for any potentially valid colors enclosed in <>
	// i.e., find color overrides
	overrides := regex.FindAllNamedRegexMatch(colorRegex, text)
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
	background = a.expandKeyword(background)
	foreground = a.expandKeyword(foreground)
	inverted := foreground == Transparent && len(background) != 0
	backgroundAnsi := a.getAnsiFromColorString(background, !inverted)
	foregroundAnsi := a.getAnsiFromColorString(foreground, false)
	return backgroundAnsi, foregroundAnsi
}

func (a *AnsiWriter) isKeyword(color string) bool {
	switch color {
	case Transparent, ParentBackground, ParentForeground, Background, Foreground:
		return true
	default:
		return false
	}
}

func (a *AnsiWriter) expandKeyword(keyword string) string {
	resolveParentColor := func(keyword string) string {
		for _, color := range a.ParentColors {
			if color == nil {
				return Transparent
			}
			switch keyword {
			case ParentBackground:
				keyword = color.Background
			case ParentForeground:
				keyword = color.Foreground
			default:
				return keyword
			}
		}
		return keyword
	}
	resolveKeyword := func(keyword string) string {
		switch {
		case keyword == Background && a.Colors != nil:
			return a.Colors.Background
		case keyword == Foreground && a.Colors != nil:
			return a.Colors.Foreground
		case (keyword == ParentBackground || keyword == ParentForeground) && a.ParentColors != nil:
			return resolveParentColor(keyword)
		default:
			return Transparent
		}
	}
	for ok := a.isKeyword(keyword); ok; ok = a.isKeyword(keyword) {
		resolved := resolveKeyword(keyword)
		if resolved == keyword {
			break
		}
		keyword = resolved
	}
	return keyword
}

func (a *AnsiWriter) String() string {
	return a.builder.String()
}

func (a *AnsiWriter) Reset() {
	a.builder.Reset()
}
