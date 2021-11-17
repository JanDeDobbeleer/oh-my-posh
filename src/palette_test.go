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

func TestPaletteShouldResolveColorFromTestPalette(t *testing.T) {
	cases := []struct {
		Case     string
		Request  string
		Expected string
	}{
		{Case: "Red", Request: "palette:red", Expected: "#FF0000"},
		{Case: "Green", Request: "palette:green", Expected: "#00FF00"},
		{Case: "Red", Request: "palette:blue", Expected: "#0000FF"},
	}

	for _, tc := range cases {
		actual, err := testPalette.resolveColor(tc.Request)

		assert.Nil(t, err, "expected no error")
		assert.Equal(t, tc.Expected, actual, "expected different color value")
	}
}
