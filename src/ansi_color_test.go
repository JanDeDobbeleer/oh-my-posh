package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAnsiFromColorString(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   string
		Color      string
		Background bool
	}{
		{Case: "Invalid background", Expected: "", Color: "invalid", Background: true},
		{Case: "Invalid background", Expected: "", Color: "invalid", Background: false},
		{Case: "Hex foreground", Expected: "48;2;170;187;204", Color: "#AABBCC", Background: false},
		{Case: "Base 8 foreground", Expected: "41", Color: "red", Background: false},
		{Case: "Base 8 background", Expected: "41", Color: "red", Background: true},
		{Case: "Base 16 foreground", Expected: "101", Color: "lightRed", Background: false},
		{Case: "Base 16 backround", Expected: "101", Color: "lightRed", Background: true},
	}
	for _, tc := range cases {
		renderer := &AnsiColor{}
		ansiColor := renderer.getAnsiFromColorString(tc.Color, true)
		assert.Equal(t, tc.Expected, ansiColor, tc.Case)
	}
}

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
			Colors:   &Color{Foreground: Inherit, Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "white"},
		},
		{
			Case:     "Inherit background",
			Input:    "test",
			Expected: "\x1b[41m\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: Inherit},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "No parent",
			Input:    "test",
			Expected: "\x1b[30mtest\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: Inherit},
		},
		{
			Case:     "Inherit override foreground",
			Input:    "hello <inherit>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[47m\x1b[33mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override background",
			Input:    "hello <black,inherit>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[41m\x1b[30mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override background, no foreground specified",
			Input:    "hello <,inherit>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[41m\x1b[30mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override both",
			Input:    "hello <inherit,inherit>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[41m\x1b[33mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
			Parent:   &Color{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit no parent foreground",
			Input:    "hello <inherit>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[47;49m\x1b[7mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Inherit no parent background",
			Input:    "hello <,inherit>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[30mworld\x1b[0m",
			Colors:   &Color{Foreground: "black", Background: "white"},
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
			Expected:           "\x1b[38;2;255;87;51m\x1b[38;2;33;47;60mtest\x1b[0m",
			Colors:             &Color{Foreground: Transparent, Background: "#FF5733"},
			TerminalBackground: "#212F3C",
		},
	}

	for _, tc := range cases {
		ansi := &ansiUtils{}
		ansi.init("pwsh")
		renderer := &AnsiColor{
			ansi:               ansi,
			Parent:             tc.Parent,
			terminalBackground: tc.TerminalBackground,
		}
		renderer.write(tc.Colors.Background, tc.Colors.Foreground, tc.Input)
		got := renderer.string()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
