package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Base struct {
	options options.Provider
	env     runtime.Environment

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

func (b *Base) Init(opts options.Provider, env runtime.Environment) {
	b.Segment = &Segment{}
	b.options = opts
	b.env = env
}

func (b *Base) CacheKey() (string, bool) {
	return "", false
}
