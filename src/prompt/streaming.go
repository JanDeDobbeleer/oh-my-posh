package prompt

import (
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

// TransientMarker prefixes a streamed record that contains the transient prompt
// rather than a primary prompt update. Shells cache such a record so rendering
// the transient prompt on Enter needs no additional CLI call.
const TransientMarker = "\x1e"

// StreamPrimary returns a channel that yields prompt updates as segments complete.
//
// The engine + terminal package globals are not thread-safe, so at most one
// StreamPrimary producer goroutine may be rendering at any given time. Callers
// that need to interrupt an in-flight cycle (e.g. a long-lived server handling
// a new render request before the previous one finished) must call Abort and
// wait for it to return before starting a new cycle - Abort blocks until the
// producer goroutine has fully exited.
func (e *Engine) StreamPrimary() <-chan string {
	// Initialize streaming infrastructure BEFORE launching goroutine
	// This ensures the channel exists when segments start timing out
	e.streamingResults = make(chan *config.Segment, 100)
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

		if e.countPendingSegments() == 0 {
			// No segment is executing in the background, so nothing can send
			// on streamingResults after this point - safe to close.
			close(e.streamingResults)
			return
		}

		// Listen for segment completions. A segment that timed out keeps
		// executing in the background (trackPendingSegment) even after this
		// loop returns; notifySegmentCompletion sends via select/default so
		// it never blocks such a goroutine, but that also means
		// streamingResults must NOT be closed here on the abort path - a
		// stray late send on a closed channel would panic. Only close it once
		// every pending segment has actually reported in (countPendingSegments
		// reaches 0), which is the one path guaranteed to have no further
		// senders. On abort, leave the channel open and let it be garbage
		// collected once the last stray sender (and this Engine) is dropped.
		for {
			select {
			case <-e.abort:
				return
			case _, ok := <-e.streamingResults:
				if !ok {
					return
				}

				if aborted() {
					continue
				}

				if !sendRecord(e.renderFromBlocks()) {
					return
				}

				if e.countPendingSegments() == 0 {
					// refresh the transient prompt now the context is fully resolved
					sendTransient()
					close(e.streamingResults)
					return
				}
			}
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

// countPendingSegments counts how many segments are marked as pending
func (e *Engine) countPendingSegments() int {
	count := 0
	e.pendingSegments.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// renderFromBlocks re-renders the complete prompt using stored block data
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

// trackPendingSegment continues execution for a timed-out segment in the background
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

// notifySegmentCompletion sends completed segment to the streaming results channel
func (e *Engine) notifySegmentCompletion(segment *config.Segment) {
	if e.streamingResults == nil {
		return
	}

	if _, ok := e.pendingSegments.LoadAndDelete(segment.Name()); ok {
		select {
		case e.streamingResults <- segment:
			// Successfully notified consumer
		default:
			// Consumer not ready or already exited
			// This can happen if segment completes after consumer finishes
		}
	}
}
