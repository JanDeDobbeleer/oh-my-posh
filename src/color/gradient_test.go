package color

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

func TestIsGradient(t *testing.T) {
	cases := []struct {
		Case     string
		Color    Ansi
		Expected bool
	}{
		{Case: "gradient", Color: "linear-gradient(#FF0000, #0000FF)", Expected: true},
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
		{Case: "single stop", Color: "linear-gradient(#FF0000)", Cells: 3},
		{Case: "not a gradient", Color: "#FF0000", Cells: 3},
		{Case: "invalid syntax", Color: "linear-gradient(#FF0000, #0000FF", Cells: 3},
		{Case: "zero cells", Color: "linear-gradient(#FF0000, #0000FF)", Cells: 0},
	}

	for _, tc := range cases {
		result := GradientCells(tc.Color, tc.Cells, &Defaults{}, false, nil, nil)
		assert.Nil(t, result, tc.Case)
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
