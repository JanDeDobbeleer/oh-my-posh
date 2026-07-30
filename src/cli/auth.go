package cli

import (
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth/tui"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var authCmd = &cmdtree.Command{
	Use:   "auth [service]",
	Short: "Authenticate against a service",
	Long: `Authenticate against a service.

Available services:

- copilot: GitHub Copilot API
- ytmda: YouTube Music Desktop App (YTMDA) API`,
	ValidArgs: []string{
		copilotServiceName,
		"ytmda",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cmdtree.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		flags := &runtime.Flags{
			Shell: os.Getenv("POSH_SHELL"),
		}

		env := &runtime.Terminal{}
		env.Init(flags)

		cache.Init(env.Shell(), cache.Persist)

		defer func() {
			cache.Close()
		}()

		switch args[0] {
		case copilotServiceName:
			authenticator := tui.NewCopilot(env)
			if err := tui.Run(authenticator); err != nil {
				log.Error(err)
				exitcode = 70
			}
		case "ytmda":
			authenticator := tui.NewYtmda(env)
			if err := tui.Run(authenticator); err != nil {
				log.Error(err)
				exitcode = 70
			}
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(authCmd)
}
