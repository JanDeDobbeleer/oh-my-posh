package ansi

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/shell"

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
			Text:      "[google](http://www.google.be) [maps (2/2)](http://maps.google.be)",
			ShellName: shell.FISH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\ \x1b]8;;http://maps.google.be\x1b\\maps (2/2)\x1b]8;;\x1b\\\x1b[0m",
		},
		{
			Text:      "in <accent>\x1b[1mpwsh \x1b[22m</> ",
			ShellName: shell.PWSH,
			Expected:  "\x1b[47m\x1b[30min \x1b[1mpwsh \x1b[22m \x1b[0m",
		},
		{Text: "[google](http://www.google.be)", ShellName: shell.ZSH, Expected: "%{\x1b[47m%}%{\x1b[30m%}%{\x1b]8;;http://www.google.be\x1b\\%}google%{\x1b]8;;\x1b\\%}%{\x1b[0m%}"},
		{Text: "[google](http://www.google.be)", ShellName: shell.PWSH, Expected: "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\\x1b[0m"},
		{
			Text:      "[google](http://www.google.be)",
			ShellName: shell.BASH,
			Expected:  "\\[\x1b[47m\\]\\[\x1b[30m\\]\\[\x1b]8;;http://www.google.be\x1b\\\\\\]google\\[\x1b]8;;\x1b\\\\\\]\\[\x1b[0m\\]",
		},
		{
			Text:      "[google](http://www.google.be) [maps](http://maps.google.be)",
			ShellName: shell.FISH,
			Expected:  "\x1b[47m\x1b[30m\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\ \x1b]8;;http://maps.google.be\x1b\\maps\x1b]8;;\x1b\\\x1b[0m",
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

func TestGenerateHyperlinkWithUrlNoName(t *testing.T) {
	cases := []struct {
		Text      string
		ShellName string
		Expected  string
	}{
		{Text: "[](http://www.google.be)", ShellName: shell.ZSH, Expected: "%{\x1b[47m%}%{\x1b[30m%}[](http://www.google.be)%{\x1b[0m%}"},
		{Text: "[](http://www.google.be)", ShellName: shell.PWSH, Expected: "\x1b[47m\x1b[30m[](http://www.google.be)\x1b[0m"},
		{Text: "[](http://www.google.be)", ShellName: shell.BASH, Expected: "\\[\x1b[47m\\]\\[\x1b[30m\\][](http://www.google.be)\\[\x1b[0m\\]"},
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
			Text:     `[Posh](file:C:/Program Files (x86)/Common Files/Microsoft Shared/Posh)`,
			Expected: "\x1b[47m\x1b[30m\x1b]8;;file:C:/Program Files (x86)/Common Files/Microsoft Shared/Posh\x1b\\Posh\x1b]8;;\x1b\\\x1b[0m",
		},
		{Text: `[Windows](file:C:/Windows)`, Expected: "\x1b[47m\x1b[30m\x1b]8;;file:C:/Windows\x1b\\Windows\x1b]8;;\x1b\\\x1b[0m"},
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
