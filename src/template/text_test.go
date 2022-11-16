package template

import (
	"oh-my-posh/mock"
	"oh-my-posh/platform"
	"testing"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestRenderTemplate(t *testing.T) {
	type Me struct {
		Name string
	}
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
		Context     interface{}
	}{
		{
			Case:     "color override with dots",
			Expected: "😺💬<#FF8000> Meow! What should I do next? ...</>",
			Template: "😺💬<#FF8000> Meow! What should I do next? ...</>",
		},
		{
			Case:     "tillig's regex",
			Expected: " ⎈ hello :: world ",
			Template: " ⎈ {{ replaceP \"([a-f0-9]{2})[a-f0-9]{6}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{10}([a-f0-9]{2})\" .Context \"$1..$2\" }}{{ if .Namespace }} :: {{ .Namespace }}{{ end }} ", //nolint:lll
			Context: struct {
				Context   string
				Namespace string
			}{
				Context:   "hello",
				Namespace: "world",
			},
		},
		{
			Case:     "Env like property name",
			Expected: "hello world",
			Template: "{{.EnvLike}} {{.Text2}}",
			Context: struct {
				EnvLike string
				Text2   string
			}{
				EnvLike: "hello",
				Text2:   "world",
			},
		},
		{
			Case:     "single property with a dot literal",
			Expected: "hello world",
			Template: "{{ if eq .Text \".Net\" }}hello world{{ end }}",
			Context:  struct{ Text string }{Text: ".Net"},
		},
		{
			Case:     "single property",
			Expected: "hello world",
			Template: "{{.Text}} world",
			Context:  struct{ Text string }{Text: "hello"},
		},
		{
			Case:     "duplicate property",
			Expected: "hello jan posh",
			Template: "hello {{ .Me.Name }} {{ .Name }}",
			Context: struct {
				Name string
				Me   Me
			}{
				Name: "posh",
				Me: Me{
					Name: "jan",
				},
			},
		},
		{
			Case:        "invalid property",
			ShouldError: true,
			Template:    "{{.Durp}} world",
			Context:     struct{ Text string }{Text: "hello"},
		},
		{
			Case:        "invalid template",
			ShouldError: true,
			Template:    "{{ if .Text }} world",
			Context:     struct{ Text string }{Text: "hello"},
		},
		{
			Case:     "if statement true",
			Expected: "hello world",
			Template: "{{ if .Text }}{{.Text}} world{{end}}",
			Context:  struct{ Text string }{Text: "hello"},
		},
		{
			Case:     "if statement false",
			Expected: "world",
			Template: "{{ if .Text }}{{.Text}} {{end}}world",
			Context:  struct{ Text string }{Text: ""},
		},
		{
			Case:     "if statement true with 2 properties",
			Expected: "hello world",
			Template: "{{.Text}}{{ if .Text2 }} {{.Text2}}{{end}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text:  "hello",
				Text2: "world",
			},
		},
		{
			Case:     "if statement false with 2 properties",
			Expected: "hello",
			Template: "{{.Text}}{{ if .Text2 }} {{.Text2}}{{end}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text: "hello",
			},
		},
		{
			Case:     "double property template",
			Expected: "hello world",
			Template: "{{.Text}} {{.Text2}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text:  "hello",
				Text2: "world",
			},
		},
		{
			Case:     "sprig - contains",
			Expected: "hello world",
			Template: "{{ if contains \"hell\" .Text }}{{.Text}} {{end}}{{.Text2}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text:  "hello",
				Text2: "world",
			},
		},
	}

	env := &mock.MockedEnvironment{}
	env.On("TemplateCache").Return(&platform.TemplateCache{
		Env: make(map[string]string),
	})
	env.On("Log", mock2.Anything, mock2.Anything, mock2.Anything)
	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  tc.Context,
			Env:      env,
		}
		text, err := tmpl.Render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}

func TestRenderTemplateEnvVar(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
		Env         map[string]string
		Context     interface{}
	}{
		{
			Case:        "nil struct with env var",
			ShouldError: true,
			Template:    "{{.Env.HELLO }} world{{ .Text}}",
			Context:     nil,
			Env:         map[string]string{"HELLO": "hello"},
		},
		{
			Case:     "map with env var",
			Expected: "hello world",
			Template: "{{.Env.HELLO}} {{.World}}",
			Context:  map[string]interface{}{"World": "world"},
			Env:      map[string]string{"HELLO": "hello"},
		},
		{
			Case:     "struct with env var",
			Expected: "hello world posh",
			Template: "{{.Env.HELLO}} world {{ .Text }}",
			Context:  struct{ Text string }{Text: "posh"},
			Env:      map[string]string{"HELLO": "hello"},
		},
		{Case: "no env var", Expected: "hello world", Template: "{{.Text}} world", Context: struct{ Text string }{Text: "hello"}},
		{Case: "map", Expected: "hello world", Template: "{{.Text}} world", Context: map[string]interface{}{"Text": "hello"}},
		{Case: "empty map", Expected: " world", Template: "{{.Text}} world", Context: map[string]string{}},
	}
	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: tc.Env,
		})
		env.On("Log", mock2.Anything, mock2.Anything, mock2.Anything)
		tmpl := &Text{
			Template: tc.Template,
			Context:  tc.Context,
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

func TestCleanTemplate(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
	}{
		{
			Case:     "Literal dots",
			Expected: " ... ",
			Template: " ... ",
		},
		{
			Case:     "Literal dot",
			Expected: "hello . what's up",
			Template: "hello . what's up",
		},
		{
			Case:     "Variable",
			Expected: "{{range $cpu := .Data.CPU}}{{round $cpu.Mhz 2 }} {{end}}",
			Template: "{{range $cpu := .CPU}}{{round $cpu.Mhz 2 }} {{end}}",
		},
		{
			Case:     "Same prefix",
			Expected: "{{ .Env.HELLO }} {{ .Data.World }} {{ .Data.WorldTrend }}",
			Template: "{{ .Env.HELLO }} {{ .World }} {{ .WorldTrend }}",
		},
		{
			Case:     "Double use of property with different child",
			Expected: "{{ .Env.HELLO }} {{ .Data.World.Trend }} {{ .Data.World.Hello }} {{ .Data.World }}",
			Template: "{{ .Env.HELLO }} {{ .World.Trend }} {{ .World.Hello }} {{ .World }}",
		},
		{
			Case:     "Hello world",
			Expected: "{{.Env.HELLO}} {{.Data.World}}",
			Template: "{{.Env.HELLO}} {{.World}}",
		},
		{
			Case:     "Multiple vars",
			Expected: "{{.Env.HELLO}} {{.Data.World}} {{.Data.World}}",
			Template: "{{.Env.HELLO}} {{.World}} {{.World}}",
		},
		{
			Case:     "Multiple vars with spaces",
			Expected: "{{ .Env.HELLO }} {{ .Data.World }} {{ .Data.World }}",
			Template: "{{ .Env.HELLO }} {{ .World }} {{ .World }}",
		},
		{
			Case:     "Braces",
			Expected: "{{ if or (.Data.Working.Changed) (.Data.Staging.Changed) }}#FF9248{{ end }}",
			Template: "{{ if or (.Working.Changed) (.Staging.Changed) }}#FF9248{{ end }}",
		},
	}
	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
		}
		tmpl.cleanTemplate()
		assert.Equal(t, tc.Expected, tmpl.Template, tc.Case)
	}
}
