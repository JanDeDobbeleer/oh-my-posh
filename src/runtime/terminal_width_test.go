package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReduceTerminalWidth(t *testing.T) {
	cases := []struct {
		Case     string
		Width    int
		Less     int
		Expected int
	}{
		{Case: "no reduction", Width: 120, Less: 0, Expected: 120},
		{Case: "negative ignored", Width: 120, Less: -4, Expected: 120},
		{Case: "reduced", Width: 120, Less: 4, Expected: 116},
		{Case: "leaves one column", Width: 5, Less: 4, Expected: 1},
		{Case: "equal to width ignored", Width: 4, Less: 4, Expected: 4},
		{Case: "larger than width ignored", Width: 4, Less: 10, Expected: 4},
		{Case: "unknown width", Width: 0, Less: 4, Expected: 0},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			assert.Equal(t, tc.Expected, reduceTerminalWidth(tc.Width, tc.Less))
		})
	}
}

func TestTerminalWidthMaxColumnsLess(t *testing.T) {
	term := &Terminal{CmdFlags: &Flags{TerminalWidth: 100, MaxColumnsLess: 4}}

	// repeated reads must not subtract more than once
	for range 3 {
		width, err := term.TerminalWidth()
		assert.NoError(t, err)
		assert.Equal(t, 96, width)
	}
}
