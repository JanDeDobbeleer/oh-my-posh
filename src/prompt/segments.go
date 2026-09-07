package prompt

import (
	"runtime/debug"
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
		if e.Env.Flags().Streaming && e.Config.Streaming > 0 {
			segment.Timeout = e.Config.Streaming
		}

		// Name memoizes on first call. Resolve it before the goroutine starts so the render and
		// Execute goroutines never both write it.
		_ = segment.Name()

		go func(segment *config.Segment, index int) {
			e.runSegment(segment)
			out <- result{segment, index}
		}(segment, i)
	}
}

// runSegment dispatches to the timeout-aware path when the segment has a deadline, and to a
// direct execution otherwise.
func (e *Engine) runSegment(segment *config.Segment) {
	if segment.Timeout > 0 {
		e.executeSegmentWithTimeout(segment)
		return
	}

	e.executeSegment(segment)
}

// executeSegment runs Execute and turns a panic into a disabled segment. Without this a
// panicking segment kills the whole process, which for the serve daemon means a dead prompt
// with no error anywhere.
func (e *Engine) executeSegment(segment *config.Segment) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		log.Errorf("segment %s panicked: %v\n%s", segment.Name(), r, debug.Stack())
		segment.Enabled = false
	}()

	segment.Execute(e.Env)
}

func (e *Engine) executeSegmentWithTimeout(segment *config.Segment) {
	done := make(chan struct{})
	gidChan := make(chan uint64, 1)

	go func() {
		gidChan <- runjobs.CurrentGID()
		defer close(done)
		e.executeSegment(segment)
	}()

	gid := <-gidChan

	timer := time.NewTimer(time.Duration(segment.Timeout) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-done:
		return
	case <-timer.C:
	}

	log.Errorf("timeout after %dms for segment: %s", segment.Timeout, segment.Name())

	if e.stream == nil {
		segment.Killed = true
		if err := runjobs.KillGoroutineChildren(gid); err != nil {
			log.Errorf("failed to kill child processes for goroutine %d (segment: %s): %v", gid, segment.Name(), err)
		}

		return
	}

	// The timed-out event is queued before this function returns and the block result is sent,
	// so the producer's absorb after drainBlockResults always sees it.
	e.stream.timedOut(segment)

	go e.awaitPendingSegment(segment, done, gid)
}

// awaitPendingSegment reports a pending segment's completion to the producer, or abandons it
// once pendingSegmentLimit passes so a hung segment cannot pin the cycle forever.
func (e *Engine) awaitPendingSegment(segment *config.Segment, done <-chan struct{}, gid uint64) {
	timer := time.NewTimer(pendingSegmentLimit)
	defer timer.Stop()

	select {
	case <-done:
		e.stream.completed(segment)
	case <-timer.C:
		log.Errorf("abandoning segment %s: still running after %s", segment.Name(), pendingSegmentLimit)
		if err := runjobs.KillGoroutineChildren(gid); err != nil {
			log.Errorf("failed to kill child processes for goroutine %d (segment: %s): %v", gid, segment.Name(), err)
		}

		e.stream.abandon(segment)
	}
}

// renderSegment renders one segment for the current pass and reports whether it occupied an
// index slot. A pending segment renders its placeholder from configuration only; an abandoned
// one is skipped entirely. Neither may touch the writer: it belongs to the Execute goroutine
// until the completed event arrives.
func (e *Engine) renderSegment(block *config.Block, segment *config.Segment, index int) bool {
	if e.segmentAbandoned(segment) {
		return false
	}

	if e.segmentPending(segment) {
		segment.RenderPlaceholder()
		e.writeSegment(block, segment, true, true)
		return true
	}

	enabled := segment.Render(index, e.forceRender)
	e.writeSegment(block, segment, false, enabled)

	return enabled
}

func (e *Engine) writeSegments(results []*config.Segment, block *config.Block, executed map[string]bool) {
	count := len(results)
	current := 0
	segmentIndex := 0

	// Render segments in index order while their dependencies are satisfied.
	// executed is fully pre-populated before rendering begins (via drainBlockResults),
	// so all resolvable cross-block and same-block dependencies are already available.
	// A pending segment's Needs must never be read here: Execute's deferred evaluateNeeds
	// may still be appending to it concurrently.
	for current < count && (e.segmentPending(results[current]) || e.canRenderSegment(results[current], executed)) {
		if e.renderSegment(block, results[current], segmentIndex) {
			segmentIndex++
		}

		current++
	}

	// Render remaining segments whose Needs could not be resolved
	for ; current < count; current++ {
		if e.renderSegment(block, results[current], segmentIndex) {
			segmentIndex++
		}
	}
}

func (e *Engine) writeSegment(block *config.Block, segment *config.Segment, pending, enabled bool) {
	if !enabled && segment.ResolveStyle() != config.Accordion {
		return
	}

	if colors, newCycle := cycle.Loop(); colors != nil {
		cycle = &newCycle
		segment.Foreground = colors.Foreground
		segment.Background = colors.Background

		// A resolved segment picks these up lazily through its color templates; the
		// placeholder already collapsed its colors in RenderPlaceholder, so re-seed them
		// or the cycle would be skipped while the segment is pending.
		if pending {
			segment.CollapseForeground(colors.Foreground)
			segment.CollapseBackground(colors.Background)
		}
	}

	if terminal.Len() == 0 && len(block.LeadingDiamond) > 0 {
		segment.LeadingDiamond = block.LeadingDiamond
	}

	e.setActiveSegment(segment, pending)
	e.renderActiveSegment(enabled)
}

func (e *Engine) canRenderSegment(segment *config.Segment, executed map[string]bool) bool {
	for _, name := range segment.Needs {
		if !executed[name] {
			return false
		}
	}

	return true
}
