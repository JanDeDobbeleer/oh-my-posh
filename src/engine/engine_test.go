package engine

import (
	"errors"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"os"
	"path/filepath"
	"strings"
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
		{Case: "Terminal > Prompt disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 100, RPromptLength: 71},
		{Case: "Prompt > Terminal enabled", Expected: true, TerminalWidth: 200, PromptLength: 300, RPromptLength: 70},
		{Case: "Prompt > Terminal disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 300, RPromptLength: 80},
		{Case: "Prompt > Terminal disabled no room", Expected: true, TerminalWidth: 200, PromptLength: 400, RPromptLength: 80},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("TerminalWidth").Return(tc.TerminalWidth, tc.TerminalWidthError)
		ansi := &color.Ansi{}
		ansi.Init(plain)
		engine := &Engine{
			Env:  env,
			Ansi: ansi,
		}
		engine.rprompt = strings.Repeat("x", tc.RPromptLength)
		engine.console.WriteString(strings.Repeat("x", tc.PromptLength))
		got := engine.canWriteRPrompt()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func BenchmarkEngineRender(b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		err = engineRender("jandedobbeleer.omp.json")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func engineRender(configPath string) error {
	testDir, err := os.Getwd()
	if err != nil {
		return err
	}

	configPath = filepath.Join(testDir, "test", configPath)

	var (
		debug    = false
		eval     = false
		shell    = "pwsh"
		plain    = false
		pwd      = ""
		pswd     = ""
		code     = 2
		execTime = 917.0
	)

	args := &environment.Args{
		Debug:         &debug,
		Config:        &configPath,
		Eval:          &eval,
		Shell:         &shell,
		Plain:         &plain,
		PWD:           &pwd,
		PSWD:          &pswd,
		ErrorCode:     &code,
		ExecutionTime: &execTime,
	}

	env := &environment.ShellEnvironment{}
	env.Init(args)
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
		Plain:        *args.Plain,
	}

	engine.Render()

	return nil
}

func BenchmarkEngineRenderPalette(b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		err = engineRender("jandedobbeleer-palette.omp.json")
		if err != nil {
			b.Fatal(err)
		}
	}
}
