package cli

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth [ytmda]",
	Short: "Authenticate against a service",
	Long: `Authenticate against a service.

Available services:

- ytmda: YouTube Music Desktop App (YTMDA) API`,
	ValidArgs: []string{
		"ytmda",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		flags := &runtime.Flags{
			Shell:     shellName,
			SaveCache: true,
		}

		env := &runtime.Terminal{}
		env.Init(flags)
		defer env.Close()

		switch args[0] {
		case "ytmda":
			authenticator := auth.NewYtmda(env)
			if err := auth.Run(authenticator); err != nil {
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
