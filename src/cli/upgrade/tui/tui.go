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
	"io"
	"os"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/ui"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

const barLabel = "  Downloading"

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
	reporter := &reporter{
		cfg:    cfg,
		writer: os.Stdout,
		status: ui.NewStatus(os.Stdout),
		bar:    ui.NewProgress(os.Stdout, barLabel),
	}

	// cli/upgrade reports through plain callbacks precisely so it never has to know what is
	// drawing - see its own report.go. This subscribes for the duration of the run and hands them
	// back afterwards, so nothing keeps writing to a terminal after the command returns.
	upgrade.SetStageReporter(reporter.stage)
	upgrade.SetProgressReporter(reporter.progress)

	defer upgrade.SetStageReporter(nil)
	defer upgrade.SetProgressReporter(nil)

	reporter.status.Start(stageMessage(cfg, upgrade.StageValidating))

	if err := upgrade.Install(cfg); err != nil {
		log.Debug("failed to install")
		reporter.fail(err)

		return err
	}

	current := fmt.Sprintf("v%s", build.Version)
	message := fmt.Sprintf("🚀 Upgraded from %s to %s", current, cfg.Latest)

	if current != cfg.Latest {
		log.Debug("new version installed, user needs to restart shell")
		message += ", restart your shell to take full advantage of the new functionality"
	}

	reporter.status.Stop(message)

	return nil
}

// reporter owns the single line the upgrade draws on. The status spinner and the progress bar
// both repaint that line, so only one may paint at a time: the spinner runs for every stage
// except the download, which hands the line to the bar. Painting both at once is what made the
// bar and the status text flicker over each other.
type reporter struct {
	status      *ui.Status
	bar         *ui.Progress
	writer      io.Writer
	cfg         *upgrade.Config
	downloading bool
}

func (r *reporter) stage(stage upgrade.Stage) {
	if stage == upgrade.StageDownloading {
		r.downloading = true
		r.status.Stop("")

		return
	}

	if r.downloading {
		r.downloading = false
		r.bar.Done()
		r.status.Start(stageMessage(r.cfg, stage))

		return
	}

	r.status.Set(stageMessage(r.cfg, stage))
}

func (r *reporter) progress(fraction float64) {
	if !r.downloading {
		return
	}

	r.bar.Set(fraction)
}

func (r *reporter) fail(err error) {
	message := fmt.Sprintf(" ❌ upgrade failed: %v", err)

	// Mid-download the status line is stopped, where Stop is a no-op that would swallow the
	// message - so the bar is cleared and the failure printed plainly instead.
	if r.downloading {
		r.bar.Done()
		fmt.Fprintln(r.writer, message)

		return
	}

	r.status.Stop(message)
}
