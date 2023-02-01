package engine

import (
	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
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

	ansiWriter := &ansi.Writer{
		TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
		AnsiColors:         cfg.MakeColors(),
		Plain:              flags.Plain,
	}
	ansiWriter.Init(env.Shell())

	eng := &Engine{
		Config: cfg,
		Env:    env,
		Writer: ansiWriter,
		Plain:  flags.Plain,
	}

	return eng
}
