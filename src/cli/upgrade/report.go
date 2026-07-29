package upgrade

import "io"

// Stage identifies where the upgrade process currently is, so that whoever is
// driving it — the terminal UI in cli/upgrade/tui, specifically — can show
// the user something meaningful. It lives here, rather than in the UI
// package, because the functions that raise it (Install, downloadAndVerify)
// belong to this package: their job is downloading and installing a new
// binary, and reporting progress is a side concern they hand off through
// stageReporter instead of depending on how (or whether) that progress gets
// displayed.
type Stage int

const (
	StageValidating Stage = iota
	StageDownloading
	StageVerifying
	StageInstalling
)

// stageReporter, when set, is notified as the upgrade moves from one Stage to
// the next. It is nil by default, which keeps this package usable (silently,
// without any progress feedback) from contexts that have no terminal to
// report to at all, such as this package compiled for wasm to supply the
// Config/CDN/Source types the config and segments packages need.
var stageReporter func(Stage)

// SetStageReporter registers the callback invoked on every Stage transition
// during Install. Pass nil to stop reporting.
func SetStageReporter(reporter func(Stage)) {
	stageReporter = reporter
}

func setState(stage Stage) {
	if stageReporter == nil {
		return
	}

	stageReporter(stage)
}

// progressReporter, when set, is notified with the fraction (0..1) of the
// current asset downloaded so far. Like stageReporter, it defaults to nil so
// Download works the same whether or not anything is listening.
var progressReporter func(percent float64)

// SetProgressReporter registers the callback invoked as an asset download
// progresses. Pass nil to stop reporting.
func SetProgressReporter(reporter func(percent float64)) {
	progressReporter = reporter
}

// downloadProgressReader wraps a response body so Download can report how
// much of it has been read so far, without this package needing to know
// anything about how that progress is displayed (or whether it is displayed
// at all). This replaces a direct dependency on cli/progress, which pulls in
// bubbletea purely to draw a progress bar — a dependency this package cannot
// afford, since config and segments import it for plain types and that
// import graph must also compile for wasm.
type downloadProgressReader struct {
	reader  io.Reader
	total   int64
	current int64
}

func (r *downloadProgressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.current += int64(n)

	if progressReporter != nil && r.total > 0 {
		progressReporter(float64(r.current) / float64(r.total))
	}

	return n, err
}
