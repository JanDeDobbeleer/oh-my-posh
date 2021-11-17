package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

var (
	testPalette = Palette{
		"red":   "#FF0000",
		"green": "#00FF00",
		"blue":  "#0000FF",
		"white": "#FFFFFF",
		"black": "#000000",
	}
)

type TestPaletteRequest struct {
	Case          string
	Request       string
	ExpectedError bool
	Expected      string
}

func TestPaletteShouldResolveColorFromTestPalette(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Palette red", Request: "palette:red", Expected: "#FF0000"},
		{Case: "Palette green", Request: "palette:green", Expected: "#00FF00"},
		{Case: "Palette blue", Request: "palette:blue", Expected: "#0000FF"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func testPaletteRequest(t *testing.T, tc TestPaletteRequest) {
	actual, err := testPalette.resolveColor(tc.Request)

	if !tc.ExpectedError {
		assert.Nil(t, err, "expected no error")
		assert.Equal(t, tc.Expected, actual, "expected different color value")
	} else {
		assert.NotNil(t, err, "expected error")
		assert.Equal(t, tc.Expected, err.Error())
	}
}

func TestPaletteShouldIgnoreNonPaletteColors(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Deep puprple", Request: "#1F1137", Expected: "#1F1137"},
		{Case: "Light red", Request: "#D55252", Expected: "#D55252"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Foreground", Request: "foreground", Expected: "foreground"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func TestPaletteShouldReturnErrorOnMissingColor(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Palette deep purple", Request: "palette:deep-purple", ExpectedError: true, Expected: "palette: requested color deep-purple does not exist in palette of colors black,blue,green,red,white"},
		{Case: "Palette cyan", Request: "palette:cyan", ExpectedError: true, Expected: "palette: requested color cyan does not exist in palette of colors black,blue,green,red,white"},
		{Case: "Palette foreground", Request: "palette:foreground", ExpectedError: true, Expected: "palette: requested color foreground does not exist in palette of colors black,blue,green,red,white"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func TestPaletteShouldHandleMixedCases(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Palette red", Request: "palette:red", Expected: "#FF0000"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Cyan", Request: "#05E6FA", Expected: "#05E6FA"},
		{Case: "Palette black", Request: "palette:black", Expected: "#000000"},
		{Case: "Palette pink", Request: "palette:pink", ExpectedError: true, Expected: "palette: requested color pink does not exist in palette of colors black,blue,green,red,white"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func TestPaletteShouldAllowShortReference(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Palette red", Request: "p:red", Expected: "#FF0000"},
		{Case: "Palette green", Request: "palette:green", Expected: "#00FF00"},
		{Case: "Palette blue", Request: "p:blue", Expected: "#0000FF"},
		{Case: "Palette white", Request: "palette:white", Expected: "#FFFFFF"},
		{Case: "Palette red", Request: "p:red", Expected: "#FF0000"},
		{Case: "Palette green", Request: "palette:green", Expected: "#00FF00"},
		{Case: "Palette blue", Request: "p:blue", Expected: "#0000FF"},
		{Case: "Palette black", Request: "palette:black", Expected: "#000000"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func TestPaletteShouldUseTransparentByDefault(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Palette magenta", Request: "p:magenta", Expected: Transparent},
		{Case: "Palette gray", Request: "p:gray", Expected: Transparent},
		{Case: "Palette rose", Request: "p:rose", Expected: Transparent},
	}

	for _, tc := range cases {
		actual := testPalette.maybeResolveColor(tc.Request)

		assert.Equal(t, tc.Expected, actual, "expected different color value")
	}
}
