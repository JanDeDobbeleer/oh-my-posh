package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type text struct {
	props properties.Properties
	env   environment.Environment

	Text string
}

func (t *text) template() string {
	return "{{ .Text }}"
}

func (t *text) enabled() bool {
	return true
}

func (t *text) init(props properties.Properties, env environment.Environment) {
	t.props = props
	t.env = env
}
