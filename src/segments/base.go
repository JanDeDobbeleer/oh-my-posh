package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Base struct {
	props properties.Properties
	env   runtime.Environment

	Segment *Segment
}

type Segment struct {
	Text  string
	Index int
}

func (b *Base) Text() string {
	return b.Segment.Text
}

func (b *Base) SetText(text string) {
	b.Segment.Text = text
}

func (b *Base) SetIndex(index int) {
	b.Segment.Index = index
}

func (b *Base) Init(props properties.Properties, env runtime.Environment) {
	b.Segment = &Segment{}
	b.props = props
	b.env = env
}

func (b *Base) CacheKey() (string, bool) {
	return "", false
}
