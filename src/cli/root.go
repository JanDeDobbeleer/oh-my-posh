package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version number of oh-my-posh
var (
	config         string
	displayVersion bool
	cliVersion     string
)

var rootCmd = &cobra.Command{
	Use:   "oh-my-posh",
	Short: "oh-my-posh is a tool to render your prompt",
	Long: `oh-my-posh is a cross platform tool to render your prompt.
It can use the same configuration everywhere to offer a consistent
experience, regardless of where you are. For a detailed guide
on getting started, have a look at the docs at https://ohmyposh.dev`,
	Run: func(cmd *cobra.Command, args []string) {
		if initialize {
			runInit(shellName)
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Backwards compatibility
var (
	shellName  string
	initialize bool
)

func init() { // nolint:gochecknoinits
	rootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "config (required)")
	rootCmd.Flags().BoolVarP(&initialize, "init", "i", false, "init (deprecated)")
	rootCmd.Flags().BoolVar(&displayVersion, "version", false, "version")
	rootCmd.Flags().StringVarP(&shellName, "shell", "s", "", "shell (deprecated)")
}
