package cli

import (
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Set up the prompt for your shell (deprecated)",
	Long:  `Set up the prompt for your shell. (deprecated)`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help()
	},
}

func init() {
	// legacy support
	promptCmd.AddCommand(initCmd)
	promptCmd.AddCommand(debugCmd)
	promptCmd.AddCommand(printCmd)
	RootCmd.AddCommand(promptCmd)
}
