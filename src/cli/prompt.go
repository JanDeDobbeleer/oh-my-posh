package cli

import (
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Set up the prompt for your shell (deprecated)",
	Long: `Set up the prompt for your shell (deprecated)
Allows to initialize one of the supported shells, or to set the prompt manually for a custom shell.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() { // nolint:gochecknoinits
	// legacy support
	promptCmd.AddCommand(initCmd)
	promptCmd.AddCommand(debugCmd)
	promptCmd.AddCommand(printCmd)
	rootCmd.AddCommand(promptCmd)
}
