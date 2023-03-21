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
	Run: func(cmd *cobra.Command, args []string) {
		env := &platform.Shell{
			CmdFlags: &platform.Flags{
				Version: cliVersion,
			},
		}
		env.Init()
		defer env.Close()

		if notice, hasNotice := upgrade.Notice(env); hasNotice {
			fmt.Println(notice)
		}
	},
}

func init() { //nolint:gochecknoinits
	RootCmd.AddCommand(noticeCmd)
}
