package prompt

import (
	"runtime"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
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

	e.writeSeparator(true)

	e.activeSegment = nil
	e.previousActiveSegment = nil

	return terminal.String()
}

func (e *Engine) writeSegments(out chan result, block *config.Block) {
	count := len(block.Segments)
	// store the current index
	current := 0
	// store the results
	results := make([]*config.Segment, count)
	// store the names of executed segments
	executed := make([]string, count)

	for {
		select {
		case res := <-out:
			results[res.index] = res.segment

			name := res.segment.Name()
			if !slices.Contains(executed, name) {
				executed = append(executed, name)
			}

			segment := results[current]

			for segment != nil {
				if !e.canRenderSegment(segment, executed) {
					break
				}

				segment.Render()
				e.writeSegment(current, block, segment)

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

func (e *Engine) writeSegment(index int, block *config.Block, segment *config.Segment) {
	if !segment.Enabled && segment.ResolveStyle() != config.Accordion {
		return
	}

	if colors, newCycle := cycle.Loop(); colors != nil {
		cycle = &newCycle
		segment.Foreground = colors.Foreground
		segment.Background = colors.Background
	}

	if index == 0 && len(block.LeadingDiamond) > 0 {
		segment.LeadingDiamond = block.LeadingDiamond
	}

	if index == len(block.Segments)-1 && len(block.TrailingDiamond) > 0 {
		segment.TrailingDiamond = block.TrailingDiamond
	}

	e.setActiveSegment(segment)
	e.renderActiveSegment()
}

func (e *Engine) canRenderSegment(segment *config.Segment, executed []string) bool {
	if !strings.Contains(segment.Template, ".Segments.") {
		return true
	}

	matches := regex.FindNamedRegexMatch(`\.Segments\.(?P<NAME>[a-zA-Z0-9]+)`, segment.Template)
	for _, name := range matches {
		if slices.Contains(executed, name) {
			continue
		}

		return false
	}

	return true
}
