package prompt

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func setupExtraPromptTest(sh string, flags *runtime.Flags) *mock.Environment {
	env := new(mock.Environment)
	env.On("Shell").Return(sh)
	env.On("Flags").Return(flags)
	// Mock accent color retrieval for both Windows and macOS. The mock
	// forwards RunCommand as Called(command, args), so the expectation
	// takes two arguments: the command and the args slice.
	env.On("RunCommand", testifymock.Anything, testifymock.Anything).Return("4", nil)
	env.On("WindowsRegistryKeyValue", testifymock.Anything).Return(&runtime.WindowsRegistryValue{ValueType: runtime.DWORD, DWord: 0xFF0078D7}, nil)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(sh)
	terminal.Colors = color.MakeColors(nil, false, "", env)

	return env
}

func TestExtraPromptTransientZSH(t *testing.T) {
	cases := []struct {
		TerminalErr   error
		Case          string
		Template      string
		RightTemplate string
		Filler        string
		Expected      string
		TerminalWidth int
		Eval          bool
	}{
		{
			Case:     "no right template, eval - byte identical to previous behavior",
			Template: "L>",
			Eval:     true,
			Expected: fmt.Sprintf("PS1=%s\nRPROMPT=''", shell.QuotePosixStr("L>")),
		},
		{
			Case:          "right template, eval",
			Template:      "L>",
			RightTemplate: "R>",
			Eval:          true,
			Expected:      fmt.Sprintf("PS1=%s\nRPROMPT=%s", shell.QuotePosixStr("L>"), shell.QuotePosixStr("R>")),
		},
		{
			Case:          "right template with quote and backslash, eval",
			Template:      "L>",
			RightTemplate: `it's a \`,
			Eval:          true,
			Expected:      fmt.Sprintf("PS1=%s\nRPROMPT=%s", shell.QuotePosixStr("L>"), shell.QuotePosixStr(`it's a \`)),
		},
		{
			Case:          "right template and filler, eval",
			Template:      "L>",
			RightTemplate: "R>",
			Filler:        "-",
			TerminalWidth: 20,
			Eval:          true,
			Expected:      fmt.Sprintf("PS1=%s\nRPROMPT=%s", shell.QuotePosixStr("L>"+strings.Repeat("-", 16)), shell.QuotePosixStr("R>")),
		},
		{
			Case:          "right template and filler, unknown terminal width, eval",
			Template:      "L>",
			RightTemplate: "R>",
			Filler:        "-",
			TerminalErr:   errors.New("burp"),
			Eval:          true,
			Expected:      fmt.Sprintf("PS1=%s\nRPROMPT=%s", shell.QuotePosixStr("L>"), shell.QuotePosixStr("R>")),
		},
		{
			Case:          "right template, no eval - raw string without RPROMPT",
			Template:      "L>",
			RightTemplate: "R>",
			Expected:      "L>",
		},
	}

	for _, tc := range cases {
		env := setupExtraPromptTest(shell.ZSH, &runtime.Flags{Eval: tc.Eval})
		env.On("TerminalWidth").Return(tc.TerminalWidth, tc.TerminalErr)

		engine := &Engine{
			Config: &config.Config{
				TransientPrompt: &config.Segment{
					Template:      tc.Template,
					RightTemplate: tc.RightTemplate,
					Filler:        tc.Filler,
				},
			},
			Env: env,
		}

		got := engine.ExtraPrompt(Transient)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestExtraPromptTransientPWSH(t *testing.T) {
	// initialize the terminal for pwsh before resolving the expected sequences
	terminal.Init(shell.PWSH)
	saveCursor := terminal.SaveCursorPosition()
	restoreCursor := terminal.RestoreCursorPosition()
	clearAfter := terminal.ClearAfter()

	cases := []struct {
		TerminalErr   error
		Case          string
		Template      string
		RightTemplate string
		Filler        string
		Expected      string
		TerminalWidth int
	}{
		{
			Case:     "no right template - byte identical to previous behavior",
			Template: "L>",
			Expected: "L>" + clearAfter,
		},
		{
			Case:          "no right template with filler - byte identical to previous behavior",
			Template:      "L>",
			Filler:        "-",
			TerminalWidth: 20,
			Expected:      "L>" + strings.Repeat("-", 18) + clearAfter,
		},
		{
			Case:          "right template",
			Template:      "L>",
			RightTemplate: "R>",
			TerminalWidth: 20,
			Expected:      "L>" + clearAfter + saveCursor + strings.Repeat(" ", 16) + "R>" + restoreCursor,
		},
		{
			Case:          "right template with filler - padding inside cursor save/restore",
			Template:      "L>",
			RightTemplate: "R>",
			Filler:        "-",
			TerminalWidth: 20,
			Expected:      "L>" + clearAfter + saveCursor + strings.Repeat("-", 16) + "R>" + restoreCursor,
		},
		{
			Case:          "right template, exact fit",
			Template:      "L>",
			RightTemplate: "R>",
			TerminalWidth: 4,
			Expected:      "L>" + clearAfter + saveCursor + "R>" + restoreCursor,
		},
		{
			Case:          "right template, insufficient width - right side omitted",
			Template:      "L>",
			RightTemplate: "R>",
			TerminalWidth: 3,
			Expected:      "L>" + clearAfter,
		},
		{
			Case:          "right template, unknown terminal width - right side omitted",
			Template:      "L>",
			RightTemplate: "R>",
			TerminalErr:   errors.New("burp"),
			Expected:      "L>" + clearAfter,
		},
	}

	for _, tc := range cases {
		env := setupExtraPromptTest(shell.PWSH, &runtime.Flags{})
		env.On("TerminalWidth").Return(tc.TerminalWidth, tc.TerminalErr)

		engine := &Engine{
			Config: &config.Config{
				TransientPrompt: &config.Segment{
					Template:      tc.Template,
					RightTemplate: tc.RightTemplate,
					Filler:        tc.Filler,
				},
			},
			Env: env,
		}

		got := engine.ExtraPrompt(Transient)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestExtraPromptTransientPWSHNewline(t *testing.T) {
	env := setupExtraPromptTest(shell.PWSH, &runtime.Flags{PromptCount: 2})
	env.On("TerminalWidth").Return(20, nil)
	env.On("CursorPosition").Return(2, 1)

	engine := &Engine{
		Config: &config.Config{
			TransientPrompt: &config.Segment{
				Template:      "L>",
				RightTemplate: "R>",
				Newline:       true,
			},
		},
		Env: env,
	}

	got := engine.ExtraPrompt(Transient)
	// the leading newline has no width and must not shift the right side
	expected := "\nL>" + terminal.ClearAfter() + terminal.SaveCursorPosition() + strings.Repeat(" ", 16) + "R>" + terminal.RestoreCursorPosition()
	assert.Equal(t, expected, got)
}

func TestExtraPromptTransientFish(t *testing.T) {
	cases := []struct {
		TerminalErr   error
		Case          string
		RightTemplate string
		Filler        string
		ExpectedLeft  string
		ExpectedRight string
		TerminalWidth int
	}{
		{
			Case:         "no right template - byte identical to previous behavior",
			ExpectedLeft: "L>",
		},
		{
			Case:          "right template is returned separately",
			RightTemplate: "R>",
			ExpectedLeft:  "L>",
			ExpectedRight: "R>",
		},
		{
			Case:          "right template and filler",
			RightTemplate: "R>",
			Filler:        "-",
			TerminalWidth: 20,
			ExpectedLeft:  "L>" + strings.Repeat("-", 16),
			ExpectedRight: "R>",
		},
		{
			Case:          "right template and filler with unknown terminal width",
			RightTemplate: "R>",
			Filler:        "-",
			TerminalErr:   errors.New("burp"),
			ExpectedLeft:  "L>",
			ExpectedRight: "R>",
		},
		{
			Case:          "right template and filler with insufficient width",
			RightTemplate: "R>",
			Filler:        "-",
			TerminalWidth: 3,
			ExpectedLeft:  "L>",
			ExpectedRight: "R>",
		},
	}

	for _, tc := range cases {
		env := setupExtraPromptTest(shell.FISH, &runtime.Flags{})
		env.On("TerminalWidth").Return(tc.TerminalWidth, tc.TerminalErr)

		engine := &Engine{
			Config: &config.Config{
				TransientPrompt: &config.Segment{
					Template:      "L>",
					RightTemplate: tc.RightTemplate,
					Filler:        tc.Filler,
				},
			},
			Env: env,
		}

		assert.Equal(t, tc.ExpectedLeft, engine.ExtraPrompt(Transient), tc.Case)
		assert.Equal(t, tc.ExpectedRight, engine.TransientRPrompt(), tc.Case)
	}
}

func TestTransientRPromptTemplateError(t *testing.T) {
	env := setupExtraPromptTest(shell.FISH, &runtime.Flags{})
	engine := &Engine{
		Config: &config.Config{
			TransientPrompt: &config.Segment{RightTemplate: "{{"},
		},
		Env: env,
	}

	assert.NotEmpty(t, engine.TransientRPrompt())
}

func TestExtraPromptRightTemplateUnsupportedShell(t *testing.T) {
	env := setupExtraPromptTest(shell.BASH, &runtime.Flags{})

	engine := &Engine{
		Config: &config.Config{
			TransientPrompt: &config.Segment{
				Template:      "L>",
				RightTemplate: "R>",
			},
		},
		Env: env,
	}

	got := engine.ExtraPrompt(Transient)
	assert.Equal(t, "L>", got)
}

func TestShouldFillNegativePadLength(t *testing.T) {
	engine := &Engine{}

	// must not panic on a negative padding length
	got, OK := engine.shouldFill("-", -5)
	assert.False(t, OK)
	assert.Empty(t, got)
}

func setupExtraStreamingTestEnv(sh string) *mock.Environment {
	env := new(mock.Environment)
	env.On("Pwd").Return("/test")
	env.On("Home").Return("/home")
	env.On("Shell").Return(sh)
	env.On("Flags").Return(&runtime.Flags{Streaming: true})
	env.On("CursorPosition").Return(1, 1)
	env.On("StatusCodes").Return(0, "0")
	env.On("TerminalWidth").Return(120, nil)
	env.On("DirMatchesOneOf", testifymock.Anything, testifymock.Anything).Return(false)
	env.On("RunCommand", testifymock.Anything, testifymock.Anything).Return("4", nil)
	env.On("WindowsRegistryKeyValue", testifymock.Anything).Return(&runtime.WindowsRegistryValue{ValueType: runtime.DWORD, DWord: 0xFF0078D7}, nil)

	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}
	template.Init(env, nil, nil)
	terminal.Init(sh)
	terminal.Colors = color.MakeColors(nil, false, "", env)

	return env
}

func TestStreamPrimary_TransientRecordSkippedForZSHRightTemplate(t *testing.T) {
	env := setupExtraStreamingTestEnv(shell.ZSH)

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
			TransientPrompt: &config.Segment{
				Template:      "L>",
				RightTemplate: "R>",
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	for _, prompt := range prompts {
		assert.False(t, strings.HasPrefix(prompt, TransientMarker), "no transient record should be streamed for zsh with a right template")
	}
}

func TestStreamPrimary_TransientRecordSentForZSHWithoutRightTemplate(t *testing.T) {
	env := setupExtraStreamingTestEnv(shell.ZSH)

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
			TransientPrompt: &config.Segment{
				Template: "L>",
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	transientRecords := 0
	for _, prompt := range prompts {
		if strings.HasPrefix(prompt, TransientMarker) {
			transientRecords++
		}
	}

	assert.Equal(t, 1, transientRecords, "the transient record should still be streamed for zsh without a right template")
}

func TestStreamPrimary_TransientRecordContainsRightTemplateForPWSH(t *testing.T) {
	env := setupExtraStreamingTestEnv(shell.PWSH)

	engine := &Engine{
		Config: &config.Config{
			Blocks: []*config.Block{},
			TransientPrompt: &config.Segment{
				Template:      "L>",
				RightTemplate: "R>",
			},
		},
		Env: env,
	}

	out := engine.StreamPrimary()
	prompts := collectChannelOutput(out, 100*time.Millisecond)

	var transient string
	for _, prompt := range prompts {
		if after, OK := strings.CutPrefix(prompt, TransientMarker); OK {
			transient = after
		}
	}

	assert.NotEmpty(t, transient, "the transient record should be streamed for pwsh")
	assert.Contains(t, transient, terminal.SaveCursorPosition(), "the streamed transient record should carry the right-aligned template")
	assert.Contains(t, transient, "R>", "the streamed transient record should carry the right-aligned template")
}
