package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

var (
	cycle *ansi.Cycle = &ansi.Cycle{}
)

type Engine struct {
	Config *Config
	Env    platform.Environment
	Writer *ansi.Writer
	Plain  bool

	console           strings.Builder
	currentLineLength int
	rprompt           string
	rpromptLength     int
}

func (e *Engine) write(text string) {
	e.console.WriteString(text)
}

func (e *Engine) string() string {
	text := e.console.String()
	e.console.Reset()
	return text
}

func (e *Engine) canWriteRightBlock(rprompt bool) bool {
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
	// cache a pointer to the color cycle
	cycle = &e.Config.Cycle
	for i, block := range e.Config.Blocks {
		var cancelNewline bool
		if i == 0 {
			row, _ := e.Env.CursorPosition()
			cancelNewline = e.Env.Flags().Cleared || e.Env.Flags().PromptCount == 1 || row == 1
		}
		e.renderBlock(block, cancelNewline)
	}
	if len(e.Config.ConsoleTitleTemplate) > 0 {
		title := e.getTitleTemplateText()
		e.write(e.Writer.FormatTitle(title))
	}
	if e.Config.FinalSpace {
		e.write(" ")
	}
	e.printPWD()
	return e.print()
}

func (e *Engine) printPWD() {
	// only print when supported
	sh := e.Env.Shell()
	if sh == shell.ELVISH || sh == shell.XONSH {
		return
	}
	// only print when relevant
	if len(e.Config.PWD) == 0 && !e.Config.OSC99 {
		return
	}

	cwd := e.Env.Pwd()
	// Backwards compatibility for deprecated OSC99
	if e.Config.OSC99 {
		e.write(e.Writer.ConsolePwd(ansi.OSC99, "", "", cwd))
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

	user := e.Env.User()
	host, _ := e.Env.Host()
	e.write(e.Writer.ConsolePwd(pwdType, user, host, cwd))
}

func (e *Engine) newline() {
	// WARP terminal will remove \n from the prompt, so we hack a newline in
	if e.isWarp() {
		e.write(e.Writer.LineBreak())
	} else {
		e.write("\n")
	}
	e.currentLineLength = 0
}

func (e *Engine) isWarp() bool {
	return e.Env.Getenv("TERM_PROGRAM") == "WarpTerminal"
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
	if lenFiller == 0 {
		return "", false
	}
	repeat := padLength / lenFiller
	return strings.Repeat(filler, repeat), true
}

func (e *Engine) getTitleTemplateText() string {
	tmpl := &template.Text{
		Template: e.Config.ConsoleTitleTemplate,
		Env:      e.Env,
	}
	if text, err := tmpl.Render(); err == nil {
		return text
	}
	return ""
}

func (e *Engine) renderBlock(block *Block, cancelNewline bool) {
	defer func() {
		// when in PowerShell, we need to clear the line after the prompt
		// to avoid the background being printed on the next line
		// when at the end of the buffer.
		// See https://github.com/JanDeDobbeleer/oh-my-posh/issues/65
		if e.Env.Shell() == shell.PWSH || e.Env.Shell() == shell.PWSH5 {
			e.write(e.Writer.ClearAfter())
		}
	}()
	// when in bash, for rprompt blocks we need to write plain
	// and wrap in escaped mode or the prompt will not render correctly
	if e.Env.Shell() == shell.BASH && block.Type == RPrompt {
		block.InitPlain(e.Env, e.Config)
	} else {
		block.Init(e.Env, e.Writer)
	}

	if !block.Enabled() {
		return
	}

	// do not print a newline to avoid a leading space
	// when we're printin the first primary prompt in
	// the shell
	if block.Newline && !cancelNewline {
		e.newline()
	}

	switch block.Type {
	// This is deprecated but we leave it in to not break configs
	// It is encouraged to used "newline": true on block level
	// rather than the standalone the linebreak block
	case LineBreak:
		// do not print a newline to avoid a leading space
		// when we're printin the first primary prompt in
		// the shell
		if !cancelNewline {
			return
		}
		e.newline()
	case Prompt:
		if block.VerticalOffset != 0 {
			e.write(e.Writer.ChangeLine(block.VerticalOffset))
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

		if !e.canWriteRightBlock(false) {
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
		prompt := e.Writer.CarriageForward()
		prompt += e.Writer.GetCursorForRightWrite(length, block.HorizontalOffset)
		prompt += text
		e.currentLineLength = 0
		e.write(prompt)
	case RPrompt:
		e.rprompt, e.rpromptLength = block.RenderSegments()
	}
}

// debug will loop through your config file and output the timings for each segments
func (e *Engine) PrintDebug(startTime time.Time, version string) string {
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Version:").Green().Bold().Plain(), version))
	e.write(log.Text("\nSegments:\n\n").Green().Bold().Plain().String())
	// console title timing
	titleStartTime := time.Now()
	title := e.getTitleTemplateText()
	consoleTitleTiming := &SegmentTiming{
		name:       "ConsoleTitle",
		nameLength: 12,
		active:     len(e.Config.ConsoleTitleTemplate) > 0,
		text:       title,
		duration:   time.Since(titleStartTime),
	}
	largestSegmentNameLength := consoleTitleTiming.nameLength
	var segmentTimings []*SegmentTiming
	segmentTimings = append(segmentTimings, consoleTitleTiming)
	// cache a pointer to the color cycle
	cycle = &e.Config.Cycle
	// loop each segments of each blocks
	for _, block := range e.Config.Blocks {
		block.Init(e.Env, e.Writer)
		longestSegmentName, timings := block.Debug()
		segmentTimings = append(segmentTimings, timings...)
		if longestSegmentName > largestSegmentNameLength {
			largestSegmentNameLength = longestSegmentName
		}
	}

	// 22 is the color for false/true and 7 is the reset color
	largestSegmentNameLength += 22 + 7
	for _, segment := range segmentTimings {
		duration := segment.duration.Milliseconds()
		var active log.Text
		if segment.active {
			active = log.Text("true").Yellow()
		} else {
			active = log.Text("false").Purple()
		}
		segmentName := fmt.Sprintf("%s(%s)", segment.name, active.Plain())
		e.write(fmt.Sprintf("%-*s - %3d ms - %s\n", largestSegmentNameLength, segmentName, duration, segment.text))
	}
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Run duration:").Green().Bold().Plain(), time.Since(startTime)))
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Cache path:").Green().Bold().Plain(), e.Env.CachePath()))
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Config path:").Green().Bold().Plain(), e.Env.Flags().Config))
	e.write(log.Text("\nLogs:\n\n").Green().Bold().Plain().String())
	e.write(e.Env.Logs())
	return e.string()
}

func (e *Engine) print() string {
	switch e.Env.Shell() {
	case shell.ZSH:
		if !e.Env.Flags().Eval {
			break
		}
		// Warp doesn't support RPROMPT so we need to write it manually
		if e.isWarp() {
			e.write(e.Writer.SaveCursorPosition())
			e.write(e.Writer.CarriageForward())
			e.write(e.Writer.GetCursorForRightWrite(e.rpromptLength, 0))
			e.write(e.rprompt)
			e.write(e.Writer.RestoreCursorPosition())
			// escape double quotes contained in the prompt
			prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), `"`, `\"`))
			return prompt
		}
		// escape double quotes contained in the prompt
		prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), `"`, `\"`))
		prompt += fmt.Sprintf("\nRPROMPT=\"%s\"", e.rprompt)
		return prompt
	case shell.PWSH, shell.PWSH5, shell.GENERIC, shell.NU:
		if !e.canWriteRightBlock(true) {
			break
		}
		e.write(e.Writer.SaveCursorPosition())
		e.write(e.Writer.CarriageForward())
		e.write(e.Writer.GetCursorForRightWrite(e.rpromptLength, 0))
		e.write(e.rprompt)
		e.write(e.Writer.RestoreCursorPosition())
	case shell.BASH:
		if !e.canWriteRightBlock(true) {
			break
		}
		// in bash, the entire rprompt needs to be escaped for the prompt to be interpreted correctly
		// see https://github.com/jandedobbeleer/oh-my-posh/pull/2398
		writer := &ansi.Writer{}
		writer.Init(shell.GENERIC)
		prompt := writer.SaveCursorPosition()
		prompt += writer.CarriageForward()
		prompt += writer.GetCursorForRightWrite(e.rpromptLength, 0)
		prompt += e.rprompt
		prompt += writer.RestoreCursorPosition()
		prompt = e.Writer.FormatText(prompt)
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
	case shell.ZSH, shell.CMD, shell.FISH, shell.GENERIC:
		block.Init(e.Env, e.Writer)
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
		e.write(e.Writer.CarriageForward())
		e.write(e.Writer.GetCursorForRightWrite(length, 0))
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
	// populate env with latest context
	e.Env.LoadTemplateCache()
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
		switch promptType { //nolint: exhaustive
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
	case shell.PWSH, shell.PWSH5, shell.CMD, shell.BASH, shell.FISH, shell.NU, shell.GENERIC:
		// Return the string and empty our buffer
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
	block.Init(e.Env, e.Writer)
	if !block.Enabled() {
		return ""
	}
	text, length := block.RenderSegments()
	e.rpromptLength = length
	return text
}
