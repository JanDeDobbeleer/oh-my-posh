package cli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSVGOutputPath pins how an export path is derived when --output is
// absent: the config's own basename, with the `.omp` marker and the config
// extension both stripped.
//
// The "keeps a name whose letters overlap .omp" cases are the regression
// this table exists for. The retired PNG renderer derived its filename with
// strings.TrimRight(name, ".omp"), which takes a cutset rather than a
// suffix — every trailing '.', 'o', 'm' and 'p' came off, so demo.json
// exported as de.svg and promo.json as pr.svg. The bug survived the port to
// this package.
func TestSVGOutputPath(t *testing.T) {
	cases := []struct {
		Case       string
		ConfigPath string
		Output     string
		Expected   string
	}{
		{Case: "omp marker", ConfigPath: "jandedobbeleer.omp.json", Expected: "jandedobbeleer.svg"},
		{Case: "omp marker, yaml", ConfigPath: "glowsticks.omp.yaml", Expected: "glowsticks.svg"},
		{Case: "no omp marker", ConfigPath: "theme.json", Expected: "theme.svg"},
		{Case: "name ending in o", ConfigPath: "demo.json", Expected: "demo.svg"},
		{Case: "name ending in mo", ConfigPath: "promo.json", Expected: "promo.svg"},
		{Case: "name ending in p", ConfigPath: "warp.json", Expected: "warp.svg"},
		{Case: "name of only cutset runes", ConfigPath: "pomp.json", Expected: "pomp.svg"},
		{Case: "dotfile", ConfigPath: ".hidden.omp.json", Expected: "hidden.svg"},
		{Case: "unrecognized name falls back", ConfigPath: "notaconfig", Expected: "prompt.svg"},
		{Case: "output wins, extension swapped", ConfigPath: "theme.omp.json", Output: "mytheme.png", Expected: "mytheme.svg"},
		{Case: "output already svg", ConfigPath: "theme.omp.json", Output: "mytheme.svg", Expected: "mytheme.svg"},
	}

	// Compared by base name: an explicit --output is resolved to an absolute
	// path by cleanOutputPath, and what matters here is the name and
	// extension it lands on, not the working directory the test ran from.
	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			assert.Equal(t, tc.Expected, filepath.Base(svgOutputPath(tc.ConfigPath, tc.Output)))
		})
	}
}
