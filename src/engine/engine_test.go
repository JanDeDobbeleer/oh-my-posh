package engine

import (
	"errors"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanWriteRPrompt(t *testing.T) {
	cases := []struct {
		Case               string
		Expected           bool
		TerminalWidth      int
		TerminalWidthError error
		PromptLength       int
		RPromptLength      int
	}{
		{Case: "Width Error", Expected: true, TerminalWidthError: errors.New("burp")},
		{Case: "Terminal > Prompt enabled", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 10},
		{Case: "Terminal > Prompt enabled edge", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 70},
		{Case: "Prompt > Terminal enabled", Expected: true, TerminalWidth: 200, PromptLength: 300, RPromptLength: 70},
		{Case: "Terminal > Prompt disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 100, RPromptLength: 71},
		{Case: "Prompt > Terminal disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 300, RPromptLength: 80},
		{Case: "Prompt > Terminal disabled no room", Expected: true, TerminalWidth: 200, PromptLength: 400, RPromptLength: 80},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("TerminalWidth").Return(tc.TerminalWidth, tc.TerminalWidthError)
		engine := &Engine{
			Env: env,
		}
		engine.rpromptLength = tc.RPromptLength
		engine.currentLineLength = tc.PromptLength
		got := engine.canWriteRPrompt()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func BenchmarkEngineRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engineRender()
	}
}

func engineRender() {
	env := &environment.ShellEnvironment{}
	env.Init(false)
	defer env.Close()

	cfg := LoadConfig(env)
	defer testClearDefaultConfig()

	ansi := &color.Ansi{}
	ansi.Init(env.Shell())
	writerColors := cfg.MakeColors(env)
	writer := &color.AnsiWriter{
		Ansi:               ansi,
		TerminalBackground: GetConsoleBackgroundColor(env, cfg.TerminalBackground),
		AnsiColors:         writerColors,
	}
	consoleTitle := &console.Title{
		Env:      env,
		Ansi:     ansi,
		Style:    cfg.ConsoleTitleStyle,
		Template: cfg.ConsoleTitleTemplate,
	}
	engine := &Engine{
		Config:       cfg,
		Env:          env,
		Writer:       writer,
		ConsoleTitle: consoleTitle,
		Ansi:         ansi,
	}

	engine.PrintPrimary()
}

func BenchmarkEngineRenderPalette(b *testing.B) {
	for i := 0; i < b.N; i++ {
		engineRender()
	}
}
