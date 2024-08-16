package segments

import (
	"math/rand"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Random struct {
	props   properties.Properties
	env     runtime.Environment
	options []string

	Text string
}

const (
	RandomOptions properties.Property = "options"
)

func (r *Random) Template() string {
	return " {{ .Text }} "
}

func (r *Random) Enabled() bool {
	r.Text = r.options[rand.Intn(len(r.options))]
	return true
}

func (r *Random) Init(props properties.Properties, env runtime.Environment) {
	r.env = env
	r.props = props
	r.options = props.GetStringArray(RandomOptions, []string{"a", "b", "c"})

	if len(r.options) < 1 {
		r.options = []string{"a", "b", "c"}
	}
}
