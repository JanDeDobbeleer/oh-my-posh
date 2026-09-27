package terminal

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// pwshLineBreak is what LineBreak renders for pwsh, the only cursor movement a prompt contains.
const pwshLineBreak = "\x1b[1000D\x1b[1B"

// CursorRow returns the row the cursor ends on after txt is printed from row 0, column 0.
// A width <= 0 disables wrapping.
func CursorRow(txt string, width int) int {
	var row, col, savedRow, savedCol int

	for i := 0; i < len(txt); {
		switch {
		case txt[i] == '\n':
			row, col = row+1, 0
			i++
		case strings.HasPrefix(txt[i:], pwshLineBreak):
			row, col = row+1, 0
			i += len(pwshLineBreak)
		case strings.HasPrefix(txt[i:], "\x1b7"):
			savedRow, savedCol = row, col
			i += 2
		case strings.HasPrefix(txt[i:], "\x1b8"):
			row, col = savedRow, savedCol
			i += 2
		case txt[i] == 0x1b:
			i += escapeLen(txt[i:])
		default:
			r, size := utf8.DecodeRuneInString(txt[i:])
			i += size

			w := runeCells(r)
			if w == 0 {
				continue
			}

			// terminals defer the wrap: a row exactly width cells wide stays put
			// until the next printable rune arrives
			if width > 0 && col+w > width {
				row, col = row+1, 0
			}

			col += w
		}
	}

	return row
}

// CursorRowMarker is stripped by omp.ps1 before PSReadLine sees the prompt; terminals
// ignore an unknown OSC, so a leaked marker is harmless.
func CursorRowMarker(row int) string {
	return "\x1b]7777;" + strconv.Itoa(row) + "\x07"
}

func escapeLen(s string) int {
	if len(s) < 2 {
		return len(s)
	}

	switch s[1] {
	case '[':
		for j := 2; j < len(s); j++ {
			if s[j] >= 0x40 && s[j] <= 0x7e {
				return j + 1
			}
		}
	case ']':
		for j := 2; j < len(s); j++ {
			if s[j] == 0x07 {
				return j + 1
			}

			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2
			}
		}
	default:
		return 2
	}

	return len(s)
}
