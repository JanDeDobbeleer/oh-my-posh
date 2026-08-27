package cli

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

var noticeCmd = &cmdtree.Command{
	Use:   "notice",
	Short: "Print the upgrade notice when a new version is available.",
	Long:  "Print the upgrade notice when a new version is available.",
	Args:  cmdtree.NoArgs,
	Run: func(_ *cmdtree.Command, _ []string) {
		env := &runtime.Terminal{}
		env.Init(&runtime.Flags{})

		cache.Init(os.Getenv("POSH_SHELL"), cache.Persist)
		defer cache.Close()

		// Skip if we already checked within the configured interval
		if _, ok := cache.Device.Get[string](upgrade.CACHEKEY); ok {
			return
		}

		cfg := config.Get(configFlag, false)

		defer func() {
			// Set the cache key after the notice check to prevent redundant checks
			cache.Device.Set(upgrade.CACHEKEY, "true", cfg.Upgrade.Interval)
		}()

		if notice, hasNotice := cfg.Upgrade.Notice(); hasNotice {
			fmt.Println(notice)
		}
	},
}

func init() {
	RootCmd.AddCommand(noticeCmd)
}
