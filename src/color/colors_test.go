package color

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
)

func TestGetAnsiFromColorString(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   Ansi
		Color      Ansi
		Background bool
		Color256   bool
	}{
		{Case: "256 color", Expected: Ansi("38;5;99"), Color: "99", Background: false},
		{Case: "256 color", Expected: Ansi("38;5;122"), Color: "122", Background: false},
		{Case: "Invalid background", Expected: emptyColor, Color: "invalid", Background: true},
		{Case: "Invalid background", Expected: emptyColor, Color: "invalid", Background: false},
		{Case: "Hex foreground", Expected: Ansi("38;2;170;187;204"), Color: "#AABBCC", Background: false},
		{Case: "Hex backgrond", Expected: Ansi("48;2;170;187;204"), Color: "#AABBCC", Background: true},
		{Case: "Base 8 foreground", Expected: Ansi("31"), Color: "red", Background: false},
		{Case: "Base 8 background", Expected: Ansi("41"), Color: "red", Background: true},
		{Case: "Base 16 foreground", Expected: Ansi("91"), Color: "lightRed", Background: false},
		{Case: "Base 16 backround", Expected: Ansi("101"), Color: "lightRed", Background: true},
		{Case: "Non true color TERM", Expected: Ansi("38;5;146"), Color: "#AABBCC", Color256: true},
	}
	for _, tc := range cases {
		ansiColors := &Defaults{}
		TrueColor = !tc.Color256
		ansiColor := ansiColors.ToAnsi(tc.Color, tc.Background)
		assert.Equal(t, tc.Expected, ansiColor, tc.Case)
	}
}

func TestMakeColors(t *testing.T) {
	env := &mock.Environment{}

	env.On("WindowsRegistryKeyValue", `HKEY_CURRENT_USER\Software\Microsoft\Windows\DWM\ColorizationColor`).Return(&runtime.WindowsRegistryValue{}, errors.New("err"))
	colors := MakeColors(nil, false, "", env)
	assert.IsType(t, &Defaults{}, colors)

	colors = MakeColors(nil, true, "", env)
	assert.IsType(t, &Cached{}, colors)
	assert.IsType(t, &Defaults{}, colors.(*Cached).ansiColors)

	colors = MakeColors(testPalette, false, "", env)
	assert.IsType(t, &PaletteColors{}, colors)
	assert.IsType(t, &Defaults{}, colors.(*PaletteColors).ansiColors)

	colors = MakeColors(testPalette, true, "", env)
	assert.IsType(t, &Cached{}, colors)
	assert.IsType(t, &PaletteColors{}, colors.(*Cached).ansiColors)
	assert.IsType(t, &Defaults{}, colors.(*Cached).ansiColors.(*PaletteColors).ansiColors)
}
