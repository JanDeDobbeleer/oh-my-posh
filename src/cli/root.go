package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/spf13/cobra"
)

var (
	configFlag     string
	displayVersion bool
)

var RootCmd = &cobra.Command{
	Use:   "oh-my-posh",
	Short: "oh-my-posh is a tool to render your prompt",
	Long:  string(*GetLongDescription()),
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
func GetLongDescription() *[]byte {
	root, _ := os.Getwd()
	data, err := os.ReadFile(root + "/cli/long.posh")
	if err != nil {
		panic(err)
	}
	if len(data) > 0 {
		return &data
	}
	return nil
}

// Backwards compatibility
var (
	shellName  string
	initialize bool
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "config file path")
	RootCmd.Flags().BoolVarP(&initialize, "init", "i", false, "init (deprecated)")
	RootCmd.Flags().BoolVar(&displayVersion, "version", false, "version")
	RootCmd.Flags().StringVarP(&shellName, "shell", "s", "", "shell (deprecated)")
}
