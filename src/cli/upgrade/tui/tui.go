// Package tui draws what oh-my-posh shows while it upgrades itself: a status line naming the
// stage, and a progress bar while the download runs.
//
// It lives one level below cli/upgrade because cli/upgrade holds the plain Config, CDN and Source
// types that the config and segments packages need, and that import graph also has to compile for
// wasm - a target with no terminal to draw to. Keeping the drawing here means importing
// cli/upgrade for its types never drags a terminal UI along; this package imports cli/upgrade,
// never the other way around.
package tui

import (
	"fmt"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/ui"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func stageMessage(cfg *upgrade.Config, stage upgrade.Stage) string {
	switch stage {
	case upgrade.StageValidating:
		return "Validating current installation"
	case upgrade.StageDownloading:
		return fmt.Sprintf("Downloading %s from %s", cfg.Latest, cfg.Source.String())
	case upgrade.StageVerifying:
		return "Verifying download"
	case upgrade.StageInstalling:
		return "Installing"
	default:
		return "Upgrading"
	}
}

func Run(cfg *upgrade.Config) error {
	status := ui.NewStatus(os.Stdout)
	bar := ui.NewProgress(os.Stdout, "  Downloading")

	// cli/upgrade reports through plain callbacks precisely so it never has to know what is
	// drawing - see its own report.go. This subscribes for the duration of the run and hands them
	// back afterwards, so nothing keeps writing to a terminal after the command returns.
	downloading := false

	upgrade.SetStageReporter(func(stage upgrade.Stage) {
		// The bar and the status line both own the same line, so only one may paint at a time.
		// Downloading is the only stage with a bar, and reaching any other stage ends it.
		if downloading && stage != upgrade.StageDownloading {
			downloading = false

			bar.Done()
		}

		if stage == upgrade.StageDownloading {
			downloading = true

			status.Set(stageMessage(cfg, stage))

			return
		}

		status.Set(stageMessage(cfg, stage))
	})

	upgrade.SetProgressReporter(func(percent float64) {
		if downloading {
			bar.Set(percent)
		}
	})

	defer upgrade.SetStageReporter(nil)
	defer upgrade.SetProgressReporter(nil)

	status.Start(stageMessage(cfg, upgrade.StageValidating))

	if err := upgrade.Install(cfg); err != nil {
		log.Debug("failed to install")
		status.Stop(fmt.Sprintf(" ❌ upgrade failed: %v", err))

		return err
	}

	current := fmt.Sprintf("v%s", build.Version)
	message := fmt.Sprintf("🚀 Upgraded from %s to %s", current, cfg.Latest)

	if current != cfg.Latest {
		log.Debug("new version installed, user needs to restart shell")
		message += ", restart your shell to take full advantage of the new functionality"
	}

	status.Stop(message)

	return nil
}
