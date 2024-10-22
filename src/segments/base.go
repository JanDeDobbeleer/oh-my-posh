package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type base struct {
	props properties.Properties
	env   runtime.Environment

	Output string `json:"Text"`
}

func (s *base) Text() string {
	return s.Output
}

func (s *base) SetText(text string) {
	s.Output = text
}

func (s *base) Init(props properties.Properties, env runtime.Environment) {
	s.props = props
	s.env = env
}
