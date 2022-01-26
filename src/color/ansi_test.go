package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLenWithoutAnsi(t *testing.T) {
	cases := []struct {
		Text      string
		ShellName string
		Expected  int
	}{
		{Text: "%{\x1b[44m%}hello%{\x1b[0m%}", ShellName: zsh, Expected: 5},
		{Text: "\x1b[44mhello\x1b[0m", ShellName: pwsh, Expected: 5},
		{Text: "\\[\x1b[44m\\]hello\\[\x1b[0m\\]", ShellName: bash, Expected: 5},
	}
	for _, tc := range cases {
		a := Ansi{}
		a.Init(tc.ShellName)
		strippedLength := a.LenWithoutANSI(tc.Text)
		assert.Equal(t, 5, strippedLength)
	}
}

func TestGenerateHyperlinkNoUrl(t *testing.T) {
	cases := []struct {
		Text      string
		ShellName string
		Expected  string
	}{
		{Text: "sample text with no url", ShellName: zsh, Expected: "sample text with no url"},
		{Text: "sample text with no url", ShellName: pwsh, Expected: "sample text with no url"},
		{Text: "sample text with no url", ShellName: bash, Expected: "sample text with no url"},
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
		{Text: "[google](http://www.google.be)", ShellName: zsh, Expected: "%{\x1b]8;;http://www.google.be\x1b\\%}google%{\x1b]8;;\x1b\\%}"},
		{Text: "[google](http://www.google.be)", ShellName: pwsh, Expected: "\x1b]8;;http://www.google.be\x1b\\google\x1b]8;;\x1b\\"},
		{Text: "[google](http://www.google.be)", ShellName: bash, Expected: "\\[\x1b]8;;http://www.google.be\x1b\\\\\\]google\\[\x1b]8;;\x1b\\\\\\]"},
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
		{Text: "[](http://www.google.be)", ShellName: zsh, Expected: "[](http://www.google.be)"},
		{Text: "[](http://www.google.be)", ShellName: pwsh, Expected: "[](http://www.google.be)"},
		{Text: "[](http://www.google.be)", ShellName: bash, Expected: "[](http://www.google.be)"},
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
	}
	for _, tc := range cases {
		a := Ansi{}
		a.Init("")
		formattedText := a.formatText(tc.Text)
		assert.Equal(t, tc.Expected, formattedText, tc.Case)
	}
}
