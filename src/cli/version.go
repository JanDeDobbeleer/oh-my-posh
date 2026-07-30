package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var (
	verbose bool
)

var versionCmd = &cmdtree.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  "Print the version number of oh-my-posh.",
	Args:  cmdtree.NoArgs,
	Run: func(_ *cmdtree.Command, _ []string) {
		if !verbose {
			fmt.Println(build.Version)
			return
		}
		fmt.Println("Version: ", build.Version)
		fmt.Println("Date:    ", build.Date)
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "write verbose output")
	RootCmd.AddCommand(versionCmd)
}
