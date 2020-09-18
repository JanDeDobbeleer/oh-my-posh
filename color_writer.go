package main

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/gookit/color"
)

//ColorWriter writes colorized strings
type ColorWriter struct {
	Buffer *bytes.Buffer
}

func (w *ColorWriter) writeColoredText(background string, foreground string, text string) {
	style := color.HEXStyle(foreground, background)
	text = style.Sprint(text)
	w.Buffer.WriteString(text)
}

func (w *ColorWriter) writeAndRemoveText(background string, foreground string, text string, textToRemove string, parentText string) string {
	w.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (w *ColorWriter) write(background string, foreground string, text string) {
	style := color.HEXStyle(foreground, background)
	text = style.Sprint(text)

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
