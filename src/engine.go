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
	writer                AnsiWriter
	utils                 *ANSIUtils
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
		e.writer.write(foreground, background, symbol)
		return
	}
	e.writer.write(background, foreground, symbol)
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
	e.writer.write(Transparent, e.activeSegment.background(), e.activeSegment.LeadingDiamond)
	e.renderText(text)
	e.writer.write(Transparent, e.activeSegment.background(), e.activeSegment.TrailingDiamond)
}

func (e *engine) renderText(text string) {
	defaultValue := " "
	if e.activeSegment.background() != "" {
		defaultValue = fmt.Sprintf("<%s>\u2588</>", e.activeSegment.background())
	}

	text = e.utils.formats.generateHyperlink(text)

	prefix := e.activeSegment.getValue(Prefix, defaultValue)
	postfix := e.activeSegment.getValue(Postfix, defaultValue)
	e.writer.write(e.activeSegment.background(), e.activeSegment.foreground(), fmt.Sprintf("%s%s%s", prefix, text, postfix))
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
	return e.writer.render()
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
	for _, block := range e.config.Blocks {
		// if line break, append a line break
		switch block.Type {
		case LineBreak:
			e.utils.write("\n")
		case Prompt:
			if block.VerticalOffset != 0 {
				e.utils.changeLine(block.VerticalOffset)
			}
			switch block.Alignment {
			case Right:
				e.utils.carriageForward()
				blockText := e.renderBlockSegments(block)
				e.utils.setCursorForRightWrite(blockText, block.HorizontalOffset)
				e.utils.write(blockText)
			case Left:
				e.utils.write(e.renderBlockSegments(block))
			}
		case RPrompt:
			e.rprompt = e.renderBlockSegments(block)
		}
	}
	if e.config.ConsoleTitle {
		e.utils.write(e.consoleTitle.getConsoleTitle())
	}
	e.utils.creset()
	if e.config.FinalSpace {
		e.utils.write(" ")
	}

	if !e.config.OSC99 {
		e.print()
		return
	}
	cwd := e.env.getcwd()
	if e.env.isWsl() {
		cwd, _ = e.env.runCommand("wslpath", "-m", cwd)
	}
	e.utils.osc99(cwd)
	e.print()
}

// debug will loop through your config file and output the timings for each segments
func (e *engine) debug() {
	var segmentTimings []SegmentTiming
	largestSegmentNameLength := 0
	e.utils.write("\n\x1b[1mHere are the timings of segments in your prompt:\x1b[0m\n\n")

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
				segmentTiming.stringValue = e.writer.render()
				e.writer.reset()
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
		e.utils.write(fmt.Sprintf("%-*s - %3d ms - %s\n", largestSegmentNameLength, segmentName, duration, segment.stringValue))
	}
	fmt.Print(e.utils.string())
}

func (e *engine) print() {
	switch e.env.getShellName() {
	case zsh:
		if *e.env.getArgs().Eval {
			// escape double quotes contained in the prompt
			fmt.Printf("PS1=\"%s\"", strings.ReplaceAll(e.utils.string(), "\"", "\"\""))
			fmt.Printf("\nRPROMPT=\"%s\"", e.rprompt)
			return
		}
	case pwsh, powershell5, bash:
		if e.rprompt != "" {
			e.utils.saveCursorPosition()
			e.utils.carriageForward()
			e.utils.setCursorForRightWrite(e.rprompt, 0)
			e.utils.write(e.rprompt)
			e.utils.restoreCursorPosition()
		}
	}
	fmt.Print(e.utils.string())
}

func (e *engine) resetBlock() {
	e.writer.reset()
	e.previousActiveSegment = nil
	e.activeBlock = nil
}
