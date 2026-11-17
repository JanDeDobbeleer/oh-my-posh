package prompt

import (
	libmaps "maps"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTooltipContext(t *testing.T) {
	cases := []struct {
		Case            string
		PrimaryAlias    string
		TooltipAlias    string
		Tip             string
		Action          config.Action
		HideIfSegment   bool
		Absent          bool
		Hidden          bool
		Empty           bool
		Left            bool
		SeparateRPrompt bool
		Multiple        bool
		OverflowHide    bool
		Narrow          bool
		Refresh         bool
		Shared          bool
		PreviousContext bool
		DisableTracking bool
		RemoveTooltips  bool
		WantTooltip     bool
		WantRPrompt     bool
	}{
		{Case: "default preserves independent tooltip", Action: config.Prepend, WantTooltip: true, WantRPrompt: true},
		{Case: "default replace does not cache right prompt", WantTooltip: true},
		{Case: "separate right prompt does not track by default", SeparateRPrompt: true, WantTooltip: true},
		{Case: "visible segment suppresses tooltip", Action: config.Prepend, HideIfSegment: true, WantRPrompt: true},
		{Case: "extend preserves unrelated right prompt", Action: config.Extend, HideIfSegment: true, WantRPrompt: true},
		{Case: "replace preserves unrelated right prompt", HideIfSegment: true, WantRPrompt: true},
		{Case: "separate right prompt is preserved", HideIfSegment: true, SeparateRPrompt: true, WantRPrompt: true},
		{Case: "separate right prompt preserves left context", HideIfSegment: true, SeparateRPrompt: true, Left: true, WantRPrompt: true},
		{
			Case: "separate right prompt refresh removes empty context", HideIfSegment: true,
			SeparateRPrompt: true, Refresh: true, WantTooltip: true,
		},
		{
			Case: "right prompt refresh preserves shared type in left prompt", HideIfSegment: true,
			SeparateRPrompt: true, Left: true, Shared: true, Refresh: true, WantRPrompt: true,
		},
		{
			Case: "right prompt refresh preserves shared alias in left prompt", HideIfSegment: true,
			PrimaryAlias: "context", TooltipAlias: "context", SeparateRPrompt: true,
			Left: true, Shared: true, Refresh: true, WantRPrompt: true,
		},
		{
			Case: "right prompt refresh preserves published context", HideIfSegment: true,
			SeparateRPrompt: true, Refresh: true, PreviousContext: true, WantRPrompt: true,
		},
		{
			Case: "primary stops tracking when disabled", HideIfSegment: true,
			Left: true, Refresh: true, DisableTracking: true, WantTooltip: true,
		},
		{
			Case: "right prompt stops tracking when disabled", HideIfSegment: true,
			SeparateRPrompt: true, Refresh: true, DisableTracking: true, WantTooltip: true,
		},
		{
			Case: "primary stops tracking without tooltips", HideIfSegment: true,
			Left: true, Refresh: true, DisableTracking: true, RemoveTooltips: true, WantRPrompt: true,
		},
		{
			Case: "right prompt stops tracking without tooltips", HideIfSegment: true,
			SeparateRPrompt: true, Refresh: true, DisableTracking: true, RemoveTooltips: true, WantRPrompt: true,
		},
		{Case: "absent segment allows tooltip", Action: config.Prepend, HideIfSegment: true, Absent: true, WantTooltip: true, WantRPrompt: true},
		{Case: "disabled segment allows tooltip", Action: config.Prepend, HideIfSegment: true, Hidden: true, WantTooltip: true, WantRPrompt: true},
		{Case: "empty segment allows tooltip", Action: config.Prepend, HideIfSegment: true, Empty: true, WantTooltip: true, WantRPrompt: true},
		{Case: "matching alias suppresses tooltip", PrimaryAlias: "context", TooltipAlias: "context", HideIfSegment: true, WantRPrompt: true},
		{Case: "different alias allows tooltip", PrimaryAlias: "context", TooltipAlias: "other", HideIfSegment: true, WantTooltip: true},
		{Case: "alias does not match type", PrimaryAlias: "context", HideIfSegment: true, WantTooltip: true},
		{Case: "left prompt also suppresses tooltip", HideIfSegment: true, Left: true, WantRPrompt: true},
		{Case: "command still must match", Tip: "other-command", HideIfSegment: true, Absent: true, WantRPrompt: true},
		{Case: "command whitespace is trimmed", Tip: " tool ", HideIfSegment: true, Absent: true, WantTooltip: true},
		{Case: "tooltips do not suppress one another", HideIfSegment: true, Absent: true, Multiple: true, WantTooltip: true},
		{
			Case: "overflow hidden block allows tooltip", Action: config.Prepend,
			HideIfSegment: true, OverflowHide: true, WantTooltip: true, WantRPrompt: true,
		},
		{
			Case: "streamed overflow refresh allows tooltip", Action: config.Prepend,
			HideIfSegment: true, OverflowHide: true, Refresh: true, WantTooltip: true, WantRPrompt: true,
		},
		{
			Case: "right prompt rejected for width allows tooltip", Action: config.Prepend,
			HideIfSegment: true, Narrow: true, WantTooltip: true, WantRPrompt: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			originalCache, originalCycle, originalColors := template.Cache, cycle, terminal.Colors
			t.Cleanup(func() {
				template.Cache, cycle, terminal.Colors = originalCache, originalCycle, originalColors
				cache.Session.Delete(cache.TEMPLATECACHE)
				cache.Session.Delete(cache.PROMPTTIMECACHE)
				cache.Session.Delete(RPromptKey)
				cache.Session.Delete(RPromptLengthKey)
				cache.Session.Delete(renderedSegmentsKey)
			})
			cache.Session.Delete(RPromptKey)
			cache.Session.Delete(RPromptLengthKey)
			cache.Session.Delete(renderedSegmentsKey)

			flags := &runtime.Flags{IsPrimary: true}
			env := new(mock.Environment)
			env.On("Flags").Return(flags)
			shellName := shell.GENERIC
			if tc.SeparateRPrompt {
				shellName = shell.FISH
			}
			env.On("Shell").Return(shellName)
			env.On("Pwd").Return("project")
			env.On("CursorPosition").Return(1, 1)
			width := 200
			if tc.Narrow || (tc.OverflowHide && !tc.Refresh) {
				width = 10
			}
			terminalWidth := env.On("TerminalWidth").Return(width, nil)
			env.On("Getenv", testifymock.Anything).Return("")
			env.On("RunCommand", testifymock.Anything, testifymock.Anything).Return("4", nil)
			env.On("WindowsRegistryKeyValue", testifymock.Anything).Return(
				&runtime.WindowsRegistryValue{ValueType: runtime.DWORD, DWord: 0xFF0078D7}, nil,
			)
			if tc.Hidden {
				env.On("DirMatchesOneOf", "project", []string{"project"}).Return(true)
			}
			env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)

			template.Cache = &cache.Template{Segments: maps.NewConcurrent[any]()}

			contextual := &config.Segment{Type: config.TEXT, Alias: tc.PrimaryAlias, Template: "CONTEXT"}
			if tc.Hidden {
				contextual.ExcludeFolders = []string{"project"}
			}
			if tc.Empty {
				contextual.Template = "{{ \"\" }}"
			}

			right := &config.Block{
				Type:      config.RPrompt,
				Alignment: config.Right,
				Segments:  []*config.Segment{{Type: config.TEXT, Alias: "unrelated", Template: "UNRELATED"}},
			}
			tooltip := &config.Segment{
				Type:          config.TEXT,
				Alias:         tc.TooltipAlias,
				Template:      "TOOLTIP",
				Tips:          []string{"tool"},
				HideIfSegment: tc.HideIfSegment,
			}
			engine := &Engine{
				Env: env,
				Config: &config.Config{
					ToolTipsAction: tc.Action,
					Tooltips:       []*config.Segment{tooltip},
					Blocks:         []*config.Block{right},
				},
			}
			switch {
			case tc.Absent:
			case tc.OverflowHide:
				engine.Config.Blocks = append([]*config.Block{{
					Type: config.Prompt, Alignment: config.Right, Overflow: config.Hide, Segments: []*config.Segment{contextual},
				}}, right)
			case tc.Left:
				engine.Config.Blocks = append([]*config.Block{{
					Type: config.Prompt, Alignment: config.Left, Segments: []*config.Segment{contextual},
				}}, right)
			default:
				right.Segments = append([]*config.Segment{contextual}, right.Segments...)
			}
			if tc.Multiple {
				engine.Config.Tooltips = append(engine.Config.Tooltips, &config.Segment{
					Type: config.TEXT, Template: "SECOND", Tips: []string{"tool"}, HideIfSegment: true,
				})
			}
			rightContextual := contextual
			if tc.Shared {
				rightContextual = &config.Segment{Type: config.TEXT, Alias: tc.PrimaryAlias, Template: "RIGHT-CONTEXT"}
				right.Segments = append([]*config.Segment{rightContextual}, right.Segments...)
			}
			engine = newEngine(engine.Config, env)

			engine.Primary()
			if tc.SeparateRPrompt {
				engine.RPrompt()
			}
			previousContext, tracked := cache.Session.Get[map[string]uint8](renderedSegmentsKey)
			assert.Equal(t, tc.HideIfSegment || tc.Multiple, tracked)
			originalContext := libmaps.Clone(previousContext)
			_, cachedRPrompt := cache.Session.Get[string](RPromptKey)
			assert.Equal(t, tc.HideIfSegment || tc.Multiple || !tc.Action.IsDefault(), cachedRPrompt)
			if tc.DisableTracking {
				tooltip.HideIfSegment = false
				right.Segments[len(right.Segments)-1].Alias = "refreshed"
				if tc.RemoveTooltips {
					engine.Config.Tooltips = nil
				}
			}
			switch {
			case tc.Refresh && tc.SeparateRPrompt:
				rightContextual.Template = `{{ "" }}`
				engine.RPrompt()
			case tc.Refresh:
				terminalWidth.Return(10, nil)
				engine.renderFromBlocks()
			}
			if tc.DisableTracking {
				currentContext, ok := cache.Session.Get[map[string]uint8](renderedSegmentsKey)
				require.True(t, ok)
				assert.Equal(t, originalContext, currentContext)
			}
			if tc.PreviousContext {
				cache.Session.Set(renderedSegmentsKey, previousContext, cache.ONEDAY)
			}
			flags.IsPrimary = false

			tip := tc.Tip
			if tip == "" {
				tip = "tool"
			}
			output := engine.Tooltip(tip)
			assert.Equal(t, tc.WantTooltip, strings.Contains(output, "TOOLTIP"))
			assert.Equal(t, tc.WantRPrompt, strings.Contains(output, "UNRELATED"))
			if tc.Multiple {
				assert.Contains(t, output, "SECOND")
			}
			if !tc.WantTooltip && tip == "tool" {
				assert.False(t, tooltip.Enabled)
			}
		})
	}
}

func TestTooltipFallback_NoCacheReturnsEmpty(t *testing.T) {
	cache.Session.Delete(RPromptKey)
	cache.Session.Delete(RPromptLengthKey)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.ZSH)

	terminal.Init(shell.ZSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Empty(t, got)
}

func TestTooltipFallback_NoMatchReturnsRPromptText(t *testing.T) {
	cache.Session.Delete(RPromptKey)
	cache.Session.Delete(RPromptLengthKey)
	cache.Session.Set(RPromptKey, "my-rprompt", cache.INFINITE)
	cache.Session.Set(RPromptLengthKey, 10, cache.INFINITE)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.ZSH)

	terminal.Init(shell.ZSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Equal(t, "my-rprompt", got)
}

func TestTooltipFallback_PwshNoMatchReturnsCursorPositionedRPrompt(t *testing.T) {
	cache.Session.Delete(RPromptKey)
	cache.Session.Delete(RPromptLengthKey)
	cache.Session.Set(RPromptKey, "my-rprompt", cache.INFINITE)
	cache.Session.Set(RPromptLengthKey, 10, cache.INFINITE)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Column: 5})
	env.On("TerminalWidth").Return(200, nil)

	terminal.Init(shell.PWSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Contains(t, got, "my-rprompt")
	assert.Contains(t, got, terminal.SaveCursorPosition())
	assert.Contains(t, got, terminal.RestoreCursorPosition())
}

func TestTooltipFallback_PwshNoRoomReturnsEmpty(t *testing.T) {
	cache.Session.Delete(RPromptKey)
	cache.Session.Delete(RPromptLengthKey)
	cache.Session.Set(RPromptKey, "my-rprompt", cache.INFINITE)
	// rprompt length exceeds available terminal space
	cache.Session.Set(RPromptLengthKey, 200, cache.INFINITE)

	env := new(mock.Environment)
	env.On("Shell").Return(shell.PWSH)
	env.On("Flags").Return(&runtime.Flags{Column: 100})
	env.On("TerminalWidth").Return(200, nil)

	terminal.Init(shell.PWSH)

	engine := &Engine{
		Env:    env,
		Config: &config.Config{},
	}

	got := engine.Tooltip("unknown-command")
	assert.Empty(t, got)
}
