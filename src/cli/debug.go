package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/spf13/cobra"
)

// debugCmd represents the prompt command
var debugCmd = createDebugCmd()

func init() {
	RootCmd.AddCommand(debugCmd)
}

func createDebugCmd() *cobra.Command {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Print the prompt in debug mode",
		Long:  "Print the prompt in debug mode.",
		Run: func(_ *cobra.Command, _ []string) {
			startTime := time.Now()

			log.Enable()
			log.Debug("debug mode enabled")

			sh := os.Getenv("POSH_SHELL")

			configFile := config.Path(configFlag)
			cfg := config.Load(configFile, sh, false)

			flags := &runtime.Flags{
				Config: configFile,
				Debug:  true,
				PWD:    pwd,
				Shell:  sh,
				Plain:  plain,
			}

			env := &runtime.Terminal{}
			env.Init(flags)

			template.Init(env, cfg.Var)

			defer func() {
				template.SaveCache()
				env.Close()
			}()

			terminal.Init(shell.GENERIC)
			terminal.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
			terminal.Colors = cfg.MakeColors(env)
			terminal.Plain = plain

			eng := &prompt.Engine{
				Config: cfg,
				Env:    env,
				Plain:  plain,
			}

			fmt.Print(eng.PrintDebug(startTime, build.Version))
		},
	}

	debugCmd.Flags().StringVar(&pwd, "pwd", "", "current working directory")
	debugCmd.Flags().BoolVarP(&plain, "plain", "p", false, "plain text output (no ANSI)")

	// Deprecated flags, should be kept to avoid breaking CLI integration.
	debugCmd.Flags().StringVar(&shellName, "shell", "", "the shell to print for")

	// Hide flags that are deprecated or for internal use only.
	_ = debugCmd.Flags().MarkHidden("shell")

	return debugCmd
}
