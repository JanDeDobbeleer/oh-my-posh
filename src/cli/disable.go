package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var disableCmd = &cmdtree.Command{
	Use:       fmt.Sprintf(toggleUse, "disable"),
	Short:     "Disable a feature",
	Long:      fmt.Sprintf(toggleLong, "Disable"),
	ValidArgs: toggleArgs,
	Args:      NoArgsOrOneValidArg,
	Run: func(cmd *cmdtree.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		toggleFeature(cmd, args[0], false)
	},
}

func init() {
	RootCmd.AddCommand(disableCmd)
}
