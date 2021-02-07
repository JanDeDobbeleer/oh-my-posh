package main

import (
	"bytes"
	"text/template"
)

const (
	// Errors to show when the template handling fails
	invalidTemplate   = "invalid template text"
	incorrectTemplate = "unable to create text based on template"
)

type textTemplate struct {
	Template string
	Context  interface{}
}

func (t *textTemplate) render() string {
	tmpl, err := template.New("title").Parse(t.Template)
	if err != nil {
		return invalidTemplate
	}
	buffer := new(bytes.Buffer)
	defer buffer.Reset()
	err = tmpl.Execute(buffer, t.Context)
	if err != nil {
		return incorrectTemplate
	}
	return buffer.String()
}
