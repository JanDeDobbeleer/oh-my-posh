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

// Reader reports how much of a download has arrived, by counting bytes as they are read.
//
// It takes a callback rather than the Progress it drives: cli/font counts bytes deep inside its
// download path, which has no business knowing whether anything is being displayed at all. The
// same reasoning already shaped cli/upgrade's reporter callbacks.
type Reader struct {
	io.Reader

	report  func(fraction float64)
	total   int64
	current int64
}

func NewReader(reader io.Reader, total int64, report func(fraction float64)) *Reader {
	return &Reader{Reader: reader, total: total, report: report}
}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)

	r.current += int64(n)

	// A server that sends no Content-Length reports -1, and dividing by it walks the bar
	// backwards. Nothing to report until the size is known.
	if r.report != nil && r.total > 0 {
		r.report(float64(r.current) / float64(r.total))
	}

	return n, err
}
