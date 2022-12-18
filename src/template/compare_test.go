package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGt(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		E1       interface{}
		E2       interface{}
	}{
		{Case: "Float vs int", Expected: false, E1: float64(3), E2: 4},
		{Case: "Int vs float", Expected: false, E1: 3, E2: float64(4)},
		{Case: "Int vs Int", Expected: false, E1: 3, E2: 4},
		{Case: "Float vs Float", Expected: false, E1: float64(3), E2: float64(4)},
		{Case: "Float vs String", Expected: true, E1: float64(3), E2: "test"},
		{Case: "Int vs String", Expected: true, E1: 3, E2: "test"},
		{Case: "String vs String", Expected: false, E1: "test", E2: "test"},
	}

	for _, tc := range cases {
		got := gt(tc.E1, tc.E2)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestLt(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		E1       interface{}
		E2       interface{}
	}{
		{Case: "Float vs int", Expected: true, E1: float64(3), E2: 4},
		{Case: "Int vs float", Expected: true, E1: 3, E2: float64(4)},
		{Case: "Int vs Int", Expected: true, E1: 3, E2: 4},
		{Case: "Float vs Float", Expected: true, E1: float64(3), E2: float64(4)},
		{Case: "Float vs String", Expected: false, E1: float64(3), E2: "test"},
		{Case: "String vs String", Expected: true, E1: "test", E2: "test"},
	}

	for _, tc := range cases {
		got := lt(tc.E1, tc.E2)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
