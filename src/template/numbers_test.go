package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
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

	env := &mock.MockedEnvironment{}
	env.On("TemplateCache").Return(&platform.TemplateCache{
		Env: make(map[string]string),
	})
	env.On("Error", mock2.Anything)
	env.On("Debug", mock2.Anything)
	env.On("DebugF", mock2.Anything, mock2.Anything).Return(nil)
	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  nil,
			Env:      env,
		}
		text, err := tmpl.Render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
