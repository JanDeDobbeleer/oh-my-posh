package cli

import (
	"fmt"
	"os"
	stdruntime "runtime"
	"slices"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
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
		var startTime time.Time

		if debug {
			startTime = time.Now()
			log.Enable()
			log.Debug("debug mode enabled")
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
		env.Init(&runtime.Flags{Debug: debug})

		defer func() {
			env.Close()

			if !debug {
				return
			}

			var builder strings.Builder

			builder.WriteString(fmt.Sprintf("%s %s\n", log.Text("Upgrade duration:").Green().Bold().Plain(), time.Since(startTime)))

			builder.WriteString(log.Text("\nLogs:\n\n").Green().Bold().Plain().String())
			builder.WriteString(env.Logs())

			fmt.Println(builder.String())
		}()

		terminal.Init(sh)
		fmt.Print(terminal.StartProgress())

		configFile := config.Path(configFlag)
		cfg := config.Load(configFile, sh, false)
		cfg.Upgrade.Cache = env.Cache()

		latest, err := cfg.Upgrade.Latest()
		if err != nil {
			log.Debug("failed to get latest version")
			log.Error(err)
			fmt.Printf("\n❌ %s\n\n%s", err, terminal.StopProgress())
			os.Exit(1)
			return
		}

		cfg.Upgrade.Version = fmt.Sprintf("v%s", latest)

		if force {
			log.Debug("forced upgrade")
			executeUpgrade(cfg.Upgrade)
			return
		}

		if upgrade.IsMajorUpgrade(build.Version, latest) {
			log.Debug("major upgrade available")
			message := terminal.StopProgress()
			message += fmt.Sprintf("\n🚨 major upgrade available: v%s -> v%s, use oh-my-posh upgrade --force to upgrade\n\n", build.Version, latest)
			fmt.Print(message)
			return
		}

		if build.Version != latest {
			executeUpgrade(cfg.Upgrade)
			return
		}

		fmt.Print(terminal.StopProgress())
	},
}

func executeUpgrade(cfg *upgrade.Config) {
	err := upgrade.Run(cfg)
	fmt.Print(terminal.StopProgress())
	if err == nil {
		return
	}

	log.Debug("failed to upgrade")
	log.Error(err)
	os.Exit(1)
}

func init() {
	upgradeCmd.Flags().BoolVarP(&force, "force", "f", false, "force the upgrade even if the version is up to date")
	upgradeCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	RootCmd.AddCommand(upgradeCmd)
}
