package main

import "oh-my-posh/environment"

type text struct {
	props Properties
	env   environment.Environment

	Text string
}

func (t *text) template() string {
	return "{{ .Text }}"
}

func (t *text) enabled() bool {
	return true
}

func (t *text) init(props Properties, env environment.Environment) {
	t.props = props
	t.env = env
}
