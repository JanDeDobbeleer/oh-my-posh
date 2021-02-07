package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
		Context  interface{}
	}{
		{Case: "single property", Expected: "hello world", Template: "{{.Text}} world", Context: struct{ Text string }{Text: "hello"}},
		{Case: "invalid property", Expected: incorrectTemplate, Template: "{{.Durp}} world", Context: struct{ Text string }{Text: "hello"}},
		{Case: "invalid template", Expected: invalidTemplate, Template: "{{ if .Text }} world", Context: struct{ Text string }{Text: "hello"}},
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
	}

	for _, tc := range cases {
		template := &textTemplate{
			Template: tc.Template,
			Context:  tc.Context,
		}
		assert.Equal(t, tc.Expected, template.render(), tc.Case)
	}
}
