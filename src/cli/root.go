package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configFlag string
	shellName  string
	silent     bool
)

var RootCmd = &cobra.Command{
	Use:   "oh-my-posh",
	Short: "oh-my-posh is a tool to render your prompt",
	Long: `oh-my-posh is a cross platform tool to render your prompt.
It can use the same configuration everywhere to offer a consistent
experience, regardless of where you are. For a detailed guide
on getting started, have a look at the docs at https://ohmyposh.dev`,
	Run: func(cmd *cobra.Command, _ []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// software error
		os.Exit(70)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "config file path")
	RootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "do not print anything")
}
