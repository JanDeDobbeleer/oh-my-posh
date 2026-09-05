package terminal

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHyperlinkNoUrl(t *testing.T) {
	cases := []struct {
		Text      string
		ShellName string
		Expected  string
	}{
		{Text: "sample text with no url", ShellName: shell.ZSH, Expected: "%{\x1b[47m%}%{\x1b[30m%}sample text with no url%{\x1b[0m%}"},
		{Text: "sample text with no url", ShellName: shell.PWSH, Expected: "\x1b[47m\x1b[30msample text with no url\x1b[0m"},
		{Text: "sample text with no url", ShellName: shell.BASH, Expected: "\\[\x1b[47m\\]\\[\x1b[30m\\]sample text with no url\\[\x1b[0m\\]"},
		{Text: "sample text with no url [test]", ShellName: shell.BASH, Expected: "\\[\x1b[47m\\]\\[\x1b[30m\\]sample text with no url [test]\\[\x1b[0m\\]"},
	}
	for _, tc := range cases {
		Init(tc.ShellName)
		Colors = &color.Defaults{}

		Write("white", "black", tc.Text)

		got, _ := String()

		assert.Equal(t, tc.Expected, got)
	}
}

func TestGenerateHyperlinkWithUrl(t *testing.T) {
	cases := []struct {
		Text      string
		ShellName string
		Expected  string
	}{
		{
			Text:      "<LINK>http://www.google.be<TEXT>google</TEXT></LINK> <LINK>http://maps.google.be<TEXT>maps (2/2)</TEXT></LINK>",
			ShellName: shell.FISH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\ \x1b]8;;http://maps.google.be\x1b\\maps (2/2)\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			Text:      "in <accent><b>pwsh </b></> ",
			ShellName: shell.PWSH,
			Expected:  "\x1b[47m\x1b[30min \x1b[49m\x1b[1mpwsh \x1b[22m\x1b[47m \x1b[0m",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT>google</TEXT></LINK>",
			ShellName: shell.ZSH,
			Expected:  "%{\x1b[47m%}%{\x1b[30m%}%{\x1b]8;;http://www.google.be\x1b\\%}google%{\x1b]8;;\x1b\\%}%{\x1b[0m%}",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT>google</TEXT></LINK>",
			ShellName: shell.PWSH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT>google</TEXT></LINK>",
			ShellName: shell.BASH,
			Expected:  "\\[\x1b[47m\\]\\[\x1b[30m\\]\\[\x1b]8;;http://www.google.be\x1b\\\\\\]google\\[\x1b]8;;\x1b\\\\\\]\\[\x1b[0m\\]",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT>google</TEXT></LINK> <LINK>http://maps.google.be<TEXT>maps</TEXT></LINK>",
			ShellName: shell.FISH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\ \x1b]8;;http://maps.google.be\x1b\\maps\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			Text:      "[]<LINK>http://www.google.be<TEXT>google</TEXT></LINK>[]",
			ShellName: shell.FISH,
			Expected:  "\x1b[47m\x1b[30m[]\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\[]\x1b[0m",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT><blue>google</></TEXT></LINK>",
			ShellName: shell.FISH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\\x1b[49m\x1b[34mgoogle\x1b[47m\x1b[30m\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT></TEXT></LINK>",
			ShellName: shell.ZSH,
			Expected:  "%{\x1b[47m%}%{\x1b[30m%}%{\x1b]8;;http://www.google.be\x1b\\%}link%{\x1b]8;;\x1b\\%}%{\x1b[0m%}",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT></TEXT></LINK>",
			ShellName: shell.PWSH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\link\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			Text:      "<LINK>http://www.google.be<TEXT></TEXT></LINK>",
			ShellName: shell.BASH,
			Expected:  "\\[\x1b[47m\\]\\[\x1b[30m\\]\\[\x1b]8;;http://www.google.be\x1b\\\\\\]link\\[\x1b]8;;\x1b\\\\\\]\\[\x1b[0m\\]",
		},
		{
			Text:      "<LINK>http://evil\x1b]0;HACKED\x07.com<TEXT>google</TEXT></LINK>",
			ShellName: shell.FISH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://evil]0;HACKED.com\x1b\\google\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			// a backslash in the OSC 8 URI is doubled for bash so @P prompt
			// expansion hands the terminal exactly one backslash; undoubled it
			// would be re-interpreted as a prompt escape (\e -> ESC)
			Text:      `<LINK>C:\temp<TEXT>x</TEXT></LINK>`,
			ShellName: shell.BASH,
			Expected:  "\\[\x1b[47m\\]\\[\x1b[30m\\]\\[\x1b]8;;C:\\\\temp\x1b\\\\\\]x\\[\x1b]8;;\x1b\\\\\\]\\[\x1b[0m\\]",
		},
		{
			// same for zsh: a % in the URI is doubled so prompt expansion
			// does not treat it as a prompt escape
			Text:      `<LINK>http://example.com/100%<TEXT>x</TEXT></LINK>`,
			ShellName: shell.ZSH,
			Expected:  "%{\x1b[47m%}%{\x1b[30m%}%{\x1b]8;;http://example.com/100%%\x1b\\%}x%{\x1b]8;;\x1b\\%}%{\x1b[0m%}",
		},
	}
	for _, tc := range cases {
		Init(tc.ShellName)
		Colors = &color.Defaults{}

		Write("white", "black", tc.Text)

		got, _ := String()

		assert.Equal(t, tc.Expected, got)
	}
}

// TestHyperlinkResetsBetweenWrites guards against isHyperlink leaking across
// Write calls. An unbalanced <LINK> - no closing </TEXT> - never clears the
// flag inside writeBody, so without an explicit reset at the top of Write it
// stays true for every later Write in the process (the serve daemon keeps the
// package-level state alive across renders). That routes subsequent runes
// through write's isHyperlink branch, which skips both length counting and
// shell-escape handling.
func TestHyperlinkResetsBetweenWrites(t *testing.T) {
	Init(shell.PWSH)
	Colors = &color.Defaults{}

	// unbalanced: <LINK> with no <TEXT>/</LINK> ever closing it.
	Write("white", "black", "<LINK>http://example.com")

	// a plain follow-up write, as engine.go does for the next segment in the
	// same block before String() is ever called.
	Write("white", "black", "abc")

	_, length := String()

	assert.Equal(t, 3, length, "isHyperlink leaked from the unbalanced <LINK> into the next Write")
}

// TestGenerateHyperlinkPlainMode pins the Plain-mode fix for OSC 8 hyperlinks.
// Guarding only the escape emitters (formats.HyperlinkStart/Center/End) is not
// enough: write()'s isHyperlink branch streams the URL between <LINK> and
// <TEXT> straight to the builder as literal runes without ever counting them
// toward length, so the URL would leak into the rendered output as invisible
// (uncounted) text. The full Plain render of a <LINK> must be exactly its
// visible link text, with length matching it exactly.
func TestGenerateHyperlinkPlainMode(t *testing.T) {
	saveGradientTestGlobals(t)

	Init(shell.PWSH)
	Plain = true
	ParentColors = nil
	Colors = &color.Defaults{}

	Write("white", "black", "<LINK>http://www.google.be<TEXT>google</TEXT></LINK>")

	got, length := String()

	assert.Equal(t, "google", got)
	assert.Equal(t, 6, length)
	assert.NotContains(t, got, "google.be", "the URL must not leak into Plain output")
}

func TestGenerateFileLink(t *testing.T) {
	cases := []struct {
		Text     string
		Expected string
	}{
		{
			Text:     `<LINK>file:C:/Program Files (x86)/Common Files/Microsoft Shared/Posh<TEXT>Posh</TEXT></LINK>`,
			Expected: "\x1b[47m\x1b[30m\x1b]8;;file:C:/Program Files (x86)/Common Files/Microsoft Shared/Posh\x1b\\Posh\x1b]8;;\x1b\\\x1b[0m",
		},
		{Text: `<LINK>file:C:/Windows<TEXT>Windows</TEXT></LINK>`, Expected: "\x1b[47m\x1b[30m\x1b]8;;file:C:/Windows\x1b\\Windows\x1b]8;;\x1b\\\x1b[0m"},
	}
	for _, tc := range cases {
		Init(shell.PWSH)
		Colors = &color.Defaults{}

		Write("white", "black", tc.Text)

		got, _ := String()

		assert.Equal(t, tc.Expected, got)
	}
}
