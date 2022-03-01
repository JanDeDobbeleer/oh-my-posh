package template

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"testing"

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
		text, err := tmpl.Render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
