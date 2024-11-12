package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
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
	env.On("Shell").Return("foo")

	Cache = new(cache.Template)
	Init(env, nil)

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
