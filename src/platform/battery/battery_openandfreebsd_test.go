//go:build openbsd || freebsd

package battery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBatteryOutput(t *testing.T) {
	cases := []struct {
		Case               string
		PercentOutput      string
		StatusOutput       string
		ExpectedState      State
		ExpectedPercentage int
		ExpectError        bool
	}{
		{
			Case:               "charging",
			PercentOutput:      "99",
			StatusOutput:       "3",
			ExpectedState:      Charging,
			ExpectedPercentage: 99,
		},
		{
			Case:               "charging 1%",
			PercentOutput:      "1",
			StatusOutput:       "3",
			ExpectedState:      Charging,
			ExpectedPercentage: 1,
		},
		{
			Case:               "removed",
			PercentOutput:      "0",
			StatusOutput:       "4",
			ExpectedState:      Unknown,
			ExpectedPercentage: 0,
		},
		{
			Case:               "charged",
			PercentOutput:      "100",
			StatusOutput:       "0",
			ExpectedState:      Full,
			ExpectedPercentage: 100,
		},
		{
			Case:               "discharging",
			PercentOutput:      "25",
			StatusOutput:       "1",
			ExpectedState:      Discharging,
			ExpectedPercentage: 25,
		},
	}
	for _, tc := range cases {
		info, err := parseBatteryOutput(tc.PercentOutput, tc.StatusOutput)
		if tc.ExpectError {
			assert.Error(t, err, tc.Case)
			return
		}
		assert.Equal(t, tc.ExpectedState, info.State, tc.Case)
		assert.Equal(t, tc.ExpectedPercentage, info.Percentage, tc.Case)
	}
}
