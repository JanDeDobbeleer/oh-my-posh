package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Root struct {
	props properties.Properties
	env   runtime.Environment
}

func (rt *Root) Template() string {
	return " \uF0E7 "
}

func (rt *Root) Enabled() bool {
	return rt.env.Root()
}

func (rt *Root) Init(props properties.Properties, env runtime.Environment) {
	rt.props = props
	rt.env = env
}
