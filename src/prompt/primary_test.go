package prompt

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
)

func TestCursorStyleTemplate(t *testing.T) {
	cases := []struct {
		Case       string
		Config     string
		POSHVIMode string
		Expected   string
	}{
		{Case: "No config"},
		{Case: "Static style", Config: terminal.SteadyBar, Expected: "\x1b[6 q"},
		{
			Case:       "Vi mode aware - normal",
			Config:     `{{ if eq .Env.POSH_VI_MODE "vicmd" }}steady_block{{ else }}steady_bar{{ end }}`,
			POSHVIMode: "vicmd",
			Expected:   "\x1b[2 q",
		},
		{
			Case:       "Vi mode aware - insert",
			Config:     `{{ if eq .Env.POSH_VI_MODE "vicmd" }}steady_block{{ else }}steady_bar{{ end }}`,
			POSHVIMode: "viins",
			Expected:   "\x1b[6 q",
		},
		{
			Case:     "Vi mode aware - unset env",
			Config:   `{{ if eq .Env.POSH_VI_MODE "vicmd" }}steady_block{{ else }}steady_bar{{ end }}`,
			Expected: "\x1b[6 q",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "POSH_VI_MODE").Return(tc.POSHVIMode)
		env.On("Shell").Return(shell.GENERIC)

		template.Cache = &cache.Template{
			Shell:    shell.GENERIC,
			Segments: maps.NewConcurrent[any](),
		}
		template.Init(env, nil, nil)

		terminal.Init(shell.GENERIC)

		engine := &Engine{
			Env: env,
			Config: &config.Config{
				CursorStyle: tc.Config,
			},
		}

		engine.writePrimaryPromptInternal(false, false)
		got := engine.string()

		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
