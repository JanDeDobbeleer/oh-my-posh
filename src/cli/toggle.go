package cli

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var toggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle a segment on/off",
	Long:  "Toggle a segment on/off on the fly.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		env := &platform.Shell{}
		env.Init()
		defer env.Close()

		cache := env.Cache()
		togglesCache, _ := cache.Get(platform.TOGGLECACHE)
		var toggles []string
		if len(togglesCache) != 0 {
			toggles = strings.Split(togglesCache, ",")
		}
		segment := args[0]

		newToggles := []string{}
		var match bool
		for _, toggle := range toggles {
			if toggle == segment {
				match = true
				continue
			}
			newToggles = append(newToggles, toggle)
		}

		if !match {
			newToggles = append(newToggles, segment)
		}

		cache.Set(platform.TOGGLECACHE, strings.Join(newToggles, ","), 1440)
	},
}

func init() { //nolint:gochecknoinits
	RootCmd.AddCommand(toggleCmd)
}
