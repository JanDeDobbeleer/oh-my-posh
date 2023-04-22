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

	e.pwd()

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
		writer := &ansi.Writer{
			TrueColor: e.Env.Flags().TrueColor,
		}
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

	foreground := prompt.ForegroundTemplates.FirstMatch(nil, e.Env, prompt.Foreground)
	background := prompt.BackgroundTemplates.FirstMatch(nil, e.Env, prompt.Background)
	e.Writer.SetColors(background, foreground)
	e.Writer.Write(background, foreground, promptText)

	str, length := e.Writer.String()
	if promptType == Transient {
		if padText, OK := e.shouldFill(prompt.Filler, length); OK {
			str += padText
		}
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
