package template

import (
	"bytes"
	"errors"
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/regex"
	"strings"
	"text/template"
)

const (
	// Errors to show when the template handling fails
	InvalidTemplate   = "invalid template text"
	IncorrectTemplate = "unable to create text based on template"
)

type Text struct {
	Template        string
	Context         interface{}
	Env             environment.Environment
	TemplatesResult string
}

type Data interface{}

type Context struct {
	*environment.TemplateCache

	// Simple container to hold ANY object
	Data
	Templates string
}

func (c *Context) init(t *Text) {
	c.Data = t.Context
	c.Templates = t.TemplatesResult
	if cache := t.Env.TemplateCache(); cache != nil {
		c.TemplateCache = cache
		return
	}
}

func (t *Text) Render() (string, error) {
	t.cleanTemplate()
	tmpl, err := template.New(t.Template).Funcs(funcMap()).Parse(t.Template)
	if err != nil {
		t.Env.Log(environment.Error, "Render", err.Error())
		return "", errors.New(InvalidTemplate)
	}
	context := &Context{}
	context.init(t)
	buffer := new(bytes.Buffer)
	defer buffer.Reset()
	err = tmpl.Execute(buffer, context)
	if err != nil {
		t.Env.Log(environment.Error, "Render", err.Error())
		return "", errors.New(IncorrectTemplate)
	}
	text := buffer.String()
	// issue with missingkey=zero ignored for map[string]interface{}
	// https://github.com/golang/go/issues/24963
	text = strings.ReplaceAll(text, "<no value>", "")
	return text, nil
}

func (t *Text) cleanTemplate() {
	unknownVariable := func(variable string, knownVariables *[]string) (string, bool) {
		variable = strings.TrimPrefix(variable, ".")
		splitted := strings.Split(variable, ".")
		if len(splitted) == 0 {
			return "", false
		}
		for _, b := range *knownVariables {
			if b == splitted[0] {
				return "", false
			}
		}
		*knownVariables = append(*knownVariables, splitted[0])
		return splitted[0], true
	}

	knownVariables := []string{
		"Root",
		"PWD",
		"Folder",
		"Shell",
		"ShellVersion",
		"UserName",
		"HostName",
		"Env",
		"Data",
		"Code",
		"OS",
		"WSL",
		"Segments",
		"Templates",
	}
	matches := regex.FindAllNamedRegexMatch(`(?: |{|\()(?P<VAR>(\.[a-zA-Z_][a-zA-Z0-9]*)+)`, t.Template)
	for _, match := range matches {
		if variable, OK := unknownVariable(match["VAR"], &knownVariables); OK {
			pattern := fmt.Sprintf(`\.%s\b`, variable)
			dataVar := fmt.Sprintf(".Data.%s", variable)
			t.Template = regex.ReplaceAllString(pattern, t.Template, dataVar)
		}
	}
	// allow literal dots in template
	t.Template = strings.ReplaceAll(t.Template, `\.`, ".")
}
