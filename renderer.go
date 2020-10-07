package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/gookit/color"
	"golang.org/x/text/unicode/norm"
)

type formats struct {
	single      string
	full        string
	transparent string
	linebreak   string
	linechange  string
	left        string
	right       string
	rANSI       string
}

//Shell indicates the shell we're currently in
type Shell string

//Renderer writes colorized strings
type Renderer struct {
	Buffer  *bytes.Buffer
	formats *formats
	shell   Shell
}

const (
	//Transparent implies a transparent color
	Transparent string = "transparent"
	zsh         Shell  = "zsh"
	universal   Shell  = "any"
)

func (r *Renderer) init(shell string) {
	r.formats = &formats{
		rANSI: "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))",
	}
	switch shell {
	case "zsh":
		r.formats.single = "%%{\x1b[%sm%%}%s%%{\x1b[0m%%}"
		r.formats.full = "%%{\x1b[%sm\x1b[%sm%%}%s%%{\x1b[0m%%}"
		r.formats.transparent = "%%{\x1b[%s;49m\x1b[7m%%}%s%%{\x1b[m\x1b[0m%%}"
		r.formats.linebreak = "\n"
		r.formats.linechange = "%%{\x1b[%d%s%%}"
		r.formats.left = "%%{\x1b[%dC%%}"
		r.formats.right = "%%{\x1b[%dD%%}"
		r.shell = zsh
	default:
		r.formats.single = "\x1b[%sm%s\x1b[0m"
		r.formats.full = "\x1b[%sm\x1b[%sm%s\x1b[0m"
		r.formats.transparent = "\x1b[%s;49m\x1b[7m%s\x1b[m\x1b[0m"
		r.formats.linebreak = "\x1b[1000C "
		r.formats.linechange = "\x1b[%d%s"
		r.formats.left = "\x1b[%dC"
		r.formats.right = "\x1b[%dD"
		r.shell = universal
	}
}

func (r *Renderer) getAnsiFromHex(hexColor string, isBackground bool) string {
	style := color.HEX(hexColor, isBackground)
	return style.Code()
}

func (r *Renderer) writeColoredText(background string, foreground string, text string) {
	var coloredText string
	if foreground == Transparent && background != "" {
		ansiColor := r.getAnsiFromHex(background, false)
		coloredText = fmt.Sprintf(r.formats.transparent, ansiColor, text)
	} else if background == "" || background == Transparent {
		ansiColor := r.getAnsiFromHex(foreground, false)
		coloredText = fmt.Sprintf(r.formats.single, ansiColor, text)
	} else if foreground != "" && background != "" {
		bgAnsiColor := r.getAnsiFromHex(background, true)
		fgAnsiColor := r.getAnsiFromHex(foreground, false)
		coloredText = fmt.Sprintf(r.formats.full, bgAnsiColor, fgAnsiColor, text)
	}
	r.Buffer.WriteString(coloredText)
}

func (r *Renderer) writeAndRemoveText(background string, foreground string, text string, textToRemove string, parentText string) string {
	r.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (r *Renderer) write(background string, foreground string, text string) {
	rex := regexp.MustCompile(`<((#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})|transparent))>(.*?)</>`)
	match := rex.FindAllStringSubmatch(text, -1)
	for i := range match {
		// get the text before the color override and write that first
		textBeforeColorOverride := strings.Split(text, match[i][0])[0]
		text = r.writeAndRemoveText(background, foreground, textBeforeColorOverride, textBeforeColorOverride, text)
		text = r.writeAndRemoveText(background, match[i][1], match[i][4], match[i][0], text)
	}
	// color the remaining part of text with background and foreground
	r.writeColoredText(background, foreground, text)
}

func (r *Renderer) lenWithoutANSI(str string) int {
	re := regexp.MustCompile(r.formats.rANSI)
	stripped := re.ReplaceAllString(str, "")
	if r.shell == zsh {
		stripped = strings.Replace(stripped, "%{", "", -1)
		stripped = strings.Replace(stripped, "%}", "", -1)
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

func (r *Renderer) lineBreak() string {
	return r.formats.linebreak
}

func (r *Renderer) carriageForward() string {
	return fmt.Sprintf(r.formats.left, 1000)
}

func (r *Renderer) setCursorForRightWrite(text string, offset int) string {
	strippedLen := r.lenWithoutANSI(text) + -offset
	return fmt.Sprintf(r.formats.right, strippedLen)
}

func (r *Renderer) changeLine(numberOfLines int) string {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	return fmt.Sprintf(r.formats.linechange, numberOfLines, position)
}

func (r *Renderer) string() string {
	return r.Buffer.String()
}

func (r *Renderer) reset() {
	r.Buffer.Reset()
}
