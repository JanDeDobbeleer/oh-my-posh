package engine

import (
	"errors"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/shell"
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
			Env:               env,
			rpromptLength:     tc.RPromptLength,
			currentLineLength: tc.PromptLength,
			rprompt:           "hello",
		}
		got := engine.canWriteRPrompt(true)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestPrintPWD(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		PWD      string
		OSC99    bool
	}{
		{Case: "Empty PWD"},
		{Case: "OSC99", PWD: color.OSC99, Expected: "\x1b]9;9;\"pwd\"\x1b\\"},
		{Case: "OSC7", PWD: color.OSC7, Expected: "\x1b]7;\"file://host/pwd\"\x1b\\"},
		{Case: "Deprecated OSC99", OSC99: true, Expected: "\x1b]9;9;\"pwd\"\x1b\\"},
		{Case: "Template (empty)", PWD: "{{ if eq .Shell \"pwsh\" }}osc7{{ end }}"},
		{Case: "Template (non empty)", PWD: "{{ if eq .Shell \"shell\" }}osc7{{ end }}", Expected: "\x1b]7;\"file://host/pwd\"\x1b\\"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Pwd").Return("pwd")
		env.On("Shell").Return("shell")
		env.On("Host").Return("host", nil)
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env:   make(map[string]string),
			Shell: "shell",
		})
		ansi := &color.Ansi{}
		ansi.InitPlain()
		engine := &Engine{
			Env: env,
			Config: &Config{
				PWD:   tc.PWD,
				OSC99: tc.OSC99,
			},
			Ansi: ansi,
		}
		engine.printPWD()
		got := engine.print()
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
	env.Init()
	defer env.Close()

	cfg := LoadConfig(env)
	defer testClearDefaultConfig()

	ansi := &color.Ansi{}
	ansi.InitPlain()
	writerColors := cfg.MakeColors(env)
	writer := &color.AnsiWriter{
		Ansi:               ansi,
		TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
		AnsiColors:         writerColors,
	}
	consoleTitle := &console.Title{
		Env:      env,
		Ansi:     ansi,
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
