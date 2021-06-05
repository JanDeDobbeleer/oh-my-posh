package main

import (
	"fmt"
	"strings"
	"time"
)

type engine struct {
	config       *Config
	env          environmentInfo
	colorWriter  colorWriter
	ansi         *ansiUtils
	consoleTitle *consoleTitle

	console strings.Builder
	rprompt string
}

func (e *engine) write(text string) {
	e.console.WriteString(text)
}

func (e *engine) string() string {
	return e.console.String()
}

func (e *engine) canWriteRPrompt() bool {
	prompt := e.string()
	consoleWidth, err := e.env.getTerminalWidth()
	if err != nil || consoleWidth == 0 {
		return true
	}
	promptWidth := e.ansi.lenWithoutANSI(prompt)
	availableSpace := consoleWidth - promptWidth
	if promptWidth > consoleWidth {
		availableSpace = promptWidth - (promptWidth % consoleWidth)
	}
	promptBreathingRoom := 30
	return (availableSpace - e.ansi.lenWithoutANSI(e.rprompt)) >= promptBreathingRoom
}

func (e *engine) render() string {
	for _, block := range e.config.Blocks {
		e.renderBlock(block)
	}
	if e.config.ConsoleTitle {
		e.write(e.consoleTitle.getConsoleTitle())
	}
	e.write(e.ansi.creset)
	if e.config.FinalSpace {
		e.write(" ")
	}

	if !e.config.OSC99 {
		return e.print()
	}
	cwd := e.env.getcwd()
	if e.env.isWsl() {
		cwd, _ = e.env.runCommand("wslpath", "-m", cwd)
	}
	e.write(e.ansi.consolePwd(cwd))
	return e.print()
}

func (e *engine) renderBlock(block *Block) {
	// when in bash, for rprompt blocks we need to write plain
	// and wrap in escaped mode or the prompt will not render correctly
	if block.Type == RPrompt && e.env.getShellName() == bash {
		block.initPlain(e.env, e.config)
	} else {
		block.init(e.env, e.colorWriter, e.ansi)
	}
	block.setStringValues()
	if !block.enabled() {
		return
	}
	if block.Newline {
		e.write("\n")
	}
	switch block.Type {
	// This is deprecated but leave if to not break current configs
	// It is encouraged to used "newline": true on block level
	// rather than the standalone the linebreak block
	case LineBreak:
		e.write("\n")
	case Prompt:
		if block.VerticalOffset != 0 {
			e.write(e.ansi.changeLine(block.VerticalOffset))
		}
		switch block.Alignment {
		case Right:
			e.write(e.ansi.carriageForward())
			blockText := block.renderSegments()
			e.write(e.ansi.getCursorForRightWrite(blockText, block.HorizontalOffset))
			e.write(blockText)
		case Left:
			e.write(block.renderSegments())
		}
	case RPrompt:
		blockText := block.renderSegments()
		if e.env.getShellName() == bash {
			blockText = fmt.Sprintf(e.ansi.bashFormat, blockText)
		}
		e.rprompt = blockText
	}
	// Due to a bug in Powershell, the end of the line needs to be cleared.
	// If this doesn't happen, the portion after the prompt gets colored in the background
	// color of the line above the new input line. Clearing the line fixes this,
	// but can hopefully one day be removed when this is resolved natively.
	if e.ansi.shell == pwsh || e.ansi.shell == powershell5 {
		e.write(e.ansi.clearEOL)
	}
}

// debug will loop through your config file and output the timings for each segments
func (e *engine) debug() string {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	e.write("\n\x1b[1mHere are the timings of segments in your prompt:\x1b[0m\n\n")

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
		block.init(e.env, e.colorWriter, e.ansi)
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
		e.write(fmt.Sprintf("%-*s - %3d ms - %s\n", largestSegmentNameLength, segmentName, duration, segment.stringValue))
	}
	return e.string()
}

func (e *engine) print() string {
	switch e.env.getShellName() {
	case zsh:
		if !*e.env.getArgs().Eval {
			break
		}
		// escape double quotes contained in the prompt
		prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), "\"", "\"\""))
		prompt += fmt.Sprintf("\nRPROMPT=\"%s\"", e.rprompt)
		return prompt
	case pwsh, powershell5, bash, plain:
		if e.rprompt == "" || !e.canWriteRPrompt() {
			break
		}
		e.write(e.ansi.saveCursorPosition)
		e.write(e.ansi.carriageForward())
		e.write(e.ansi.getCursorForRightWrite(e.rprompt, 0))
		e.write(e.rprompt)
		e.write(e.ansi.restoreCursorPosition)
	}
	return e.string()
}

func (e *engine) renderTooltip(tip string) string {
	tip = strings.Trim(tip, " ")
	var tooltip *Segment
	for _, tp := range e.config.Tooltips {
		if !tp.shouldInvokeWithTip(tip) {
			continue
		}
		tooltip = tp
	}
	if tooltip == nil {
		return ""
	}
	if err := tooltip.mapSegmentWithWriter(e.env); err != nil {
		return ""
	}
	if !tooltip.enabled() {
		return ""
	}
	tooltip.stringValue = tooltip.string()
	// little hack to reuse the current logic
	block := &Block{
		Alignment: Right,
		Segments:  []*Segment{tooltip},
	}
	switch e.env.getShellName() {
	case zsh:
		block.init(e.env, e.colorWriter, e.ansi)
		return block.renderSegments()
	case pwsh, powershell5:
		block.initPlain(e.env, e.config)
		tooltipText := block.renderSegments()
		e.write(e.ansi.clearEOL)
		e.write(e.ansi.carriageForward())
		e.write(e.ansi.getCursorForRightWrite(tooltipText, 0))
		e.write(tooltipText)
		return e.string()
	}
	return ""
}
