package ansi

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/regex"
	"github.com/jandedobbeleer/oh-my-posh/shell"
)

func (w *Writer) GenerateHyperlink(text string) string {
	const (
		LINK  = "link"
		TEXT  = "text"
		OTHER = "plain"
	)

	// do not do this when we do not need to
	anchorCount := strings.Count(text, "[") + strings.Count(text, "]") + strings.Count(text, "(") + strings.Count(text, ")")
	if anchorCount < 4 {
		return text
	}

	var result, hyperlink strings.Builder
	var squareIndex, roundCount int
	state := OTHER

	for i, s := range text {
		if s == '[' && state == OTHER {
			state = TEXT
			hyperlink.WriteRune(s)
			continue
		}

		if state == OTHER {
			result.WriteRune(s)
			continue
		}

		hyperlink.WriteRune(s)

		switch s {
		case ']':
			// potential end of text part of hyperlink
			squareIndex = i
		case '(':
			// split into link part
			if squareIndex == i-1 {
				state = LINK
			}
			if state == LINK {
				roundCount++
			}
		case ')':
			if state != LINK {
				continue
			}
			roundCount--
			if roundCount != 0 {
				continue
			}
			// end of link part
			result.WriteString(w.replaceHyperlink(hyperlink.String()))
			hyperlink.Reset()
			state = OTHER
		}
	}

	result.WriteString(hyperlink.String())
	return result.String()
}

func (w *Writer) replaceHyperlink(text string) string {
	// hyperlink matching
	results := regex.FindNamedRegexMatch("(?P<ALL>(?:\\[(?P<TEXT>.+)\\])(?:\\((?P<URL>.*)\\)))", text)
	if len(results) != 3 {
		return text
	}
	linkText := w.escapeLinkTextForFishShell(results["TEXT"])
	// build hyperlink ansi
	hyperlink := fmt.Sprintf(w.hyperlink, results["URL"], linkText)
	// replace original text by the new onex
	return strings.Replace(text, results["ALL"], hyperlink, 1)
}

func (w *Writer) escapeLinkTextForFishShell(text string) string {
	if w.shell != shell.FISH {
		return text
	}
	escapeChars := map[string]string{
		`c`: `\c`,
		`a`: `\a`,
		`b`: `\b`,
		`e`: `\e`,
		`f`: `\f`,
		`n`: `\n`,
		`r`: `\r`,
		`t`: `\t`,
		`v`: `\v`,
		`$`: `\$`,
		`*`: `\*`,
		`?`: `\?`,
		`~`: `\~`,
		`%`: `\%`,
		`#`: `\#`,
		`(`: `\(`,
		`)`: `\)`,
		`{`: `\{`,
		`}`: `\}`,
		`[`: `\[`,
		`]`: `\]`,
		`<`: `\<`,
		`>`: `\>`,
		`^`: `\^`,
		`&`: `\&`,
		`;`: `\;`,
		`"`: `\"`,
		`'`: `\'`,
		`x`: `\x`,
		`X`: `\X`,
		`0`: `\0`,
		`u`: `\u`,
		`U`: `\U`,
	}
	if val, ok := escapeChars[text[0:1]]; ok {
		return val + text[1:]
	}
	return text
}
