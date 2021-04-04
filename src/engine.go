package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type engine struct {
	config                *Config
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
	return e.previousActiveSegment.background()
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
		e.writePowerLineSeparator(e.getPowerlineColor(false), e.previousActiveSegment.background(), true)
	}
}

func (e *engine) renderPowerLineSegment(text string) {
	e.writePowerLineSeparator(e.activeSegment.background(), e.getPowerlineColor(true), false)
	e.renderText(text)
}

func (e *engine) renderPlainSegment(text string) {
	e.renderText(text)
}

func (e *engine) renderDiamondSegment(text string) {
	e.color.write(Transparent, e.activeSegment.background(), e.activeSegment.LeadingDiamond)
	e.renderText(text)
	e.color.write(Transparent, e.activeSegment.background(), e.activeSegment.TrailingDiamond)
}

func (e *engine) renderText(text string) {
	text = e.color.formats.generateHyperlink(text)
	defaultValue := " "
	prefix := e.activeSegment.getValue(Prefix, defaultValue)
	postfix := e.activeSegment.getValue(Postfix, defaultValue)
	e.color.write(e.activeSegment.background(), e.activeSegment.foreground(), fmt.Sprintf("%s%s%s", prefix, text, postfix))
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
		e.renderSegmentText(segment.stringValue)
	}
	if e.previousActiveSegment != nil && e.previousActiveSegment.Style == Powerline {
		e.writePowerLineSeparator(Transparent, e.previousActiveSegment.background(), true)
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

func (e *engine) render() string {
	for _, block := range e.config.Blocks {
		// if line break, append a line break
		switch block.Type {
		case LineBreak:
			e.renderer.write("\n")
		case Prompt:
			if block.VerticalOffset != 0 {
				e.renderer.changeLine(block.VerticalOffset)
			}
			switch block.Alignment {
			case Right:
				e.renderer.carriageForward()
				blockText := e.renderBlockSegments(block)
				e.renderer.setCursorForRightWrite(blockText, block.HorizontalOffset)
				e.renderer.write(blockText)
			case Left:
				e.renderer.write(e.renderBlockSegments(block))
			}
		case RPrompt:
			e.rprompt = e.renderBlockSegments(block)
		}
	}
	if e.config.ConsoleTitle {
		e.renderer.write(e.consoleTitle.getConsoleTitle())
	}
	e.renderer.creset()
	if e.config.FinalSpace {
		e.renderer.write(" ")
	}

	if !e.config.OSC99 {
		return e.print()
	}
	cwd := e.env.getcwd()
	if e.env.isWsl() {
		cwd, _ = e.env.runCommand("wslpath", "-m", cwd)
	}
	e.renderer.osc99(cwd)
	return e.print()
}

// debug will loop through your config file and output the timings for each segments
func (e *engine) debug() string {
	var segmentTimings []SegmentTiming
	largestSegmentNameLength := 0
	e.renderer.write("\n\x1b[1mHere are the timings of segments in your prompt:\x1b[0m\n\n")

	// console title timing
	start := time.Now()
	consoleTitle := e.consoleTitle.getTemplateText()
	duration := time.Since(start)
	segmentTiming := SegmentTiming{
		name:            "ConsoleTitle",
		nameLength:      12,
		enabled:         e.config.ConsoleTitle,
		stringValue:     consoleTitle,
		enabledDuration: 0,
		stringDuration:  duration,
	}
	segmentTimings = append(segmentTimings, segmentTiming)
	// loop each segments of each blocks
	for _, block := range e.config.Blocks {
		for _, segment := range block.Segments {
			err := segment.mapSegmentWithWriter(e.env)
			if err != nil || !segment.shouldIncludeFolder(e.env.getcwd()) {
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
				e.renderSegmentText(segmentTiming.stringValue)
				if e.activeSegment.Style == Powerline {
					e.writePowerLineSeparator(Transparent, e.activeSegment.background(), true)
				}
				segmentTiming.stringValue = e.color.string()
				e.color.builder.Reset()
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
		e.renderer.write(fmt.Sprintf("%-*s - %3d ms - %s\n", largestSegmentNameLength, segmentName, duration, segment.stringValue))
	}
	return e.renderer.string()
}

func (e *engine) print() string {
	switch e.env.getShellName() {
	case zsh:
		if *e.env.getArgs().Eval {
			// escape double quotes contained in the prompt
			prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.renderer.string(), "\"", "\"\""))
			prompt += fmt.Sprintf("\nRPROMPT=\"%s\"", e.rprompt)
			return prompt
		}
	case pwsh, powershell5, bash, shelly:
		if e.rprompt != "" {
			e.renderer.saveCursorPosition()
			e.renderer.carriageForward()
			e.renderer.setCursorForRightWrite(e.rprompt, 0)
			e.renderer.write(e.rprompt)
			e.renderer.restoreCursorPosition()
		}
	}
	return e.renderer.string()
}

func (e *engine) resetBlock() {
	e.color.reset()
	e.previousActiveSegment = nil
	e.activeBlock = nil
}
