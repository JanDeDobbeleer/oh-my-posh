package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/gookit/color"
)

//ColorWriter writes colorized strings
type ColorWriter struct {
	Buffer *bytes.Buffer
}

const (
	//Transparent implies a transparent color
	Transparent string = "transparent"
)

func (w *ColorWriter) writeColoredText(background string, foreground string, text string) {
	var coloredText string
	if foreground == Transparent {
		style := color.HEX(background, false)
		colorCodes := style.Code()
		// this takes the colors and inverts them so the foreground becomes transparent
		coloredText = fmt.Sprintf("\x1b[%s;49m\x1b[7m%s\x1b[m\x1b[0m", colorCodes, text)
	} else if background == "" || background == Transparent {
		style := color.HEX(foreground)
		coloredText = style.Sprint(text)
	} else {
		style := color.HEXStyle(foreground, background)
		coloredText = style.Sprint(text)
	}
	w.Buffer.WriteString(coloredText)
}

func (w *ColorWriter) writeAndRemoveText(background string, foreground string, text string, textToRemove string, parentText string) string {
	w.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (w *ColorWriter) write(background string, foreground string, text string) {
	r := regexp.MustCompile(`<\s*(?P<color>#[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})>(?P<text>.*?)<\s*/\s*>`)
	match := r.FindAllStringSubmatch(text, -1)
	for i := range match {
		// get the text before the color override and write that first
		textBeforeColorOverride := strings.Split(text, match[i][0])[0]
		text = w.writeAndRemoveText(background, foreground, textBeforeColorOverride, textBeforeColorOverride, text)
		text = w.writeAndRemoveText(background, match[i][1], match[i][2], match[i][0], text)
	}
	// color the remaining part of text with background and foreground
	w.writeColoredText(background, foreground, text)
}

func (w *ColorWriter) string() string {
	return w.Buffer.String()
}

func (w *ColorWriter) reset() {
	w.Buffer.Reset()
}
