package template

import (
	"bytes"
	"errors"
	"oh-my-posh/platform"
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
	Env             platform.Environment
	TemplatesResult string
}

type Data interface{}

type Context struct {
	*platform.TemplateCache

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
	if !strings.Contains(t.Template, "{{") || !strings.Contains(t.Template, "}}") {
		return t.Template, nil
	}
	t.cleanTemplate()
	tmpl, err := template.New(t.Template).Funcs(funcMap()).Parse(t.Template)
	if err != nil {
		t.Env.Error("Render", err)
		return "", errors.New(InvalidTemplate)
	}
	context := &Context{}
	context.init(t)
	buffer := new(bytes.Buffer)
	defer buffer.Reset()
	err = tmpl.Execute(buffer, context)
	if err != nil {
		t.Env.Error("Render", err)
		msg := regex.FindNamedRegexMatch(`at (?P<MSG><.*)$`, err.Error())
		if len(msg) == 0 {
			return "", errors.New(IncorrectTemplate)
		}
		return "", errors.New(msg["MSG"])
	}
	text := buffer.String()
	// issue with missingkey=zero ignored for map[string]interface{}
	// https://github.com/golang/go/issues/24963
	text = strings.ReplaceAll(text, "<no value>", "")
	return text, nil
}

func (t *Text) cleanTemplate() {
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

	knownVariable := func(variable string) bool {
		variable = strings.TrimPrefix(variable, ".")
		splitted := strings.Split(variable, ".")
		if len(splitted) == 0 {
			return true
		}
		variable = splitted[0]
		// check if alphanumeric
		if !regex.MatchString(`^[a-zA-Z0-9]+$`, variable) {
			return true
		}
		for _, b := range knownVariables {
			if variable == b {
				return true
			}
		}
		return false
	}

	var result, property string
	var inProperty, inTemplate bool
	for i, char := range t.Template {
		// define start or end of template
		if !inTemplate && char == '{' {
			if i-1 >= 0 && rune(t.Template[i-1]) == '{' {
				inTemplate = true
			}
		} else if inTemplate && char == '}' {
			if i-1 >= 0 && rune(t.Template[i-1]) == '}' {
				inTemplate = false
			}
		}
		if !inTemplate {
			result += string(char)
			continue
		}
		switch char {
		case '.':
			var lastChar rune
			if len(result) > 0 {
				lastChar = rune(result[len(result)-1])
			}
			// only replace if we're in a valid property start
			// with a space, { or ( character
			switch lastChar {
			case ' ', '{', '(':
				property += string(char)
				inProperty = true
			default:
				result += string(char)
			}
		case ' ', '}', ')': // space or }
			if !inProperty {
				result += string(char)
				continue
			}
			// end of a variable, needs to be appended
			if !knownVariable(property) {
				result += ".Data" + property
			} else {
				result += property
			}
			property = ""
			result += string(char)
			inProperty = false
		default:
			if inProperty {
				property += string(char)
				continue
			}
			result += string(char)
		}
	}

	// return the result and remaining unresolved property
	t.Template = result + property
}
