package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config edit",
	Short: "Interact with the config",
	Long: `Interact with the config.

You can export, migrate or edit the config (via the editor specified in the environment variable "EDITOR").`,
	ValidArgs: []string{
		"export",
		"migrate",
		"edit",
		"get",
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
		case "get":
			// only here for backwards compatibility
			fmt.Print(time.Now().UnixNano() / 1000000)
		default:
			_ = cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
}
