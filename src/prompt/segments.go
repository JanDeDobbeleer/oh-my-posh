package prompt

import (
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

const (
	// Threshold for using goroutine pool vs individual goroutines
	goroutinePoolThreshold = 10
	// Size of the worker pool
	workerPoolSize = 4
)

type result struct {
	segment *config.Segment
	index   int
}

func (e *Engine) writeBlockSegments(block *config.Block) (string, int) {
	length := len(block.Segments)

	if length == 0 {
		return "", 0
	}

	out := make(chan result, length)

	// Use goroutine pool for large numbers of segments to reduce overhead
	if length > goroutinePoolThreshold {
		e.writeSegmentsWithPool(block.Segments, out)
	} else {
		e.writeSegmentsConcurrently(block.Segments, out)
	}

	e.writeSegments(out, block)

	if e.activeSegment != nil && len(block.TrailingDiamond) > 0 {
		e.activeSegment.TrailingDiamond = block.TrailingDiamond
	}

	e.writeSeparator(true)

	e.activeSegment = nil
	e.previousActiveSegment = nil

	return terminal.String()
}

// writeSegmentsConcurrently uses individual goroutines for each segment
func (e *Engine) writeSegmentsConcurrently(segments []*config.Segment, out chan result) {
	for i, segment := range segments {
		go func(segment *config.Segment, index int) {
			segment.Execute(e.Env)
			out <- result{segment, index}
		}(segment, i)
	}
}

// writeSegmentsWithPool uses a worker pool to process segments
func (e *Engine) writeSegmentsWithPool(segments []*config.Segment, out chan result) {
	tasks := make(chan result, len(segments))
	var wg sync.WaitGroup

	// Start worker pool
	for range workerPoolSize {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				task.segment.Execute(e.Env)
				out <- task
			}
		}()
	}

	// Send tasks to workers
	go func() {
		defer close(tasks)
		for i, segment := range segments {
			tasks <- result{segment, i}
		}
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(out)
	}()
}

func (e *Engine) writeSegments(out chan result, block *config.Block) {
	count := len(block.Segments)
	current := 0
	executedCount := 0
	results := make([]*config.Segment, count)
	// Pre-allocate map with known capacity to reduce allocations
	executed := make(map[string]bool, count)
	segmentIndex := 0

	// Process results as they come in, eliminating busy waiting
	for executedCount < count {
		res := <-out // Block until result is available
		executedCount++

		results[res.index] = res.segment
		executed[res.segment.Name()] = true

		// Process segments that can now be rendered
		for current < count && results[current] != nil {
			segment := results[current]
			if !e.canRenderSegment(segment, executed) {
				break
			}

			if segment.Render(segmentIndex, e.forceRender) {
				segmentIndex++
			}

			e.writeSegment(block, segment)
			current++
		}
	}

	// render all remaining segments where the needs can't be resolved
	for current < executedCount {
		segment := results[current]
		if segment.Render(segmentIndex, e.forceRender) {
			segmentIndex++
		}

		e.writeSegment(block, segment)
		current++
	}
}

func (e *Engine) writeSegment(block *config.Block, segment *config.Segment) {
	if !segment.Enabled && segment.ResolveStyle() != config.Accordion {
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

// canRenderSegment now uses map for O(1) lookups instead of O(n) slice search
func (e *Engine) canRenderSegment(segment *config.Segment, executed map[string]bool) bool {
	for _, name := range segment.Needs {
		if !executed[name] {
			return false
		}
	}

	return true
}
