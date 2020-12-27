package main

import (
	"bytes"
	"fmt"
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

// TitleTemplateContext defines what can be aded to the title when using the template
type TitleTemplateContext struct {
	Root   bool
	Path   string
	Folder string
	Shell  string
}

func (t *consoleTitle) getConsoleTitle() string {
	var title string
	switch t.settings.ConsoleTitleStyle {
	case FullPath:
		title = t.env.getcwd()
	case Template:
		title = t.getTemplateText()
	case FolderName:
		fallthrough
	default:
		title = base(t.env.getcwd(), t.env)
	}
	return fmt.Sprintf(t.formats.title, title)
}

func (t *consoleTitle) getTemplateText() string {
	context := &TitleTemplateContext{
		Root:   t.env.isRunningAsRoot(),
		Path:   t.env.getcwd(),
		Folder: base(t.env.getcwd(), t.env),
		Shell:  t.env.getShellName(),
	}
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
