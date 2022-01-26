package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Root struct {
	props properties.Properties
	env   environment.Environment
}

func (rt *Root) template() string {
	return "\uF0E7"
}

func (rt *Root) enabled() bool {
	return rt.env.Root()
}

func (rt *Root) init(props properties.Properties, env environment.Environment) {
	rt.props = props
	rt.env = env
}
