package main

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
		a := ansiFormats{}
		a.init(tc.ShellName)
		strippedLength := a.lenWithoutANSI(tc.Text)
		assert.Equal(t, 5, strippedLength)
	}
}
