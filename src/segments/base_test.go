package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

// TestBaseTextCrossSegmentKeepsMarkup guards against text/template preferring
// the promoted Text() method over the Segment.Text field it wraps: a plain
// field reference like {{ .Segments.Foo.Text }} resolves to the method
// (Base has no promoted Text field, only Text() and a Segment pointer), so
// the method itself must return Markup or the anchors come back escaped.
func TestBaseTextCrossSegmentKeepsMarkup(t *testing.T) {
	writer := &Base{}
	writer.Init(nil, nil)
	writer.SetText("<red>x</>")

	segmentsMap := maps.NewConcurrent[any]()
	segmentsMap.Set("Foo", writer)

	origCache := template.Cache
	t.Cleanup(func() { template.Cache = origCache })
	template.Cache = &cache.Template{Segments: segmentsMap}

	env := new(mock.Environment)
	text := renderTemplate(env, "{{ .Segments.Foo.Text }}", nil)

	assert.Equal(t, "<red>x</>", text)
}
