package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
)

var (
	toggleHelpText = `%s one of the following features:
`
	toggleArgs = []string{
		config.UPGRADENOTICE,
		config.AUTOUPGRADE,
		config.RELOAD,
	}
	toggleUse  = fmt.Sprintf("%%s [%s]", strings.Join(toggleArgs, "|"))
	toggleLong = strings.Join(append([]string{toggleHelpText}, toggleArgs...), "\n- ")
)

var enableCmd = &cmdtree.Command{
	Use:       fmt.Sprintf(toggleUse, "enable"),
	Short:     "Enable a feature",
	Long:      fmt.Sprintf(toggleLong, "Enable"),
	ValidArgs: toggleArgs,
	Args:      NoArgsOrOneValidArg,
	Run: func(cmd *cmdtree.Command, args []string) {
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

func toggleFeature(cmd *cmdtree.Command, feature string, enable bool) {
	if feature == "" {
		_ = cmd.Help()
		return
	}

	cache.Init(os.Getenv("POSH_SHELL"), cache.Persist)
	cache.Set(cache.Device, feature, enable, cache.INFINITE)
	cache.Close()
}
