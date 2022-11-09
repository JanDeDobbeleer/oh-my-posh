package segments

import (
	"oh-my-posh/platform"
	"oh-my-posh/properties"
)

type Root struct {
	props properties.Properties
	env   platform.Environment
}

func (rt *Root) Template() string {
	return " \uF0E7 "
}

func (rt *Root) Enabled() bool {
	return rt.env.Root()
}

func (rt *Root) Init(props properties.Properties, env platform.Environment) {
	rt.props = props
	rt.env = env
}
