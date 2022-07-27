package engine

import (
	"fmt"
	"oh-my-posh/color"
	"oh-my-posh/console"
	"oh-my-posh/environment"
	"oh-my-posh/shell"
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

func (e *Engine) canWriteRPrompt(rprompt bool) bool {
	if rprompt && (e.rprompt == "" || e.Plain) {
		return false
	}
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
	promptBreathingRoom := 5
	if rprompt {
		promptBreathingRoom = 30
	}
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
	e.printPWD()
	return e.print()
}

func (e *Engine) printPWD() {
	if len(e.Config.PWD) == 0 && !e.Config.OSC99 {
		return
	}
	cwd := e.Env.Pwd()
	// Backwards compatibility for deprecated OSC99
	if e.Config.OSC99 {
		e.writeANSI(e.Ansi.ConsolePwd(color.OSC99, "", cwd))
		return
	}
	// Allow template logic to define when to enable the PWD (when supported)
	tmpl := &template.Text{
		Template: e.Config.PWD,
		Env:      e.Env,
	}
	pwdType, err := tmpl.Render()
	if err != nil || len(pwdType) == 0 {
		return
	}
	host, _ := e.Env.Host()
	e.writeANSI(e.Ansi.ConsolePwd(pwdType, host, cwd))
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
	if err != nil || terminalWidth == 0 {
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
	defer func() {
		// Due to a bug in Powershell, the end of the line needs to be cleared.
		// If this doesn't happen, the portion after the prompt gets colored in the background
		// color of the line above the new input line. Clearing the line fixes this,
		// but can hopefully one day be removed when this is resolved natively.
		if e.Env.Shell() == shell.PWSH || e.Env.Shell() == shell.PWSH5 {
			e.writeANSI(e.Ansi.ClearAfter())
		}
	}()
	// when in bash, for rprompt blocks we need to write plain
	// and wrap in escaped mode or the prompt will not render correctly
	if e.Env.Shell() == shell.BASH && (block.Type == RPrompt || block.Alignment == Right) {
		block.InitPlain(e.Env, e.Config)
	} else {
		block.Init(e.Env, e.Writer, e.Ansi)
	}
	if !block.Enabled() {
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

		if block.Alignment == Left {
			text, length := block.RenderSegments()
			e.currentLineLength += length
			e.write(text)
			return
		}

		if block.Alignment != Right {
			return
		}

		text, length := block.RenderSegments()
		e.rpromptLength = length

		if !e.canWriteRPrompt(false) {
			switch block.Overflow {
			case Break:
				e.newline()
			case Hide:
				// make sure to fill if needed
				if padText, OK := e.shouldFill(block, 0); OK {
					e.write(padText)
				}
				return
			}
		}

		if padText, OK := e.shouldFill(block, length); OK {
			// in this case we can print plain
			e.write(padText)
			e.write(text)
			return
		}
		// this can contain ANSI escape sequences
		ansi := e.Ansi
		if e.Env.Shell() == shell.BASH {
			ansi = &color.Ansi{}
			ansi.InitPlain()
		}
		prompt := ansi.CarriageForward()
		prompt += ansi.GetCursorForRightWrite(length, block.HorizontalOffset)
		prompt += text
		e.currentLineLength = 0
		if e.Env.Shell() == shell.BASH {
			prompt = e.Ansi.FormatText(prompt)
		}
		e.write(prompt)
	case RPrompt:
		e.rprompt, e.rpromptLength = block.RenderSegments()
	}
}

// debug will loop through your config file and output the timings for each segments
func (e *Engine) PrintDebug(startTime time.Time, version string) string {
	var segmentTimings []*SegmentTiming
	largestSegmentNameLength := 0
	e.write(fmt.Sprintf("\n\x1b[1mVersion:\x1b[0m %s\n", version))
	e.write("\n\x1b[1mSegments:\x1b[0m\n\n")
	// console title timing
	titleStartTime := time.Now()
	title := e.ConsoleTitle.GetTitle()
	title = strings.TrimPrefix(title, "\x1b]0;")
	title = strings.TrimSuffix(title, "\a")
	segmentTiming := &SegmentTiming{
		name:       "ConsoleTitle",
		nameLength: 12,
		active:     len(e.Config.ConsoleTitleTemplate) > 0,
		text:       title,
		duration:   time.Since(titleStartTime),
	}
	segmentTimings = append(segmentTimings, segmentTiming)
	// loop each segments of each blocks
	for _, block := range e.Config.Blocks {
		block.Init(e.Env, e.Writer, e.Ansi)
		longestSegmentName, timings := block.Debug()
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
	e.write(fmt.Sprintf("\n\x1b[1mRun duration:\x1b[0m %s\n", time.Since(startTime)))
	e.write(fmt.Sprintf("\n\x1b[1mCache path:\x1b[0m %s\n", e.Env.CachePath()))
	e.write(fmt.Sprintf("\n\x1b[1mConfig path:\x1b[0m %s\n", e.Env.Flags().Config))
	e.write("\n\x1b[1mLogs:\x1b[0m\n\n")
	e.write(e.Env.Logs())
	return e.string()
}

func (e *Engine) print() string {
	switch e.Env.Shell() {
	case shell.ZSH:
		if !e.Env.Flags().Eval {
			break
		}
		// escape double quotes contained in the prompt
		prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), `"`, `\"`))
		prompt += fmt.Sprintf("\nRPROMPT=\"%s\"", e.rprompt)
		return prompt
	case shell.PWSH, shell.PWSH5, shell.PLAIN, shell.NU:
		if !e.canWriteRPrompt(true) {
			break
		}
		e.write(e.Ansi.SaveCursorPosition())
		e.write(e.Ansi.CarriageForward())
		e.write(e.Ansi.GetCursorForRightWrite(e.rpromptLength, 0))
		e.write(e.rprompt)
		e.write(e.Ansi.RestoreCursorPosition())
	case shell.BASH:
		if !e.canWriteRPrompt(true) {
			break
		}
		// in bash, the entire rprompt needs to be escaped for the prompt to be interpreted correctly
		// see https://github.com/JanDeDobbeleer/oh-my-posh/pull/2398
		ansi := &color.Ansi{}
		ansi.InitPlain()
		prompt := ansi.SaveCursorPosition()
		prompt += ansi.CarriageForward()
		prompt += ansi.GetCursorForRightWrite(e.rpromptLength, 0)
		prompt += e.rprompt
		prompt += ansi.RestoreCursorPosition()
		prompt = e.Ansi.FormatText(prompt)
		e.write(prompt)
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
	tooltip.Enabled = true
	// little hack to reuse the current logic
	block := &Block{
		Alignment: Right,
		Segments:  []*Segment{tooltip},
	}
	switch e.Env.Shell() {
	case shell.ZSH, shell.CMD, shell.FISH:
		block.Init(e.Env, e.Writer, e.Ansi)
		if !block.Enabled() {
			return ""
		}
		text, _ := block.RenderSegments()
		return text
	case shell.PWSH, shell.PWSH5:
		block.InitPlain(e.Env, e.Config)
		if !block.Enabled() {
			return ""
		}
		text, length := block.RenderSegments()
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
	var prompt *Segment
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
		prompt = &Segment{}
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
	foreground := prompt.ForegroundTemplates.FirstMatch(nil, e.Env, prompt.Foreground)
	background := prompt.BackgroundTemplates.FirstMatch(nil, e.Env, prompt.Background)
	e.Writer.SetColors(background, foreground)
	e.Writer.Write(background, foreground, promptText)
	switch e.Env.Shell() {
	case shell.ZSH:
		// escape double quotes contained in the prompt
		str, _ := e.Writer.String()
		if promptType == Transient {
			prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(str, "\"", "\"\""))
			// empty RPROMPT
			prompt += "\nRPROMPT=\"\""
			return prompt
		}
		return str
	case shell.PWSH, shell.PWSH5, shell.CMD, shell.BASH, shell.FISH, shell.NU:
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
	block.Init(e.Env, e.Writer, e.Ansi)
	if !block.Enabled() {
		return ""
	}
	text, length := block.RenderSegments()
	e.rpromptLength = length
	return text
}
