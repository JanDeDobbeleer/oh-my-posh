package color

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/shell"
	"oh-my-posh/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeasureText(t *testing.T) {
	cases := []struct {
		Case     string
		Template string
		Expected int
	}{
		{
			Case:     "Simple text",
			Template: "src",
			Expected: 3,
		},
		{
			Case:     "Hyperlink",
			Template: `{{ url "link" "https://ohmyposh.dev" }}`,
			Expected: 4,
		},
	}
	env := new(mock.MockedEnvironment)
	env.On("TemplateCache").Return(&environment.TemplateCache{
		Env: make(map[string]string),
	})
	shells := []string{shell.BASH, shell.ZSH, shell.PLAIN}
	for _, shell := range shells {
		for _, tc := range cases {
			ansi := &Ansi{}
			ansi.Init(shell)
			tmpl := &template.Text{
				Template: tc.Template,
				Env:      env,
			}
			text, _ := tmpl.Render()
			text = ansi.GenerateHyperlink(text)
			got := ansi.MeasureText(text)
			assert.Equal(t, tc.Expected, got, fmt.Sprintf("%s: %s", shell, tc.Case))
		}
	}
}
