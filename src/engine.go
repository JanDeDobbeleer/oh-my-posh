package main

import (
	"fmt"
	"sync"
	"time"
)

type engine struct {
	settings              *Settings
	env                   environmentInfo
	color                 *AnsiColor
	renderer              *AnsiRenderer
	consoleTitle          *consoleTitle
	activeBlock           *Block
	activeSegment         *Segment
	previousActiveSegment *Segment
	rprompt               string
}

// SegmentTiming holds the timing context for a segment
type SegmentTiming struct {
	name            string
	nameLength      int
	enabled         bool
	stringValue     string
	enabledDuration time.Duration
	stringDuration  time.Duration
}

func (e *engine) getPowerlineColor(foreground bool) string {
	if e.previousActiveSegment == nil {
		return Transparent
	}
	if !foreground && e.activeSegment.Style != Powerline {
		return Transparent
	}
	if foreground && e.previousActiveSegment.Style != Powerline {
		return Transparent
	}
	return e.previousActiveSegment.Background
}

func (e *engine) writePowerLineSeparator(background, foreground string, end bool) {
	symbol := e.activeSegment.PowerlineSymbol
	if end {
		symbol = e.previousActiveSegment.PowerlineSymbol
	}
	if e.activeSegment.InvertPowerline {
		e.color.write(foreground, background, symbol)
		return
	}
	e.color.write(background, foreground, symbol)
}

func (e *engine) endPowerline() {
	if e.activeSegment != nil &&
		e.activeSegment.Style != Powerline &&
		e.previousActiveSegment != nil &&
		e.previousActiveSegment.Style == Powerline {
		e.writePowerLineSeparator(e.getPowerlineColor(false), e.previousActiveSegment.Background, true)
	}
}

func (e *engine) renderPowerLineSegment(text string) {
	e.writePowerLineSeparator(e.activeSegment.Background, e.getPowerlineColor(true), false)
	e.renderText(text)
}

func (e *engine) renderPlainSegment(text string) {
	e.renderText(text)
}

func (e *engine) renderDiamondSegment(text string) {
	e.color.write(Transparent, e.activeSegment.Background, e.activeSegment.LeadingDiamond)
	e.renderText(text)
	e.color.write(Transparent, e.activeSegment.Background, e.activeSegment.TrailingDiamond)
}

func (e *engine) renderText(text string) {
	defaultValue := " "
	if e.activeSegment.Background != "" {
		defaultValue = fmt.Sprintf("<%s>\u2588</>", e.activeSegment.Background)
	}
	prefix := e.activeSegment.getValue(Prefix, defaultValue)
	postfix := e.activeSegment.getValue(Postfix, defaultValue)
	e.color.write(e.activeSegment.Background, e.activeSegment.Foreground, fmt.Sprintf("%s%s%s", prefix, text, postfix))
}

func (e *engine) renderSegmentText(text string) {
	switch e.activeSegment.Style {
	case Plain:
		e.renderPlainSegment(text)
	case Diamond:
		e.renderDiamondSegment(text)
	case Powerline:
		e.renderPowerLineSegment(text)
	}
	e.previousActiveSegment = e.activeSegment
}

func (e *engine) renderBlockSegments(block *Block) string {
	defer e.resetBlock()
	e.activeBlock = block
	e.setStringValues(block.Segments)
	for _, segment := range block.Segments {
		if !segment.active {
			continue
		}
		e.activeSegment = segment
		e.endPowerline()
		text := segment.stringValue
		e.activeSegment.Background = segment.props.background
		e.activeSegment.Foreground = segment.props.foreground
		e.renderSegmentText(text)
	}
	if e.previousActiveSegment != nil && e.previousActiveSegment.Style == Powerline {
		e.writePowerLineSeparator(Transparent, e.previousActiveSegment.Background, true)
	}
	return e.color.string()
}

func (e *engine) setStringValues(segments []*Segment) {
	wg := sync.WaitGroup{}
	wg.Add(len(segments))
	defer wg.Wait()
	cwd := e.env.getcwd()
	for _, segment := range segments {
		go func(s *Segment) {
			defer wg.Done()
			s.setStringValue(e.env, cwd)
		}(segment)
	}
}

func (e *engine) render() {
	for _, block := range e.settings.Blocks {
		// if line break, append a line break
		switch block.Type {
		case LineBreak:
			e.renderer.print("\n")
		case Prompt:
			if block.VerticalOffset != 0 {
				e.renderer.changeLine(block.VerticalOffset)
			}
			switch block.Alignment {
			case Right:
				e.renderer.carriageForward()
				blockText := e.renderBlockSegments(block)
				e.renderer.setCursorForRightWrite(blockText, block.HorizontalOffset)
				e.renderer.print(blockText)
			case Left:
				e.renderer.print(e.renderBlockSegments(block))
			}
		case RPrompt:
			e.rprompt = e.renderBlockSegments(block)
		}
	}
	if e.settings.ConsoleTitle {
		e.renderer.print(e.consoleTitle.getConsoleTitle())
	}
	e.renderer.creset()
	if e.settings.FinalSpace {
		e.renderer.print(" ")
	}
	e.write()
}

// debug will lool through your config file and output the timings for each segments
func (e *engine) debug() {
	var segmentTimings []SegmentTiming
	largestSegmentNameLength := 0
	e.renderer.print("\n\x1b[1mHere are the timings of segments in your prompt:\x1b[0m\n\n")
	// loop each segments of each blocks
	for _, block := range e.settings.Blocks {
		for _, segment := range block.Segments {
			err := segment.mapSegmentWithWriter(e.env)
			if err != nil || segment.shouldIgnoreFolder(e.env.getcwd()) {
				continue
			}
			var segmentTiming SegmentTiming
			segmentTiming.name = string(segment.Type)
			segmentTiming.nameLength = len(segmentTiming.name)
			if segmentTiming.nameLength > largestSegmentNameLength {
				largestSegmentNameLength = segmentTiming.nameLength
			}
			// enabled() timing
			start := time.Now()
			segmentTiming.enabled = segment.enabled()
			segmentTiming.enabledDuration = time.Since(start)
			// string() timing
			if segmentTiming.enabled {
				start = time.Now()
				segmentTiming.stringValue = segment.string()
				segmentTiming.stringDuration = time.Since(start)
				e.previousActiveSegment = nil
				e.activeSegment = segment
				e.activeSegment.Background = segment.props.background
				e.activeSegment.Foreground = segment.props.foreground
				e.renderSegmentText(segmentTiming.stringValue)
				if e.activeSegment.Style == Powerline {
					e.writePowerLineSeparator(Transparent, e.activeSegment.Background, true)
				}
				segmentTiming.stringValue = e.color.string()
				e.color.buffer.Reset()
			}
			segmentTimings = append(segmentTimings, segmentTiming)
		}
	}
	// pad the output so the tabs render correctly
	largestSegmentNameLength += 7
	for _, segment := range segmentTimings {
		duration := segment.enabledDuration.Milliseconds()
		if segment.enabled {
			duration += segment.stringDuration.Milliseconds()
		}
		segmentName := fmt.Sprintf("%s(%t)", segment.name, segment.enabled)
		e.renderer.print(fmt.Sprintf("%-*s - %3d ms - %s\n", largestSegmentNameLength, segmentName, duration, segment.stringValue))
	}
	fmt.Print(e.renderer.string())
}

func (e *engine) write() {
	switch e.env.getShellName() {
	case zsh:
		if *e.env.getArgs().Eval {
			fmt.Printf("PS1=\"%s\"", e.renderer.string())
			fmt.Printf("\nRPROMPT=\"%s\"", e.rprompt)
			return
		}
	case pwsh, powershell5, bash:
		if e.rprompt != "" {
			e.renderer.saveCursorPosition()
			e.renderer.carriageForward()
			e.renderer.setCursorForRightWrite(e.rprompt, 0)
			e.renderer.print(e.rprompt)
			e.renderer.restoreCursorPosition()
		}
	}
	fmt.Print(e.renderer.string())
}

func (e *engine) resetBlock() {
	e.color.reset()
	e.previousActiveSegment = nil
	e.activeBlock = nil
}
