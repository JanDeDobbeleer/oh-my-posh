package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
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
	Env      Environment
}

type Data interface{}

type Context struct {
	TemplateCache

	// Simple container to hold ANY object
	Data
}

type TemplateCache struct {
	Root     bool
	PWD      string
	Folder   string
	Shell    string
	UserName string
	HostName string
	Code     int
	Env      map[string]string
	OS       string
	WSL      bool
}

func (c *Context) init(t *textTemplate) {
	c.Data = t.Context
	if cache := t.Env.TemplateCache(); cache != nil {
		c.TemplateCache = *cache
		return
	}
}

func (t *textTemplate) render() (string, error) {
	t.cleanTemplate()
	tmpl, err := template.New("title").Funcs(funcMap()).Parse(t.Template)
	if err != nil {
		return "", errors.New(invalidTemplate)
	}
	context := &Context{}
	context.init(t)
	buffer := new(bytes.Buffer)
	defer buffer.Reset()
	err = tmpl.Execute(buffer, context)
	if err != nil {
		return "", errors.New(incorrectTemplate)
	}
	text := buffer.String()
	// issue with missingkey=zero ignored for map[string]interface{}
	// https://github.com/golang/go/issues/24963
	text = strings.ReplaceAll(text, "<no value>", "")
	return text, nil
}

func (t *textTemplate) cleanTemplate() {
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
	knownVariables := []string{"Root", "PWD", "Folder", "Shell", "UserName", "HostName", "Env", "Data", "Code", "OS", "WSL"}
	matches := findAllNamedRegexMatch(`(?: |{|\()(?P<var>(\.[a-zA-Z_][a-zA-Z0-9]*)+)`, t.Template)
	for _, match := range matches {
		if variable, OK := unknownVariable(match["var"], &knownVariables); OK {
			pattern := fmt.Sprintf(`\.%s\b`, variable)
			dataVar := fmt.Sprintf(".Data.%s", variable)
			t.Template = replaceAllString(pattern, t.Template, dataVar)
		}
	}
}
