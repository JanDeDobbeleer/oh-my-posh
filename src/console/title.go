package console

import (
	"oh-my-posh/color"
	"oh-my-posh/environment"
	"oh-my-posh/template"
)

type Title struct {
	Env      environment.Environment
	Ansi     *color.Ansi
	Template string
}

func (t *Title) GetTitle() string {
	title := t.getTitleTemplateText()
	title = color.TrimAnsi(title)
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
