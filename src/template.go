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
	// nostruct          = "unable to create map from non-struct type"

	templateEnvRegex = `\.Env\.(?P<ENV>[^ \.}]*)`
)

type textTemplate struct {
	Template string
	Context  interface{}
	Env      Environment
}

type Data interface{}

type Context struct {
	Root   bool
	PWD    string
	Folder string
	Shell  string
	User   string
	Host   string
	Code   int
	Env    map[string]string

	// Simple container to hold ANY object
	Data
}

func (c *Context) init(t *textTemplate) {
	c.Data = t.Context
	if t.Env == nil {
		return
	}
	c.Root = t.Env.isRunningAsRoot()
	pwd := t.Env.getcwd()
	pwd = strings.Replace(pwd, t.Env.homeDir(), "~", 1)
	c.PWD = pwd
	c.Folder = base(c.PWD, t.Env)
	c.Shell = t.Env.getShellName()
	c.User = t.Env.getCurrentUser()
	if host, err := t.Env.getHostName(); err == nil {
		c.Host = host
	}
	c.Code = t.Env.lastErrorCode()
	if strings.Contains(t.Template, ".Env.") {
		c.Env = map[string]string{}
		matches := findAllNamedRegexMatch(templateEnvRegex, t.Template)
		for _, match := range matches {
			c.Env[match["ENV"]] = t.Env.getenv(match["ENV"])
		}
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
	knownVariables := []string{"Root", "PWD", "Folder", "Shell", "User", "Host", "Env", "Data", "Code"}
	matches := findAllNamedRegexMatch(`(?: |{)(?P<var>(\.[a-zA-Z_][a-zA-Z0-9]*)+)`, t.Template)
	for _, match := range matches {
		if variable, OK := unknownVariable(match["var"], &knownVariables); OK {
			pattern := fmt.Sprintf(`\.%s\b`, variable)
			dataVar := fmt.Sprintf(".Data.%s", variable)
			t.Template = replaceAllString(pattern, t.Template, dataVar)
		}
	}
}
