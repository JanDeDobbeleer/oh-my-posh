package ansi

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestGetAnsiFromColorString(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   Color
		Color      string
		Background bool
	}{
		{Case: "256 color", Expected: Color("38;5;99"), Color: "99", Background: false},
		{Case: "256 color", Expected: Color("38;5;122"), Color: "122", Background: false},
		{Case: "Invalid background", Expected: emptyColor, Color: "invalid", Background: true},
		{Case: "Invalid background", Expected: emptyColor, Color: "invalid", Background: false},
		{Case: "Hex foreground", Expected: Color("38;2;170;187;204"), Color: "#AABBCC", Background: false},
		{Case: "Hex backgrond", Expected: Color("48;2;170;187;204"), Color: "#AABBCC", Background: true},
		{Case: "Base 8 foreground", Expected: Color("31"), Color: "red", Background: false},
		{Case: "Base 8 background", Expected: Color("41"), Color: "red", Background: true},
		{Case: "Base 16 foreground", Expected: Color("91"), Color: "lightRed", Background: false},
		{Case: "Base 16 backround", Expected: Color("101"), Color: "lightRed", Background: true},
	}
	for _, tc := range cases {
		ansiColors := &DefaultColors{}
		ansiColor := ansiColors.ToColor(tc.Color, tc.Background)
		assert.Equal(t, tc.Expected, ansiColor, tc.Case)
	}
}

func TestMakeColors(t *testing.T) {
	colors := MakeColors(nil, false, "", nil)
	assert.IsType(t, &DefaultColors{}, colors)

	colors = MakeColors(nil, true, "", nil)
	assert.IsType(t, &CachedColors{}, colors)
	assert.IsType(t, &DefaultColors{}, colors.(*CachedColors).ansiColors)

	colors = MakeColors(testPalette, false, "", nil)
	assert.IsType(t, &PaletteColors{}, colors)
	assert.IsType(t, &DefaultColors{}, colors.(*PaletteColors).ansiColors)

	colors = MakeColors(testPalette, true, "", nil)
	assert.IsType(t, &CachedColors{}, colors)
	assert.IsType(t, &PaletteColors{}, colors.(*CachedColors).ansiColors)
	assert.IsType(t, &DefaultColors{}, colors.(*CachedColors).ansiColors.(*PaletteColors).ansiColors)
}
