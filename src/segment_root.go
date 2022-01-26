package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type root struct {
	props properties.Properties
	env   environment.Environment
}

func (rt *root) template() string {
	return "\uF0E7"
}

func (rt *root) enabled() bool {
	return rt.env.Root()
}

func (rt *root) init(props properties.Properties, env environment.Environment) {
	rt.props = props
	rt.env = env
}
