package color

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/shell"

	"github.com/stretchr/testify/assert"
)

func TestWriteANSIColors(t *testing.T) {
	cases := []struct {
		Case               string
		Expected           string
		Input              string
		Colors             *Color
		Parent             *Color
		TerminalBackground string
	}{
		{
			Case:     "No color override",
			Input:    "test",
			Expected: "\x1b[47m\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Inherit foreground",
			Input:    "test",
			Expected: "\x1b[47m\x1b[33mtest\x1b[0m",
			Colors:   &Color{Foreground: ParentForeground, Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "white"},
		},
		{
			Case:     "Inherit background",
			Input:    "test",
			Expected: "\x1b[41m\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: ParentBackground},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "No parent",
			Input:    "test",
			Expected: "\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Inherit override foreground",
			Input:    "hello <parentForeground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[47m\x1b[33mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override background",
			Input:    "hello <black,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[41m\x1b[30mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override background, no foreground specified",
			Input:    "hello <,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[41m\x1b[30mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit no parent foreground",
			Input:    "hello <parentForeground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[47;49m\x1b[7mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Inherit no parent background",
			Input:    "hello <,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[30mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Inherit override both",
			Input:    "hello <parentForeground,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[41m\x1b[33mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override both inverted",
			Input:    "hello <parentBackground,parentForeground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[43m\x1b[31mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inline override",
			Input:    "hello, <red>world</>, rabbit",
			Expected: "\x1b[47m\x1b[30mhello, \x1b[0m\x1b[47m\x1b[31mworld\x1b[0m\x1b[47m\x1b[30m, rabbit\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Transparent background",
			Input:    "hello world",
			Expected: "\x1b[37mhello world\x1b[0m",
			Colors:   &Color{Foreground: "white", Background: Transparent},
		},
		{
			Case:     "Transparent foreground override",
			Input:    "hello <#ffffff>world</>",
			Expected: "\x1b[32mhello \x1b[0m\x1b[38;2;255;255;255mworld\x1b[0m",
			Colors:   &Color{Foreground: "green", Background: Transparent},
		},
		{
			Case:     "Double override",
			Input:    "<#ffffff>jan</>@<#ffffff>Jans-MBP</>",
			Expected: "\x1b[48;2;255;87;51m\x1b[38;2;255;255;255mjan\x1b[0m\x1b[48;2;255;87;51m\x1b[32m@\x1b[0m\x1b[48;2;255;87;51m\x1b[38;2;255;255;255mJans-MBP\x1b[0m",
			Colors:   &Color{Foreground: "green", Background: "#FF5733"},
		},
		{
			Case:     "No foreground",
			Input:    "test",
			Expected: "\x1b[48;2;255;87;51m\x1b[37mtest\x1b[0m",
			Colors:   &Color{Foreground: "", Background: "#FF5733"},
		},
		{
			Case:     "Transparent foreground",
			Input:    "test",
			Expected: "\x1b[38;2;255;87;51;49m\x1b[7mtest\x1b[0m",
			Colors:   &Color{Foreground: Transparent, Background: "#FF5733"},
		},
		{
			Case:               "Transparent foreground, terminal background set",
			Input:              "test",
			Expected:           "\x1b[48;2;255;87;51m\x1b[38;2;33;47;60mtest\x1b[0m",
			Colors:             &Color{Foreground: Transparent, Background: "#FF5733"},
			TerminalBackground: "#212F3C",
		},
		{
			Case:     "Foreground for foreground override",
			Input:    "<foreground>test</>",
			Expected: "\x1b[47m\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Foreground for background override",
			Input:    "<background>test</>",
			Expected: "\x1b[47m\x1b[37mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Foreground for background vice versa override",
			Input:    "<background,foreground>test</>",
			Expected: "\x1b[40m\x1b[37mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Background for background override",
			Input:    "<,background>test</>",
			Expected: "\x1b[47m\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Background for foreground override",
			Input:    "<,foreground>test</>",
			Expected: "\x1b[40m\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Google",
			Input:    "<blue,white>G</><red,white>o</><yellow,white>o</><blue,white>g</><green,white>l</><red,white>e</>",
			Expected: "\x1b[47m\x1b[34mG\x1b[0m\x1b[47m\x1b[31mo\x1b[0m\x1b[47m\x1b[33mo\x1b[0m\x1b[47m\x1b[34mg\x1b[0m\x1b[47m\x1b[32ml\x1b[0m\x1b[47m\x1b[31me\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "black"},
		},
	}

	for _, tc := range cases {
		ansi := &Ansi{}
		ansi.Init(shell.PWSH)
		renderer := &AnsiWriter{
			Ansi:               ansi,
			ParentColors:       []*Color{tc.Parent},
			Colors:             tc.Colors,
			TerminalBackground: tc.TerminalBackground,
			AnsiColors:         &DefaultColors{},
		}
		renderer.Write(tc.Colors.Background, tc.Colors.Foreground, tc.Input)
		got, _ := renderer.String()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
