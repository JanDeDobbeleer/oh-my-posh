package cli

import (
	"fmt"
	"os"
	stdruntime "runtime"
	"slices"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
	"github.com/spf13/cobra"
)

var (
	force bool
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade when a new version is available.",
	Long:  "Upgrade when a new version is available.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		var startTime time.Time

		if debug {
			startTime = time.Now()
			log.Enable(plain)
		}

		if upgrade.IsPackagedInstallation() {
			msg := "upgrade is not supported when installed as a MSIX package"
			log.Debug(msg)
			fmt.Printf("\n  ❌ %s\n\n", msg)
			return
		}

		supportedPlatforms := []string{
			runtime.WINDOWS,
			runtime.DARWIN,
			runtime.LINUX,
		}

		if !slices.Contains(supportedPlatforms, stdruntime.GOOS) {
			log.Debug("unsupported platform")
			return
		}

		sh := os.Getenv("POSH_SHELL")

		env := &runtime.Terminal{}
		env.Init(&runtime.Flags{
			Debug: debug,
		})

		cache.Init(sh, cache.Persist)

		if _, OK := cache.Get[string](cache.Device, upgrade.CACHEKEY); OK && !force {
			log.Debug("upgrade check already performed recently, skipping")
			return
		}

		terminal.Init(sh)
		fmt.Print(terminal.StartProgress())

		cfg := config.Get(configFlag, false)

		defer func() {
			fmt.Print(terminal.StopProgress())

			// always reset the cache key so we respect the interval no matter what the outcome
			cache.Set(cache.Device, upgrade.CACHEKEY, "true", cfg.Upgrade.Interval)

			cache.Close()

			if !debug {
				return
			}

			sb := text.NewBuilder()

			sb.WriteString(fmt.Sprintf("%s %s\n", log.Text("Upgrade duration:").Green().Bold().Plain(), time.Since(startTime)))

			sb.WriteString(log.Text("\nLogs:\n\n").Green().Bold().Plain().String())
			sb.WriteString(env.Logs())

			fmt.Println(sb.String())
		}()

		latest, err := cfg.Upgrade.FetchLatest()
		if err != nil {
			log.Debug("failed to get latest version")
			log.Error(err)
			fmt.Printf("\n  ❌ %s\n\n", err)

			exitcode = 1
			return
		}

		log.Debugf("current version: v%s, latest version: v%s", build.Version, latest)

		if force {
			log.Debug("forced upgrade")
			exitcode = executeUpgrade(cfg.Upgrade)
			return
		}

		if upgrade.IsMajorUpgrade(build.Version, latest) {
			log.Debug("major upgrade available")
			message := fmt.Sprintf("\n  🚨 major upgrade available: v%s -> v%s, use oh-my-posh upgrade --force to upgrade\n\n", build.Version, latest)
			fmt.Print(message)
			return
		}

		if build.Version != latest {
			log.Debug("upgrade available")
			exitcode = executeUpgrade(cfg.Upgrade)
			return
		}

		log.Debug("already on the latest version")
	},
}

func executeUpgrade(cfg *upgrade.Config) int {
	err := upgrade.Run(cfg)
	if err == nil {
		return 0
	}

	log.Debug("failed to upgrade")
	log.Error(err)

	return 1
}

func init() {
	upgradeCmd.Flags().BoolVarP(&force, "force", "f", false, "force the upgrade even if the version is up to date")
	upgradeCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	RootCmd.AddCommand(upgradeCmd)
}
