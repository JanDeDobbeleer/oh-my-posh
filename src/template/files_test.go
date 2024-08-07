package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestGlob(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{Case: "valid glob", Expected: "OK", Template: `{{ if glob "*.go" }}OK{{ else }}NOK{{ end }}`},
		{Case: "invalid glob", Expected: "NOK", Template: `{{ if glob "package.json" }}OK{{ else }}NOK{{ end }}`},
		{Case: "multiple glob", Expected: "NOK", Template: `{{ if or (glob "package.json") (glob "node_modules") }}OK{{ else }}NOK{{ end }}`},
	}

	env := &mock.Environment{}
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
	env.On("TemplateCache").Return(&cache.Template{})
	env.On("Shell").Return("foo")

	Init(env)

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
