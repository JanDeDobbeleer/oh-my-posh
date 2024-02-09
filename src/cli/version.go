package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  "Print the version number of oh-my-posh.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		if !verbose {
			fmt.Println(build.Version)
			return
		}
		fmt.Println("Version: ", build.Version)
		fmt.Println("Date:    ", build.Date)
	},
}

func init() { //nolint:gochecknoinits
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "write verbose output")
	RootCmd.AddCommand(versionCmd)
}
