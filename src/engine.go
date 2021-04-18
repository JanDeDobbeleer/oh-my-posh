package main

import (
	"fmt"
	"strings"
	"time"
)

type engine struct {
	config       *Config
	env          environmentInfo
	color        *AnsiColor
	renderer     *AnsiRenderer
	consoleTitle *consoleTitle
	// activeBlock           *Block
	// activeSegment         *Segment
	// previousActiveSegment *Segment
	rprompt string
}

func (e *engine) render() string {
	for _, block := range e.config.Blocks {
		e.renderBlock(block)
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

func (e *engine) renderBlock(block *Block) {
	block.init(e.env, e.color)
	block.setStringValues()
	defer e.color.reset()
	if !block.enabled() {
		return
	}
	if block.Newline {
		e.renderer.write("\n")
	}
	switch block.Type {
	// This is deprecated but leave if to not break current configs
	// It is encouraged to used "newline": true on block level
	// rather than the standalone the linebreak block
	case LineBreak:
		e.renderer.write("\n")
	case Prompt:
		if block.VerticalOffset != 0 {
			e.renderer.changeLine(block.VerticalOffset)
		}
		switch block.Alignment {
		case Right:
			e.renderer.carriageForward()
			blockText := block.renderSegments()
			e.renderer.setCursorForRightWrite(blockText, block.HorizontalOffset)
			e.renderer.write(blockText)
		case Left:
			e.renderer.write(block.renderSegments())
		}
	case RPrompt:
		e.rprompt = block.renderSegments()
	}
}

// debug will loop through your config file and output the timings for each segments
func (e *engine) debug() string {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	e.renderer.write("\n\x1b[1mHere are the timings of segments in your prompt:\x1b[0m\n\n")

	// console title timing
	start := time.Now()
	consoleTitle := e.consoleTitle.getTemplateText()
	duration := time.Since(start)
	segmentTiming := &SegmentTiming{
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
		block.init(e.env, e.color)
		longestSegmentName, timings := block.debug()
		segmentTimings = append(segmentTimings, timings...)
		if longestSegmentName > largestSegmentNameLength {
			largestSegmentNameLength = longestSegmentName
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
