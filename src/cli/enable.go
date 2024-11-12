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
	toggleHelpText = `%s one of the following features:
`
	toggleArgs = []string{
		config.UPGRADENOTICE,
		config.AUTOUPGRADE,
	}
	toggleUse  = fmt.Sprintf("%%s [%s]", strings.Join(toggleArgs, "|"))
	toggleLong = strings.Join(append([]string{toggleHelpText}, toggleArgs...), "\n- ")
)

// getCmd represents the get command
var enableCmd = &cobra.Command{
	Use:       fmt.Sprintf(toggleUse, "enable"),
	Short:     "Enable a feature",
	Long:      fmt.Sprintf(toggleLong, "Enable"),
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
	flags := &runtime.Flags{
		Shell:     shellName,
		SaveCache: true,
	}

	env := &runtime.Terminal{}
	env.Init(flags)
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
