package battery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBatteryOutput(t *testing.T) {
	cases := []struct {
		Case               string
		Output             string
		ExpectedState      State
		ExpectedPercentage int
		ExpectError        bool
	}{
		{
			Case:               "charging",
			Output:             "99%; charging;",
			ExpectedState:      Charging,
			ExpectedPercentage: 99,
		},
		{
			Case:               "charging 1%",
			Output:             "1%; charging;",
			ExpectedState:      Charging,
			ExpectedPercentage: 1,
		},
		{
			Case:               "not charging 80%",
			Output:             "81%; AC attached;",
			ExpectedState:      NotCharging,
			ExpectedPercentage: 81,
		},
		{
			Case:               "charged",
			Output:             "100%; charged;",
			ExpectedState:      Full,
			ExpectedPercentage: 100,
		},
		{
			Case:               "discharging, but not",
			Output:             "100%; discharging;",
			ExpectedState:      Full,
			ExpectedPercentage: 100,
		},
	}
	for _, tc := range cases {
		info, err := parseBatteryOutput(tc.Output)
		if tc.ExpectError {
			assert.Error(t, err, tc.Case)
			return
		}
		assert.Equal(t, tc.ExpectedState, info.State, tc.Case)
		assert.Equal(t, tc.ExpectedPercentage, info.Percentage, tc.Case)
	}
}
