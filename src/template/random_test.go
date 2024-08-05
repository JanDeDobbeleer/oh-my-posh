package template

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	cases := []struct {
		Input       interface{}
		Case        string
		ShouldError bool
	}{
		{
			Case:  "valid slice",
			Input: []int{1, 2, 3, 4, 5},
		},
		{
			Case:  "valid array",
			Input: [5]int{1, 2, 3, 4, 5},
		},
		{
			Case:        "empty slice",
			Input:       []int{},
			ShouldError: true,
		},
		{
			Case:        "not a slice or array",
			Input:       "not a slice",
			ShouldError: true,
		},
		{
			Case:  "valid string slice",
			Input: []string{"a", "b", "c"},
		},
		{
			Case:  "valid float slice",
			Input: []float64{1.1, 2.2, 3.3},
		},
		{
			Case: "interface with multiple types",
			Input: []interface{}{
				"a",
				1,
				true,
			},
		},
		{
			Case:  "valid struct slice",
			Input: []struct{ Name string }{{Name: "Alice"}, {Name: "Bob"}},
		},
	}

	for _, tc := range cases {
		result, err := random(tc.Input)
		if tc.ShouldError {
			assert.Error(t, err, tc.Case)
		} else {
			assert.NoError(t, err, tc.Case)
			assert.Contains(t, fmt.Sprintf("%v", tc.Input), result, tc.Case)
		}
	}
}
