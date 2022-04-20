package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [export|migrate|edit]",
	Short: "Interact with the configuration",
	Long: `Interact with the configuration
It allows to export, migrate or edit the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		switch args[0] {
		case "edit":
			editFileWithEditor(os.Getenv("POSH_THEME"))
		case "get":
			// only here for backwards compatibility
			fmt.Print(time.Now().UnixNano() / 1000000)
		default:
			_ = cmd.Help()
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(configCmd)
}
