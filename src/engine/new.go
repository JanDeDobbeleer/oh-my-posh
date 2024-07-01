package engine

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
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

	if cfg.PatchPwshBleed {
		patchPowerShellBleed(env.Shell(), flags)
	}

	env.Var = cfg.Var
	flags.HasTransient = cfg.TransientPrompt != nil

	terminal.Init(env.Shell())
	terminal.BackgroundColor = shell.ConsoleBackgroundColor(env, cfg.TerminalBackground)
	terminal.AnsiColors = cfg.MakeColors()
	terminal.Plain = flags.Plain
	terminal.TrueColor = env.CmdFlags.TrueColor

	eng := &Engine{
		Config: cfg,
		Env:    env,
		Plain:  flags.Plain,
	}

	return eng
}

func patchPowerShellBleed(sh string, flags *platform.Flags) {
	// when in PowerShell, and force patching the bleed bug
	// we need to reduce the terminal width by 1 so the last
	// character isn't cut off by the ANSI escape sequences
	// See https://github.com/JanDeDobbeleer/oh-my-posh/issues/65
	if sh != shell.PWSH && sh != shell.PWSH5 {
		return
	}

	// only do this when relevant
	if flags.TerminalWidth <= 0 {
		return
	}

	flags.TerminalWidth--
}
