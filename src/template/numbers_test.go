package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHResult(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{Case: "Windows exit code", Expected: "0x8A150014", Template: `{{ hresult -1978335212 }}`},
		{Case: "Not a number", Template: `{{ hresult "no number" }}`, ShouldError: true},
	}

	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  nil,
		}

		text, err := tmpl.Render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}

		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
