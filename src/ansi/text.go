package ansi

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/regex"

	"github.com/mattn/go-runewidth"
)

func init() { //nolint:gochecknoinits
	runewidth.DefaultCondition.EastAsianWidth = false
}

func (a *Writer) MeasureText(text string) int {
	// skip strings with ANSI
	if !strings.Contains(text, "\x1b") {
		text = a.TrimEscapeSequences(text)
		length := runewidth.StringWidth(text)
		return length
	}
	if strings.Contains(text, "\x1b]8;;") {
		matches := regex.FindAllNamedRegexMatch(a.hyperlinkRegex, text)
		for _, match := range matches {
			text = strings.ReplaceAll(text, match["STR"], match["TEXT"])
		}
	}
	text = a.trimAnsi(text)
	text = a.TrimEscapeSequences(text)
	length := runewidth.StringWidth(text)
	return length
}

func (a *Writer) trimAnsi(text string) string {
	if len(text) == 0 || !strings.Contains(text, "\x1b") {
		return text
	}
	return regex.ReplaceAllString(AnsiRegex, text, "")
}

func (a *Writer) TrimEscapeSequences(text string) string {
	if len(text) == 0 {
		return text
	}
	text = strings.ReplaceAll(text, a.escapeLeft, "")
	text = strings.ReplaceAll(text, a.escapeRight, "")
	return text
}
