package prompt

import (
	"runtime"
	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
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

	for i, segment := range block.Segments {
		go func(segment *config.Segment) {
			segment.Execute(e.Env)
			out <- result{segment, i}
		}(segment)
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

func (e *Engine) writeSegments(out chan result, block *config.Block) {
	count := len(block.Segments)
	// store the current index
	current := 0
	// keep track of what we already executed
	executedCount := 0
	// store the results
	results := make([]*config.Segment, count)
	// store the unique names of executed segments
	executed := make([]string, count)
	// store the actual redered index
	segmentIndex := 0

	for {
		select {
		case res := <-out:
			executedCount++

			finished := executedCount == count

			results[res.index] = res.segment

			name := res.segment.Name()
			if !slices.Contains(executed, name) {
				executed = append(executed, name)
			}

			segment := results[current]

			for segment != nil {
				if !e.canRenderSegment(segment, executed) && !finished {
					break
				}

				if segment.Render(segmentIndex) {
					segmentIndex++
				}

				e.writeSegment(block, segment)

				if current == count-1 {
					return
				}

				current++
				segment = results[current]
			}
		default:
			runtime.Gosched()
		}
	}
}

func (e *Engine) writeSegment(block *config.Block, segment *config.Segment) bool {
	if !segment.Enabled && segment.ResolveStyle() != config.Accordion {
		return false
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

	return true
}

func (e *Engine) canRenderSegment(segment *config.Segment, executed []string) bool {
	for _, name := range segment.Needs {
		if slices.Contains(executed, name) {
			continue
		}

		return false
	}

	return true
}
