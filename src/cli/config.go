package cli

import (
	"fmt"
	"os"

	configDSC "github.com/jandedobbeleer/oh-my-posh/src/config/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [export|migrate|edit|get|set|test|schema|export]",
	Short: "Interact with the config",
	Long: `Interact with the config.

You can export, migrate or edit the config (via the editor specified in the environment variable "EDITOR").`,
	ValidArgs: []string{
		"export",
		"migrate",
		"edit",
		"get",
		"set",
		"test",
		"schema",
		"export",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		if args[0] == "edit" {
			exitcode = editFileWithEditor(os.Getenv("POSH_THEME"))
			return
		}

		err := dsc.Run[*configDSC.State](args[0], state)
		if err == nil {
			return
		}

		exitcode = 1
		fmt.Println(err.Error())
	},
}

func init() {
	configCmd.Flags().StringVar(&state, "state", "", "State configuration to set")
	RootCmd.AddCommand(configCmd)
}
