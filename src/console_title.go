package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type consoleTitle struct {
	env      environmentInfo
	settings *Settings
	formats  *ansiFormats
}

// ConsoleTitleStyle defines how to show the title in the console window
type ConsoleTitleStyle string

const (
	// FolderName show the current folder name
	FolderName ConsoleTitleStyle = "folder"
	// FullPath show the current path
	FullPath ConsoleTitleStyle = "path"
	// Template allows a more powerful custom string
	Template ConsoleTitleStyle = "template"
	// Errors to show when the template handling fails
	invalidTitleTemplate   = "invalid template title text"
	incorrectTitleTemplate = "unable to create title based on template"
)

func (t *consoleTitle) getConsoleTitle() string {
	var title string
	switch t.settings.ConsoleTitleStyle {
	case FullPath:
		title = t.getPwd()
	case Template:
		title = t.getTemplateText()
	case FolderName:
		fallthrough
	default:
		title = base(t.getPwd(), t.env)
	}
	return fmt.Sprintf(t.formats.title, title)
}

func (t *consoleTitle) getTemplateText() string {
	context := make(map[string]interface{})

	context["Root"] = t.env.isRunningAsRoot()
	context["Path"] = t.getPwd()
	context["Folder"] = base(t.getPwd(), t.env)
	context["Shell"] = t.env.getShellName()

	// load environment variables into the map
	envVars := map[string]string{}
	matches := findAllNamedRegexMatch(`\.Env\.(?P<ENV>[^ \.}]*)`, t.settings.ConsoleTitleTemplate)
	for _, match := range matches {
		envVars[match["ENV"]] = t.env.getenv(match["ENV"])
	}
	context["Env"] = envVars

	tmpl, err := template.New("title").Parse(t.settings.ConsoleTitleTemplate)
	if err != nil {
		return invalidTitleTemplate
	}
	buffer := new(bytes.Buffer)
	defer buffer.Reset()
	err = tmpl.Execute(buffer, context)
	if err != nil {
		return incorrectTitleTemplate
	}
	return buffer.String()
}

func (t *consoleTitle) getPwd() string {
	pwd := t.env.getcwd()
	pwd = strings.Replace(pwd, t.env.homeDir(), "~", 1)
	return pwd
}
