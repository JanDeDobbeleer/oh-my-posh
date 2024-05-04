package cli

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"
	"github.com/spf13/cobra"
)

// noticeCmd represents the get command
var noticeCmd = &cobra.Command{
	Use:   "notice",
	Short: "Print the upgrade notice when a new version is available.",
	Long:  "Print the upgrade notice when a new version is available.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		env := &platform.Shell{
			CmdFlags: &platform.Flags{},
		}
		env.Init()
		defer env.Close()

		if notice, hasNotice := upgrade.Notice(env); hasNotice {
			fmt.Println(notice)
		}
	},
}

func init() {
	RootCmd.AddCommand(noticeCmd)
}
