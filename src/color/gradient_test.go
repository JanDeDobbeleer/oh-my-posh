package color

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/lucasb-eyer/go-colorful"
)

func TestIsGradient(t *testing.T) {
	cases := []struct {
		Case     string
		Color    Ansi
		Expected bool
	}{
		{Case: "gradient", Color: "linear-gradient(#FF0000, #0000FF)", Expected: true},
		{Case: "dark-gradient", Color: "dark-gradient(#3465A4)", Expected: true},
		{Case: "light-gradient", Color: "light-gradient(#3465A4)", Expected: true},
		{Case: "hex", Color: "#FF0000", Expected: false},
		{Case: "empty", Color: "", Expected: false},
		{Case: "keyword", Color: Background, Expected: false},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, tc.Color.IsGradient(), tc.Case)
	}
}

func TestGradientStops(t *testing.T) {
	cases := []struct {
		Case     string
		Color    Ansi
		Expected []Ansi
	}{
		{Case: "two stops", Color: "linear-gradient(#FF0000, #0000FF)", Expected: []Ansi{"#FF0000", "#0000FF"}},
		{Case: "three stops", Color: "linear-gradient(#FF0000, #00FF00, #0000FF)", Expected: []Ansi{"#FF0000", "#00FF00", "#0000FF"}},
		{Case: "whitespace variants", Color: "linear-gradient( #FF0000  ,  #0000FF )", Expected: []Ansi{"#FF0000", "#0000FF"}},
		{Case: "palette ref stops", Color: "linear-gradient(p:red, p:blue)", Expected: []Ansi{"p:red", "p:blue"}},
		{Case: "dark-gradient single stop", Color: "dark-gradient(#3465A4)", Expected: []Ansi{"#3465A4"}},
		{Case: "light-gradient single stop", Color: "light-gradient(#3465A4)", Expected: []Ansi{"#3465A4"}},
		{Case: "not a gradient", Color: "#FF0000", Expected: nil},
		{Case: "empty string", Color: "", Expected: nil},
		{Case: "no closing paren", Color: "linear-gradient(#FF0000, #0000FF", Expected: nil},
		{Case: "nested gradient", Color: "linear-gradient(linear-gradient(#FF0000, #00FF00), #0000FF)", Expected: nil},
		{Case: "empty stop between commas", Color: "linear-gradient(#FF0000,,#0000FF)", Expected: nil},
		{Case: "trailing comma", Color: "linear-gradient(#FF0000, #0000FF,)", Expected: nil},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, tc.Color.GradientStops(), tc.Case)
	}
}

func TestGradientFirstAndLast(t *testing.T) {
	cases := []struct {
		Case          string
		Color         Ansi
		ExpectedFirst Ansi
		ExpectedLast  Ansi
	}{
		{Case: "two stops", Color: "linear-gradient(#FF0000, #0000FF)", ExpectedFirst: "#FF0000", ExpectedLast: "#0000FF"},
		{Case: "three stops", Color: "linear-gradient(#FF0000, #00FF00, #0000FF)", ExpectedFirst: "#FF0000", ExpectedLast: "#0000FF"},
		{Case: "not a gradient", Color: "#ABCDEF", ExpectedFirst: "#ABCDEF", ExpectedLast: "#ABCDEF"},
		{
			Case:          "invalid gradient syntax",
			Color:         "linear-gradient(#FF0000",
			ExpectedFirst: "linear-gradient(#FF0000",
			ExpectedLast:  "linear-gradient(#FF0000",
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.ExpectedFirst, tc.Color.GradientFirst(), tc.Case)
		assert.Equal(t, tc.ExpectedLast, tc.Color.GradientLast(), tc.Case)
	}
}

func TestGradientCellsSingleCellReturnsFirstStop(t *testing.T) {
	result := GradientCells("linear-gradient(#FF0000, #0000FF)", 1, &Defaults{}, false, nil, nil)
	assert.Equal(t, []Ansi{"38;2;255;0;0"}, result)
}

func TestGradientCellsInterpolation(t *testing.T) {
	cases := []struct {
		Resolver      String
		Case          string
		Color         Ansi
		ExpectedFirst Ansi
		ExpectedLast  Ansi
		Cells         int
		IsBackground  bool
	}{
		{
			Case:          "endpoints exact, foreground",
			Color:         "linear-gradient(#000000, #FFFFFF)",
			Resolver:      &Defaults{},
			Cells:         5,
			ExpectedFirst: "38;2;0;0;0",
			ExpectedLast:  "38;2;255;255;255",
		},
		{
			Case:          "endpoints exact, background",
			Color:         "linear-gradient(#FF0000, #0000FF)",
			Resolver:      &Defaults{},
			Cells:         3,
			IsBackground:  true,
			ExpectedFirst: "48;2;255;0;0",
			ExpectedLast:  "48;2;0;0;255",
		},
		{
			Case:          "palette ref stops resolve",
			Color:         "linear-gradient(p:red, p:blue)",
			Resolver:      &PaletteColors{ansiColors: &Defaults{}, palette: testPalette},
			Cells:         2,
			ExpectedFirst: "38;2;255;0;0",
			ExpectedLast:  "38;2;0;0;255",
		},
	}

	for _, tc := range cases {
		result := GradientCells(tc.Color, tc.Cells, tc.Resolver, tc.IsBackground, nil, nil)
		assert.Equal(t, tc.Cells, len(result), tc.Case)
		assert.Equal(t, tc.ExpectedFirst, result[0], tc.Case)
		assert.Equal(t, tc.ExpectedLast, result[len(result)-1], tc.Case)
	}
}

// TestGradientCellsMonotonicProgression checks the red channel never regresses across a
// black-to-white ramp, where R, G and B move in lockstep.
func TestGradientCellsMonotonicProgression(t *testing.T) {
	result := GradientCells("linear-gradient(#000000, #FFFFFF)", 5, &Defaults{}, false, nil, nil)
	assert.Len(t, result, 5)

	prev := -1
	for i, cell := range result {
		var r, g, b int
		_, err := fmt.Sscanf(cell.String(), "38;2;%d;%d;%d", &r, &g, &b)
		assert.Nil(t, err, fmt.Sprintf("cell %d: %s", i, cell))

		assert.True(t, r >= prev, fmt.Sprintf("cell %d regressed: %d < %d", i, r, prev))
		prev = r
	}
}

func TestGradientCellsInvalidReturnsNil(t *testing.T) {
	cases := []struct {
		Color Ansi
		Case  string
		Cells int
	}{
		{Case: "one invalid stop of two", Color: "linear-gradient(#FF0000, notacolor)", Cells: 3},
		{Case: "linear-gradient with a single stop needs a real second stop", Color: "linear-gradient(#FF0000)", Cells: 3},
		{Case: "dark-gradient unresolvable stop", Color: "dark-gradient(notacolor)", Cells: 3},
		{Case: "light-gradient unresolvable stop", Color: "light-gradient(notacolor)", Cells: 3},
		{Case: "dark-gradient rejects more than one stop", Color: "dark-gradient(#FF0000, #0000FF)", Cells: 3},
		{Case: "not a gradient", Color: "#FF0000", Cells: 3},
		{Case: "invalid syntax", Color: "linear-gradient(#FF0000, #0000FF", Cells: 3},
		{Case: "zero cells", Color: "linear-gradient(#FF0000, #0000FF)", Cells: 0},
	}

	for _, tc := range cases {
		result := GradientCells(tc.Color, tc.Cells, &Defaults{}, false, nil, nil)
		assert.Nil(t, result, tc.Case)
	}
}

// TestGradientCellsAutoShade verifies dark-gradient(#color)/light-gradient(#color) — a
// single explicit stop — spreads into a two-stop gradient running from the exact
// configured color to a darker/lighter shade of it, instead of collapsing like an
// ordinary invalid (< 2 stop) linear-gradient. The first cell must be the unmodified
// configured color (matching GradientFirst) and the last must match
// GradientLastForCells for the SAME cell count, so separators/caps line up.
func TestGradientCellsAutoShade(t *testing.T) {
	cases := []struct {
		Case string
		Kind string
	}{
		{Case: "dark-gradient darkens", Kind: "dark"},
		{Case: "light-gradient lightens", Kind: "light"},
	}

	for _, tc := range cases {
		gradient := Ansi(tc.Kind + "-gradient(#3465A4)")
		result := GradientCells(gradient, 5, &Defaults{}, false, nil, nil)
		assert.Len(t, result, 5, tc.Case)

		assert.Equal(t, Ansi("38;2;52;101;164"), result[0], tc.Case+": the first cell must be the unmodified configured color")

		expectedLast := GradientCells(Ansi("linear-gradient(#3465A4, "+gradient.GradientLastForCells(5).String()+")"), 5, &Defaults{}, false, nil, nil)
		assert.Equal(t, expectedLast[len(expectedLast)-1], result[len(result)-1], tc.Case+": the last cell must match GradientLastForCells(5), the color separators/caps use")

		assert.NotEqual(t, result[0], result[len(result)-1], tc.Case+": the segment must actually shade, not render solid")
	}
}

// TestGradientCellsAutoShadeScalesWithWidth verifies the total base-to-shade delta
// grows with the segment's cell count instead of staying fixed: a wide segment must
// end further from its base color than a narrow one, so a wide gradient still reads
// as a clear effect instead of fading into an imperceptibly fine ramp - the report
// behind this fix was that a wide segment's gradient looked completely flat.
func TestGradientCellsAutoShadeScalesWithWidth(t *testing.T) {
	narrow := GradientCells("dark-gradient(#179299)", 3, &Defaults{}, true, nil, nil)
	wide := GradientCells("dark-gradient(#179299)", 15, &Defaults{}, true, nil, nil)

	base, err := colorful.Hex("#179299")
	assert.Nil(t, err)

	narrowLast, ok := parseTrueColor(narrow[len(narrow)-1])
	assert.True(t, ok)

	wideLast, ok := parseTrueColor(wide[len(wide)-1])
	assert.True(t, ok)

	assert.True(t, base.DistanceLab(wideLast) > base.DistanceLab(narrowLast), "a 15-cell segment must end further from the base color than a 3-cell one")
}

// TestGradientCellsAutoShadeSingleCell verifies a single-cell segment (too narrow to
// show any blend) renders the configured color unmodified, not a shaded endpoint.
func TestGradientCellsAutoShadeSingleCell(t *testing.T) {
	result := GradientCells("dark-gradient(#3465A4)", 1, &Defaults{}, false, nil, nil)
	assert.Equal(t, []Ansi{"38;2;52;101;164"}, result)
}

// TestGradientLastAutoShade verifies GradientLast (width unknown, the gentlest single-
// step shade) and GradientLastForCells (matching GradientCells for a given cell count)
// darken for dark-gradient and lighten for light-gradient, and fall back to the raw
// stop for one that can't be shaded without a resolver (keyword, palette reference).
func TestGradientLastAutoShade(t *testing.T) {
	unshadeable := "can't be shaded without a resolver, so it passes through unchanged"

	assert.Equal(t, Ansi("#2b5f9d"), Ansi("dark-gradient(#3465A4)").GradientLast(), "dark-gradient's width-unknown shade uses the gentlest (single-step) delta")
	assert.Equal(t, Ansi("#3f6eae"), Ansi("light-gradient(#3465A4)").GradientLast(), "light-gradient's width-unknown shade uses the gentlest (single-step) delta")
	assert.Equal(t, Ansi("#245a98"), Ansi("dark-gradient(#3465A4)").GradientLastForCells(5), "dark-gradient at 5 cells shades further than the width-unknown fallback")
	assert.Equal(t, Ansi("parentBackground"), Ansi("dark-gradient(parentBackground)").GradientLast(), "a keyword "+unshadeable)
	assert.Equal(t, Ansi("p:red"), Ansi("dark-gradient(p:red)").GradientLast(), "a palette reference "+unshadeable)

	// linear-gradient with a single stop is not an auto-shade request; it degrades like
	// any other invalid (< 2 stop) gradient, returning the lone stop unshaded.
	assert.Equal(t, Ansi("#3465A4"), Ansi("linear-gradient(#3465A4)").GradientLast(), "a single-stop linear-gradient is not auto-shaded")
}

// TestWithGradientStops verifies the rebuilt string keeps c's own prefix, so a palette
// reference resolved inside a dark-gradient/light-gradient stays that same kind instead
// of silently becoming a plain linear-gradient.
func TestWithGradientStops(t *testing.T) {
	cases := []struct {
		Case     string
		Color    Ansi
		Expected Ansi
		Stops    []Ansi
	}{
		{Case: "linear-gradient", Color: "linear-gradient(#FF0000, #0000FF)", Stops: []Ansi{"#111111", "#222222"}, Expected: "linear-gradient(#111111, #222222)"},
		{Case: "dark-gradient", Color: "dark-gradient(p:teal)", Stops: []Ansi{"#179299"}, Expected: "dark-gradient(#179299)"},
		{Case: "light-gradient", Color: "light-gradient(p:teal)", Stops: []Ansi{"#179299"}, Expected: "light-gradient(#179299)"},
		{Case: "not a gradient", Color: "#FF0000", Stops: []Ansi{"#111111"}, Expected: "#FF0000"},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, tc.Color.WithGradientStops(tc.Stops), tc.Case)
	}
}

func TestGradientCellsColor256Fallback(t *testing.T) {
	origTrueColor := TrueColor
	t.Cleanup(func() { TrueColor = origTrueColor })
	TrueColor = false

	result := GradientCells("linear-gradient(#AABBCC, #AABBCC)", 3, &Defaults{}, false, nil, nil)

	assert.Len(t, result, 3)
	for i, cell := range result {
		assert.Equal(t, Ansi("38;5;146"), cell, fmt.Sprintf("cell %d", i))
	}
}

// TestGradientCellsKeywordStops verifies keyword stops resolve against the segment
// context before interpolation: parentBackground picks up the parent's color (a parent
// gradient collapses to its last stop), and keywords without a hex resolution
// invalidate the stop.
func TestGradientCellsKeywordStops(t *testing.T) {
	parents := []*Set{{Background: "#112233", Foreground: "#445566"}}
	gradientParents := []*Set{{Background: "linear-gradient(#FF0000, #0000FF)", Foreground: "#445566"}}
	current := &Set{Background: "#778899", Foreground: "#AABBCC"}

	cases := []struct {
		Current  *Set
		Case     string
		Color    Ansi
		Expected Ansi
		Parents  []*Set
	}{
		{
			Case:     "parentBackground stop resolves to parent color",
			Color:    "linear-gradient(parentBackground, #FFFFFF)",
			Parents:  parents,
			Expected: "38;2;17;34;51",
		},
		{
			Case:     "parentBackground stop collapses a parent gradient to its last stop",
			Color:    "linear-gradient(parentBackground, #FFFFFF)",
			Parents:  gradientParents,
			Expected: "38;2;0;0;255",
		},
		{
			Case:     "foreground stop resolves to the current foreground",
			Color:    "linear-gradient(foreground, #000000)",
			Current:  current,
			Expected: "38;2;170;187;204",
		},
	}

	for _, tc := range cases {
		result := GradientCells(tc.Color, 2, &Defaults{}, false, tc.Current, tc.Parents)
		assert.Len(t, result, 2, tc.Case)
		if len(result) == 2 {
			assert.Equal(t, tc.Expected, result[0], tc.Case)
		}
	}

	// a keyword with no context resolves to transparent, not a hex color
	assert.Nil(t, GradientCells("linear-gradient(parentBackground, #FFFFFF)", 2, &Defaults{}, false, nil, nil), "keyword without context invalidates the stop")
	assert.Nil(t, GradientCells("linear-gradient(transparent, #FFFFFF)", 2, &Defaults{}, false, current, parents), "transparent is never a valid stop")
}

// TestGradientCellsAccentStop verifies the accent keyword works as a stop: it resolves
// through ToAnsi to a truecolor payload, which parseTrueColor recovers. An unresolved
// accent (empty Set) invalidates the stop instead of erroring hard.
func TestGradientCellsAccentStop(t *testing.T) {
	resolver := &Defaults{
		accent: &Set{
			Foreground: "38;2;0;120;215",
			Background: "48;2;0;120;215",
		},
	}

	result := GradientCells("linear-gradient(accent, #FFFFFF)", 2, resolver, false, nil, nil)
	assert.Len(t, result, 2)
	if len(result) == 2 {
		assert.Equal(t, Ansi("38;2;0;120;215"), result[0], "first cell must be the accent color")
		assert.Equal(t, Ansi("38;2;255;255;255"), result[1])
	}

	assert.Nil(t, GradientCells("linear-gradient(accent, #FFFFFF)", 2, &Defaults{}, false, nil, nil), "unresolved accent invalidates the stop")
}

func TestParseTrueColor(t *testing.T) {
	cases := []struct {
		Case  string
		Input Ansi
		OK    bool
	}{
		{Case: "foreground payload", Input: "38;2;255;0;128", OK: true},
		{Case: "background payload", Input: "48;2;0;120;215", OK: true},
		{Case: "256-color payload", Input: "38;5;146", OK: false},
		{Case: "hex string", Input: "#FFFFFF", OK: false},
		{Case: "out of range", Input: "38;2;300;0;0", OK: false},
		{Case: "empty", Input: "", OK: false},
	}

	for _, tc := range cases {
		_, ok := parseTrueColor(tc.Input)
		assert.Equal(t, tc.OK, ok, tc.Case)
	}
}
