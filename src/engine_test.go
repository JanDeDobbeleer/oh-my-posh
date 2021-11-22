package main

import (
	"errors"
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
		env := new(MockedEnvironment)
		env.On("getTerminalWidth", nil).Return(tc.TerminalWidth, tc.TerminalWidthError)
		ansi := &ansiUtils{}
		ansi.init(plain)
		engine := &engine{
			env:  env,
			ansi: ansi,
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

	configPath = filepath.Join(testDir, "testdata", configPath)

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

	args := &args{
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

	env := &environment{}
	env.init(args)
	defer env.close()

	cfg := GetConfig(env)
	defer testClearDefaultConfig()

	ansi := &ansiUtils{}
	ansi.init(env.getShellName())
	writerColors := MakeColors(env, cfg)
	writer := &AnsiWriter{
		ansi:               ansi,
		terminalBackground: getConsoleBackgroundColor(env, cfg.TerminalBackground),
		ansiColors:         writerColors,
	}
	title := &consoleTitle{
		env:    env,
		config: cfg,
		ansi:   ansi,
	}
	engine := &engine{
		config:       cfg,
		env:          env,
		writer:       writer,
		consoleTitle: title,
		ansi:         ansi,
		plain:        *args.Plain,
	}

	engine.render()

	return nil
}
