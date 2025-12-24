package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercentageGauge(t *testing.T) {
	cases := []struct {
		Case          string
		ExpectedGauge string
		Percent       Percentage
	}{
		{
			Case:          "0 percent used (100% remaining)",
			Percent:       Percentage(0),
			ExpectedGauge: "▰▰▰▰▰",
		},
		{
			Case:          "20 percent used (80% remaining - 4 blocks)",
			Percent:       Percentage(20),
			ExpectedGauge: "▰▰▰▰▱",
		},
		{
			Case:          "40 percent used (60% remaining - 3 blocks)",
			Percent:       Percentage(40),
			ExpectedGauge: "▰▰▰▱▱",
		},
		{
			Case:          "60 percent used (40% remaining - 2 blocks)",
			Percent:       Percentage(60),
			ExpectedGauge: "▰▰▱▱▱",
		},
		{
			Case:          "80 percent used (20% remaining - 1 block)",
			Percent:       Percentage(80),
			ExpectedGauge: "▰▱▱▱▱",
		},
		{
			Case:          "100 percent used (0% remaining - 0 blocks)",
			Percent:       Percentage(100),
			ExpectedGauge: "▱▱▱▱▱",
		},
		{
			Case:          "50 percent used (50% remaining - 2.5 rounds to 2 blocks)",
			Percent:       Percentage(50),
			ExpectedGauge: "▰▰▱▱▱",
		},
		{
			Case:          "Negative percent clamps to 0",
			Percent:       Percentage(-10),
			ExpectedGauge: "▰▰▰▰▰",
		},
		{
			Case:          "Over 100 percent clamps to 100",
			Percent:       Percentage(120),
			ExpectedGauge: "▱▱▱▱▱",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			result := tc.Percent.Gauge()
			assert.Equal(t, tc.ExpectedGauge, result, tc.Case)
		})
	}
}

func TestPercentageString(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Percent  Percentage
	}{
		{
			Case:     "Zero percent",
			Percent:  Percentage(0),
			Expected: "0",
		},
		{
			Case:     "50 percent",
			Percent:  Percentage(50),
			Expected: "50",
		},
		{
			Case:     "100 percent",
			Percent:  Percentage(100),
			Expected: "100",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			result := tc.Percent.String()
			assert.Equal(t, tc.Expected, result, tc.Case)
		})
	}
}
