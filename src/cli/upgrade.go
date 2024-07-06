package cli

import (
	"fmt"
	"os"
	stdruntime "runtime"
	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"
	"github.com/spf13/cobra"
)

var force bool

// noticeCmd represents the get command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade when a new version is available.",
	Long:  "Upgrade when a new version is available.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		supportedPlatforms := []string{
			runtime.WINDOWS,
			runtime.DARWIN,
			runtime.LINUX,
		}

		if !slices.Contains(supportedPlatforms, stdruntime.GOOS) {
			fmt.Print("\n⚠️ upgrade is not supported on this platform\n\n")
			return
		}

		env := &runtime.Terminal{}
		env.Init()
		defer env.Close()

		terminal.Init(env.Shell())
		fmt.Print(terminal.StartProgress())

		latest, err := upgrade.Latest(env)
		if err != nil {
			fmt.Printf("\n❌ %s\n\n", err)
			terminal.StopProgress()
			os.Exit(1)
			return
		}

		if force {
			executeUpgrade(latest)
			return
		}

		cfg := config.Load(env)

		version := fmt.Sprintf("v%s", build.Version)

		if version == latest {
			if !cfg.DisableNotice {
				fmt.Print("\n✅ no new version available\n\n")
			}

			terminal.StopProgress()
			return
		}

		executeUpgrade(latest)
	},
}

func executeUpgrade(latest string) {
	err := upgrade.Run(latest)
	if err == nil {
		return
	}

	fmt.Printf("\n❌ %s\n\n", err)
	terminal.StopProgress()
	os.Exit(1)
}

func init() {
	upgradeCmd.Flags().BoolVarP(&force, "force", "f", false, "force the upgrade even if the version is up to date")
	RootCmd.AddCommand(upgradeCmd)
}
