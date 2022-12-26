package engine

import (
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/platform"
	"oh-my-posh/shell"
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
	ansi.Init(env.Shell())

	var writer color.Writer
	if flags.Plain {
		ansi.InitPlain()
		writer = &color.PlainWriter{
			Ansi: ansi,
		}
	} else {
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

// AddSegment allows to add a user-implemented prompt segment writer to the engine.
func (eng *Engine) AddSegment(stype SegmentType, segment SegmentWriter) {
	if segment == nil {
		return
	}

	// Else, simply add the writer to the list of user-defined ones.
	// It can now be used by any previously or to-be loaded configuration.
	userSegments[stype] = segment
}
