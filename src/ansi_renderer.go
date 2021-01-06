package main

import (
	"fmt"
	"strings"
)

// AnsiRenderer exposes functionality using ANSI
type AnsiRenderer struct {
	builder strings.Builder
	formats *ansiFormats
}

func (r *AnsiRenderer) carriageForward() {
	r.builder.WriteString(fmt.Sprintf(r.formats.left, 1000))
}

func (r *AnsiRenderer) setCursorForRightWrite(text string, offset int) {
	strippedLen := r.formats.lenWithoutANSI(text) + -offset
	r.builder.WriteString(fmt.Sprintf(r.formats.right, strippedLen))
}

func (r *AnsiRenderer) changeLine(numberOfLines int) {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	r.builder.WriteString(fmt.Sprintf(r.formats.linechange, numberOfLines, position))
}

func (r *AnsiRenderer) creset() {
	r.builder.WriteString(r.formats.creset)
}

func (r *AnsiRenderer) print(text string) {
	r.builder.WriteString(text)
	// Due to a bug in Powershell, the end of the line needs to be cleared.
	// If this doesn't happen, the portion after the prompt gets colored in the background
	// color of the line above the new input line. Clearing the line fixes this,
	// but can hopefully one day be removed when this is resolved natively.
	if r.formats.shell == pwsh || r.formats.shell == powershell5 {
		r.builder.WriteString(r.formats.clearEOL)
	}
}

func (r *AnsiRenderer) string() string {
	return r.builder.String()
}

func (r *AnsiRenderer) saveCursorPosition() {
	r.builder.WriteString(r.formats.saveCursorPosition)
}

func (r *AnsiRenderer) restoreCursorPosition() {
	r.builder.WriteString(r.formats.restoreCursorPosition)
}
