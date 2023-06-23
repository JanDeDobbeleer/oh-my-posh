package ansi

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

func TestGetAnsiFromColorString(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   Color
		Color      string
		Background bool
		Color256   bool
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
		{Case: "Non true color TERM", Expected: Color("38;5;146"), Color: "#AABBCC", Color256: true},
	}
	for _, tc := range cases {
		ansiColors := &DefaultColors{}
		ansiColor := ansiColors.ToColor(tc.Color, tc.Background, !tc.Color256)
		assert.Equal(t, tc.Expected, ansiColor, tc.Case)
	}
}

func TestMakeColors(t *testing.T) {
	env := &mock.MockedEnvironment{}
	env.On("Flags").Return(&platform.Flags{
		TrueColor: true,
	})
	env.On("WindowsRegistryKeyValue", `HKEY_CURRENT_USER\Software\Microsoft\Windows\DWM\ColorizationColor`).Return(&platform.WindowsRegistryValue{}, errors.New("err"))
	colors := MakeColors(nil, false, "", env)
	assert.IsType(t, &DefaultColors{}, colors)

	colors = MakeColors(nil, true, "", env)
	assert.IsType(t, &CachedColors{}, colors)
	assert.IsType(t, &DefaultColors{}, colors.(*CachedColors).ansiColors)

	colors = MakeColors(testPalette, false, "", env)
	assert.IsType(t, &PaletteColors{}, colors)
	assert.IsType(t, &DefaultColors{}, colors.(*PaletteColors).ansiColors)

	colors = MakeColors(testPalette, true, "", env)
	assert.IsType(t, &CachedColors{}, colors)
	assert.IsType(t, &PaletteColors{}, colors.(*CachedColors).ansiColors)
	assert.IsType(t, &DefaultColors{}, colors.(*CachedColors).ansiColors.(*PaletteColors).ansiColors)
}
