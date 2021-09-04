package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemory(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		ExpectDisabled bool
		Memory         memory
		Precision      int
	}{
		{Case: "50", ExpectedString: "50", Memory: memory{FreeMemory: 50, TotalMemory: 100}},
		{Case: "50.0", ExpectedString: "50.0", Memory: memory{FreeMemory: 50, TotalMemory: 100}, Precision: 1},
		{Case: "not enabled", ExpectDisabled: true, Memory: memory{FreeMemory: 0, TotalMemory: 0}},
	}

	for _, tc := range cases {
		tc.Memory.env = new(MockedEnvironment)
		tc.Memory.props = &properties{
			values: map[Property]interface{}{
				Precision: tc.Precision,
			},
		}
		if tc.ExpectDisabled {
			assert.Equal(t, false, tc.Memory.enabled(), tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedString, tc.Memory.string(), tc.Case)
		}
	}
}
