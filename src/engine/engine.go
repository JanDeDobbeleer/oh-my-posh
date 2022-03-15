package engine

import (
	"fmt"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"oh-my-posh/template"
	"strings"
	"time"
)

type Engine struct {
	Config       *Config
	Env          environment.Environment
	Writer       color.Writer
	Ansi         *color.Ansi
	ConsoleTitle *console.Title
	Plain        bool

	console           strings.Builder
	currentLineLength int
	rprompt           string
	rpromptLength     int
}

func (e *Engine) write(text string) {
	e.console.WriteString(text)
}

func (e *Engine) writeANSI(text string) {
	if e.Plain {
		return
	}
	e.console.WriteString(text)
}

func (e *Engine) string() string {
	return e.console.String()
}

func (e *Engine) canWriteRPrompt() bool {
	consoleWidth, err := e.Env.TerminalWidth()
	if err != nil || consoleWidth == 0 {
		return true
	}
	promptWidth := e.currentLineLength
	availableSpace := consoleWidth - promptWidth
	// spanning multiple lines
	if availableSpace < 0 {
		overflow := promptWidth % consoleWidth
		availableSpace = consoleWidth - overflow
	}
	promptBreathingRoom := 30
	canWrite := (availableSpace - e.rpromptLength) >= promptBreathingRoom
	return canWrite
}

func (e *Engine) PrintPrimary() string {
	for _, block := range e.Config.Blocks {
		e.renderBlock(block)
	}
	if len(e.Config.ConsoleTitleTemplate) > 0 {
		e.writeANSI(e.ConsoleTitle.GetTitle())
	}
	e.writeANSI(e.Ansi.ColorReset())
	if e.Config.FinalSpace {
		e.write(" ")
	}
	if !e.Config.OSC99 {
		return e.print()
	}
	cwd := e.Env.Pwd()
	e.writeANSI(e.Ansi.ConsolePwd(cwd))
	return e.print()
}

func (e *Engine) newline() {
	e.write("\n")
	e.currentLineLength = 0
}

func (e *Engine) shouldFill(block *Block, length int) (string, bool) {
	if len(block.Filler) == 0 {
		return "", false
	}
	terminalWidth, err := e.Env.TerminalWidth()
	if err != nil && terminalWidth == 0 {
		return "", false
	}
	padLength := terminalWidth - e.currentLineLength - length
	if padLength <= 0 {
		return "", false
	}
	e.Writer.Write("", "", block.Filler)
	filler, lenFiller := e.Writer.String()
	e.Writer.Reset()
	if lenFiller == 0 {
		return "", false
	}
	repeat := padLength / lenFiller
	return strings.Repeat(filler, repeat), true
}

func (e *Engine) renderBlock(block *Block) {
	// when in bash, for rprompt blocks we need to write plain
	// and wrap in escaped mode or the prompt will not render correctly
	if block.Type == RPrompt && e.Env.Shell() == bash {
		block.initPlain(e.Env, e.Config)
	} else {
		block.init(e.Env, e.Writer, e.Ansi)
	}
	block.renderSegmentsText()
	if !block.enabled() {
		return
	}
	if block.Newline {
		e.newline()
	}
	switch block.Type {
	// This is deprecated but leave if to not break current configs
	// It is encouraged to used "newline": true on block level
	// rather than the standalone the linebreak block
	case LineBreak:
		e.newline()
	case Prompt:
		if block.VerticalOffset != 0 {
			e.writeANSI(e.Ansi.ChangeLine(block.VerticalOffset))
		}
		switch block.Alignment {
		case Right:
			text, length := block.renderSegments()
			if padText, OK := e.shouldFill(block, length); OK {
				e.write(padText)
			}
			e.writeANSI(e.Ansi.CarriageForward())
			e.writeANSI(e.Ansi.GetCursorForRightWrite(length, block.HorizontalOffset))
			e.currentLineLength = 0
			e.write(text)
		case Left:
			text, length := block.renderSegments()
			e.currentLineLength += length
			e.write(text)
		}
	case RPrompt:
		text, length := block.renderSegments()
		e.rpromptLength = length
		if e.Env.Shell() == bash {
			text = e.Ansi.FormatText(text)
		}
		e.rprompt = text
	}
	// Due to a bug in Powershell, the end of the line needs to be cleared.
	// If this doesn't happen, the portion after the prompt gets colored in the background
	// color of the line above the new input line. Clearing the line fixes this,
	// but can hopefully one day be removed when this is resolved natively.
	if e.Env.Shell() == pwsh || e.Env.Shell() == powershell5 {
		e.writeANSI(e.Ansi.ClearAfter())
	}
}

// debug will loop through your config file and output the timings for each segments
func (e *Engine) PrintDebug(version string) string {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	e.write(fmt.Sprintf("\n\x1b[1mVersion:\x1b[0m %s\n", version))
	e.write("\n\x1b[1mSegments:\x1b[0m\n\n")
	// console title timing
	start := time.Now()
	title := e.ConsoleTitle.GetTitle()
	title = strings.TrimPrefix(title, "\x1b]0;")
	title = strings.TrimSuffix(title, "\a")
	duration := time.Since(start)
	segmentTiming := &SegmentTiming{
		name:       "ConsoleTitle",
		nameLength: 12,
		active:     len(e.Config.ConsoleTitleTemplate) > 0,
		text:       title,
		duration:   duration,
	}
	segmentTimings = append(segmentTimings, segmentTiming)
	// loop each segments of each blocks
	for _, block := range e.Config.Blocks {
		block.init(e.Env, e.Writer, e.Ansi)
		longestSegmentName, timings := block.debug()
		segmentTimings = append(segmentTimings, timings...)
		if longestSegmentName > largestSegmentNameLength {
			largestSegmentNameLength = longestSegmentName
		}
	}

	// pad the output so the tabs render correctly
	largestSegmentNameLength += 7
	for _, segment := range segmentTimings {
		duration := segment.duration.Milliseconds()
		segmentName := fmt.Sprintf("%s(%t)", segment.name, segment.active)
		e.write(fmt.Sprintf("%-*s - %3d ms - %s\n", largestSegmentNameLength, segmentName, duration, segment.text))
	}
	e.write(fmt.Sprintf("\n\x1b[1mRun duration:\x1b[0m %s\n", time.Since(start)))
	e.write(fmt.Sprintf("\n\x1b[1mCache path:\x1b[0m %s\n", e.Env.CachePath()))
	e.write("\n\x1b[1mLogs:\x1b[0m\n\n")
	e.write(e.Env.Logs())
	return e.string()
}

func (e *Engine) print() string {
	switch e.Env.Shell() {
	case zsh:
		if !e.Env.Flags().Eval {
			break
		}
		// escape double quotes contained in the prompt
		prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), "\"", "\"\""))
		prompt += fmt.Sprintf("\nRPROMPT=\"%s\"", e.rprompt)
		return prompt
	case pwsh, powershell5, bash, plain:
		if e.rprompt == "" || !e.canWriteRPrompt() || e.Plain {
			break
		}
		e.write(e.Ansi.SaveCursorPosition())
		e.write(e.Ansi.CarriageForward())
		e.write(e.Ansi.GetCursorForRightWrite(e.rpromptLength, 0))
		e.write(e.rprompt)
		e.write(e.Ansi.RestoreCursorPosition())
	}
	return e.string()
}

func (e *Engine) PrintTooltip(tip string) string {
	tip = strings.Trim(tip, " ")
	var tooltip *Segment
	for _, tp := range e.Config.Tooltips {
		if !tp.shouldInvokeWithTip(tip) {
			continue
		}
		tooltip = tp
	}
	if tooltip == nil {
		return ""
	}
	if err := tooltip.mapSegmentWithWriter(e.Env); err != nil {
		return ""
	}
	if !tooltip.writer.Enabled() {
		return ""
	}
	tooltip.text = tooltip.string()
	tooltip.enabled = true
	// little hack to reuse the current logic
	block := &Block{
		Alignment: Right,
		Segments:  []*Segment{tooltip},
	}
	switch e.Env.Shell() {
	case zsh, winCMD:
		block.init(e.Env, e.Writer, e.Ansi)
		text, _ := block.renderSegments()
		return text
	case pwsh, powershell5:
		block.initPlain(e.Env, e.Config)
		text, length := block.renderSegments()
		e.write(e.Ansi.ClearAfter())
		e.write(e.Ansi.CarriageForward())
		e.write(e.Ansi.GetCursorForRightWrite(length, 0))
		e.write(text)
		return e.string()
	}
	return ""
}

type ExtraPromptType int

const (
	Transient ExtraPromptType = iota
	Valid
	Error
	Secondary
	Debug
)

func (e *Engine) PrintExtraPrompt(promptType ExtraPromptType) string {
	var prompt *ExtraPrompt
	switch promptType {
	case Debug:
		prompt = e.Config.DebugPrompt
	case Transient:
		prompt = e.Config.TransientPrompt
	case Valid:
		prompt = e.Config.ValidLine
	case Error:
		prompt = e.Config.ErrorLine
	case Secondary:
		prompt = e.Config.SecondaryPrompt
	}
	if prompt == nil {
		return ""
	}
	getTemplate := func(template string) string {
		if len(template) != 0 {
			return template
		}
		switch promptType { // nolint: exhaustive
		case Debug:
			return "[DBG]: "
		case Transient:
			return "{{ .Shell }}> "
		case Secondary:
			return "> "
		default:
			return ""
		}
	}
	tmpl := &template.Text{
		Template: getTemplate(prompt.Template),
		Env:      e.Env,
	}
	promptText, err := tmpl.Render()
	if err != nil {
		promptText = err.Error()
	}
	e.Writer.SetColors(prompt.Background, prompt.Foreground)
	e.Writer.Write(prompt.Background, prompt.Foreground, promptText)
	switch e.Env.Shell() {
	case zsh:
		// escape double quotes contained in the prompt
		str, _ := e.Writer.String()
		if promptType == Transient {
			prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(str, "\"", "\"\""))
			// empty RPROMPT
			prompt += "\nRPROMPT=\"\""
			return prompt
		}
		return str
	case pwsh, powershell5, winCMD, bash:
		str, _ := e.Writer.String()
		return str
	}
	return ""
}

func (e *Engine) PrintRPrompt() string {
	filterRPromptBlock := func(blocks []*Block) *Block {
		for _, block := range blocks {
			if block.Type == RPrompt {
				return block
			}
		}
		return nil
	}
	block := filterRPromptBlock(e.Config.Blocks)
	if block == nil {
		return ""
	}
	block.init(e.Env, e.Writer, e.Ansi)
	block.renderSegmentsText()
	if !block.enabled() {
		return ""
	}
	text, length := block.renderSegments()
	e.rpromptLength = length
	return text
}
