package prompt

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

func TestRenderBlock(t *testing.T) {
	engine := New(&runtime.Flags{
		IsPrimary: true,
	})
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
	assert.Equal(t, "\x1b[44m\x1b[31mHello\x1b[0m\x1b[44m\x1b[31mWorld\x1b[0m", prompt)
	assert.Equal(t, 10, length)
}

func TestRenderBlockRestartCycle(t *testing.T) {
	buildCycle := func() color.Cycle {
		return color.Cycle{
			{Foreground: "#000001", Background: "#000011"},
			{Foreground: "#000002", Background: "#000022"},
			{Foreground: "#000003", Background: "#000033"},
		}
	}

	cases := []struct {
		Case               string
		ExpectedForeground color.Ansi
		ExpectedBackground color.Ansi
		RestartCycle       bool
	}{
		{
			Case:               "No restart carries the cycle over from the previous block",
			RestartCycle:       false,
			ExpectedForeground: "#000003",
			ExpectedBackground: "#000033",
		},
		{
			Case:               "RestartCycle resets the block to the cycle's first color",
			RestartCycle:       true,
			ExpectedForeground: "#000001",
			ExpectedBackground: "#000011",
		},
	}

	for _, c := range cases {
		t.Run(c.Case, func(t *testing.T) {
			origCycle := cycle
			t.Cleanup(func() { cycle = origCycle })

			engine := New(&runtime.Flags{
				IsPrimary: true,
			})
			engine.Config.Cycle = buildCycle()

			// mirror writePrimaryPromptInternal, which points the shared
			// cycle at the pristine config cycle before the first block renders
			cycle = &engine.Config.Cycle

			firstBlock := &config.Block{
				Segments: []*config.Segment{
					{Type: "text", Template: "A"},
					{Type: "text", Template: "B"},
				},
			}
			secondBlock := &config.Block{
				RestartCycle: c.RestartCycle,
				Segments: []*config.Segment{
					{Type: "text", Template: "C"},
				},
			}

			engine.writeBlockSegments(firstBlock)
			engine.writeBlockSegments(secondBlock)

			assert.Equal(t, c.ExpectedForeground, secondBlock.Segments[0].Foreground, c.Case)
			assert.Equal(t, c.ExpectedBackground, secondBlock.Segments[0].Background, c.Case)
		})
	}
}

func TestCanRenderSegment(t *testing.T) {
	cases := []struct {
		Case     string
		Executed map[string]bool
		Needs    []string
		Expected bool
	}{
		{
			Case:     "No cross segment dependencies",
			Expected: true,
		},
		{
			Case:     "Cross segment dependencies, nothing executed",
			Expected: false,
			Needs:    []string{"Foo"},
		},
		{
			Case:     "Cross segment dependencies, available",
			Expected: true,
			Executed: map[string]bool{
				"Foo": true,
			},
			Needs: []string{"Foo"},
		},
	}
	for _, c := range cases {
		segment := &config.Segment{
			Type:  "text",
			Needs: c.Needs,
		}

		engine := &Engine{}
		got := engine.canRenderSegment(segment, c.Executed)

		assert.Equal(t, c.Expected, got, c.Case)
	}
}

// TestRenderPlaceholder_ThenRenderShowsRealText covers the streaming placeholder contract: a
// pending segment's Text() serves the "..." placeholder without touching the writer, and once
// the segment completes, Render replaces it with the segment's real, template-derived text.
func TestRenderPlaceholder_ThenRenderShowsRealText(t *testing.T) {
	env := new(mock.Environment)
	env.On("Flags").Return(&runtime.Flags{})
	env.On("Shell").Return(shell.GENERIC)

	// Render writes through the template cache, so it has to exist rather than
	// be inherited from whichever test happened to run first.
	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)

	segment := &config.Segment{
		Type:     "text",
		Template: "actual content",
	}

	err := segment.MapSegmentWithWriter(env)
	assert.NoError(t, err)

	segment.RenderPlaceholder()
	assert.Equal(t, "...", segment.Text(), "a pending segment should show the placeholder text")

	segment.Render(0, true)
	assert.Equal(t, "actual content", segment.Text(), "a completed segment should show its real text")
}
