package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	config         string
	displayVersion bool

	// Version number of oh-my-posh
	cliVersion string
)

var RootCmd = &cobra.Command{
	Use:   "oh-my-posh",
	Short: "oh-my-posh is a tool to render your prompt",
	Long: `oh-my-posh is a cross platform tool to render your prompt.
It can use the same configuration everywhere to offer a consistent
experience, regardless of where you are. For a detailed guide
on getting started, have a look at the docs at https://ohmyposh.dev`,
	Run: func(cmd *cobra.Command, args []string) {
		if initialize {
			runInit(strings.ToLower(shellName))
			return
		}
		if displayVersion {
			fmt.Println(cliVersion)
			return
		}
		_ = cmd.Help()
	},
}

func Execute(version string) {
	cliVersion = version
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Backwards compatibility
var (
	shellName  string
	initialize bool
)

func init() { //nolint:gochecknoinits
	RootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "config file path")
	RootCmd.Flags().BoolVarP(&initialize, "init", "i", false, "init (deprecated)")
	RootCmd.Flags().BoolVar(&displayVersion, "version", false, "version")
	RootCmd.Flags().StringVarP(&shellName, "shell", "s", "", "shell (deprecated)")
}
