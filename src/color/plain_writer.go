package color

import (
	"oh-my-posh/regex"
	"strings"
)

// PlainWriter writes a plain string
type PlainWriter struct {
	builder strings.Builder
}

func (a *PlainWriter) SetColors(background, foreground string)       {}
func (a *PlainWriter) SetParentColors(background, foreground string) {}
func (a *PlainWriter) ClearParentColors()                            {}

func (a *PlainWriter) Write(background, foreground, text string) {
	if len(text) == 0 {
		return
	}
	writeAndRemoveText := func(text, textToRemove, parentText string) string {
		a.builder.WriteString(text)
		return strings.Replace(parentText, textToRemove, "", 1)
	}
	match := regex.FindAllNamedRegexMatch(colorRegex, text)
	for i := range match {
		escapedTextSegment := match[i]["text"]
		innerText := match[i]["content"]
		textBeforeColorOverride := strings.Split(text, escapedTextSegment)[0]
		text = writeAndRemoveText(textBeforeColorOverride, textBeforeColorOverride, text)
		text = writeAndRemoveText(innerText, escapedTextSegment, text)
	}
	a.builder.WriteString(text)
}

func (a *PlainWriter) String() string {
	return a.builder.String()
}

func (a *PlainWriter) Reset() {
	a.builder.Reset()
}
