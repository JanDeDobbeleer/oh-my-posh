package prompt

import (
	"runtime/debug"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

// TransientMarker prefixes a streamed record that contains the transient prompt
// rather than a primary prompt update. Shells cache such a record so rendering
// the transient prompt on Enter needs no additional CLI call.
const TransientMarker = "\x1e"

// pendingSegmentLimit bounds how long a timed-out segment may stay pending before the cycle
// abandons it and renders without it. A package variable so tests can shorten it.
var pendingSegmentLimit = 30 * time.Second

type segmentEventKind uint8

const (
	segmentTimedOut segmentEventKind = iota
	segmentCompleted
	segmentAbandoned
)

type segmentEvent struct {
	segment *config.Segment
	kind    segmentEventKind
}

// streamCycle holds the state of one StreamPrimary run. events is the only channel between
// segment goroutines and the producer goroutine. pending and abandoned are owned by the producer
// goroutine: nothing else reads or writes them, so they need no lock.
type streamCycle struct {
	events    chan segmentEvent
	pending   map[*config.Segment]struct{}
	abandoned map[*config.Segment]struct{}
	abort     chan struct{}
	done      chan struct{}
}

// newStreamCycle creates the event queue for one StreamPrimary run. events is sized to
// 2*segmentCount + 1: every segment emits at most two events (timedOut, then exactly one of
// completed/abandoned), so a send can never block and is never dropped. Every send on this
// channel is a plain send, never a select/default, because the buffer is provably big enough.
func newStreamCycle(segmentCount int) *streamCycle {
	return &streamCycle{
		events:    make(chan segmentEvent, 2*segmentCount+1),
		pending:   make(map[*config.Segment]struct{}),
		abandoned: make(map[*config.Segment]struct{}),
		abort:     make(chan struct{}),
		done:      make(chan struct{}),
	}
}

func (c *streamCycle) timedOut(segment *config.Segment) {
	c.events <- segmentEvent{segment: segment, kind: segmentTimedOut}
}

func (c *streamCycle) completed(segment *config.Segment) {
	c.events <- segmentEvent{segment: segment, kind: segmentCompleted}
}

// abandon queues a segmentAbandoned event. Named to avoid colliding with the abandoned field.
func (c *streamCycle) abandon(segment *config.Segment) {
	c.events <- segmentEvent{segment: segment, kind: segmentAbandoned}
}

// apply folds one event into the pending/abandoned sets. Producer goroutine only.
func (c *streamCycle) apply(ev segmentEvent) {
	switch ev.kind {
	case segmentTimedOut:
		if _, ok := c.abandoned[ev.segment]; ok {
			return
		}

		c.pending[ev.segment] = struct{}{}
	case segmentCompleted:
		delete(c.pending, ev.segment)
	case segmentAbandoned:
		delete(c.pending, ev.segment)
		c.abandoned[ev.segment] = struct{}{}
	}
}

// absorb drains every event already queued without blocking. Producer goroutine only.
func (c *streamCycle) absorb() {
	for {
		select {
		case ev := <-c.events:
			c.apply(ev)
		default:
			return
		}
	}
}

func (c *streamCycle) isPending(segment *config.Segment) bool {
	_, ok := c.pending[segment]
	return ok
}

func (c *streamCycle) isAbandoned(segment *config.Segment) bool {
	_, ok := c.abandoned[segment]
	return ok
}

// The engine + terminal package globals are not thread-safe, so at most one
// StreamPrimary producer goroutine may be rendering at any given time. Callers
// that need to interrupt an in-flight cycle (e.g. a long-lived server handling
// a new render request before the previous one finished) must call Abort and
// wait for it to return before starting a new cycle - Abort blocks until the
// producer goroutine has fully exited.
func (e *Engine) StreamPrimary() <-chan string {
	e.stream = newStreamCycle(e.segmentCount())
	out := make(chan string, 10)

	// sendRecord delivers a record unless the cycle gets aborted. A plain
	// channel send could block forever once the buffer fills against a
	// stalled consumer, keeping the producer from ever observing the abort
	// and deadlocking Abort() - which waits for this goroutine to exit.
	// Returns false when the cycle is aborted; the producer must stop then.
	sendRecord := func(record string) bool {
		select {
		case out <- record:
			return true
		case <-e.stream.abort:
			return false
		}
	}

	// aborted reports whether the cycle has been asked to stop. Once true,
	// the producer must not touch the engine's prompt builder or the
	// terminal package globals again - those are shared with the next cycle.
	aborted := func() bool {
		select {
		case <-e.stream.abort:
			return true
		default:
			return false
		}
	}

	// The transient prompt must render in the same goroutine as the primary
	// updates: both write to the engine's prompt builder and the terminal
	// package's global state.
	sendTransient := func() {
		if aborted() {
			return
		}

		// The zsh script caches a streamed transient record as PS1 only and
		// resets RPROMPT (see _omp_zle-line-init in omp.zsh), so the record
		// cannot carry a right-aligned template. Skip it to make the script
		// fall back to the eval path which sets both PS1 and RPROMPT.
		if e.Env.Shell() == shell.ZSH && e.Config.TransientPrompt != nil && len(e.Config.TransientPrompt.RightTemplate) != 0 {
			return
		}

		// The zsh script renders the transient prompt one column narrower to avoid
		// a redundant blank line when a filler is configured and the input is empty
		// (see _omp_zle-line-init in omp.zsh), mirror that for the streamed record.
		if e.Env.Shell() == shell.ZSH {
			e.rectifyTerminalWidth(-1)
			defer e.rectifyTerminalWidth(1)
		}

		sendRecord(TransientMarker + e.ExtraPrompt(Transient))
	}

	go func() {
		defer close(e.stream.done)
		defer close(out)
		// Registered last so it runs first during unwinding: a panic in
		// segment/render code then costs this one cycle instead of the whole
		// process - which matters for the long-lived serve daemon. The closes
		// above still run afterwards, so Abort() and the record consumer both
		// observe a normally-ended cycle. The recover MUST log; a silent
		// recover previously hid the very bug this rewrite fixes.
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			log.Errorf("streaming: render cycle panicked: %v\n%s", r, debug.Stack())
		}()

		if aborted() {
			return
		}

		if !sendRecord(e.Primary()) {
			return
		}

		sendTransient()

		refreshed := false

		for len(e.stream.pending) > 0 {
			select {
			case <-e.stream.abort:
				return
			case ev := <-e.stream.events:
				e.stream.apply(ev)
			}

			// Coalesce whatever else already queued so simultaneous completions cost one render.
			e.stream.absorb()

			if !sendRecord(e.renderFromBlocks()) {
				return
			}

			refreshed = true
		}

		if refreshed {
			sendTransient()
		}
	}()

	return out
}

// Abort signals the active StreamPrimary cycle (if any) to stop rendering and
// blocks until its producer goroutine has fully exited, so the caller can
// safely start a new cycle (on a new Engine) immediately after Abort returns.
// It is safe to call multiple times and safe to call when no cycle is active
// or the cycle has already finished on its own.
//
// Abort does not wait for segments still executing in the background after a
// per-segment timeout (see awaitPendingSegment) - those belong to this Engine
// instance only and are expected to be abandoned along with it.
func (e *Engine) Abort() {
	if e.stream == nil {
		return
	}

	select {
	case <-e.stream.abort:
		// already aborted
	default:
		close(e.stream.abort)
	}

	<-e.stream.done
}

// renderFromBlocks resets the prompt builder state a fresh render pass needs and re-renders
// the primary prompt from the segments' already-executed state (the cache path).
func (e *Engine) renderFromBlocks() string {
	// Reset prompt builder
	e.prompt.Reset()
	e.currentLineLength = 0
	e.activeSegment = nil
	e.previousActiveSegment = nil
	e.rprompt = ""
	e.rpromptLength = 0

	return e.primaryInternal(true)
}

func (e *Engine) segmentPending(segment *config.Segment) bool {
	return e.stream != nil && e.stream.isPending(segment)
}

func (e *Engine) segmentAbandoned(segment *config.Segment) bool {
	return e.stream != nil && e.stream.isAbandoned(segment)
}

func (e *Engine) segmentCount() int {
	count := 0
	for _, block := range e.Config.Blocks {
		count += len(block.Segments)
	}

	return count
}
