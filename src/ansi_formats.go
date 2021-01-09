package main

import (
	"fmt"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type ansiFormats struct {
	shell                 string
	linechange            string
	left                  string
	right                 string
	creset                string
	clearEOL              string
	saveCursorPosition    string
	restoreCursorPosition string
	title                 string
	colorSingle           string
	colorFull             string
	colorTransparent      string
	escapeLeft            string
	escapeRight           string
	hyperlink             string
}

func (a *ansiFormats) init(shell string) {
	a.shell = shell
	switch shell {
	case zsh:
		a.linechange = "%%{\x1b[%d%s%%}"
		a.left = "%%{\x1b[%dC%%}"
		a.right = "%%{\x1b[%dD%%}"
		a.creset = "%{\x1b[0m%}"
		a.clearEOL = "%{\x1b[K%}"
		a.saveCursorPosition = "%{\x1b7%}"
		a.restoreCursorPosition = "%{\x1b8%}"
		a.title = "%%{\033]0;%s\007%%}"
		a.colorSingle = "%%{\x1b[%sm%%}%s%%{\x1b[0m%%}"
		a.colorFull = "%%{\x1b[%sm\x1b[%sm%%}%s%%{\x1b[0m%%}"
		a.colorTransparent = "%%{\x1b[%s;49m\x1b[7m%%}%s%%{\x1b[m\x1b[0m%%}"
		a.escapeLeft = "%{"
		a.escapeRight = "%}"
		a.hyperlink = "%%{\x1b]8;;%s\x1b\\%%}%s%%{\x1b]8;;\x1b\\%%}"
	case bash:
		a.linechange = "\\[\x1b[%d%s\\]"
		a.left = "\\[\x1b[%dC\\]"
		a.right = "\\[\x1b[%dD\\]"
		a.creset = "\\[\x1b[0m\\]"
		a.clearEOL = "\\[\x1b[K\\]"
		a.saveCursorPosition = "\\[\x1b7\\]"
		a.restoreCursorPosition = "\\[\x1b8\\]"
		a.title = "\\[\033]0;%s\007\\]"
		a.colorSingle = "\\[\x1b[%sm\\]%s\\[\x1b[0m\\]"
		a.colorFull = "\\[\x1b[%sm\x1b[%sm\\]%s\\[\x1b[0m\\]"
		a.colorTransparent = "\\[\x1b[%s;49m\x1b[7m\\]%s\\[\x1b[m\x1b[0m\\]"
		a.escapeLeft = "\\["
		a.escapeRight = "\\]"
		a.hyperlink = "\\[\x1b]8;;%s\x1b\\\\\\]%s\\[\x1b]8;;\x1b\\\\\\]"
	default:
		a.linechange = "\x1b[%d%s"
		a.left = "\x1b[%dC"
		a.right = "\x1b[%dD"
		a.creset = "\x1b[0m"
		a.clearEOL = "\x1b[K"
		a.saveCursorPosition = "\x1b7"
		a.restoreCursorPosition = "\x1b8"
		a.title = "\033]0;%s\007"
		a.colorSingle = "\x1b[%sm%s\x1b[0m"
		a.colorFull = "\x1b[%sm\x1b[%sm%s\x1b[0m"
		a.colorTransparent = "\x1b[%s;49m\x1b[7m%s\x1b[m\x1b[0m"
		a.escapeLeft = ""
		a.escapeRight = ""
		a.hyperlink = "\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\"
	}
}

func (a *ansiFormats) lenWithoutANSI(text string) int {
	rANSI := "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	stripped := replaceAllString(rANSI, text, "")
	stripped = strings.ReplaceAll(stripped, a.escapeLeft, "")
	stripped = strings.ReplaceAll(stripped, a.escapeRight, "")
	var i norm.Iter
	i.InitString(norm.NFD, stripped)
	var count int
	for !i.Done() {
		i.Next()
		count++
	}
	return count
}

func (a *ansiFormats) generateHyperlink(text string) string {
	// hyperlink matching
	results := findNamedRegexMatch("(?P<all>(?:\\[(?P<name>.+)\\])(?:\\((?P<url>.*)\\)))", text)
	if len(results) != 3 {
		return text
	}
	// build hyperlink ansi
	hyperlink := fmt.Sprintf(a.hyperlink, results["url"], results["name"])
	// replace original text by the new one
	return strings.Replace(text, results["all"], hyperlink, 1)
}
