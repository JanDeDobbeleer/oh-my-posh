package main

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"text/template"

	"oh-my-posh/regex"

	"oh-my-posh/runtime"

	"github.com/Masterminds/sprig"
)

const (
	// Errors to show when the template handling fails
	invalidTemplate   = "invalid template text"
	incorrectTemplate = "unable to create text based on template"

	templateEnvRegex = `\.Env\.(?P<ENV>[^ \.}]*)`
)

type textTemplate struct {
	Template string
	Context  interface{}
	Env      runtime.Environment
}

func (t *textTemplate) renderPlainContextTemplate(context map[string]interface{}) string {
	if context == nil {
		context = make(map[string]interface{})
	}
	context["Root"] = t.Env.IsRunningAsRoot()
	pwd := t.Env.Getcwd()
	pwd = strings.Replace(pwd, t.Env.HomeDir(), "~", 1)
	context["Path"] = pwd
	context["Folder"] = base(pwd, t.Env)
	context["Shell"] = t.Env.GetShellName()
	context["User"] = t.Env.GetCurrentUser()
	context["Host"] = ""
	if host, err := t.Env.GetHostName(); err == nil {
		context["Host"] = host
	}
	t.Context = context
	text, err := t.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (t *textTemplate) render() (string, error) {
	tmpl, err := template.New("title").Funcs(sprig.TxtFuncMap()).Parse(t.Template)
	if err != nil {
		return "", errors.New(invalidTemplate)
	}
	if strings.Contains(t.Template, ".Env") {
		t.loadEnvVars()
	}
	buffer := new(bytes.Buffer)
	defer buffer.Reset()
	err = tmpl.Execute(buffer, t.Context)
	if err != nil {
		return "", errors.New(incorrectTemplate)
	}
	text := buffer.String()
	// issue with missingkey=zero ignored for map[string]interface{}
	// https://github.com/golang/go/issues/24963
	text = strings.ReplaceAll(text, "<no value>", "")
	return text, nil
}

func (t *textTemplate) loadEnvVars() {
	context := make(map[string]interface{})
	switch v := t.Context.(type) {
	case map[string]interface{}:
		context = v
	default:
		// we currently only support structs
		if !t.isStruct() {
			break
		}
		context = t.structToMap()
	}
	envVars := map[string]string{}
	matches := regex.FindAllNamedRegexMatch(templateEnvRegex, t.Template)
	for _, match := range matches {
		envVars[match["ENV"]] = t.Env.Getenv(match["ENV"])
	}
	context["Env"] = envVars
	t.Context = context
}

func (t *textTemplate) isStruct() bool {
	v := reflect.TypeOf(t.Context)
	if v == nil {
		return false
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Invalid {
		return false
	}
	return v.Kind() == reflect.Struct
}

func (t *textTemplate) structToMap() map[string]interface{} {
	context := make(map[string]interface{})
	v := reflect.ValueOf(t.Context)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	strct := v.Type()
	for i := 0; i < strct.NumField(); i++ {
		sf := strct.Field(i)
		if !v.Field(i).CanInterface() {
			continue
		}
		name := sf.Name
		value := v.Field(i).Interface()
		context[name] = value
	}
	return context
}
