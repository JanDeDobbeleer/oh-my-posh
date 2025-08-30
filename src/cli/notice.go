package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/spf13/cobra"
)

// noticeCmd represents the notice command
var noticeCmd = &cobra.Command{
	Use:   "notice",
	Short: "Print the upgrade notice when a new version is available.",
	Long:  "Print the upgrade notice when a new version is available.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		flags := &runtime.Flags{
			SaveCache: true,
		}

		env := &runtime.Terminal{}
		env.Init(flags)
		defer env.Close()

		cfg := config.Load(configFlag, false)

		if notice, hasNotice := cfg.Upgrade.Notice(); hasNotice {
			fmt.Println(notice)
		}
	},
}

func init() {
	RootCmd.AddCommand(noticeCmd)
}
