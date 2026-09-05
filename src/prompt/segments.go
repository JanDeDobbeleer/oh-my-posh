package prompt

import (
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	runjobs "github.com/jandedobbeleer/oh-my-posh/src/runtime/jobs"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

type result struct {
	segment *config.Segment
	index   int
}

func (e *Engine) writeBlockSegments(block *config.Block) (string, int) {
	out := e.launchBlockSegments(block)
	if out == nil {
		return "", 0
	}

	// A single standalone block (RPrompt/Tooltip) only ever needs to resolve
	// dependencies against its own segments, matching a fresh executed map.
	executed := make(map[string]bool, len(block.Segments))
	results := drainBlockResults(out, len(block.Segments), executed)

	text, length, _ := e.renderBlockSegments(results, block, executed)
	return text, length
}

// Callers may consume the channel immediately or defer consumption to allow
// other blocks' segments to execute concurrently in the meantime. Returns nil
// when the block has no segments.
func (e *Engine) launchBlockSegments(block *config.Block) chan result {
	length := len(block.Segments)

	if length == 0 {
		return nil
	}

	out := make(chan result, length)

	e.writeSegmentsConcurrently(block.Segments, out)

	return out
}

// Calling this for every block before any rendering begins ensures the
// executed map is fully populated, so cross-block .Segments.X dependencies
// resolve in both directions (an earlier block can reference a later block's
// segment and vice versa).
func drainBlockResults(out chan result, count int, executed map[string]bool) []*config.Segment {
	results := make([]*config.Segment, count)
	for range count {
		res := <-out
		results[res.index] = res.segment
		executed[res.segment.Name()] = true
	}
	return results
}

// Rendering is strictly sequential. For multi-block prompts, executed must be
// fully populated for all blocks before this is called (via drainBlockResults),
// so that cross-block .Segments.X dependencies resolve in both directions.
func (e *Engine) renderBlockSegments(results []*config.Segment, block *config.Block, executed map[string]bool) (string, int, []terminal.Run) {
	if block.RestartCycle {
		cycle = &e.Config.Cycle
	}

	e.writeSegments(results, block, executed)

	if e.activeSegment != nil && len(block.TrailingDiamond) > 0 {
		e.activeSegment.TrailingDiamond = block.TrailingDiamond
	}

	e.writeSeparator(true)

	e.activeSegment = nil
	e.captureBlockTailColors()
	e.previousActiveSegment = nil

	// captureBlockRuns must run before terminal.String(): String's own defer
	// resets the run stream (runsState.Runs[:0]) before String returns to
	// this caller, so capturing terminal.Runs() after the call would see it
	// already truncated.
	runs := e.captureBlockRuns()

	text, length := terminal.String()

	return text, length, runs
}

func (e *Engine) writeSegmentsConcurrently(segments []*config.Segment, out chan result) {
	for i, segment := range segments {
		// In streaming mode, pre-register all segments as pending
		// This ensures countPendingSegments() sees them before timeout occurs.
		// Without a positive streaming timeout no segment can ever time out
		// into the pending state, so pre-registering would leak entries (the
		// cleanup below only runs for segment.Timeout > 0) and keep the
		// StreamPrimary producer waiting forever.
		if e.Env.Flags().Streaming && e.Config.Streaming > 0 {
			segment.Timeout = e.Config.Streaming
			e.pendingSegments.Store(segment.Name(), true)
		}

		go func(segment *config.Segment, index int) {
			timedOut := false
			if segment.Timeout > 0 {
				timedOut = e.executeSegmentWithTimeout(segment)
			} else {
				segment.Execute(e.Env)
			}

			// A segment that finished in time sends no completion notification,
			// so it leaves the pending set here, before its result is handed
			// over. The producer starts waiting for notifications as soon as
			// every result is in; an entry removed after that point would be
			// waited on forever. A segment that timed out stays registered
			// until the producer consumes its notification.
			if e.Env.Flags().Streaming && segment.Timeout > 0 && !timedOut {
				e.pendingSegments.Delete(segment.Name())
			}

			out <- result{segment, index}
		}(segment, i)
	}
}

// executeSegmentWithTimeout runs the segment and reports whether it hit its
// timeout. In streaming mode a timed-out segment keeps executing in the
// background and is handed to trackPendingSegment.
func (e *Engine) executeSegmentWithTimeout(segment *config.Segment) bool {
	done := make(chan bool)
	gidChan := make(chan uint64, 1)

	go func() {
		gidChan <- runjobs.CurrentGID()
		segment.Execute(e.Env)
		close(done)
	}()

	gid := <-gidChan

	timer := time.NewTimer(time.Duration(segment.Timeout) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-done:
		return false
	case <-timer.C:
	}

	log.Errorf("timeout after %dms for segment: %s", segment.Timeout, segment.Name())

	// Streaming lets the segment keep running and renders a placeholder
	// until it reports back. Only the pending flag is set here: Enabled is
	// still being written by Execute on the other goroutine.
	if e.Env.Flags().Streaming {
		segment.SetPending(true)
		e.trackPendingSegment(segment, done)
		return true
	}

	segment.Killed = true
	if err := runjobs.KillGoroutineChildren(gid); err != nil {
		log.Errorf("failed to kill child processes for goroutine %d (segment: %s): %v", gid, segment.Name(), err)
	}

	return true
}

func (e *Engine) writeSegments(results []*config.Segment, block *config.Block, executed map[string]bool) {
	count := len(results)
	current := 0
	segmentIndex := 0

	// Render segments in index order while their dependencies are satisfied.
	// executed is fully pre-populated before rendering begins (via drainBlockResults),
	// so all resolvable cross-block and same-block dependencies are already available.
	for current < count && e.canRenderSegment(results[current], executed) {
		segment := results[current]
		if segment.Render(segmentIndex, e.forceRender) {
			segmentIndex++
		}

		e.writeSegment(block, segment)
		current++
	}

	// Render remaining segments whose Needs could not be resolved
	for ; current < count; current++ {
		segment := results[current]
		if segment.Render(segmentIndex, e.forceRender) {
			segmentIndex++
		}

		e.writeSegment(block, segment)
	}
}

func (e *Engine) writeSegment(block *config.Block, segment *config.Segment) {
	// Allow pending segments to render (they show "..." text)
	if !segment.IsPending() && !segment.Enabled && segment.ResolveStyle() != config.Accordion {
		return
	}

	if colors, newCycle := cycle.Loop(); colors != nil {
		cycle = &newCycle
		segment.Foreground = colors.Foreground
		segment.Background = colors.Background
	}

	if terminal.Len() == 0 && len(block.LeadingDiamond) > 0 {
		segment.LeadingDiamond = block.LeadingDiamond
	}

	e.setActiveSegment(segment)
	e.renderActiveSegment()
}

func (e *Engine) canRenderSegment(segment *config.Segment, executed map[string]bool) bool {
	for _, name := range segment.Needs {
		if !executed[name] {
			return false
		}
	}

	return true
}
