package main

import (
	"fmt"
	"oh-my-posh/environment"
	"strings"
	"time"
)

type engine struct {
	config       *Config
	env          environment.Environment
	writer       promptWriter
	ansi         *ansiUtils
	consoleTitle *consoleTitle
	plain        bool

	console strings.Builder
	rprompt string
}

func (e *engine) write(text string) {
	e.console.WriteString(text)
}

func (e *engine) writeANSI(text string) {
	if e.plain {
		return
	}
	e.console.WriteString(text)
}

func (e *engine) string() string {
	return e.console.String()
}

func (e *engine) canWriteRPrompt() bool {
	prompt := e.string()
	consoleWidth, err := e.env.TerminalWidth()
	if err != nil || consoleWidth == 0 {
		return true
	}
	promptWidth := e.ansi.lenWithoutANSI(prompt)
	availableSpace := consoleWidth - promptWidth
	// spanning multiple lines
	if availableSpace < 0 {
		overflow := promptWidth % consoleWidth
		availableSpace = consoleWidth - overflow
	}
	promptBreathingRoom := 30
	canWrite := (availableSpace - e.ansi.lenWithoutANSI(e.rprompt)) >= promptBreathingRoom
	return canWrite
}

func (e *engine) render() string {
	for _, block := range e.config.Blocks {
		e.renderBlock(block)
	}
	if e.config.ConsoleTitle {
		e.writeANSI(e.consoleTitle.getConsoleTitle())
	}
	e.writeANSI(e.ansi.creset)
	if e.config.FinalSpace {
		e.write(" ")
	}

	if !e.config.OSC99 {
		return e.print()
	}
	cwd := e.env.Pwd()
	e.writeANSI(e.ansi.consolePwd(cwd))
	return e.print()
}

func (e *engine) renderBlock(block *Block) {
	// when in bash, for rprompt blocks we need to write plain
	// and wrap in escaped mode or the prompt will not render correctly
	if block.Type == RPrompt && e.env.Shell() == bash {
		block.initPlain(e.env, e.config)
	} else {
		block.init(e.env, e.writer, e.ansi)
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
			e.writeANSI(e.ansi.changeLine(block.VerticalOffset))
		}
		switch block.Alignment {
		case Right:
			e.writeANSI(e.ansi.carriageForward())
			blockText := block.renderSegments()
			e.writeANSI(e.ansi.getCursorForRightWrite(blockText, block.HorizontalOffset))
			e.write(blockText)
		case Left:
			e.write(block.renderSegments())
		}
	case RPrompt:
		blockText := block.renderSegments()
		if e.env.Shell() == bash {
			blockText = fmt.Sprintf(e.ansi.bashFormat, blockText)
		}
		e.rprompt = blockText
	}
	// Due to a bug in Powershell, the end of the line needs to be cleared.
	// If this doesn't happen, the portion after the prompt gets colored in the background
	// color of the line above the new input line. Clearing the line fixes this,
	// but can hopefully one day be removed when this is resolved natively.
	if e.ansi.shell == pwsh || e.ansi.shell == powershell5 {
		e.writeANSI(e.ansi.clearAfter())
	}
}

// debug will loop through your config file and output the timings for each segments
func (e *engine) debug() string {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	e.write(fmt.Sprintf("\n\x1b[1mVersion:\x1b[0m %s\n", Version))
	e.write("\n\x1b[1mSegments:\x1b[0m\n\n")
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
		block.init(e.env, e.writer, e.ansi)
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
	e.write(fmt.Sprintf("\n\x1b[1mRun duration:\x1b[0m %s\n", time.Since(start)))
	e.write(fmt.Sprintf("\n\x1b[1mCache path:\x1b[0m %s\n", e.env.CachePath()))
	e.write("\n\x1b[1mLogs:\x1b[0m\n\n")
	e.write(e.env.Logs())
	return e.string()
}

func (e *engine) print() string {
	switch e.env.Shell() {
	case zsh:
		if !*e.env.Args().Eval {
			break
		}
		// escape double quotes contained in the prompt
		prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), "\"", "\"\""))
		prompt += fmt.Sprintf("\nRPROMPT=\"%s\"", e.rprompt)
		return prompt
	case pwsh, powershell5, bash, plain:
		if e.rprompt == "" || !e.canWriteRPrompt() || e.plain {
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
	switch e.env.Shell() {
	case zsh, winCMD:
		block.init(e.env, e.writer, e.ansi)
		return block.renderSegments()
	case pwsh, powershell5:
		block.initPlain(e.env, e.config)
		tooltipText := block.renderSegments()
		e.write(e.ansi.clearAfter())
		e.write(e.ansi.carriageForward())
		e.write(e.ansi.getCursorForRightWrite(tooltipText, 0))
		e.write(tooltipText)
		return e.string()
	}
	return ""
}

func (e *engine) renderTransientPrompt() string {
	if e.config.TransientPrompt == nil {
		return ""
	}
	promptTemplate := e.config.TransientPrompt.Template
	if len(promptTemplate) == 0 {
		promptTemplate = "{{ .Shell }}> "
	}
	template := &textTemplate{
		Template: promptTemplate,
		Env:      e.env,
	}
	prompt, err := template.render()
	if err != nil {
		prompt = err.Error()
	}
	e.writer.setColors(e.config.TransientPrompt.Background, e.config.TransientPrompt.Foreground)
	e.writer.write(e.config.TransientPrompt.Background, e.config.TransientPrompt.Foreground, prompt)
	switch e.env.Shell() {
	case zsh:
		// escape double quotes contained in the prompt
		prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.writer.string(), "\"", "\"\""))
		prompt += "\nRPROMPT=\"\""
		return prompt
	case pwsh, powershell5, winCMD:
		return e.writer.string()
	}
	return ""
}

func (e *engine) renderRPrompt() string {
	filterRPromptBlock := func(blocks []*Block) *Block {
		for _, block := range blocks {
			if block.Type == RPrompt {
				return block
			}
		}
		return nil
	}
	block := filterRPromptBlock(e.config.Blocks)
	if block == nil {
		return ""
	}
	block.init(e.env, e.writer, e.ansi)
	block.setStringValues()
	if !block.enabled() {
		return ""
	}
	return block.renderSegments()
}
