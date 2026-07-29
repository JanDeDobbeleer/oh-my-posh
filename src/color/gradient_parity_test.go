package color

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GradientCells and GradientCellsRGB resolve stops, auto-shade and interpolate
// the same way, one returning printable escapes and the other raw channels. They
// are deliberately separate so the escape path stays untouched, which means
// nothing but this test stops them drifting apart — and a drift here is silent:
// the prompt keeps rendering correctly while anything drawing from the channels
// interpolates a different ramp.
func TestGradientCellsRGBMatchesGradientCells(t *testing.T) {
	original := TrueColor
	TrueColor = true

	t.Cleanup(func() { TrueColor = original })

	cases := []struct {
		gradient Ansi
		cells    int
	}{
		{"linear-gradient(#FF0000, #0000FF)", 8},
		{"linear-gradient(#FF0000, #00FF00, #0000FF)", 12},
		{"linear-gradient(red, blue)", 5},
		{"linear-gradient(#FFFFFF, #000000)", 1},
		{"linear-gradient(#123456, #654321)", 32},
	}

	resolver := &Defaults{}

	for _, tc := range cases {
		for _, isBackground := range []bool{true, false} {
			name := fmt.Sprintf("%s/%d/bg=%v", tc.gradient, tc.cells, isBackground)

			t.Run(name, func(t *testing.T) {
				ansiCells := GradientCells(tc.gradient, tc.cells, resolver, isBackground, nil, nil)
				rgbCells := GradientCellsRGB(tc.gradient, tc.cells, resolver, nil, nil)

				require.Len(t, rgbCells, len(ansiCells), "the two must produce the same number of cells")

				for i := range ansiCells {
					channel := "38"
					if isBackground {
						channel = "48"
					}

					expected := Ansi(fmt.Sprintf("%s;2;%d;%d;%d", channel, rgbCells[i].R, rgbCells[i].G, rgbCells[i].B))

					assert.Equal(t, expected, ansiCells[i], "cell %d disagrees between the escape and channel paths", i)
				}
			})
		}
	}
}
