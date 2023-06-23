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

	if s == '«' && w.hyperlinkState == OTHER {
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
	case '»':
		// potential end of text part of hyperlink
		w.bracketIndex = i
	case '(':
		// split into link part
		if w.bracketIndex == i-1 {
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
	results := regex.FindNamedRegexMatch("(?P<ALL>(?:«(?P<TEXT>.+)»)(?:\\((?P<URL>.*)\\)))", text)
	if len(results) != 3 {
		return text
	}

	linkText := results["TEXT"]

	// this isn't supported for elvish and xonsh
	if w.shell == shell.ELVISH || w.shell == shell.XONSH {
		return strings.Replace(text, results["ALL"], linkText, 1)
	}

	// we only care about the length of the text part
	w.length += runewidth.StringWidth(linkText)

	if w.Plain {
		return linkText
	}

	// build hyperlink ansi
	hyperlink := fmt.Sprintf(w.hyperlink, results["URL"], linkText)
	// replace original text by the new ones
	return strings.Replace(text, results["ALL"], hyperlink, 1)
}
