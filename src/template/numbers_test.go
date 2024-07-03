package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
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

	env := &mock.Environment{}
	env.On("TemplateCache").Return(&cache.Template{
		Env: make(map[string]string),
	})
	env.On("Error", testify_.Anything)
	env.On("Debug", testify_.Anything)
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
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
