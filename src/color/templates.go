package color

import (
	"oh-my-posh/environment"
	"oh-my-posh/template"
)

type Templates []string

func (t Templates) Resolve(context interface{}, env environment.Environment, defaultColor string) string {
	if len(t) == 0 {
		return defaultColor
	}
	txtTemplate := &template.Text{
		Context: context,
		Env:     env,
	}
	for _, tmpl := range t {
		txtTemplate.Template = tmpl
		value, err := txtTemplate.Render()
		if err != nil || value == "" {
			continue
		}
		return value
	}
	return defaultColor
}
