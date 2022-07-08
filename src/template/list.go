package template

import (
	"oh-my-posh/environment"
	"strings"
)

type Logic string

const (
	FirstMatch Logic = "first_match"
	Join       Logic = "join"
)

type List []string

func (l List) Empty() bool {
	return len(l) == 0
}

func (l List) Resolve(context interface{}, env environment.Environment, defaultValue string, logic Logic) string {
	switch logic {
	case FirstMatch:
		return l.FirstMatch(context, env, defaultValue)
	case Join:
		fallthrough
	default:
		return l.Join(context, env)
	}
}

func (l List) Join(context interface{}, env environment.Environment) string {
	if len(l) == 0 {
		return ""
	}
	txtTemplate := &Text{
		Context: context,
		Env:     env,
	}
	var buffer strings.Builder
	for _, tmpl := range l {
		txtTemplate.Template = tmpl
		value, err := txtTemplate.Render()
		if err != nil || len(strings.TrimSpace(value)) == 0 {
			continue
		}
		buffer.WriteString(value)
	}
	return buffer.String()
}

func (l List) FirstMatch(context interface{}, env environment.Environment, defaultValue string) string {
	if len(l) == 0 {
		return defaultValue
	}
	txtTemplate := &Text{
		Context: context,
		Env:     env,
	}
	for _, tmpl := range l {
		txtTemplate.Template = tmpl
		value, err := txtTemplate.Render()
		if err != nil || len(strings.TrimSpace(value)) == 0 {
			continue
		}
		return value
	}
	return defaultValue
}
