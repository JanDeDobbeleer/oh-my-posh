package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemory(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Memory         memory
		Precision      int
	}{
		{Case: "50", ExpectedString: "50", Memory: memory{FreeMemory: 50, TotalMemory: 100}},
		{Case: "50.0", ExpectedString: "50.0", Memory: memory{FreeMemory: 50, TotalMemory: 100}, Precision: 1},
	}

	for _, tc := range cases {
		tc.Memory.env = new(MockedEnvironment)
		tc.Memory.props = &properties{
			values: map[Property]interface{}{
				Precision: tc.Precision,
			},
		}
		assert.Equal(t, tc.ExpectedString, tc.Memory.string(), tc.Case)
	}
}
