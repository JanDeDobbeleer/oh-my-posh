package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/text"
	"github.com/stretchr/testify/assert"
)

func TestGt(t *testing.T) {
	cases := []struct {
		E1       any
		E2       any
		Case     string
		Expected bool
	}{
		{Case: "Float vs int", Expected: false, E1: float64(3), E2: 4},
		{Case: "Int vs float", Expected: false, E1: 3, E2: float64(4)},
		{Case: "Int vs Int", Expected: false, E1: 3, E2: 4},
		{Case: "Int64 vs Int", Expected: false, E1: int64(3), E2: 4},
		{Case: "Float vs Float", Expected: false, E1: float64(3), E2: float64(4)},
		{Case: "Float vs String", Expected: true, E1: float64(3), E2: "test"},
		{Case: "Int vs String", Expected: true, E1: 3, E2: "test"},
		{Case: "String vs String", Expected: false, E1: "test", E2: "test"},
		// Named int types (e.g. text.Percentage)
		{Case: "Percentage(60) gt 50 = true", Expected: true, E1: text.Percentage(60), E2: 50},
		{Case: "Percentage(40) gt 50 = false", Expected: false, E1: text.Percentage(40), E2: 50},
		{Case: "Percentage(50) gt 50 = false", Expected: false, E1: text.Percentage(50), E2: 50},
		{Case: "Int gt Percentage = true", Expected: true, E1: 60, E2: text.Percentage(50)},
		{Case: "Int gt Percentage = false", Expected: false, E1: 40, E2: text.Percentage(50)},
	}

	for _, tc := range cases {
		got := gt(tc.E1, tc.E2)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestLt(t *testing.T) {
	cases := []struct {
		E1       any
		E2       any
		Case     string
		Expected bool
	}{
		{Case: "Float vs int", Expected: true, E1: float64(3), E2: 4},
		{Case: "Int vs float", Expected: true, E1: 3, E2: float64(4)},
		{Case: "Int vs Int", Expected: true, E1: 3, E2: 4},
		{Case: "Float vs Float", Expected: true, E1: float64(3), E2: float64(4)},
		{Case: "Float vs String", Expected: false, E1: float64(3), E2: "test"},
		{Case: "String vs String", Expected: false, E1: "test", E2: "test"},
		// Named int types (e.g. text.Percentage)
		{Case: "Percentage(40) lt 50 = true", Expected: true, E1: text.Percentage(40), E2: 50},
		{Case: "Percentage(60) lt 50 = false", Expected: false, E1: text.Percentage(60), E2: 50},
		{Case: "Int lt Percentage = true", Expected: true, E1: 40, E2: text.Percentage(50)},
		{Case: "Int lt Percentage = false", Expected: false, E1: 60, E2: text.Percentage(50)},
	}

	for _, tc := range cases {
		got := lt(tc.E1, tc.E2)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
