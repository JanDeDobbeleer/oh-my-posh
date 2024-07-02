package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Text struct {
	props properties.Properties
	env   runtime.Environment
}

func (t *Text) Template() string {
	return "  "
}

func (t *Text) Enabled() bool {
	return true
}

func (t *Text) Init(props properties.Properties, env runtime.Environment) {
	t.props = props
	t.env = env
}
