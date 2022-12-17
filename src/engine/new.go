package engine

import (
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/platform"
	"oh-my-posh/shell"
)

// NewEngine returns a prompt engine initialized with the
// given configuration options, and is ready to print any
// of the prompt components.
func NewEngine(config string) *Engine {
	env := &platform.Shell{
		CmdFlags: &platform.Flags{
			Config: config,
			Shell:  "shell",
		},
	}

	env.Init()
	defer env.Close()
	cfg := LoadConfig(env)
	ansi := &color.Ansi{}
	ansi.Init(env.Shell())

	// Note that here plain (no ANSI) is not supported,
	// maybe one would want to still be able to specify it ?
	var writer color.Writer
	writerColors := cfg.MakeColors()
	writer = &color.AnsiWriter{
		Ansi:               ansi,
		TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
		AnsiColors:         writerColors,
	}

	consoleTitle := &console.Title{
		Env:      env,
		Ansi:     ansi,
		Template: cfg.ConsoleTitleTemplate,
	}

	eng := &Engine{
		Config:       cfg,
		Env:          env,
		Writer:       writer,
		ConsoleTitle: consoleTitle,
		Ansi:         ansi,
		Plain:        false, // No plain support
	}

	return eng
}
