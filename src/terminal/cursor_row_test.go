package terminal

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/stretchr/testify/assert"
)

func TestCursorRow(t *testing.T) {
	cases := []struct {
		Case     string
		Text     string
		Width    int
		Expected int
	}{
		{
			Case:     "single line",
			Text:     "hello world",
			Width:    80,
			Expected: 0,
		},
		{
			Case:     "multiple newlines",
			Text:     "one\ntwo\nthree",
			Width:    80,
			Expected: 2,
		},
		{
			Case:     "empty string",
			Text:     "",
			Width:    80,
			Expected: 0,
		},
		{
			Case:     "trailing newline counts the new row",
			Text:     "abc\n",
			Width:    80,
			Expected: 1,
		},
		{
			Case:     "line exactly width stays on one row",
			Text:     "abcde",
			Width:    5,
			Expected: 0,
		},
		{
			Case:     "width plus one wraps",
			Text:     "abcdef",
			Width:    5,
			Expected: 1,
		},
		{
			Case:     "long line spans three rows",
			Text:     strings.Repeat("x", 12),
			Width:    5,
			Expected: 2,
		},
		{
			Case:     "SGR colored text still wraps on visible cells",
			Text:     "\x1b[31mhello\x1b[0mworld",
			Width:    5,
			Expected: 1,
		},
		{
			Case:     "OSC 8 hyperlink with long URL does not count toward width",
			Text:     "\x1b]8;;http://example.com/very/long/path/that/would/wrap\x1b\\click\x1b]8;;\x1b\\",
			Width:    10,
			Expected: 0,
		},
		{
			Case:     "rprompt fills exactly the width and restores the saved column",
			Text:     "LEFT " + "\x1b7" + strings.Repeat(" ", 10) + "RIGHT" + "\x1b8",
			Width:    20,
			Expected: 0,
		},
		{
			Case:     "rprompt on a line below the top still restores to the saved row",
			Text:     "top\n" + "LEFT " + "\x1b7" + strings.Repeat(" ", 10) + "RIGHT" + "\x1b8",
			Width:    20,
			Expected: 1,
		},
		{
			Case:     "wide CJK rune straddling the edge wraps as a whole",
			Text:     "abcd" + "中",
			Width:    5,
			Expected: 1,
		},
		{
			Case:     "Warp line break",
			Text:     "a" + pwshLineBreak + "b",
			Width:    80,
			Expected: 1,
		},
		{
			Case:     "width zero disables wrapping",
			Text:     strings.Repeat("x", 100),
			Width:    0,
			Expected: 0,
		},
		{
			Case:     "unterminated OSC at end of string just ends the scan",
			Text:     "abc\x1b]8;;http://example.com",
			Width:    80,
			Expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			got := CursorRow(tc.Text, tc.Width)

			assert.Equal(t, tc.Expected, got, tc.Case)
		})
	}
}

func TestCursorRowLineBreakMatchesPwsh(t *testing.T) {
	Init(shell.PWSH)

	assert.Equal(t, LineBreak(), pwshLineBreak)
}
