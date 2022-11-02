package cli

import (
	"fmt"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/engine"
	"oh-my-posh/environment"
	"oh-my-posh/shell"
	"time"

	"github.com/spf13/cobra"
)

// debugCmd represents the prompt command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Print the prompt in debug mode",
	Long:  "Print the prompt in debug mode.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()
		env := &environment.ShellEnvironment{
			Version: cliVersion,
			CmdFlags: &environment.Flags{
				Config: config,
				Debug:  true,
			},
		}
		env.Init()
		defer env.Close()
		cfg := engine.LoadConfig(env)
		ansi := &color.Ansi{}
		ansi.InitPlain()
		writerColors := cfg.MakeColors()
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
		eng := &engine.Engine{
			Config:       cfg,
			Env:          env,
			Writer:       writer,
			ConsoleTitle: consoleTitle,
			Ansi:         ansi,
			Plain:        plain,
		}
		fmt.Print(eng.PrintDebug(startTime, cliVersion))
	},
}

func init() { //nolint:gochecknoinits
	RootCmd.AddCommand(debugCmd)
}
