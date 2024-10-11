package prompt

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/stretchr/testify/assert"
)

func TestRenderBlock(t *testing.T) {
	engine := New(&runtime.Flags{})
	block := &config.Block{
		Segments: []*config.Segment{
			{
				Type:       "text",
				Template:   "Hello",
				Foreground: "red",
				Background: "blue",
			},
			{
				Type:       "text",
				Template:   "World",
				Foreground: "red",
				Background: "blue",
			},
		},
	}

	prompt, length := engine.writeBlockSegments(block)
	assert.NotEmpty(t, prompt)
	assert.Equal(t, 10, length)
}

func TestCanRenderSegment(t *testing.T) {
	cases := []struct {
		Case             string
		Template         string
		ExecutedSegments []string
		Expected         bool
	}{
		{
			Case:     "No cross segment dependencies",
			Expected: true,
			Template: "Hello",
		},
		{
			Case:     "Cross segment dependencies, nothing executed",
			Expected: false,
			Template: "Hello {{ .Segments.Foo.World }} {{ .Segments.Foo.Bar }}",
		},
		{
			Case:     "Cross segment dependencies, available",
			Expected: true,
			Template: "Hello {{ .Segments.Foo.World }}",
			ExecutedSegments: []string{
				"Foo",
			},
		},
	}
	for _, c := range cases {
		segment := &config.Segment{
			Type:     "text",
			Template: c.Template,
		}

		engine := &Engine{}
		got := engine.canRenderSegment(segment, c.ExecutedSegments)

		assert.Equal(t, c.Expected, got, c.Case)
	}
}
