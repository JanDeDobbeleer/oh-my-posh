package color

import (
	"oh-my-posh/regex"
	"strings"
	"unicode/utf8"
)

func measureText(text string) int {
	// skip strings with ANSI
	if !strings.Contains(text, "\x1b") {
		return utf8.RuneCountInString(text)
	}
	if strings.Contains(text, "\x1b]8;;") {
		matches := regex.FindAllNamedRegexMatch(regex.LINK, text)
		for _, match := range matches {
			text = strings.ReplaceAll(text, match["STR"], match["TEXT"])
		}
	}
	text = TrimAnsi(text)
	return utf8.RuneCountInString(text)
}

func TrimAnsi(text string) string {
	if len(text) == 0 || !strings.Contains(text, "\x1b") {
		return text
	}
	return regex.ReplaceAllString(AnsiRegex, text, "")
}
