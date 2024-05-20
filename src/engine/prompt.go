package engine

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/ansi"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type ExtraPromptType int

const (
	Transient ExtraPromptType = iota
	Valid
	Error
	Secondary
	Debug
)

func (e *Engine) Primary() string {
	if e.Config.ShellIntegration {
		exitCode, _ := e.Env.StatusCodes()
		e.write(e.Writer.CommandFinished(exitCode, e.Env.Flags().NoExitCode))
		e.write(e.Writer.PromptStart())
	}

	// cache a pointer to the color cycle
	cycle = &e.Config.Cycle
	var cancelNewline, didRender bool

	for i, block := range e.Config.Blocks {
		// do not print a leading newline when we're at the first row and the prompt is cleared
		if i == 0 {
			row, _ := e.Env.CursorPosition()
			cancelNewline = e.Env.Flags().Cleared || e.Env.Flags().PromptCount == 1 || row == 1
		}

		// skip setting a newline when we didn't print anything yet
		if i != 0 {
			cancelNewline = !didRender
		}

		// only render rprompt for shells where we need it from the primary prompt
		renderRPrompt := true
		switch e.Env.Shell() {
		case shell.ELVISH, shell.FISH, shell.NU, shell.XONSH, shell.CMD:
			renderRPrompt = false
		}

		if block.Type == RPrompt && !renderRPrompt {
			continue
		}

		if e.renderBlock(block, cancelNewline) {
			didRender = true
		}
	}

	if len(e.Config.ConsoleTitleTemplate) > 0 && !e.Env.Flags().Plain {
		title := e.getTitleTemplateText()
		e.write(e.Writer.FormatTitle(title))
	}

	if e.Config.FinalSpace {
		e.write(" ")
		e.currentLineLength++
	}

	if e.Config.ITermFeatures != nil && e.isIterm() {
		host, _ := e.Env.Host()
		e.write(e.Writer.RenderItermFeatures(e.Config.ITermFeatures, e.Env.Shell(), e.Env.Pwd(), e.Env.User(), host))
	}

	if e.Config.ShellIntegration && e.Config.TransientPrompt == nil {
		e.write(e.Writer.CommandStart())
	}

	e.pwd()

	switch e.Env.Shell() {
	case shell.ZSH:
		if !e.Env.Flags().Eval {
			break
		}
		// Warp doesn't support RPROMPT so we need to write it manually
		if e.isWarp() {
			e.writeRPrompt()
			// escape double quotes contained in the prompt
			prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), `"`, `\"`))
			return prompt
		}
		// escape double quotes contained in the prompt
		prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(e.string(), `"`, `\"`))
		prompt += fmt.Sprintf("\nRPROMPT=\"%s\"", e.rprompt)
		return prompt
	case shell.PWSH, shell.PWSH5, shell.GENERIC:
		e.writeRPrompt()
	case shell.BASH:
		space, OK := e.canWriteRightBlock(true)
		if !OK {
			break
		}
		// in bash, the entire rprompt needs to be escaped for the prompt to be interpreted correctly
		// see https://github.com/jandedobbeleer/oh-my-posh/pull/2398
		writer := &ansi.Writer{
			TrueColor: e.Env.Flags().TrueColor,
		}
		writer.Init(shell.GENERIC)
		prompt := writer.SaveCursorPosition()
		prompt += strings.Repeat(" ", space)
		prompt += e.rprompt
		prompt += writer.RestoreCursorPosition()
		prompt = e.Writer.FormatText(prompt)
		e.write(prompt)
	}

	return e.string()
}

func (e *Engine) ExtraPrompt(promptType ExtraPromptType) string {
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

	if promptType == Transient && e.Config.ShellIntegration {
		exitCode, _ := e.Env.StatusCodes()
		e.write(e.Writer.CommandFinished(exitCode, e.Env.Flags().NoExitCode))
		e.write(e.Writer.PromptStart())
	}

	foreground := prompt.ForegroundTemplates.FirstMatch(nil, e.Env, prompt.Foreground)
	background := prompt.BackgroundTemplates.FirstMatch(nil, e.Env, prompt.Background)
	e.Writer.SetColors(background, foreground)
	e.Writer.Write(background, foreground, promptText)

	str, length := e.Writer.String()
	if promptType == Transient {
		consoleWidth, err := e.Env.TerminalWidth()
		if err == nil || consoleWidth != 0 {
			if padText, OK := e.shouldFill(prompt.Filler, consoleWidth, length); OK {
				str += padText
			}
		}
	}

	if promptType == Transient && e.Config.ShellIntegration {
		str += e.Writer.CommandStart()
	}

	switch e.Env.Shell() {
	case shell.ZSH:
		// escape double quotes contained in the prompt
		if promptType == Transient {
			prompt := fmt.Sprintf("PS1=\"%s\"", strings.ReplaceAll(str, "\"", "\"\""))
			// empty RPROMPT
			prompt += "\nRPROMPT=\"\""
			return prompt
		}
		return str
	case shell.PWSH, shell.PWSH5:
		// Return the string and empty our buffer
		// clear the line afterwards to prevent text from being written on the same line
		// see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3628
		return str + e.Writer.ClearAfter()
	case shell.CMD, shell.BASH, shell.FISH, shell.NU, shell.GENERIC:
		// Return the string and empty our buffer
		return str
	}

	return ""
}

func (e *Engine) RPrompt() string {
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
