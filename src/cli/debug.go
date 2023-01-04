package cli

import (
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/ansi"
	"github.com/jandedobbeleer/oh-my-posh/engine"
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/shell"

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
		env := &platform.Shell{
			Version: cliVersion,
			CmdFlags: &platform.Flags{
				Config: config,
				Debug:  true,
				PWD:    pwd,
				Shell:  shellName,
			},
		}
		env.Init()
		defer env.Close()
		cfg := engine.LoadConfig(env)
		writerColors := cfg.MakeColors()
		writer := &ansi.Writer{
			TerminalBackground: shell.ConsoleBackgroundColor(env, cfg.TerminalBackground),
			AnsiColors:         writerColors,
		}
		writer.Init(shell.GENERIC)
		eng := &engine.Engine{
			Config: cfg,
			Env:    env,
			Writer: writer,
			Plain:  plain,
		}
		fmt.Print(eng.PrintDebug(startTime, cliVersion))
	},
}

func init() { //nolint:gochecknoinits
	debugCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	debugCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")
	RootCmd.AddCommand(debugCmd)
}
