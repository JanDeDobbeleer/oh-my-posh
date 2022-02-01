package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Text struct {
	props properties.Properties
	env   environment.Environment

	Text string
}

func (t *Text) Template() string {
	return " {{ .Text }} "
}

func (t *Text) Enabled() bool {
	return true
}

func (t *Text) Init(props properties.Properties, env environment.Environment) {
	t.props = props
	t.env = env
}
