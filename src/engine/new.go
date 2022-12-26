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
// This segment is added according to the following logic:
// If the currently loaded configuration contains a segment (in one or more blocks)
// matching this segment type, it will use the provided segment writer for this type.
// If no segment type is found, either in all blocks or in the specified one, a new
// segment is created with this type and writer, and appended to the specified block.
//
// NOTE to @JanDeDobbeleer: The logic specified above assumes that when the configuration
// is parsed, it does not consider whether or not a given segment indeed has an exiting
// SegmentWriter: it just creates the &Segment{} and adds it to the block's list.
// It is only, to my understanding, when the said segment has to render itself that it
// will look into the map of existing segment writers.
// So, if I'm right:
// - Either the config specifies a segment, without this method being called before
//   rendering, and the segment will silently fail.
// - Or the config specifies the segment, this method is then called, and the segment
//   will correctly render itself.
//
// NOTE2: Do we even need this function to be a method of the Engine ?
func (eng *Engine) AddSegment(stype SegmentType, segment SegmentWriter) {
	if segment == nil {
		return
	}

	// Else, simply add the writer to the list of user-defined ones.
	// It can now be used by any previously or to-be loaded configuration.
	userSegments[stype] = segment
}
