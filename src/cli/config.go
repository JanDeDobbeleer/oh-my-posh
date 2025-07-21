package cli

import (
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config edit",
	Short: "Interact with the config",
	Long: `Interact with the config.

You can export, migrate or edit the config (via the editor specified in the environment variable "EDITOR").`,
	ValidArgs: []string{
		"edit",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		switch args[0] {
		case "edit":
			exitcode = editFileWithEditor(os.Getenv("POSH_THEME"))
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	configCmd.AddCommand(dsc.Command(config.DSC()))
	RootCmd.AddCommand(configCmd)
}
