package main

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func lenWithoutANSI(text, shell string) int {
	rANSI := "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	stripped := replaceAllString(rANSI, text, "")
	switch shell {
	case zsh:
		stripped = strings.ReplaceAll(stripped, "%{", "")
		stripped = strings.ReplaceAll(stripped, "%}", "")
	case bash:
		stripped = strings.ReplaceAll(stripped, "\\[", "")
		stripped = strings.ReplaceAll(stripped, "\\]", "")
	}
	var i norm.Iter
	i.InitString(norm.NFD, stripped)
	var count int
	for !i.Done() {
		i.Next()
		count++
	}
	return count
}

type formats struct {
	linechange            string
	left                  string
	right                 string
	title                 string
	creset                string
	clearOEL              string
	saveCursorPosition    string
	restoreCursorPosition string
}

// AnsiRenderer exposes functionality using ANSI
type AnsiRenderer struct {
	buffer  *bytes.Buffer
	formats *formats
	shell   string
}

const (
	zsh         = "zsh"
	bash        = "bash"
	pwsh        = "pwsh"
	powershell5 = "powershell"
)

func (r *AnsiRenderer) init(shell string) {
	r.shell = shell
	r.formats = &formats{}
	switch shell {
	case zsh:
		r.formats.linechange = "%%{\x1b[%d%s%%}"
		r.formats.left = "%%{\x1b[%dC%%}"
		r.formats.right = "%%{\x1b[%dD%%}"
		r.formats.title = "%%{\033]0;%s\007%%}"
		r.formats.creset = "%{\x1b[0m%}"
		r.formats.clearOEL = "%{\x1b[K%}"
		r.formats.saveCursorPosition = "%{\x1b7%}"
		r.formats.restoreCursorPosition = "%{\x1b8%}"
	case bash:
		r.formats.linechange = "\\[\x1b[%d%s\\]"
		r.formats.left = "\\[\x1b[%dC\\]"
		r.formats.right = "\\[\x1b[%dD\\]"
		r.formats.title = "\\[\033]0;%s\007\\]"
		r.formats.creset = "\\[\x1b[0m\\]"
		r.formats.clearOEL = "\\[\x1b[K\\]"
		r.formats.saveCursorPosition = "\\[\x1b7\\]"
		r.formats.restoreCursorPosition = "\\[\x1b8\\]"
	default:
		r.formats.linechange = "\x1b[%d%s"
		r.formats.left = "\x1b[%dC"
		r.formats.right = "\x1b[%dD"
		r.formats.title = "\033]0;%s\007"
		r.formats.creset = "\x1b[0m"
		r.formats.clearOEL = "\x1b[K"
		r.formats.saveCursorPosition = "\x1b7"
		r.formats.restoreCursorPosition = "\x1b8"
	}
}

func (r *AnsiRenderer) carriageForward() {
	r.buffer.WriteString(fmt.Sprintf(r.formats.left, 1000))
}

func (r *AnsiRenderer) setCursorForRightWrite(text string, offset int) {
	strippedLen := lenWithoutANSI(text, r.shell) + -offset
	r.buffer.WriteString(fmt.Sprintf(r.formats.right, strippedLen))
}

func (r *AnsiRenderer) changeLine(numberOfLines int) {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	r.buffer.WriteString(fmt.Sprintf(r.formats.linechange, numberOfLines, position))
}

func (r *AnsiRenderer) setConsoleTitle(title string) {
	r.buffer.WriteString(fmt.Sprintf(r.formats.title, title))
}

func (r *AnsiRenderer) creset() {
	r.buffer.WriteString(r.formats.creset)
}

func (r *AnsiRenderer) print(text string) {
	r.buffer.WriteString(text)
	r.clearEOL()
}

func (r *AnsiRenderer) clearEOL() {
	r.buffer.WriteString(r.formats.clearOEL)
}

func (r *AnsiRenderer) string() string {
	return r.buffer.String()
}

func (r *AnsiRenderer) saveCursorPosition() {
	r.buffer.WriteString(r.formats.saveCursorPosition)
}

func (r *AnsiRenderer) restoreCursorPosition() {
	r.buffer.WriteString(r.formats.restoreCursorPosition)
}
