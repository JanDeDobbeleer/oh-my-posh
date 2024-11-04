package cli

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/spf13/cobra"
)

var (
	toggleUse  = "%s [%s]"
	toggleLong = `%s a feature

This command is used to %s one of the following features:

- upgradenotice`
	toggleArgs = []string{
		config.UPGRADENOTICE,
		config.AUTOUPGRADE,
	}
)

// getCmd represents the get command
var enableCmd = &cobra.Command{
	Use:       fmt.Sprintf(toggleUse, "enable", strings.Join(toggleArgs, "|")),
	Short:     "Enable a feature",
	Long:      fmt.Sprintf(toggleLong, "Enable", "enable"),
	ValidArgs: toggleArgs,
	Args:      NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		toggleFeature(cmd, args[0], true)
	},
}

func init() {
	RootCmd.AddCommand(enableCmd)
}

func toggleFeature(cmd *cobra.Command, feature string, enable bool) {
	env := &runtime.Terminal{
		CmdFlags: &runtime.Flags{
			Shell:     shellName,
			SaveCache: true,
		},
	}

	env.Init()
	defer env.Close()

	if len(feature) == 0 {
		_ = cmd.Help()
		return
	}

	if enable {
		env.Cache().Set(feature, "true", cache.INFINITE)
		return
	}

	env.Cache().Delete(feature)
}
