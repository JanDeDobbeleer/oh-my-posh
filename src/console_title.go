package main

import (
	"fmt"
	"strings"
)

type consoleTitle struct {
	env    Environment
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
	template := &textTemplate{
		Template: t.config.ConsoleTitleTemplate,
		Env:      t.env,
	}
	if text, err := template.render(); err == nil {
		return text
	}
	return ""
}

func (t *consoleTitle) getPwd() string {
	pwd := t.env.pwd()
	pwd = strings.Replace(pwd, t.env.homeDir(), "~", 1)
	return pwd
}
