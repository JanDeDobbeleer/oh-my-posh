package image

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
)

// loadTestFont returns a font face backed by the Go Mono TTF bundled with
// golang.org/x/image: monospaced like the real Hack Nerd Font, so a column
// corresponds to one consistent advance, without needing network access.
func loadTestFont(t *testing.T) font.Face {
	t.Helper()

	fnt, err := opentype.Parse(gomono.TTF)
	if err != nil {
		t.Fatalf("failed to parse test font: %v", err)
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{Size: 24, DPI: 144})
	if err != nil {
		t.Fatalf("failed to create test font face: %v", err)
	}

	return face
}

// newLayoutTestRenderer builds a Renderer with the given content and column
// count, wired up with a real (test) font face and the ANSI regex map, but
// without the file/watermark side effects of Init/cleanContent, so layout
// tests can assert on exactly the content they provide.
func newLayoutTestRenderer(t *testing.T, columns int, content string) *Renderer {
	t.Helper()

	face := loadTestFont(t)

	r := &Renderer{
		Settings:   Settings{Columns: columns},
		AnsiString: content,
	}
	r.regular, r.bold, r.italic = face, face, face
	r.initDefaults()

	return r
}

func testSpaceAdvance(t *testing.T) float64 {
	t.Helper()

	drawer := &font.Drawer{Face: loadTestFont(t)}
	return float64(drawer.MeasureString(" ") >> 6)
}

func glyphString(glyphs []glyphPlan) string {
	runes := make([]rune, len(glyphs))
	for i, g := range glyphs {
		runes[i] = g.r
	}
	return string(runes)
}

func TestMeasureContentIsAlwaysFixedWidth(t *testing.T) {
	spaceAdvance := testSpaceAdvance(t)

	cases := []struct {
		name    string
		content string
	}{
		{name: "short content", content: "hi"},
		{name: "empty content", content: ""},
		{name: "long content that would have widened the old canvas", content: strings.Repeat("x", 300)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newLayoutTestRenderer(t, 40, tc.content)

			width, _ := r.measureContent()

			assert.Equal(t, float64(40)*spaceAdvance, width, "canvas width must always be exactly Columns x spaceAdvance")
		})
	}
}

func TestLongLineWithoutPaddingWrapsAndAddsARow(t *testing.T) {
	r := newLayoutTestRenderer(t, 20, strings.Repeat("x", 60))

	plan := r.buildLayout()

	assert.Greater(t, len(plan.rows), 1, "a line with no alignment padding that exceeds the column count must wrap onto extra rows")

	maxWidth := float64(20) * plan.spaceAdvance

	var rebuilt strings.Builder
	for _, row := range plan.rows {
		assert.LessOrEqual(t, row.width(), maxWidth+0.001, "no wrapped row may exceed the fixed canvas width")
		rebuilt.WriteString(glyphString(row.glyphs))
	}

	assert.Equal(t, strings.Repeat("x", 60), rebuilt.String(), "wrapping must not drop or duplicate runes")
}

func TestAlignedLineWithPaddingRunCollapsesInsteadOfWrapping(t *testing.T) {
	left := "left"
	right := "right"
	content := left + strings.Repeat(" ", 20) + right

	r := newLayoutTestRenderer(t, 20, content)

	plan := r.buildLayout()

	assert.Len(t, plan.rows, 1, "an engine-aligned line with a long padding run must collapse, not wrap onto a new row")

	maxWidth := float64(20) * plan.spaceAdvance
	row := plan.rows[0]

	assert.LessOrEqual(t, row.width(), maxWidth, "the collapsed row must fit within the fixed canvas width")
	assert.Greater(t, row.width(), maxWidth-plan.spaceAdvance, "the collapsed row should end close to the right edge, not far short of it")

	result := glyphString(row.glyphs)
	assert.True(t, strings.HasPrefix(result, left), "left-aligned content must be preserved")
	assert.True(t, strings.HasSuffix(result, right), "the right-aligned segment must stay flush against the right edge")
}

func TestDoubleWidthGlyphNeverSplitsAcrossWrapBoundary(t *testing.T) {
	icon := '' // Font Awesome range, double width (see doubleWidthRunes)
	content := strings.Repeat("a", 15) + string(icon) + strings.Repeat("b", 15)

	r := newLayoutTestRenderer(t, 10, content)

	plan := r.buildLayout()

	assert.Greater(t, len(plan.rows), 1, "content this long at 10 columns must wrap")

	maxWidth := float64(10) * plan.spaceAdvance

	var rebuilt strings.Builder
	iconCount := 0

	for _, row := range plan.rows {
		assert.LessOrEqual(t, row.width(), maxWidth+0.001, "no wrapped row may exceed the fixed canvas width")

		for _, g := range row.glyphs {
			if g.r == icon {
				iconCount++
			}
		}

		rebuilt.WriteString(glyphString(row.glyphs))
	}

	assert.Equal(t, content, rebuilt.String(), "wrapping must not drop or duplicate runes")
	assert.Equal(t, 1, iconCount, "the double-width glyph must appear exactly once, whole, never split across rows")
}

func TestResolvedColumnsDefaultsTo120WhenUnset(t *testing.T) {
	cases := []struct {
		name     string
		columns  int
		expected int
	}{
		{name: "unset falls back to default", columns: 0, expected: defaultColumns},
		{name: "negative falls back to default", columns: -1, expected: defaultColumns},
		{name: "explicit value is respected", columns: 90, expected: 90},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &Renderer{Settings: Settings{Columns: tc.columns}}
			assert.Equal(t, tc.expected, r.resolvedColumns())
		})
	}
}

func TestSettingsWithoutColumnsStillLoadsWithDefault120(t *testing.T) {
	tempFile := createTempFile(t, `{"author": "no columns specified"}`)
	defer os.Remove(tempFile)

	settings, err := LoadSettings(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, 0, settings.Columns, "Columns is the JSON zero value when absent from an older settings file")

	r := &Renderer{Settings: *settings}
	assert.Equal(t, defaultColumns, r.resolvedColumns(), "a settings file predating Columns must still render at the default width")
}

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
