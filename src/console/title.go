package console

import (
	"github.com/jandedobbeleer/oh-my-posh/color"
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/template"
)

type Title struct {
	Env      platform.Environment
	Ansi     *color.Ansi
	Template string
}

func (t *Title) GetTitle() string {
	title := t.getTitleTemplateText()
	title = t.Ansi.TrimAnsi(title)
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
