package prompt

import (
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

// TransientMarker prefixes a streamed record that contains the transient prompt
// rather than a primary prompt update. Shells cache such a record so rendering
// the transient prompt on Enter needs no additional CLI call.
const TransientMarker = "\x1e"

// The engine + terminal package globals are not thread-safe, so at most one
// StreamPrimary producer goroutine may be rendering at any given time. Callers
// that need to interrupt an in-flight cycle (e.g. a long-lived server handling
// a new render request before the previous one finished) must call Abort and
// wait for it to return before starting a new cycle - Abort blocks until the
// producer goroutine has fully exited.
func (e *Engine) StreamPrimary() <-chan string {
	// Initialize streaming infrastructure BEFORE launching goroutine
	// This ensures the channel exists when segments start timing out.
	// Every segment notifies at most once, so a buffer sized to the segment
	// count guarantees the non-blocking send in notifySegmentCompletion
	// never drops a completion while the producer is still consuming.
	e.streamingResults = make(chan *config.Segment, max(e.segmentCount(), 1))
	e.allBlocks = e.Config.Blocks
	e.abort = make(chan struct{})
	e.done = make(chan struct{})

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
		case <-e.abort:
			return false
		}
	}

	// aborted reports whether the cycle has been asked to stop. Once true,
	// the producer must not touch the engine's prompt builder or the
	// terminal package globals again - those are shared with the next cycle.
	aborted := func() bool {
		select {
		case <-e.abort:
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
		defer close(e.done)
		defer close(out)
		// Registered last so it runs first during unwinding: a panic in
		// segment/render code then costs this one cycle instead of the whole
		// process - which matters for the long-lived serve daemon. The closes
		// above still run afterwards, so Abort() and the record consumer both
		// observe a normally-ended cycle.
		defer func() {
			_ = recover()
		}()

		if aborted() {
			return
		}

		// Render and send initial prompt with pending segments
		if !sendRecord(e.Primary()) {
			return
		}

		sendTransient()

		// Re-render on every completion until nothing is pending AND no
		// completion is queued. Checking the pending count alone is a race:
		// a segment that finishes right after a render read its Pending flag
		// (and printed the placeholder) deregisters itself before this check
		// runs, so the count reads 0 while its notification sits unread in
		// the buffer - the prompt then stays on "..." forever. The queued
		// notification is what forces the render that picks up the result.
		//
		// streamingResults is never closed: a segment that timed out keeps
		// executing in the background (trackPendingSegment) and notifies
		// whenever it finishes, possibly after this goroutine returned - a
		// send on a closed channel would panic in a goroutine with no
		// recover, which takes down the serve daemon. notifySegmentCompletion
		// sends via select/default so a late sender never blocks either; the
		// channel is garbage collected with the Engine.
		refreshed := false

		for e.countPendingSegments() > 0 || len(e.streamingResults) > 0 {
			var segment *config.Segment

			select {
			case <-e.abort:
				return
			case segment = <-e.streamingResults:
			}

			if aborted() {
				return
			}

			// Only the consumer deregisters a timed-out segment: it stays
			// registered from timeout until its completion has been taken off
			// the channel, so the loop condition above can never observe
			// "nothing registered, nothing queued" while a render is still owed.
			e.pendingSegments.Delete(segment.Name())

			if !sendRecord(e.renderFromBlocks()) {
				return
			}

			refreshed = true
		}

		if refreshed {
			// refresh the transient prompt now the context is fully resolved
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
// per-segment timeout (see trackPendingSegment) - those belong to this
// Engine instance only and are expected to be abandoned along with it; they
// will report to a now-unread streamingResults channel (via
// notifySegmentCompletion's non-blocking send) until they finish on their own
// and get garbage collected with this Engine.
func (e *Engine) Abort() {
	if e.abort == nil {
		return
	}

	select {
	case <-e.abort:
		// already aborted
	default:
		close(e.abort)
	}

	if e.done != nil {
		<-e.done
	}
}

func (e *Engine) countPendingSegments() int {
	count := 0
	e.pendingSegments.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

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

func (e *Engine) trackPendingSegment(segment *config.Segment, done chan bool) {
	if e.streamingResults == nil {
		return
	}

	// Segment is already pre-registered in pendingSegments map
	go func() {
		<-done
		segment.Pending = false
		e.notifySegmentCompletion(segment)
	}()
}

// notifySegmentCompletion queues a re-render for a segment that finished
// after timing out. It deliberately leaves the segment registered: the
// producer deregisters it when it consumes the notification, so there is
// no moment at which the segment is neither registered nor queued.
func (e *Engine) notifySegmentCompletion(segment *config.Segment) {
	if e.streamingResults == nil {
		return
	}

	if _, ok := e.pendingSegments.Load(segment.Name()); !ok {
		return
	}

	select {
	case e.streamingResults <- segment:
	default:
		// The buffer holds one slot per segment, so this only triggers for
		// an engine whose cycle already ended and will never consume again.
	}
}

func (e *Engine) segmentCount() int {
	count := 0
	for _, block := range e.Config.Blocks {
		count += len(block.Segments)
	}

	return count
}
