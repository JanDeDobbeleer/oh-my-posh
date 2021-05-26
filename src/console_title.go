package main

import (
	"fmt"
	"strings"
)

type consoleTitle struct {
	env    environmentInfo
	config *Config
	ansi   *ansiUtils
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
)

func (t *consoleTitle) getConsoleTitle() string {
	var title string
	switch t.config.ConsoleTitleStyle {
	case FullPath:
		title = t.getPwd()
	case Template:
		title = t.getTemplateText()
	case FolderName:
		fallthrough
	default:
		title = base(t.getPwd(), t.env)
	}
	title = t.ansi.escapeText(title)
	return fmt.Sprintf(t.ansi.title, title)
}

func (t *consoleTitle) getTemplateText() string {
	context := make(map[string]interface{})

	context["Root"] = t.env.isRunningAsRoot()
	context["Path"] = t.getPwd()
	context["Folder"] = base(t.getPwd(), t.env)
	context["Shell"] = t.env.getShellName()
	context["User"] = t.env.getCurrentUser()
	context["Host"] = ""
	if host, err := t.env.getHostName(); err == nil {
		context["Host"] = host
	}

	template := &textTemplate{
		Template: t.config.ConsoleTitleTemplate,
		Context:  context,
		Env:      t.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (t *consoleTitle) getPwd() string {
	pwd := t.env.getcwd()
	pwd = strings.Replace(pwd, t.env.homeDir(), "~", 1)
	return pwd
}
