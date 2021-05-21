package main

import (
	"fmt"
	"strings"
)

const (
	ansiRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
)

type ansiUtils struct {
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
	osc99                 string
	bold                  string
	italic                string
	underline             string
	strikethrough         string
}

func (a *ansiUtils) init(shell string) {
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
		a.title = "%%{\x1b]0;%s\007%%}"
		a.colorSingle = "%%{\x1b[%sm%%}%s%%{\x1b[0m%%}"
		a.colorFull = "%%{\x1b[%sm\x1b[%sm%%}%s%%{\x1b[0m%%}"
		a.colorTransparent = "%%{\x1b[%s;49m\x1b[7m%%}%s%%{\x1b[0m%%}"
		a.escapeLeft = "%{"
		a.escapeRight = "%}"
		a.hyperlink = "%%{\x1b]8;;%s\x1b\\%%}%s%%{\x1b]8;;\x1b\\%%}"
		a.osc99 = "%%{\x1b]9;9;\"%s\"\x1b\\%%}"
		a.bold = "%%{\x1b[1m%%}%s%%{\x1b[22m%%}"
		a.italic = "%%{\x1b[3m%%}%s%%{\x1b[23m%%}"
		a.underline = "%%{\x1b[4m%%}%s%%{\x1b[24m%%}"
		a.strikethrough = "%%{\x1b[9m%%}%s%%{\x1b[29m%%}"
	case bash:
		a.linechange = "\\[\x1b[%d%s\\]"
		a.left = "\\[\x1b[%dC\\]"
		a.right = "\\[\x1b[%dD\\]"
		a.creset = "\\[\x1b[0m\\]"
		a.clearEOL = "\\[\x1b[K\\]"
		a.saveCursorPosition = "\\[\x1b7\\]"
		a.restoreCursorPosition = "\\[\x1b8\\]"
		a.title = "\\[\x1b]0;%s\007\\]"
		a.colorSingle = "\\[\x1b[%sm\\]%s\\[\x1b[0m\\]"
		a.colorFull = "\\[\x1b[%sm\x1b[%sm\\]%s\\[\x1b[0m\\]"
		a.colorTransparent = "\\[\x1b[%s;49m\x1b[7m\\]%s\\[\x1b[0m\\]"
		a.escapeLeft = "\\["
		a.escapeRight = "\\]"
		a.hyperlink = "\\[\x1b]8;;%s\x1b\\\\\\]%s\\[\x1b]8;;\x1b\\\\\\]"
		a.osc99 = "\\[\x1b]9;9;\"%s\"\x1b\\\\\\]"
		a.bold = "\\[\x1b[1m\\]%s\\[\x1b[22m\\]"
		a.italic = "\\[\x1b[3m\\]%s\\[\x1b[23m\\]"
		a.underline = "\\[\x1b[4m\\]%s\\[\x1b[24m\\]"
		a.strikethrough = "\\[\x1b[9m\\]%s\\[\x1b[29m\\]"
	default:
		a.linechange = "\x1b[%d%s"
		a.left = "\x1b[%dC"
		a.right = "\x1b[%dD"
		a.creset = "\x1b[0m"
		a.clearEOL = "\x1b[K"
		a.saveCursorPosition = "\x1b7"
		a.restoreCursorPosition = "\x1b8"
		a.title = "\x1b]0;%s\007"
		a.colorSingle = "\x1b[%sm%s\x1b[0m"
		a.colorFull = "\x1b[%sm\x1b[%sm%s\x1b[0m"
		a.colorTransparent = "\x1b[%s;49m\x1b[7m%s\x1b[0m"
		a.escapeLeft = ""
		a.escapeRight = ""
		a.hyperlink = "\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\"
		a.osc99 = "\x1b]9;9;\"%s\"\x1b\\"
		a.bold = "\x1b[1m%s\x1b[22m"
		a.italic = "\x1b[3m%s\x1b[23m"
		a.underline = "\x1b[4m%s\x1b[24m"
		a.strikethrough = "\x1b[9m%s\x1b[29m"
	}
}

func (a *ansiUtils) lenWithoutANSI(text string) int {
	if len(text) == 0 {
		return 0
	}
	// replace hyperlinks
	matches := findAllNamedRegexMatch(`(?P<STR>\x1b]8;;file:\/\/(.+)\x1b\\(?P<URL>.+)\x1b]8;;\x1b\\)`, text)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match[str], match[url])
	}
	// replace console title
	matches = findAllNamedRegexMatch(`(?P<STR>\x1b\]0;(.+)\007)`, text)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match[str], "")
	}
	stripped := replaceAllString(ansiRegex, text, "")
	stripped = strings.ReplaceAll(stripped, a.escapeLeft, "")
	stripped = strings.ReplaceAll(stripped, a.escapeRight, "")
	runeText := []rune(stripped)
	return len(runeText)
}

func (a *ansiUtils) generateHyperlink(text string) string {
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

func (a *ansiUtils) formatText(text string) string {
	results := findAllNamedRegexMatch("(?P<context><(?P<format>[buis])>(?P<text>[^<]+)</[buis]>)", text)
	for _, result := range results {
		var formatted string
		switch result["format"] {
		case "b":
			formatted = fmt.Sprintf(a.bold, result["text"])
		case "u":
			formatted = fmt.Sprintf(a.underline, result["text"])
		case "i":
			formatted = fmt.Sprintf(a.italic, result["text"])
		case "s":
			formatted = fmt.Sprintf(a.strikethrough, result["text"])
		}
		text = strings.Replace(text, result["context"], formatted, 1)
	}
	return text
}

func (a *ansiUtils) carriageForward() string {
	return fmt.Sprintf(a.left, 1000)
}

func (a *ansiUtils) getCursorForRightWrite(text string, offset int) string {
	strippedLen := a.lenWithoutANSI(text) + -offset
	return fmt.Sprintf(a.right, strippedLen)
}

func (a *ansiUtils) changeLine(numberOfLines int) string {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	return fmt.Sprintf(a.linechange, numberOfLines, position)
}

func (a *ansiUtils) consolePwd(pwd string) string {
	if strings.HasSuffix(pwd, ":") {
		pwd += "\\"
	}
	return fmt.Sprintf(a.osc99, pwd)
}
