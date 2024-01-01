package ansi

import (
	"testing"

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
		a := Writer{
			AnsiColors: &DefaultColors{},
		}
		a.Init(tc.ShellName)
		a.Write("white", "black", tc.Text)
		hyperlinkText, _ := a.String()
		assert.Equal(t, tc.Expected, hyperlinkText)
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
	}
	for _, tc := range cases {
		a := Writer{
			AnsiColors: &DefaultColors{},
		}
		a.Init(tc.ShellName)
		a.Write("white", "black", tc.Text)
		hyperlinkText, _ := a.String()
		assert.Equal(t, tc.Expected, hyperlinkText)
	}
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
		a := Writer{
			AnsiColors: &DefaultColors{},
		}
		a.Init(shell.PWSH)
		a.Write("white", "black", tc.Text)
		hyperlinkText, _ := a.String()
		assert.Equal(t, tc.Expected, hyperlinkText)
	}
}
