package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminalWidthMaxColumnsLess(t *testing.T) {
	cases := []struct {
		Case     string
		Env      string
		Width    int
		Less     int
		Expected int
	}{
		{Case: "no reduction", Width: 120, Expected: 120},
		{Case: "negative ignored", Width: 120, Less: -4, Expected: 120},
		{Case: "reduced", Width: 120, Less: 4, Expected: 116},
		{Case: "leaves one column", Width: 5, Less: 4, Expected: 1},
		{Case: "equal to width ignored", Width: 4, Less: 4, Expected: 4},
		{Case: "larger than width ignored", Width: 4, Less: 10, Expected: 4},
		{Case: "uses the variable", Width: 100, Env: "4", Expected: 96},
		{Case: "flag wins over the variable", Width: 100, Env: "4", Less: 10, Expected: 90},
		{Case: "ignores an invalid variable", Width: 100, Env: "wide", Expected: 100},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			t.Setenv(maxColumnsLessEnv, tc.Env)

			term := &Terminal{}
			term.Init(&Flags{TerminalWidth: tc.Width, MaxColumnsLess: tc.Less})

			// repeated reads must not subtract more than once
			for range 3 {
				width, err := term.TerminalWidth()

				assert.NoError(t, err)
				assert.Equal(t, tc.Expected, width)
			}
		})
	}
}
