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
func (e *Engine) StreamPrimary() <-chan string {
	// Initialize streaming infrastructure BEFORE launching goroutine
	// This ensures the channel exists when segments start timing out
	e.streamingResults = make(chan *config.Segment, 100)
	e.allBlocks = e.Config.Blocks

	out := make(chan string, 10)

	// The transient prompt must render in the same goroutine as the primary
	// updates: both write to the engine's prompt builder and the terminal
	// package's global state.
	sendTransient := func() {
		// The zsh script renders the transient prompt one column narrower to avoid
		// a redundant blank line when a filler is configured and the input is empty
		// (see _omp_zle-line-init in omp.zsh), mirror that for the streamed record.
		if e.Env.Shell() == shell.ZSH {
			e.rectifyTerminalWidth(-1)
			defer e.rectifyTerminalWidth(1)
		}

		out <- TransientMarker + e.ExtraPrompt(Transient)
	}

	go func() {
		defer close(out)
		defer close(e.streamingResults)

		// Render and send initial prompt with pending segments
		initialPrompt := e.Primary()
		out <- initialPrompt
		sendTransient()

		if e.countPendingSegments() == 0 {
			return
		}

		// Listen for segment completions
		for range e.streamingResults {
			out <- e.renderFromBlocks()

			if e.countPendingSegments() == 0 {
				// refresh the transient prompt now the context is fully resolved
				sendTransient()
				return
			}
		}
	}()

	return out
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
