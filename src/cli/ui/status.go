// Package ui draws the handful of interactive things oh-my-posh's commands need: a status line
// that spins while something is happening, a progress bar, and a picker.
//
// It exists to keep a TUI framework out of a prompt engine. bubbletea and its dependencies were
// linked into every build, and Go runs a linked package's init before main whether or not the
// subcommand that needs it was invoked - so someone typing at a prompt paid for the font
// installer's UI. Measured: 1.84MB of binary and 44ms of every single prompt render, most of it
// go-runewidth building a 2.2MB lookup table that nothing in the prompt path reads.
//
// Everything here writes in cooked mode. A status line and a progress bar are just a carriage
// return and an erase-line away from repainting themselves, which means no raw mode, no termios,
// no Windows console mode juggling, and nothing to restore if the process dies badly - the class
// of bug that makes a terminal unusable after a crash.
package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// spinnerFrames is the Braille spinner, which reads as motion at any width and needs no font
// beyond what a terminal already draws for box characters.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const (
	spinnerInterval = 100 * time.Millisecond

	// eraseLine clears the whole line and parks the cursor at its start, which is what makes a
	// repaint leave nothing of the longer message it replaced.
	eraseLine = "\r\x1b[2K"

	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
)

// Status is a single line that repaints in place: a spinner, then whatever the caller last said
// is happening. Safe to drive from several goroutines, which every caller does - the work being
// reported on runs somewhere other than the ticker.
type Status struct {
	writer  io.Writer
	done    chan struct{}
	stopped chan struct{}
	message string
	frame   int
	mutex   sync.Mutex
}

func NewStatus(writer io.Writer) *Status {
	return &Status{writer: writer}
}

// Start begins repainting. Stop must be called, and a defer is the only sane way to do it: the
// cursor is hidden until then.
func (s *Status) Start(message string) {
	s.mutex.Lock()
	s.message = message
	s.mutex.Unlock()

	s.done = make(chan struct{})
	s.stopped = make(chan struct{})

	fmt.Fprint(s.writer, hideCursor)

	go s.loop()
}

func (s *Status) loop() {
	defer close(s.stopped)

	ticker := time.NewTicker(spinnerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.paint()
		}
	}
}

func (s *Status) paint() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.frame = (s.frame + 1) % len(spinnerFrames)

	fmt.Fprintf(s.writer, "%s%s %s", eraseLine, spinnerFrames[s.frame], s.message)
}

// Set changes what the line says without interrupting the spin.
func (s *Status) Set(message string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.message = message
}

// Stop ends the repaint and leaves final on its own line, or clears the line entirely when final
// is empty. It waits for the ticker to exit first, so nothing can repaint over what it wrote.
func (s *Status) Stop(final string) {
	if s.done == nil {
		return
	}

	close(s.done)
	<-s.stopped

	s.done = nil

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if final == "" {
		fmt.Fprint(s.writer, eraseLine+showCursor)
		return
	}

	fmt.Fprintf(s.writer, "%s%s%s\n", eraseLine, final, showCursor)
}

// Print writes a line above the status line without disturbing it: erase, write, and let the next
// tick repaint the spinner underneath.
func (s *Status) Print(line string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	fmt.Fprintf(s.writer, "%s%s\n", eraseLine, strings.TrimRight(line, "\n"))
}
