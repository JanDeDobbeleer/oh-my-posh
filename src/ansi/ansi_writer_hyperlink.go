package ansi

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/regex"
	"github.com/jandedobbeleer/oh-my-posh/shell"
)

func (w *Writer) write(i int, s rune) {
	// ignore the logic when there is no hyperlink
	if !w.hasHyperlink {
		w.builder.WriteRune(s)
		return
	}

	if s == '[' && w.state == OTHER {
		w.state = TEXT
		w.hyperlinkBuilder.WriteRune(s)
		return
	}

	if w.state == OTHER {
		w.builder.WriteRune(s)
		return
	}

	w.hyperlinkBuilder.WriteRune(s)

	switch s {
	case ']':
		// potential end of text part of hyperlink
		w.squareIndex = i
	case '(':
		// split into link part
		if w.squareIndex == i-1 {
			w.state = LINK
		}
		if w.state == LINK {
			w.roundCount++
		}
	case ')':
		if w.state != LINK {
			return
		}
		w.roundCount--
		if w.roundCount != 0 {
			return
		}
		// end of link part
		w.builder.WriteString(w.replaceHyperlink(w.hyperlinkBuilder.String()))
		w.hyperlinkBuilder.Reset()
		w.state = OTHER
	}
}

func (w *Writer) replaceHyperlink(text string) string {
	// hyperlink matching
	results := regex.FindNamedRegexMatch("(?P<ALL>(?:\\[(?P<TEXT>.+)\\])(?:\\((?P<URL>.*)\\)))", text)
	if len(results) != 3 {
		return text
	}

	if w.Plain {
		return results["TEXT"]
	}

	linkText := w.escapeLinkTextForFishShell(results["TEXT"])
	// build hyperlink ansi
	hyperlink := fmt.Sprintf(w.hyperlink, results["URL"], linkText)
	// replace original text by the new ones
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
