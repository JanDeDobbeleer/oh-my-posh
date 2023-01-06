package ansi

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/mattn/go-runewidth"
)

func (w *Writer) write(i int, s rune) {
	// ignore processing when invisible (<transparent,transparent>)
	if w.invisible {
		return
	}
	// ignore the logic when there is no hyperlink or things arent't visible
	if !w.hasHyperlink {
		w.length += runewidth.RuneWidth(s)
		w.builder.WriteRune(s)
		return
	}

	if s == '[' && w.hyperlinkState == OTHER {
		w.hyperlinkState = TEXT
		w.hyperlinkBuilder.WriteRune(s)
		return
	}

	if w.hyperlinkState == OTHER {
		w.length += runewidth.RuneWidth(s)
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
			w.hyperlinkState = LINK
		}
		if w.hyperlinkState == LINK {
			w.roundCount++
		}
	case ')':
		if w.hyperlinkState != LINK {
			return
		}
		w.roundCount--
		if w.roundCount != 0 {
			return
		}
		// end of link part
		w.builder.WriteString(w.replaceHyperlink(w.hyperlinkBuilder.String()))
		w.hyperlinkBuilder.Reset()
		w.hyperlinkState = OTHER
	}
}

func (w *Writer) replaceHyperlink(text string) string {
	// hyperlink matching
	results := regex.FindNamedRegexMatch("(?P<ALL>(?:\\[(?P<TEXT>.+)\\])(?:\\((?P<URL>.*)\\)))", text)
	if len(results) != 3 {
		return text
	}

	// we only care about the length of the text part
	w.length += runewidth.StringWidth(results["TEXT"])

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
