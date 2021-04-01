package main

import (
	"fmt"
	"strings"
)

// ANSIUtils exposes functionality using ANSI
type ANSIUtils struct {
	builder strings.Builder
	formats *ansiFormats
}

func (r *ANSIUtils) carriageForward() {
	r.builder.WriteString(fmt.Sprintf(r.formats.left, 1000))
}

func (r *ANSIUtils) setCursorForRightWrite(text string, offset int) {
	strippedLen := r.formats.lenWithoutANSI(text) + -offset
	r.builder.WriteString(fmt.Sprintf(r.formats.right, strippedLen))
}

func (r *ANSIUtils) changeLine(numberOfLines int) {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	r.builder.WriteString(fmt.Sprintf(r.formats.linechange, numberOfLines, position))
}

func (r *ANSIUtils) creset() {
	r.builder.WriteString(r.formats.creset)
}

func (r *ANSIUtils) write(text string) {
	r.builder.WriteString(text)
	// Due to a bug in Powershell, the end of the line needs to be cleared.
	// If this doesn't happen, the portion after the prompt gets colored in the background
	// color of the line above the new input line. Clearing the line fixes this,
	// but can hopefully one day be removed when this is resolved natively.
	if r.formats.shell == pwsh || r.formats.shell == powershell5 {
		r.builder.WriteString(r.formats.clearEOL)
	}
}

func (r *ANSIUtils) string() string {
	return r.builder.String()
}

func (r *ANSIUtils) saveCursorPosition() {
	r.builder.WriteString(r.formats.saveCursorPosition)
}

func (r *ANSIUtils) restoreCursorPosition() {
	r.builder.WriteString(r.formats.restoreCursorPosition)
}

func (r *ANSIUtils) osc99(pwd string) {
	r.builder.WriteString(fmt.Sprintf(r.formats.osc99, pwd))
}
