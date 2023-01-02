package engine

import (
	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/console"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

// New returns a prompt engine initialized with the
// given configuration options, and is ready to print any
// of the prompt components.
func New(flags *platform.Flags) *Engine {
	env := &platform.Shell{
		CmdFlags: flags,
	}

	env.Init()
	cfg := LoadConfig(env)
	ansi := &color.Ansi{}

	var writer color.Writer
	if flags.Plain {
		ansi.InitPlain()
		writer = &color.PlainWriter{
			Ansi: ansi,
		}
	} else {
		ansi.Init(env.Shell())
		writerColors := cfg.MakeColors()
		writer = &color.AnsiWriter{
			Ansi:               ansi,
			TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
			AnsiColors:         writerColors,
		}
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
		Plain:        flags.Plain,
	}

	return eng
}
