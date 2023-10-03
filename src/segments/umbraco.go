package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Umbraco struct {
	props properties.Properties
	env   platform.Environment

	Text string
}

const (
	//NewProp enables something
	NewProp properties.Property = "newprop"
)

func (u *Umbraco) Enabled() bool {
	return true
}

func (u *Umbraco) Template() string {
	return " {{.Text}} world "
}

func (u *Umbraco) Init(props properties.Properties, env platform.Environment) {
	u.props = props
	u.env = env

	u.Text = props.GetString(NewProp, "Hello")
}
