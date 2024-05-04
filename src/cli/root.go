package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/spf13/cobra"
)

var (
	config         string
	displayVersion bool
)

var RootCmd = &cobra.Command{
	Use:   "oh-my-posh",
	Short: "oh-my-posh is a tool to render your prompt",
	Long: `oh-my-posh is a cross platform tool to render your prompt.
It can use the same configuration everywhere to offer a consistent
experience, regardless of where you are. For a detailed guide
on getting started, have a look at the docs at https://ohmyposh.dev`,
	Run: func(cmd *cobra.Command, _ []string) {
		if initialize {
			runInit(strings.ToLower(shellName))
			return
		}
		if displayVersion {
			fmt.Println(build.Version)
			return
		}
		_ = cmd.Help()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// software error
		os.Exit(70)
	}
}

// Backwards compatibility
var (
	shellName  string
	initialize bool
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "config file path")
	RootCmd.Flags().BoolVarP(&initialize, "init", "i", false, "init (deprecated)")
	RootCmd.Flags().BoolVar(&displayVersion, "version", false, "version")
	RootCmd.Flags().StringVarP(&shellName, "shell", "s", "", "shell (deprecated)")
}
