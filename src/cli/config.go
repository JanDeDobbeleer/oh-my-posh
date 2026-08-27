package cli

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	basedsc "github.com/jandedobbeleer/oh-my-posh/src/dsc"
)

var configCmd = &cmdtree.Command{
	Use:   "config edit",
	Short: "Interact with the config",
	Long: `Interact with the config.

You can export or edit the config (via the editor specified in the environment variable "EDITOR").`,
	ValidArgs: []string{
		"edit",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cmdtree.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		switch args[0] {
		case "edit":
			cache.Init(os.Getenv("POSH_SHELL"))
			if configPath, OK := cache.Session.Get[string](config.SourceKey); OK {
				exitcode = editFileWithEditor(configPath)
				return
			}

			fmt.Println("no config found in session cache")
			exitcode = 666
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	configCmd.AddCommand(basedsc.Command(dsc.ConfigDSC()))
	RootCmd.AddCommand(configCmd)
}
