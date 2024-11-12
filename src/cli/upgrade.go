package cli

import (
	"fmt"
	"os"
	stdruntime "runtime"
	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
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
			return
		}

		env := &runtime.Terminal{}
		env.Init(nil)
		defer env.Close()

		terminal.Init(env.Shell())
		fmt.Print(terminal.StartProgress())

		latest, err := upgrade.Latest(env)
		if err != nil {
			fmt.Printf("\n❌ %s\n\n%s", err, terminal.StopProgress())
			os.Exit(1)
			return
		}

		if force {
			executeUpgrade(latest)
			return
		}

		version := fmt.Sprintf("v%s", build.Version)

		if upgrade.IsMajorUpgrade(version, latest) {
			message := terminal.StopProgress()
			message += fmt.Sprintf("\n🚨 major upgrade available: %s -> %s, use oh-my-posh upgrade --force to upgrade\n\n", version, latest)
			fmt.Print(message)
			return
		}

		if version != latest {
			executeUpgrade(latest)
			return
		}

		fmt.Print(terminal.StopProgress())
	},
}

func executeUpgrade(latest string) {
	err := upgrade.Run(latest)
	fmt.Print(terminal.StopProgress())
	if err == nil {
		return
	}

	os.Exit(1)
}

func init() {
	upgradeCmd.Flags().BoolVarP(&force, "force", "f", false, "force the upgrade even if the version is up to date")
	RootCmd.AddCommand(upgradeCmd)
}
