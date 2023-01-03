package color

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
		{Text: "sample text with no url", ShellName: shell.ZSH, Expected: "sample text with no url"},
		{Text: "sample text with no url", ShellName: shell.PWSH, Expected: "sample text with no url"},
		{Text: "sample text with no url", ShellName: shell.BASH, Expected: "sample text with no url"},
	}
	for _, tc := range cases {
		a := AnsiWriter{}
		a.Init(tc.ShellName)
		hyperlinkText := a.GenerateHyperlink(tc.Text)
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
			Expected:  "\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\ \x1b]8;;http://maps.google.be\x1b\\maps (2/2)\x1b]8;;\x1b\\",
		},
		{
			Text:      "in <accent>\x1b[1mpwsh \x1b[22m</> ",
			ShellName: shell.PWSH,
			Expected:  "in <accent>\x1b[1mpwsh \x1b[22m</> ",
		},
		{Text: "[google](http://www.google.be)", ShellName: shell.ZSH, Expected: "%{\x1b]8;;http://www.google.be\x1b\\%}google%{\x1b]8;;\x1b\\%}"},
		{Text: "[google](http://www.google.be)", ShellName: shell.PWSH, Expected: "\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\"},
		{Text: "[google](http://www.google.be)", ShellName: shell.BASH, Expected: "\\[\x1b]8;;http://www.google.be\x1b\\\\\\]google\\[\x1b]8;;\x1b\\\\\\]"},
		{
			Text:      "[google](http://www.google.be) [maps](http://maps.google.be)",
			ShellName: shell.FISH,
			Expected:  "\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\ \x1b]8;;http://maps.google.be\x1b\\maps\x1b]8;;\x1b\\",
		},
	}
	for _, tc := range cases {
		a := AnsiWriter{}
		a.Init(tc.ShellName)
		hyperlinkText := a.GenerateHyperlink(tc.Text)
		assert.Equal(t, tc.Expected, hyperlinkText)
	}
}

func TestGenerateHyperlinkWithUrlNoName(t *testing.T) {
	cases := []struct {
		Text      string
		ShellName string
		Expected  string
	}{
		{Text: "[](http://www.google.be)", ShellName: shell.ZSH, Expected: "[](http://www.google.be)"},
		{Text: "[](http://www.google.be)", ShellName: shell.PWSH, Expected: "[](http://www.google.be)"},
		{Text: "[](http://www.google.be)", ShellName: shell.BASH, Expected: "[](http://www.google.be)"},
	}
	for _, tc := range cases {
		a := AnsiWriter{}
		a.Init(tc.ShellName)
		hyperlinkText := a.GenerateHyperlink(tc.Text)
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
			Expected: "\x1b]8;;file:C:/Program Files (x86)/Common Files/Microsoft Shared/Posh\x1b\\Posh\x1b]8;;\x1b\\",
		},
		{Text: `[Windows](file:C:/Windows)`, Expected: "\x1b]8;;file:C:/Windows\x1b\\Windows\x1b]8;;\x1b\\"},
	}
	for _, tc := range cases {
		a := AnsiWriter{}
		a.Init(shell.PWSH)
		hyperlinkText := a.GenerateHyperlink(tc.Text)
		assert.Equal(t, tc.Expected, hyperlinkText)
	}
}
