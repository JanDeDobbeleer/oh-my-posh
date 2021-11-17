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
	Case     string
	Request  string
	Expected string
}

func TestPaletteShouldResolveColorFromTestPalette(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Red", Request: "palette:red", Expected: "#FF0000"},
		{Case: "Green", Request: "palette:green", Expected: "#00FF00"},
		{Case: "Red", Request: "palette:blue", Expected: "#0000FF"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func testPaletteRequest(t *testing.T, tc TestPaletteRequest) {
	actual, err := testPalette.resolveColor(tc.Request)

	assert.Nil(t, err, "expected no error")
	assert.Equal(t, tc.Expected, actual, "expected different color value")
}

func TestPaletteShouldIgnoreNonPaletteColors(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Deep puprple", Request: "#1F1137", Expected: "#1F1137"},
		{Case: "Palette red", Request: "palette:red", Expected: "#FF0000"},
		{Case: "Light red", Request: "#D55252", Expected: "#D55252"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Palette black", Request: "palette:black", Expected: "#000000"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}
