package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestTrunc(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{Case: "5 length integer", Expected: "Hello", Template: `{{ trunc 5 "Hello World" }}`},
		{Case: "5 length stringteger", Expected: "Hello", Template: `{{ trunc "5" "Hello World" }}`},
		{Case: "5 length float", Expected: "Hello", Template: `{{ trunc 5.0 "Hello World" }}`},
		{Case: "invalid", ShouldError: true, Template: `{{ trunc "foo" "Hello World" }}`},
		{Case: "smaller than length", Expected: "Hello World", Template: `{{ trunc 20 "Hello World" }}`},
		{Case: "negative", Expected: "ld", Template: `{{ trunc -2 "Hello World" }}`},
	}

	env := &mock.Environment{}
	env.On("TemplateCache").Return(&cache.Template{
		Env: make(map[string]string),
	})
	env.On("Error", testify_.Anything)
	env.On("Debug", testify_.Anything)
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
	env.On("Flags").Return(&runtime.Flags{})

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
