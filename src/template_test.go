package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
		Context     interface{}
	}{
		{Case: "single property", Expected: "hello world", Template: "{{.Text}} world", Context: struct{ Text string }{Text: "hello"}},
		{Case: "invalid property", ShouldError: true, Template: "{{.Durp}} world", Context: struct{ Text string }{Text: "hello"}},
		{Case: "invalid template", ShouldError: true, Template: "{{ if .Text }} world", Context: struct{ Text string }{Text: "hello"}},
		{Case: "if statement true", Expected: "hello world", Template: "{{ if .Text }}{{.Text}} world{{end}}", Context: struct{ Text string }{Text: "hello"}},
		{Case: "if statement false", Expected: "world", Template: "{{ if .Text }}{{.Text}} {{end}}world", Context: struct{ Text string }{Text: ""}},
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

	env := &MockedEnvironment{}
	env.On("TemplateCache").Return(&TemplateCache{
		Env: make(map[string]string),
	})
	for _, tc := range cases {
		template := &textTemplate{
			Template: tc.Template,
			Context:  tc.Context,
			Env:      env,
		}
		text, err := template.render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
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
		env := &MockedEnvironment{}
		env.On("TemplateCache").Return(&TemplateCache{
			Env: tc.Env,
		})
		template := &textTemplate{
			Template: tc.Template,
			Context:  tc.Context,
			Env:      env,
		}
		text, err := template.render()
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
		template := &textTemplate{
			Template: tc.Template,
		}
		template.cleanTemplate()
		assert.Equal(t, tc.Expected, template.Template, tc.Case)
	}
}
