package image

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetOutputPath(t *testing.T) {
	cases := []struct {
		Case     string
		Config   string
		Path     string
		Expected string
	}{
		{Case: "default config", Expected: "prompt.png"},
		{Case: "hidden file", Config: ".posh.omp.json", Expected: "posh.png"},
		{Case: "hidden file toml", Config: ".posh.omp.toml", Expected: "posh.png"},
		{Case: "hidden file yaml", Config: ".posh.omp.yaml", Expected: "posh.png"},
		{Case: "hidden file yml", Config: ".posh.omp.yml", Expected: "posh.png"},
		{Case: "path provided", Path: "mytheme.png", Expected: "mytheme.png"},
		{Case: "relative, no omp", Config: "~/jandedobbeleer.json", Expected: "jandedobbeleer.png"},
		{Case: "relative path", Config: "~/jandedobbeleer.omp.json", Expected: "jandedobbeleer.png"},
		{Case: "invalid config name", Config: "~/jandedobbeleer.omp.foo", Expected: "prompt.png"},
	}

	for _, tc := range cases {
		image := &Renderer{
			Path: tc.Path,
		}

		image.setOutputPath(tc.Config)

		assert.Equal(t, tc.Expected, image.Path, tc.Case)
	}
}

func TestHexToRGB(t *testing.T) {
	cases := []struct {
		expected *RGB
		name     string
		hex      HexColor
		hasError bool
	}{
		{
			name:     "Valid hex with hash",
			hex:      "#FF0000",
			expected: &RGB{255, 0, 0},
			hasError: false,
		},
		{
			name:     "Valid hex without hash",
			hex:      "00FF00",
			expected: &RGB{0, 255, 0},
			hasError: false,
		},
		{
			name:     "Valid hex blue",
			hex:      "#0000FF",
			expected: &RGB{0, 0, 255},
			hasError: false,
		},
		{
			name:     "Invalid hex too short",
			hex:      "#FFF",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Invalid hex too long",
			hex:      "#FFFFFFF",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Invalid hex characters",
			hex:      "#GGGGGG",
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.hex.RGB()

			if tc.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetColorNameFromCode(t *testing.T) {
	cases := []struct {
		expected  string
		colorCode int
	}{
		{"black", 30},
		{"red", 31},
		{"green", 32},
		{"yellow", 33},
		{"blue", 34},
		{"magenta", 35},
		{"cyan", 36},
		{"white", 37},
		{"black", 40}, // background
		{"red", 41},   // background
		{"darkGray", 90},
		{"lightRed", 91},
		{"lightGreen", 92},
		{"lightYellow", 93},
		{"lightBlue", 94},
		{"lightMagenta", 95},
		{"lightCyan", 96},
		{"lightWhite", 97},
		{"", 999}, // invalid code
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			result := colorNameFromCode(tc.colorCode)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSetBase16Color(t *testing.T) {
	cases := []struct {
		colorOverrides     map[string]HexColor
		expectedForeground *RGB
		expectedBackground *RGB
		name               string
		colorCode          string
	}{
		{
			name:               "Red foreground with override",
			colorCode:          "31",
			colorOverrides:     map[string]HexColor{"red": "#FF6B6B", "blue": "#4ECDC4"},
			expectedForeground: &RGB{255, 107, 107},
			expectedBackground: nil,
		},
		{
			name:               "Blue background with override",
			colorCode:          "44",
			colorOverrides:     map[string]HexColor{"red": "#FF6B6B", "blue": "#4ECDC4"},
			expectedForeground: nil,
			expectedBackground: &RGB{78, 205, 196},
		},
		{
			name:               "Green foreground without override",
			colorCode:          "32",
			colorOverrides:     map[string]HexColor{"red": "#FF6B6B", "blue": "#4ECDC4"},
			expectedForeground: &RGB{57, 181, 74},
			expectedBackground: nil,
		},
		{
			name:               "Red foreground without any overrides",
			colorCode:          "31",
			colorOverrides:     nil,
			expectedForeground: &RGB{222, 56, 43},
			expectedBackground: nil,
		},
		{
			name:               "Blue background without any overrides",
			colorCode:          "44",
			colorOverrides:     nil,
			expectedForeground: nil,
			expectedBackground: &RGB{0, 111, 184},
		},
		{
			name:               "Invalid color code",
			colorCode:          "invalid",
			colorOverrides:     nil,
			expectedForeground: &RGB{255, 255, 255},
			expectedBackground: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := &Renderer{
				defaultForegroundColor: &RGB{255, 255, 255},
				Settings: Settings{
					Colors: tc.colorOverrides,
				},
			}

			renderer.setBase16Color(tc.colorCode)

			if tc.expectedForeground != nil {
				assert.Equal(t, tc.expectedForeground, renderer.foregroundColor)
			}

			if tc.expectedBackground != nil {
				assert.Equal(t, tc.expectedBackground, renderer.backgroundColor)
			}
		})
	}
}

func TestProcessAnsiSequence(t *testing.T) {
	cases := []struct {
		expectedForegroundColor *RGB
		expectedBackgroundColor *RGB
		colorOverrides          map[string]HexColor
		name                    string
		ansiString              string
		expectedAnsiString      string
		expectedStyle           string
	}{
		{
			name:               "Regular character",
			ansiString:         "hello",
			expectedAnsiString: "hello",
		},
		{
			name:                    "Inverted color",
			ansiString:              "\x1b[38;2;255;0;0;49m\x1b[7mtest",
			expectedAnsiString:      "test",
			expectedForegroundColor: &RGB{21, 21, 21}, // defaultBackgroundColor
			expectedBackgroundColor: &RGB{255, 0, 0},
		},
		{
			name:                    "Inverted color single",
			ansiString:              "\x1b[31;49m\x1b[7mtest",
			expectedAnsiString:      "test",
			expectedForegroundColor: &RGB{21, 21, 21},  // defaultBackgroundColor
			expectedBackgroundColor: &RGB{222, 56, 43}, // red background (31 + 10 = 41)
		},
		{
			name:                    "Full color",
			ansiString:              "\x1b[48;2;100;200;50m\x1b[38;2;255;0;0mtest",
			expectedAnsiString:      "test",
			expectedBackgroundColor: &RGB{100, 200, 50},
			expectedForegroundColor: &RGB{255, 0, 0},
		},
		{
			name:                    "Foreground color",
			ansiString:              "\x1b[38;2;255;128;0mtest",
			expectedAnsiString:      "test",
			expectedForegroundColor: &RGB{255, 128, 0},
		},
		{
			name:                    "Background color",
			ansiString:              "\x1b[48;2;0;255;128mtest",
			expectedAnsiString:      "test",
			expectedBackgroundColor: &RGB{0, 255, 128},
		},
		{
			name:                    "Reset sequence",
			ansiString:              "\x1b[0mtest",
			expectedAnsiString:      "test",
			expectedForegroundColor: &RGB{255, 255, 255}, // defaultForegroundColor
			expectedBackgroundColor: nil,
		},
		{
			name:                    "Background reset",
			ansiString:              "\x1b[49mtest",
			expectedAnsiString:      "test",
			expectedBackgroundColor: nil,
		},
		{
			name:               "Bold style",
			ansiString:         "\x1b[1mtest",
			expectedAnsiString: "test",
			expectedStyle:      "bold",
		},
		{
			name:               "Italic style",
			ansiString:         "\x1b[3mtest",
			expectedAnsiString: "test",
			expectedStyle:      "italic",
		},
		{
			name:               "Underline style",
			ansiString:         "\x1b[4mtest",
			expectedAnsiString: "test",
			expectedStyle:      "underline",
		},
		{
			name:               "Overline style",
			ansiString:         "\x1b[53mtest",
			expectedAnsiString: "test",
			expectedStyle:      "overline",
		},
		{
			name:               "Bold reset",
			ansiString:         "\x1b[22mtest",
			expectedAnsiString: "test",
			expectedStyle:      "",
		},
		{
			name:               "Italic reset",
			ansiString:         "\x1b[23mtest",
			expectedAnsiString: "test",
			expectedStyle:      "",
		},
		{
			name:               "Underline reset",
			ansiString:         "\x1b[24mtest",
			expectedAnsiString: "test",
			expectedStyle:      "",
		},
		{
			name:               "Overline reset",
			ansiString:         "\x1b[55mtest",
			expectedAnsiString: "test",
			expectedStyle:      "",
		},
		{
			name:               "Strikethrough",
			ansiString:         "\x1b[9mtest",
			expectedAnsiString: "test",
		},
		{
			name:               "Strikethrough reset",
			ansiString:         "\x1b[29mtest",
			expectedAnsiString: "test",
		},
		{
			name:               "Left cursor movement",
			ansiString:         "\x1b[5Dtest",
			expectedAnsiString: "test",
		},
		{
			name:               "Line change",
			ansiString:         "\x1b[2Ftest",
			expectedAnsiString: "test",
		},
		{
			name:               "Console title",
			ansiString:         "\x1b]0;My Title\007test",
			expectedAnsiString: "test",
		},
		{
			name:                    "Base16 red color",
			ansiString:              "\x1b[31mtest",
			expectedAnsiString:      "test",
			expectedForegroundColor: &RGB{222, 56, 43},
		},
		{
			name:                    "Base16 blue background",
			ansiString:              "\x1b[44mtest",
			expectedAnsiString:      "test",
			expectedBackgroundColor: &RGB{0, 111, 184},
		},
		{
			name:                    "Base16 red with override",
			ansiString:              "\x1b[31mtest",
			expectedAnsiString:      "test",
			expectedForegroundColor: &RGB{255, 107, 107},
			colorOverrides:          map[string]HexColor{"red": "#FF6B6B"},
		},
		{
			name:               "Link sequence",
			ansiString:         "\x1b]8;;https://example.com\x1b\\Click here\x1b]8;;\x1b\\test",
			expectedAnsiString: "Click heretest",
		},
		{
			name:               "No matching sequence",
			ansiString:         "plain text",
			expectedAnsiString: "plain text",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			renderer := &Renderer{
				AnsiString: tc.ansiString,
				Settings: Settings{
					Colors: tc.colorOverrides,
				},
			}

			renderer.initDefaults()

			var result bool
			for !result {
				result = renderer.processAnsiSequence()
			}

			assert.Equal(t, tc.expectedAnsiString, renderer.AnsiString)

			if tc.expectedForegroundColor != nil {
				assert.Equal(t, tc.expectedForegroundColor, renderer.foregroundColor)
			}

			if tc.expectedBackgroundColor != nil {
				assert.Equal(t, tc.expectedBackgroundColor, renderer.backgroundColor)
			}

			if tc.expectedStyle != "" {
				assert.Equal(t, tc.expectedStyle, renderer.style)
			}
		})
	}
}
