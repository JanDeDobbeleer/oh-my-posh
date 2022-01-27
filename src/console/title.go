package console

import (
	"oh-my-posh/color"
	"oh-my-posh/environment"
	"oh-my-posh/template"
	"strings"
)

type Title struct {
	Env      environment.Environment
	Ansi     *color.Ansi
	Style    Style
	Template string
}

// Style defines how to show the title in the console window
type Style string

const (
	// FolderName show the current folder name
	FolderName Style = "folder"
	// FullPath show the current path
	FullPath Style = "path"
	// Template allows a more powerful custom string
	Template Style = "template"
)

func (t *Title) GetTitle() string {
	var title string
	switch t.Style {
	case FullPath:
		title = t.getPwd()
	case Template:
		title = t.getTitleTemplateText()
	case FolderName:
		fallthrough
	default:
		title = environment.Base(t.Env, t.getPwd())
	}
	title = t.Ansi.EscapeText(title)
	return t.Ansi.Title(title)
}

func (t *Title) getTitleTemplateText() string {
	tmpl := &template.Text{
		Template: t.Template,
		Env:      t.Env,
	}
	if text, err := tmpl.Render(); err == nil {
		return text
	}
	return ""
}

func (t *Title) getPwd() string {
	pwd := t.Env.Pwd()
	pwd = strings.Replace(pwd, t.Env.Home(), "~", 1)
	return pwd
}
