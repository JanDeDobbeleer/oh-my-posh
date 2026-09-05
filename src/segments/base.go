package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Base struct {
	options options.Provider
	env     runtime.Environment

	Segment *Segment
}

type Segment struct {
	// Text is the fully rendered markup (anchors intact), stored for
	// cross-segment references like {{ .Segments.Git.Segment.Text }} in
	// transient and tooltip templates. It must be Markup or a re-render
	// escapes its anchors back into literal text.
	Text  template.Markup
	Index int
}

func (b *Base) Text() string {
	return b.Segment.Text.String()
}

func (b *Base) SetText(text string) {
	b.Segment.Text = template.RawMarkup(text)
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

// Activation satisfies the SegmentWriter contract with the ungated default:
// the segment always executes. Writers with a cheap, provable precondition
// override this to let the engine skip their Enabled() probe entirely.
func (b *Base) Activation() Activation {
	return Activation{Always: true}
}
