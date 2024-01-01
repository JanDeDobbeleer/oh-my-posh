package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

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
		{Case: "valid url", Expected: "<LINK>https://ohmyposh.dev<TEXT>link</TEXT></LINK>", Template: `{{ url "link" "https://ohmyposh.dev" }}`},
		{Case: "invalid url", Expected: "", Template: `{{ url "link" "Foo" }}`, ShouldError: true},
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

func TestPath(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
	}{
		{Case: "valid path", Expected: "<LINK>file:/test/test<TEXT>link</TEXT></LINK>", Template: `{{ path "link" "/test/test" }}`},
	}

	env := &mock.MockedEnvironment{}
	env.On("DebugF", mock2.Anything, mock2.Anything).Return(nil)
	env.On("TemplateCache").Return(&platform.TemplateCache{
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
