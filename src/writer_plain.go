package main

import (
	"strings"
)

// PlainWriter writes a plain string
type PlainWriter struct {
	builder strings.Builder
}

func (a *PlainWriter) setColors(background, foreground string)       {}
func (a *PlainWriter) setParentColors(background, foreground string) {}
func (a *PlainWriter) clearParentColors()                            {}

func (a *PlainWriter) write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}
	writeAndRemoveText := func(text, textToRemove, parentText string) string {
		a.builder.WriteString(text)
		return strings.Replace(parentText, textToRemove, "", 1)
	}
	match := findAllNamedRegexMatch(colorRegex, text)
	for i := range match {
		escapedTextSegment := match[i]["text"]
		innerText := match[i]["content"]
		textBeforeColorOverride := strings.Split(text, escapedTextSegment)[0]
		text = writeAndRemoveText(textBeforeColorOverride, textBeforeColorOverride, text)
		text = writeAndRemoveText(innerText, escapedTextSegment, text)
	}
	a.builder.WriteString(text)
}

func (a *PlainWriter) string() string {
	return a.builder.String()
}

func (a *PlainWriter) reset() {
	a.builder.Reset()
}
