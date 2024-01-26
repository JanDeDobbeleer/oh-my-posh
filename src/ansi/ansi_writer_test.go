package ansi

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

func TestWriteANSIColors(t *testing.T) {
	cases := []struct {
		Case               string
		Expected           string
		Input              string
		Colors             *Colors
		Parent             *Colors
		TerminalBackground string
	}{
		{
			Case:     "Inline override identical",
			Input:    "\ue0a0saturnay <red>↑</>1",
			Expected: "\x1b[31m\ue0a0saturnay ↑1\x1b[0m",
			Colors:   &Colors{Foreground: "red", Background: Transparent},
		},
		{
			Case:     "Bold",
			Input:    "<b>test</b>",
			Expected: "\x1b[1m\x1b[30mtest\x1b[22m\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Bold with color override",
			Input:    "<b><#ffffff>test</></b>",
			Expected: "\x1b[1m\x1b[30m\x1b[38;2;255;255;255mtest\x1b[30m\x1b[22m\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Bold with color override, flavor 2",
			Input:    "<#ffffff><b>test</b></>",
			Expected: "\x1b[38;2;255;255;255m\x1b[1mtest\x1b[22m\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Double override",
			Input:    "<#ffffff>jan</>@<#ffffff>Jans-MBP</>",
			Expected: "\x1b[48;2;255;87;51m\x1b[38;2;255;255;255mjan\x1b[32m@\x1b[38;2;255;255;255mJans-MBP\x1b[0m",
			Colors:   &Colors{Foreground: "green", Background: "#FF5733"},
		},
		{
			Case:     "No color override",
			Input:    "test",
			Expected: "\x1b[47m\x1b[30mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
			Parent:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Inherit foreground",
			Input:    "test",
			Expected: "\x1b[47m\x1b[33mtest\x1b[0m",
			Colors:   &Colors{Foreground: ParentForeground, Background: "white"},
			Parent:   &Colors{Foreground: "yellow", Background: "white"},
		},
		{
			Case:     "Inherit background",
			Input:    "test",
			Expected: "\x1b[41m\x1b[30mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
			Parent:   &Colors{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "No parent",
			Input:    "test",
			Expected: "\x1b[30mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Inherit override foreground",
			Input:    "hello <parentForeground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[33mworld\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
			Parent:   &Colors{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override background",
			Input:    "hello <black,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[41mworld\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
			Parent:   &Colors{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override background, no foreground specified",
			Input:    "hello <,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[41mworld\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
			Parent:   &Colors{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit no parent foreground",
			Input:    "hello <parentForeground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[0m\x1b[37;49m\x1b[7mworld\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Inherit no parent background",
			Input:    "hello <,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[49mworld\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Inherit override both",
			Input:    "hello <parentForeground,parentBackground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[41m\x1b[33mworld\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
			Parent:   &Colors{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inherit override both inverted",
			Input:    "hello <parentBackground,parentForeground>world</>",
			Expected: "\x1b[47m\x1b[30mhello \x1b[43m\x1b[31mworld\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
			Parent:   &Colors{Foreground: "yellow", Background: "red"},
		},
		{
			Case:     "Inline override",
			Input:    "hello, <red>world</>, rabbit",
			Expected: "\x1b[47m\x1b[30mhello, \x1b[31mworld\x1b[30m, rabbit\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Transparent background",
			Input:    "hello world",
			Expected: "\x1b[37mhello world\x1b[0m",
			Colors:   &Colors{Foreground: "white", Background: Transparent},
		},
		{
			Case:     "Transparent foreground override",
			Input:    "hello <#ffffff>world</>",
			Expected: "\x1b[32mhello \x1b[38;2;255;255;255mworld\x1b[0m",
			Colors:   &Colors{Foreground: "green", Background: Transparent},
		},
		{
			Case:     "No foreground",
			Input:    "test",
			Expected: "\x1b[48;2;255;87;51m\x1b[37mtest\x1b[0m",
			Colors:   &Colors{Foreground: "", Background: "#FF5733"},
		},
		{
			Case:     "Transparent foreground",
			Input:    "test",
			Expected: "\x1b[0m\x1b[38;2;255;87;51;49m\x1b[7mtest\x1b[0m",
			Colors:   &Colors{Foreground: Transparent, Background: "#FF5733"},
		},
		{
			Case:               "Transparent foreground, terminal background set",
			Input:              "test",
			Expected:           "\x1b[38;2;33;47;60m\x1b[48;2;255;87;51mtest\x1b[0m",
			Colors:             &Colors{Foreground: Transparent, Background: "#FF5733"},
			TerminalBackground: "#212F3C",
		},
		{
			Case:     "Foreground for foreground override",
			Input:    "<foreground>test</>",
			Expected: "\x1b[47m\x1b[30mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Background for background override",
			Input:    "<,background>test</>",
			Expected: "\x1b[47m\x1b[30mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Google",
			Input:    "<blue,white>G</><red,white>o</><yellow,white>o</><blue,white>g</><green,white>l</><red,white>e</>",
			Expected: "\x1b[47m\x1b[34mG\x1b[40m\x1b[30m\x1b[47m\x1b[31mo\x1b[40m\x1b[30m\x1b[47m\x1b[33mo\x1b[40m\x1b[30m\x1b[47m\x1b[34mg\x1b[40m\x1b[30m\x1b[47m\x1b[32ml\x1b[40m\x1b[30m\x1b[47m\x1b[31me\x1b[0m", //nolint: lll
			Colors:   &Colors{Foreground: "black", Background: "black"},
		},
		{
			Case:     "Foreground for background override",
			Input:    "<background>test</>",
			Expected: "\x1b[47m\x1b[37mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Foreground for background vice versa override",
			Input:    "<background,foreground>test</>",
			Expected: "\x1b[40m\x1b[37mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Background for foreground override",
			Input:    "<,foreground>test</>",
			Expected: "\x1b[40m\x1b[30mtest\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Nested override",
			Input:    "hello, <red>world, <white>rabbit</> hello</>",
			Expected: "\x1b[47m\x1b[30mhello, \x1b[31mworld, \x1b[37mrabbit\x1b[31m hello\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Transparent override",
			Input:    "home<transparent> / </>code<transparent> / </>src ",
			Expected: "\x1b[47m\x1b[30mhome\x1b[0m\x1b[37;49m\x1b[7m / \x1b[27m\x1b[47m\x1b[30mcode\x1b[0m\x1b[37;49m\x1b[7m / \x1b[27m\x1b[47m\x1b[30msrc \x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
		{
			Case:     "Larger than and less than symbols",
			Input:    "<red><</><orange>></><yellow><</>",
			Expected: "\x1b[47m\x1b[31m<\x1b[30m>\x1b[33m<\x1b[0m",
			Colors:   &Colors{Foreground: "black", Background: "white"},
		},
	}

	for _, tc := range cases {
		renderer := &Writer{
			ParentColors:       []*Colors{tc.Parent},
			Colors:             tc.Colors,
			TerminalBackground: tc.TerminalBackground,
			AnsiColors:         &DefaultColors{},
			TrueColor:          true,
		}
		renderer.Init(shell.GENERIC)
		renderer.Write(tc.Colors.Background, tc.Colors.Foreground, tc.Input)
		got, _ := renderer.String()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestWriteLength(t *testing.T) {
	cases := []struct {
		Case     string
		Expected int
		Input    string
		Colors   *Colors
	}{
		{
			Case:     "Bold",
			Input:    "<b>test</b>",
			Expected: 4,
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Bold with color override",
			Input:    "<b><#ffffff>test</></b>",
			Expected: 4,
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Bold with color override and link",
			Input:    "<b><#ffffff>test</></b> <LINK>https://example.com<TEXT>url</TEXT></LINK>",
			Expected: 8,
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Bold with color override and link and leading/trailing spaces",
			Input:    " <b><#ffffff>test</></b> <LINK>https://example.com<TEXT>url</TEXT></LINK> ",
			Expected: 10,
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
		{
			Case:     "Bold with color override and link without text and leading/trailing spaces",
			Input:    " <b><#ffffff>test</></b> <LINK>https://example.com<TEXT></TEXT></LINK> ",
			Expected: 11,
			Colors:   &Colors{Foreground: "black", Background: ParentBackground},
		},
	}

	for _, tc := range cases {
		renderer := &Writer{
			ParentColors: []*Colors{},
			Colors:       tc.Colors,
			AnsiColors:   &DefaultColors{},
		}
		renderer.Init(shell.GENERIC)
		renderer.Write(tc.Colors.Background, tc.Colors.Foreground, tc.Input)
		_, got := renderer.String()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
