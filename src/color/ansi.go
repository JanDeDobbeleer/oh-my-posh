package color

import (
	"fmt"
	"oh-my-posh/regex"
	"strings"
)

const (
	ansiRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	zsh  = "zsh"
	bash = "bash"
	pwsh = "pwsh"

	str = "STR"
	url = "URL"
)

type Ansi struct {
	title                 string
	shell                 string
	linechange            string
	left                  string
	right                 string
	creset                string
	clearBelow            string
	clearLine             string
	saveCursorPosition    string
	restoreCursorPosition string
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
	format                string
	shellReservedKeywords []shellKeyWordReplacement
}

type shellKeyWordReplacement struct {
	text        string
	replacement string
}

func (a *Ansi) Init(shell string) {
	a.shell = shell
	switch shell {
	case zsh:
		a.format = "%%{%s%%}"
		a.linechange = "%%{\x1b[%d%s%%}"
		a.right = "%%{\x1b[%dC%%}"
		a.left = "%%{\x1b[%dD%%}"
		a.creset = "%{\x1b[0m%}"
		a.clearBelow = "%{\x1b[0J%}"
		a.clearLine = "%{\x1b[K%}"
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
		// escape double quotes and variable expansion
		a.shellReservedKeywords = append(a.shellReservedKeywords, shellKeyWordReplacement{"\\", "\\\\"}, shellKeyWordReplacement{"%", "%%"})
	case bash:
		a.format = "\\[%s\\]"
		a.linechange = "\\[\x1b[%d%s\\]"
		a.right = "\\[\x1b[%dC\\]"
		a.left = "\\[\x1b[%dD\\]"
		a.creset = "\\[\x1b[0m\\]"
		a.clearBelow = "\\[\x1b[0J\\]"
		a.clearLine = "\\[\x1b[K\\]"
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
		// escape backslashes to avoid replacements
		// https://tldp.org/HOWTO/Bash-Prompt-HOWTO/bash-prompt-escape-sequences.html
		a.shellReservedKeywords = append(a.shellReservedKeywords, shellKeyWordReplacement{"\\", "\\\\"})
	default:
		a.format = "%s"
		a.linechange = "\x1b[%d%s"
		a.right = "\x1b[%dC"
		a.left = "\x1b[%dD"
		a.creset = "\x1b[0m"
		a.clearBelow = "\x1b[0J"
		a.clearLine = "\x1b[K"
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
	// common replacement for all shells
	a.shellReservedKeywords = append(a.shellReservedKeywords, shellKeyWordReplacement{"`", "'"})
}

func (a *Ansi) LenWithoutANSI(text string) int {
	if len(text) == 0 {
		return 0
	}
	// replace hyperlinks(file/http/https)
	matches := regex.FindAllNamedRegexMatch(`(?P<STR>\x1b]8;;(file|http|https):\/\/(.+?)\x1b\\(?P<URL>.+?)\x1b]8;;\x1b\\)`, text)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match[str], match[url])
	}
	// replace console title
	matches = regex.FindAllNamedRegexMatch(`(?P<STR>\x1b\]0;(.+)\007)`, text)
	for _, match := range matches {
		text = strings.ReplaceAll(text, match[str], "")
	}
	stripped := regex.ReplaceAllString(ansiRegex, text, "")
	stripped = strings.ReplaceAll(stripped, a.escapeLeft, "")
	stripped = strings.ReplaceAll(stripped, a.escapeRight, "")
	runeText := []rune(stripped)
	return len(runeText)
}

func (a *Ansi) generateHyperlink(text string) string {
	// hyperlink matching
	results := regex.FindNamedRegexMatch("(?P<all>(?:\\[(?P<name>.+)\\])(?:\\((?P<url>.*)\\)))", text)
	if len(results) != 3 {
		return text
	}
	// build hyperlink ansi
	hyperlink := fmt.Sprintf(a.hyperlink, results["url"], results["name"])
	// replace original text by the new one
	return strings.Replace(text, results["all"], hyperlink, 1)
}

func (a *Ansi) formatText(text string) string {
	results := regex.FindAllNamedRegexMatch("(?P<context><(?P<format>[buis])>(?P<text>[^<]+)</[buis]>)", text)
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

func (a *Ansi) CarriageForward() string {
	return fmt.Sprintf(a.right, 1000)
}

func (a *Ansi) GetCursorForRightWrite(text string, offset int) string {
	strippedLen := a.LenWithoutANSI(text) + -offset
	return fmt.Sprintf(a.left, strippedLen)
}

func (a *Ansi) ChangeLine(numberOfLines int) string {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	return fmt.Sprintf(a.linechange, numberOfLines, position)
}

func (a *Ansi) ConsolePwd(pwd string) string {
	if strings.HasSuffix(pwd, ":") {
		pwd += "\\"
	}
	return fmt.Sprintf(a.osc99, pwd)
}

func (a *Ansi) ClearAfter() string {
	return a.clearLine + a.clearBelow
}

func (a *Ansi) EscapeText(text string) string {
	// what to escape/replace is different per shell
	for _, s := range a.shellReservedKeywords {
		text = strings.ReplaceAll(text, s.text, s.replacement)
	}
	return text
}

func (a *Ansi) Title(title string) string {
	return fmt.Sprintf(a.title, title)
}

func (a *Ansi) ColorReset() string {
	return a.creset
}

func (a *Ansi) FormatText(text string) string {
	return fmt.Sprintf(a.format, text)
}

func (a *Ansi) SaveCursorPosition() string {
	return a.saveCursorPosition
}

func (a *Ansi) RestoreCursorPosition() string {
	return a.restoreCursorPosition
}
