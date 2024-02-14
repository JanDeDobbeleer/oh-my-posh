package cli

import (
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/engine"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

// debugCmd represents the prompt command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Print the prompt in debug mode",
	Long:  "Print the prompt in debug mode.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		startTime := time.Now()

		env := &platform.Shell{
			CmdFlags: &platform.Flags{
				Config: config,
				Debug:  true,
				PWD:    pwd,
				Shell:  shellName,
				Plain:  plain,
			},
		}

		env.Init()
		defer env.Close()

		cfg := engine.LoadConfig(env)

		// add variables to the environment
		env.Var = cfg.Var

		writerColors := cfg.MakeColors()
		writer := &ansi.Writer{
			TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
			AnsiColors:         writerColors,
			Plain:              plain,
			TrueColor:          env.CmdFlags.TrueColor,
		}

		writer.Init(shell.GENERIC)
		eng := &engine.Engine{
			Config: cfg,
			Env:    env,
			Writer: writer,
			Plain:  plain,
		}

		fmt.Print(eng.PrintDebug(startTime, build.Version))
	},
}

func init() { //nolint:gochecknoinits
	debugCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	debugCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")
	debugCmd.Flags().BoolVarP(&plain, "plain", "p", false, "plain text output (no ANSI)")
	RootCmd.AddCommand(debugCmd)
}
