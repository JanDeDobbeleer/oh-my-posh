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
	actual, err := testPalette.ResolveColor(tc.Request)

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
		{
			Case:          "Palette deep purple",
			Request:       "palette:deep-purple",
			ExpectedError: true,
			Expected:      "palette: requested color deep-purple does not exist in palette of colors black,blue,green,red,white",
		},
		{
			Case:          "Palette cyan",
			Request:       "palette:cyan",
			ExpectedError: true,
			Expected:      "palette: requested color cyan does not exist in palette of colors black,blue,green,red,white",
		},
		{
			Case:          "Palette foreground",
			Request:       "palette:foreground",
			ExpectedError: true,
			Expected:      "palette: requested color foreground does not exist in palette of colors black,blue,green,red,white",
		},
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
		actual := testPalette.MaybeResolveColor(tc.Request)

		assert.Equal(t, tc.Expected, actual, "expected different color value")
	}
}

func TestPaletteShouldNotResolveRecursiveReference(t *testing.T) {
	tp := Palette{
		"light-blue": "#CAF0F8",
		"dark-blue":  "#023E8A",
		"background": "p:dark-blue",
		"foreground": "p:light-blue",
	}

	cases := []TestPaletteRequest{
		{
			Case:     "Palette light-blue",
			Request:  "p:light-blue",
			Expected: "#CAF0F8",
		},
		{
			Case:          "Palette background",
			Request:       "p:background",
			ExpectedError: true,
			Expected:      "palette: resolution of color background returned palette reference p:dark-blue; recursive references are not supported",
		},
		{
			Case:          "Palette foreground",
			Request:       "p:foreground",
			ExpectedError: true,
			Expected:      "palette: resolution of color foreground returned palette reference p:light-blue; recursive references are not supported",
		},
	}

	for _, tc := range cases {
		actual, err := tp.ResolveColor(tc.Request)

		if !tc.ExpectedError {
			assert.Nil(t, err, "expected no error")
			assert.Equal(t, tc.Expected, actual, "expected different color value")
		} else {
			assert.NotNil(t, err, "expected error")
			assert.Equal(t, tc.Expected, err.Error())
		}
	}
}

func TestPaletteShouldHandleEmptyKey(t *testing.T) {
	tp := Palette{
		"": "#000000",
	}

	actual, err := tp.ResolveColor("palette:")

	assert.Nil(t, err, "expected no error")
	assert.Equal(t, "#000000", actual, "expected different color value")
}

func BenchmarkPaletteMixedCaseResolution(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkPaletteMixedCaseResolution()
	}
}

func benchmarkPaletteMixedCaseResolution() {
	cases := []TestPaletteRequest{
		{Case: "Palette red", Request: "palette:red", Expected: "#FF0000"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Cyan", Request: "#05E6FA", Expected: "#05E6FA"},
		{Case: "Palette black", Request: "palette:black", Expected: "#000000"},
		{Case: "Palette pink", Request: "palette:pink", ExpectedError: true, Expected: "palette: requested color pink does not exist in palette of colors black,blue,green,red,white"},
		{Case: "Palette blue", Request: "p:blue", Expected: "#0000FF"},
		// repeating the same set to have longer benchmarks
		{Case: "Palette red", Request: "palette:red", Expected: "#FF0000"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Cyan", Request: "#05E6FA", Expected: "#05E6FA"},
		{Case: "Palette black", Request: "palette:black", Expected: "#000000"},
		{Case: "Palette pink", Request: "palette:pink", ExpectedError: true, Expected: "palette: requested color pink does not exist in palette of colors black,blue,green,red,white"},
		{Case: "Palette blue", Request: "p:blue", Expected: "#0000FF"},
	}

	for _, tc := range cases {
		// both value and error values are irrelevant, but such assignment calms down
		// golangci-lint "return value of `testPalette.ResolveColor` is not checked" error
		_, _ = testPalette.ResolveColor(tc.Request)
	}
}
