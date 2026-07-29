package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

// progressWidth is the bar's width in cells, matching what the bubbles bar defaulted to closely
// enough that a download looks the same length as it used to.
const progressWidth = 40

// Progress is a bar that repaints in place, and reports the same percentage to the terminal's own
// taskbar indicator - the OSC 9;4 sequence Windows Terminal and ConEmu read, which oh-my-posh
// already knows how to write (terminal.SetProgress) and which the old bar appended by hand.
type Progress struct {
	writer  io.Writer
	label   string
	percent float64
	mutex   sync.Mutex
}

func NewProgress(writer io.Writer, label string) *Progress {
	return &Progress{writer: writer, label: label}
}

// Set repaints at the given fraction, clamped to 0..1. Callers feed it straight from a byte count
// over a content length, which overshoots when a server lies about the length.
func (p *Progress) Set(fraction float64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.percent = min(max(fraction, 0), 1)

	filled := int(p.percent * progressWidth)

	fmt.Fprintf(p.writer, "%s%s %s%s %3.0f%%%s",
		eraseLine,
		p.label,
		strings.Repeat("█", filled),
		strings.Repeat("░", progressWidth-filled),
		p.percent*100,
		terminal.SetProgress(int(p.percent*100)),
	)
}

// Done clears the bar and tells the terminal the task is over, so a taskbar button stops showing
// progress for something that already finished.
func (p *Progress) Done() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	fmt.Fprintf(p.writer, "%s%s", eraseLine, terminal.SetProgress(100))
}
