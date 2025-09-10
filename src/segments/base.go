package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type base struct {
	props properties.Properties
	env   runtime.Environment

	Segment *Segment
}

type Segment struct {
	Text  string
	Index int
}

func (b *base) Text() string {
	return b.Segment.Text
}

func (b *base) SetText(text string) {
	b.Segment.Text = text
}

func (b *base) SetIndex(index int) {
	b.Segment.Index = index
}

func (b *base) Init(props properties.Properties, env runtime.Environment) {
	b.Segment = &Segment{}
	b.props = props
	b.env = env
}

func (b *base) CacheKey() (string, bool) {
	return "", false
}
