package cli

import (
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

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

		env := &runtime.Terminal{
			CmdFlags: &runtime.Flags{
				Config: configFlag,
				Debug:  true,
				PWD:    pwd,
				Shell:  shellName,
				Plain:  plain,
			},
		}

		env.Init()
		defer env.Close()

		cfg := config.Load(env)

		// add variables to the environment
		env.Var = cfg.Var

		terminal.Init(shell.GENERIC)
		terminal.BackgroundColor = shell.ConsoleBackgroundColor(env, cfg.TerminalBackground)
		terminal.Colors = cfg.MakeColors()
		terminal.Plain = plain

		eng := &prompt.Engine{
			Config: cfg,
			Env:    env,
			Plain:  plain,
		}

		fmt.Print(eng.PrintDebug(startTime, build.Version))
	},
}

func init() {
	debugCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	debugCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")
	debugCmd.Flags().BoolVarP(&plain, "plain", "p", false, "plain text output (no ANSI)")
	RootCmd.AddCommand(debugCmd)
}
