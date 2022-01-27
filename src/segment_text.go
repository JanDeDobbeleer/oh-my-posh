package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Text struct {
	props properties.Properties
	env   environment.Environment

	Text string
}

func (t *Text) template() string {
	return "{{ .Text }}"
}

func (t *Text) enabled() bool {
	return true
}

func (t *Text) init(props properties.Properties, env environment.Environment) {
	t.props = props
	t.env = env
}
