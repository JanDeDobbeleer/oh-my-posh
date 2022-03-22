package color

import (
	"oh-my-posh/shell"
	"testing"

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
		a := Ansi{}
		a.Init(tc.ShellName)
		hyperlinkText := a.generateHyperlink(tc.Text)
		assert.Equal(t, tc.Expected, hyperlinkText)
	}
}

func TestGenerateHyperlinkWithUrl(t *testing.T) {
	cases := []struct {
		Text      string
		ShellName string
		Expected  string
	}{
		{Text: "[google](http://www.google.be)", ShellName: shell.ZSH, Expected: "%{\x1b]8;;http://www.google.be\x1b\\%}google%{\x1b]8;;\x1b\\%}"},
		{Text: "[google](http://www.google.be)", ShellName: shell.PWSH, Expected: "\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\"},
		{Text: "[google](http://www.google.be)", ShellName: shell.BASH, Expected: "\\[\x1b]8;;http://www.google.be\x1b\\\\\\]google\\[\x1b]8;;\x1b\\\\\\]"},
	}
	for _, tc := range cases {
		a := Ansi{}
		a.Init(tc.ShellName)
		hyperlinkText := a.generateHyperlink(tc.Text)
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
		a := Ansi{}
		a.Init(tc.ShellName)
		hyperlinkText := a.generateHyperlink(tc.Text)
		assert.Equal(t, tc.Expected, hyperlinkText)
	}
}

func TestFormatText(t *testing.T) {
	cases := []struct {
		Case     string
		Text     string
		Expected string
	}{
		{Case: "single format", Text: "This <b>is</b> white", Expected: "This \x1b[1mis\x1b[22m white"},
		{Case: "double format", Text: "This <b>is</b> white, this <b>is</b> orange", Expected: "This \x1b[1mis\x1b[22m white, this \x1b[1mis\x1b[22m orange"},
		{Case: "underline", Text: "This <u>is</u> white", Expected: "This \x1b[4mis\x1b[24m white"},
		{Case: "italic", Text: "This <i>is</i> white", Expected: "This \x1b[3mis\x1b[23m white"},
		{Case: "strikethrough", Text: "This <s>is</s> white", Expected: "This \x1b[9mis\x1b[29m white"},
		{Case: "dimmed", Text: "This <d>is</d> white", Expected: "This \x1b[2mis\x1b[22m white"},
		{Case: "flash", Text: "This <f>is</f> white", Expected: "This \x1b[5mis\x1b[25m white"},
		{Case: "reversed", Text: "This <r>is</r> white", Expected: "This \x1b[7mis\x1b[27m white"},
		{Case: "double", Text: "This <i><f>is</f></i> white", Expected: "This \x1b[3m\x1b[5mis\x1b[25m\x1b[23m white"},
	}
	for _, tc := range cases {
		a := Ansi{}
		a.Init("")
		formattedText := a.formatText(tc.Text)
		assert.Equal(t, tc.Expected, formattedText, tc.Case)
	}
}
