package color

import (
	"oh-my-posh/regex"
	"strings"
	"unicode/utf8"
)

func measureText(text string) int {
	// skip hyperlinks
	if !strings.Contains(text, "\x1b]8;;") {
		return utf8.RuneCountInString(text)
	}
	matches := regex.FindAllNamedRegexMatch(regex.LINK, text)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match["STR"], match["TEXT"])
	}
	length := utf8.RuneCountInString(text)
	return length
}
