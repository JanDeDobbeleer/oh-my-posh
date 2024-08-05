package template

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"text/template"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

const (
	// Errors to show when the template handling fails
	InvalidTemplate   = "invalid template text"
	IncorrectTemplate = "unable to create text based on template"

	globalRef = ".$"

	elvish = "elvish"
	xonsh  = "xonsh"
)

var (
	knownVariables = []string{
		"Root",
		"PWD",
		"AbsolutePWD",
		"PSWD",
		"Folder",
		"Shell",
		"ShellVersion",
		"UserName",
		"HostName",
		"Code",
		"Env",
		"OS",
		"WSL",
		"PromptCount",
		"Segments",
		"SHLVL",
		"Templates",
		"Var",
		"Data",
		"Jobs",
	}

	shell string

	tmplFunc = template.New("cache").Funcs(funcMap())
)

type Text struct {
	Template        string
	Context         any
	Env             runtime.Environment
	TemplatesResult string
}

type Data any

type Context struct {
	*cache.Template

	// Simple container to hold ANY object
	Data
	Templates string
}

func (c *Context) init(t *Text) {
	c.Data = t.Context
	c.Templates = t.TemplatesResult
	if tmplCache := t.Env.TemplateCache(); tmplCache != nil {
		c.Template = tmplCache
		return
	}
}

func (t *Text) Render() (string, error) {
	t.Env.DebugF("rendering template: %s", t.Template)

	shell = t.Env.Flags().Shell

	if !strings.Contains(t.Template, "{{") || !strings.Contains(t.Template, "}}") {
		return t.Template, nil
	}

	t.cleanTemplate()

	tmpl, err := tmplFunc.Parse(t.Template)
	if err != nil {
		t.Env.Error(err)
		return "", errors.New(InvalidTemplate)
	}

	context := &Context{}
	context.init(t)

	buffer := new(bytes.Buffer)
	defer buffer.Reset()

	err = tmpl.Execute(buffer, context)
	if err != nil {
		t.Env.Error(err)
		msg := regex.FindNamedRegexMatch(`at (?P<MSG><.*)$`, err.Error())
		if len(msg) == 0 {
			return "", errors.New(IncorrectTemplate)
		}

		return "", errors.New(msg["MSG"])
	}

	text := buffer.String()
	// issue with missingkey=zero ignored for map[string]any
	// https://github.com/golang/go/issues/24963
	text = strings.ReplaceAll(text, "<no value>", "")

	return text, nil
}

func (t *Text) cleanTemplate() {
	isKnownVariable := func(variable string) bool {
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

	fields := make(fields)
	fields.init(t.Context)

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
			if !isKnownVariable(property) { //nolint: gocritic
				result += ".Data" + property
			} else if strings.HasPrefix(property, ".Segments") && !strings.HasSuffix(property, ".Contains") {
				// as we can't provide a clean way to access the list
				// of segments, we need to replace the property with
				// the list of segments so they can be accessed directly
				property = strings.Replace(property, ".Segments", ".Segments.ToSimple", 1)
				result += property
			} else {
				// check if we have the same property in Data
				// and replace it with the Data property so it
				// can take precedence
				if fields.hasField(property) {
					property = ".Data" + property
				}
				// remove the global reference so we can use it directly
				property = strings.TrimPrefix(property, globalRef)
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

type fields map[string]bool

func (f *fields) init(data any) {
	if data == nil {
		return
	}

	val := reflect.TypeOf(data)
	switch val.Kind() { //nolint:exhaustive
	case reflect.Struct:
		fieldsNum := val.NumField()
		for i := 0; i < fieldsNum; i++ {
			(*f)[val.Field(i).Name] = true
		}
	case reflect.Map:
		m, ok := data.(map[string]any)
		if !ok {
			return
		}
		for key := range m {
			(*f)[key] = true
		}
	case reflect.Ptr:
		f.init(reflect.ValueOf(data).Elem().Interface())
	}
}

func (f fields) hasField(field string) bool {
	field = strings.TrimPrefix(field, ".")
	_, ok := f[field]
	return ok
}
