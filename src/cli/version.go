package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  "Print the version number of oh-my-posh.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cliVersion)
	},
}

func init() { //nolint:gochecknoinits
	RootCmd.AddCommand(versionCmd)
}
