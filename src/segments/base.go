package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type base struct {
	props properties.Properties
	env   runtime.Environment

	Segment *Segment `json:"Segment"`
}

type Segment struct {
	Text  string `json:"Text"`
	Index int    `json:"Index"`
}

func (s *base) Text() string {
	return s.Segment.Text
}

func (s *base) SetText(text string) {
	s.Segment.Text = text
}

func (s *base) SetIndex(index int) {
	s.Segment.Index = index
}

func (s *base) Init(props properties.Properties, env runtime.Environment) {
	s.Segment = &Segment{}
	s.props = props
	s.env = env
}
