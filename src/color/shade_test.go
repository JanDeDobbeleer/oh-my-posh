package color

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/lucasb-eyer/go-colorful"
)

func TestShadeArgs(t *testing.T) {
	cases := []struct {
		Case            string
		Color           Ansi
		ExpectedInner   Ansi
		ExpectedDir     shadeDirection
		ExpectedPercent float64
		ExpectedOK      bool
	}{
		{Case: "darken", Color: "darken(#3465A4, 20)", ExpectedOK: true, ExpectedDir: shadeDark, ExpectedInner: "#3465A4", ExpectedPercent: 20},
		{Case: "lighten", Color: "lighten(#3465A4, 20)", ExpectedOK: true, ExpectedDir: shadeLight, ExpectedInner: "#3465A4", ExpectedPercent: 20},
		{Case: "palette reference", Color: "darken(p:accent, 20)", ExpectedOK: true, ExpectedDir: shadeDark, ExpectedInner: "p:accent", ExpectedPercent: 20},
		{Case: "whitespace variants", Color: "darken( #3465A4  ,  20 )", ExpectedOK: true, ExpectedDir: shadeDark, ExpectedInner: "#3465A4", ExpectedPercent: 20},
		{Case: "decimal percent", Color: "darken(#3465A4, 12.5)", ExpectedOK: true, ExpectedDir: shadeDark, ExpectedInner: "#3465A4", ExpectedPercent: 12.5},
		{Case: "percent 0", Color: "darken(#3465A4, 0)", ExpectedOK: true, ExpectedDir: shadeDark, ExpectedInner: "#3465A4", ExpectedPercent: 0},
		{Case: "percent 100", Color: "darken(#3465A4, 100)", ExpectedOK: true, ExpectedDir: shadeDark, ExpectedInner: "#3465A4", ExpectedPercent: 100},
		{Case: "not a shade call", Color: "#3465A4"},
		{Case: "gradient", Color: "linear-gradient(#FF0000, #0000FF)"},
		{Case: "missing closing paren", Color: "darken(#3465A4, 20"},
		{Case: "nested parens", Color: "darken(darken(#3465A4, 10), 20)"},
		{Case: "missing percent", Color: "darken(#3465A4)"},
		{Case: "too many args", Color: "darken(#3465A4, 20, 30)"},
		{Case: "non-numeric percent", Color: "darken(#3465A4, abc)"},
		{Case: "percent below 0", Color: "darken(#3465A4, -1)"},
		{Case: "percent above 100", Color: "darken(#3465A4, 101)"},
		{Case: "empty color arg", Color: "darken(, 20)"},
	}

	for _, tc := range cases {
		dir, inner, percent, ok := tc.Color.ShadeArgs()

		assert.Equal(t, tc.ExpectedOK, ok, tc.Case)

		if !tc.ExpectedOK {
			continue
		}

		assert.Equal(t, tc.ExpectedDir, dir, tc.Case)
		assert.Equal(t, tc.ExpectedInner, inner, tc.Case)
		assert.Equal(t, tc.ExpectedPercent, percent, tc.Case)
	}
}

func TestIsShade(t *testing.T) {
	cases := []struct {
		Case     string
		Color    Ansi
		Expected bool
	}{
		{Case: "darken", Color: "darken(#3465A4, 20)", Expected: true},
		{Case: "lighten", Color: "lighten(#3465A4, 20)", Expected: true},
		{Case: "gradient", Color: "linear-gradient(#FF0000, #0000FF)", Expected: false},
		{Case: "dark-gradient", Color: "dark-gradient(#3465A4)", Expected: false},
		{Case: "hex", Color: "#3465A4", Expected: false},
		{Case: "empty", Color: "", Expected: false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, tc.Color.IsShade(), tc.Case)
	}
}

func TestShadeHex(t *testing.T) {
	base := "#3465A4"

	darker, ok := shadeHex(Ansi(base), shadeDark, 20)
	assert.True(t, ok, "darken should resolve a hex color")

	lighter, ok := shadeHex(Ansi(base), shadeLight, 20)
	assert.True(t, ok, "lighten should resolve a hex color")

	baseColor, _ := colorful.Hex(base)
	darkerColor, _ := colorful.Hex(darker)
	lighterColor, _ := colorful.Hex(lighter)

	_, _, baseLightness := baseColor.Hcl()
	_, _, darkerLightness := darkerColor.Hcl()
	_, _, lighterLightness := lighterColor.Hcl()

	assert.True(t, darkerLightness < baseLightness, "darken should reduce lightness")
	assert.True(t, lighterLightness > baseLightness, "lighten should increase lightness")

	unchanged, ok := shadeHex(Ansi(base), shadeDark, 0)
	assert.True(t, ok, "0%% shade should still resolve")
	assert.Equal(t, "#3465a4", unchanged, "0%% darken should leave the color unchanged")

	black, ok := shadeHex(Ansi(base), shadeDark, 100)
	assert.True(t, ok, "100%% darken should resolve")
	assert.Equal(t, "#000000", black, "100%% darken should reach black")

	white, ok := shadeHex(Ansi(base), shadeLight, 100)
	assert.True(t, ok, "100%% lighten should resolve")

	whiteColor, err := colorful.Hex(white)
	assert.Nil(t, err, "100%% lighten should still be a valid hex color")
	r, g, b := whiteColor.RGB255()
	// the gamut-walk loop only guarantees IsValid, not zero chroma, so 100% lighten
	// lands very close to white rather than exactly #ffffff
	assert.True(t, r >= 250 && g >= 250 && b >= 250, "100%% lighten should reach near-white, got "+white)

	_, ok = shadeHex("green", shadeDark, 20)
	assert.False(t, ok, "named ANSI colors have no known RGB and cannot be shaded")

	_, ok = shadeHex("146", shadeDark, 20)
	assert.False(t, ok, "256 color indices have no known RGB and cannot be shaded")

	_, ok = shadeHex("not-a-color", shadeDark, 20)
	assert.False(t, ok, "an unparseable hex string should not resolve")
}

func TestDefaultsToAnsiShade(t *testing.T) {
	expectedHex, _ := shadeHex("#3465A4", shadeDark, 20)

	cases := []struct {
		Case       string
		Color      Ansi
		Expected   Ansi
		Background bool
	}{
		{Case: "darken hex foreground", Color: "darken(#3465A4, 20)", Background: false, Expected: (&Defaults{}).ToAnsi(Ansi(expectedHex), false)},
		{Case: "darken hex background", Color: "darken(#3465A4, 20)", Background: true, Expected: (&Defaults{}).ToAnsi(Ansi(expectedHex), true)},
		{Case: "named ANSI color is rejected", Color: "darken(green, 20)", Background: false, Expected: emptyColor},
		{Case: "256 color index is rejected", Color: "darken(146, 20)", Background: false, Expected: emptyColor},
		{Case: "keyword is rejected", Color: "darken(foreground, 20)", Background: false, Expected: emptyColor},
	}

	for _, tc := range cases {
		ansiColors := &Defaults{}
		assert.Equal(t, tc.Expected, ansiColors.ToAnsi(tc.Color, tc.Background), tc.Case)
	}
}

func TestPaletteColorsResolveShade(t *testing.T) {
	colors := &PaletteColors{ansiColors: &Defaults{}, palette: testPalette}

	expectedHex, _ := shadeHex("#FF0000", shadeDark, 30)
	expected := (&Defaults{}).ToAnsi(Ansi(expectedHex), false)

	assert.Equal(t, expected, colors.ToAnsi("darken(p:red, 30)", false), "darken should resolve a palette reference before shading")

	resolved, err := colors.Resolve("darken(p:red, 30)")
	assert.Nil(t, err)
	assert.Equal(t, withShadeCall(shadeDark, "#FF0000", 30), resolved, "Resolve should rebuild the call with the palette reference resolved")

	assert.Equal(t, emptyColor, colors.ToAnsi("darken(p:missing, 30)", false), "an unresolvable palette reference should fail closed")
}
