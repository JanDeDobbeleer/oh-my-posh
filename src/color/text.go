package color

import (
	"oh-my-posh/regex"
	"strings"
	"unicode/utf8"
)

func (ansi *Ansi) MeasureText(text string) int {
	// skip strings with ANSI
	if !strings.Contains(text, "\x1b") {
		text = ansi.TrimEscapeSequences(text)
		length := utf8.RuneCountInString(text)
		return length
	}
	if strings.Contains(text, "\x1b]8;;") {
		matches := regex.FindAllNamedRegexMatch(ansi.hyperlinkRegex, text)
		for _, match := range matches {
			text = strings.ReplaceAll(text, match["STR"], match["TEXT"])
		}
	}
	text = ansi.TrimAnsi(text)
	text = ansi.TrimEscapeSequences(text)
	length := utf8.RuneCountInString(text)
	return length
}

func (ansi *Ansi) TrimAnsi(text string) string {
	if len(text) == 0 || !strings.Contains(text, "\x1b") {
		return text
	}
	return regex.ReplaceAllString(AnsiRegex, text, "")
}

func (ansi *Ansi) TrimEscapeSequences(text string) string {
	if len(text) == 0 {
		return text
	}
	text = strings.ReplaceAll(text, ansi.escapeLeft, "")
	text = strings.ReplaceAll(text, ansi.escapeRight, "")
	return text
}
