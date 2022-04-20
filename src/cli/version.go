package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  "Print oh-my-posh version and build information.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cliVersion)
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(versionCmd)
}
