package template

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestUrl(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{Case: "valid url", Expected: "[link](https://ohmyposh.dev)", Template: `{{ url "link" "https://ohmyposh.dev" }}`},
		{Case: "invalid url", Expected: "", Template: `{{ url "link" "Foo" }}`, ShouldError: true},
	}

	env := &mock.MockedEnvironment{}
	env.On("TemplateCache").Return(&environment.TemplateCache{
		Env: make(map[string]string),
	})
	env.On("Log", mock2.Anything, mock2.Anything, mock2.Anything)
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

func TestPath(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
	}{
		{Case: "valid path", Expected: "[link](file:/test/test)", Template: `{{ path "link" "/test/test" }}`},
	}

	env := &mock.MockedEnvironment{}
	env.On("TemplateCache").Return(&environment.TemplateCache{
		Env: make(map[string]string),
	})
	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  nil,
			Env:      env,
		}
		text, _ := tmpl.Render()
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
