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
		{Case: "Palette red", Request: "p:red", Expected: "#FF0000"},
		{Case: "Palette green", Request: "p:green", Expected: "#00FF00"},
		{Case: "Palette blue", Request: "p:blue", Expected: "#0000FF"},
		{Case: "Palette white", Request: "p:white", Expected: "#FFFFFF"},
		{Case: "Palette black", Request: "p:black", Expected: "#000000"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func testPaletteRequest(t *testing.T, tc TestPaletteRequest) {
	actual, err := testPalette.ResolveColor(tc.Request)

	if !tc.ExpectedError {
		assert.Nil(t, err, tc.Case)
		assert.Equal(t, tc.Expected, actual, "expected different color value")
	} else {
		assert.NotNil(t, err, tc.Case)
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
			Request:       "p:deep-purple",
			ExpectedError: true,
			Expected:      "palette: requested color deep-purple does not exist in palette of colors black,blue,green,red,white",
		},
		{
			Case:          "Palette cyan",
			Request:       "p:cyan",
			ExpectedError: true,
			Expected:      "palette: requested color cyan does not exist in palette of colors black,blue,green,red,white",
		},
		{
			Case:          "Palette foreground",
			Request:       "p:foreground",
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
		{Case: "Palette red", Request: "p:red", Expected: "#FF0000"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Cyan", Request: "#05E6FA", Expected: "#05E6FA"},
		{Case: "Palette black", Request: "p:black", Expected: "#000000"},
		{Case: "Palette pink", Request: "p:pink", ExpectedError: true, Expected: "palette: requested color pink does not exist in palette of colors black,blue,green,red,white"},
	}

	for _, tc := range cases {
		testPaletteRequest(t, tc)
	}
}

func TestPaletteShouldUseEmptyColorByDefault(t *testing.T) {
	cases := []TestPaletteRequest{
		{Case: "Palette magenta", Request: "p:magenta", Expected: ""},
		{Case: "Palette gray", Request: "p:gray", Expected: ""},
		{Case: "Palette rose", Request: "p:rose", Expected: ""},
	}

	for _, tc := range cases {
		actual := testPalette.MaybeResolveColor(tc.Request)

		assert.Equal(t, tc.Expected, actual, "expected different color value")
	}
}

func TestPaletteShouldResolveRecursiveReference(t *testing.T) {
	tp := Palette{
		"light-blue": "#CAF0F8",
		"dark-blue":  "#023E8A",
		"foreground": "p:light-blue",
		"background": "p:dark-blue",
		"text":       "p:foreground",
		"icon":       "p:background",
		"void":       "p:void", // infinite recursion - error
		"1":          "white",
		"2":          "p:1",
		"3":          "p:2",
		"4":          "p:3", // 3 recursive lookups - allowed
		"5":          "p:4", // 4 recursive lookups - error
	}

	cases := []TestPaletteRequest{
		{
			Case:     "Palette light-blue",
			Request:  "p:light-blue",
			Expected: "#CAF0F8",
		},
		{
			Case:     "Palette foreground",
			Request:  "p:foreground",
			Expected: "#CAF0F8",
		},
		{
			Case:     "Palette background",
			Request:  "p:background",
			Expected: "#023E8A",
		},
		{
			Case:     "Palette text (2 recursive lookups)",
			Request:  "p:text",
			Expected: "#CAF0F8",
		},
		{
			Case:     "Palette icon (2 recursive lookups)",
			Request:  "p:icon",
			Expected: "#023E8A",
		},
		{
			Case:          "Palette void (infinite recursion)",
			Request:       "p:void",
			ExpectedError: true,
			Expected:      "palette: recursive resolution of color p:void returned palette reference p:void and reached recursion depth 4",
		},
		{
			Case:     "Palette p:4 (3 recursive lookups)",
			Request:  "p:4",
			Expected: "white",
		},
		{
			Case:          "Palette p:5 (4 recursive lookups)",
			Request:       "p:5",
			ExpectedError: true,
			Expected:      "palette: recursive resolution of color p:5 returned palette reference p:1 and reached recursion depth 4",
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

	actual, err := tp.ResolveColor("p:")

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
		{Case: "Palette red", Request: "p:red", Expected: "#FF0000"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Cyan", Request: "#05E6FA", Expected: "#05E6FA"},
		{Case: "Palette black", Request: "p:black", Expected: "#000000"},
		{Case: "Palette pink", Request: "p:pink", ExpectedError: true, Expected: "palette: requested color pink does not exist in palette of colors black,blue,green,red,white"},
		{Case: "Palette blue", Request: "p:blue", Expected: "#0000FF"},
		// repeating the same set to have longer benchmarks
		{Case: "Palette red", Request: "p:red", Expected: "#FF0000"},
		{Case: "ANSI black", Request: "black", Expected: "black"},
		{Case: "Cyan", Request: "#05E6FA", Expected: "#05E6FA"},
		{Case: "Palette black", Request: "p:black", Expected: "#000000"},
		{Case: "Palette pink", Request: "p:pink", ExpectedError: true, Expected: "palette: requested color pink does not exist in palette of colors black,blue,green,red,white"},
		{Case: "Palette blue", Request: "p:blue", Expected: "#0000FF"},
	}

	for _, tc := range cases {
		// both value and error values are irrelevant, but such assignment calms down
		// golangci-lint "return value of `testPalette.ResolveColor` is not checked" error
		_, _ = testPalette.ResolveColor(tc.Request)
	}
}
