package prompt

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// newBenchEngine builds a minimal Engine with two left-aligned plain-text segments
// in one block, using a real (but empty) runtime.Terminal so no network/git calls
// are made. Template and terminal state are reset each call so iterations are
// independent.
func newBenchEngine() *Engine {
	flags := &runtime.Flags{
		Shell:     shell.GENERIC,
		IsPrimary: true,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	template.Cache = &cache.Template{
		SimpleTemplate: cache.SimpleTemplate{
			Shell: shell.GENERIC,
		},
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)

	terminal.Init(shell.GENERIC)
	terminal.Colors = &color.Defaults{}

	cfg := &config.Config{
		Blocks: []*config.Block{
			{
				Type:      config.Prompt,
				Alignment: config.Left,
				Segments: []*config.Segment{
					{
						Type:       "text",
						Template:   "hello",
						Foreground: "white",
						Background: "blue",
					},
					{
						Type:       "text",
						Template:   "world",
						Foreground: "white",
						Background: "cyan",
					},
				},
			},
		},
	}

	return &Engine{
		Config: cfg,
		Env:    env,
	}
}

// BenchmarkEnginePrimary measures a full engine.Primary() call with a small
// deterministic config — no real git/network — using a plain shell.
func BenchmarkEnginePrimary(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		eng := newBenchEngine()
		eng.Primary()
		eng.string() // flush prompt builder
	}
}
