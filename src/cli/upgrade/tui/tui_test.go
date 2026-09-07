package tui

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cli/ui"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/upgrade"

	"github.com/stretchr/testify/assert"
)

// lockedBuffer serializes the spinner goroutine's writes with the test's reads - a plain
// bytes.Buffer races the moment a running spinner repaints while the test inspects the output.
type lockedBuffer struct {
	bytes.Buffer
	mutex sync.Mutex
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.Buffer.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.Buffer.String()
}

func (b *lockedBuffer) Reset() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.Buffer.Reset()
}

func newReporter(out *lockedBuffer) *reporter {
	return &reporter{
		cfg:    &upgrade.Config{},
		writer: out,
		status: ui.NewStatus(out),
		bar:    ui.NewProgress(out, barLabel),
	}
}

// TestReporterHandsTheLineToTheBar pins the fix for the flicker between the status text and the
// progress bar: while the download runs, the spinner is stopped and only the bar may repaint the
// line, so without bar activity nothing gets written at all.
func TestReporterHandsTheLineToTheBar(t *testing.T) {
	var out lockedBuffer

	r := newReporter(&out)
	r.status.Start("Validating current installation")

	r.stage(upgrade.StageDownloading)

	out.Reset()
	r.progress(0.4)
	painted := out.String()

	// Outlives several spinner ticks: a running spinner would have repainted by now.
	time.Sleep(250 * time.Millisecond)

	assert.Contains(t, painted, "40%")
	assert.Equal(t, painted, out.String(), "the spinner must stay silent while the bar owns the line")
}

// TestReporterResumesTheSpinnerAfterDownloading covers the end of the download: the bar is done
// and the status line spins again for the stages that follow.
func TestReporterResumesTheSpinnerAfterDownloading(t *testing.T) {
	var out lockedBuffer

	r := newReporter(&out)
	r.status.Start("Validating current installation")
	defer r.status.Stop("")

	r.stage(upgrade.StageDownloading)
	r.stage(upgrade.StageInstalling)

	out.Reset()

	// A running spinner repaints every 100ms; poll rather than sleep so a loaded runner
	// decides how long that takes.
	assert.Eventually(t, func() bool {
		return out.String() != ""
	}, time.Second, 10*time.Millisecond, "the spinner must repaint again once the bar is done")
}

// TestReporterFailWhileDownloading pins that an error mid-download is still shown: the status
// line is stopped then, so stopping it with a final message would print nothing.
func TestReporterFailWhileDownloading(t *testing.T) {
	var out lockedBuffer

	r := newReporter(&out)
	r.status.Start("Validating current installation")

	r.stage(upgrade.StageDownloading)
	r.progress(0.4)
	r.fail(errors.New("boom"))

	assert.Contains(t, out.String(), "upgrade failed: boom")
}
