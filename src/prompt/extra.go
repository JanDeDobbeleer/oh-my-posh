package prompt

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

type ExtraPromptType int

const (
	Transient ExtraPromptType = iota
	Valid
	Error
	Secondary
	Debug
)

func (e *Engine) ExtraPrompt(promptType ExtraPromptType) string {
	var prompt *config.Segment

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
		prompt = &config.Segment{}
	}

	getTemplate := func(template string) string {
		if len(template) != 0 {
			return template
		}
		switch promptType {
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

	promptText, err := template.Render(getTemplate(prompt.Template), nil)
	if err != nil {
		promptText = err.Error()
	}

	if promptType == Transient && prompt.Newline && !e.cancelNewline() {
		promptText = fmt.Sprintf("%s%s", e.getNewline(), promptText)
	}

	if promptType == Transient && e.Config.ShellIntegration {
		exitCode, _ := e.Env.StatusCodes()
		e.write(terminal.CommandFinished(exitCode, e.Env.Flags().NoExitCode))
		e.write(terminal.PromptStart())
	}

	foreground := color.Ansi(prompt.ForegroundTemplates.FirstMatch(nil, string(prompt.Foreground)))
	background := color.Ansi(prompt.BackgroundTemplates.FirstMatch(nil, string(prompt.Background)))
	terminal.SetColors(background, foreground)
	terminal.Write(background, foreground, promptText)

	str, length := terminal.String()

	if promptType != Transient {
		return str
	}

	rightStr, rightLength := e.renderRightTemplate(prompt, background, foreground)

	var padText string
	if len(prompt.Filler) != 0 {
		consoleWidth, err := e.Env.TerminalWidth()
		if err == nil || consoleWidth != 0 {
			padText, _ = e.shouldFill(prompt.Filler, consoleWidth-length-rightLength)
		}
	}

	// for pwsh, the padding moves inside the cursor save/restore sequence
	// when a right-aligned template is rendered, see transientPWSH
	if e.Env.Shell() != shell.PWSH || rightLength == 0 {
		str += padText
	}

	switch e.Env.Shell() {
	case shell.ZSH:
		if !e.Env.Flags().Eval {
			return str
		}

		return e.transientZSH(str, rightStr)
	case shell.PWSH:
		return e.transientPWSH(str, padText, rightStr, length, rightLength)
	}

	return str
}

// transientZSH returns the transient prompt as an eval statement setting
// both PS1 and RPROMPT, letting zsh align the right-aligned template natively.
func (e *Engine) transientZSH(str, rightStr string) string {
	// Warp doesn't support RPROMPT
	if e.isWarp() {
		rightStr = ""
	}

	prompt := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(str))
	prompt += fmt.Sprintf("\nRPROMPT=%s", shell.QuotePosixStr(rightStr))
	return prompt
}

// transientPWSH appends the right-aligned template to the transient prompt by
// writing the gap and the right-aligned text, then restoring the cursor to
// right after the left part so the accepted command is drawn there,
// mirroring writePrimaryRightPrompt.
func (e *Engine) transientPWSH(str, padText, rightStr string, length, rightLength int) string {
	// clear the line afterwards to prevent text from being written on the same line
	// see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3628
	str += terminal.ClearAfter()

	if rightLength == 0 {
		return str
	}

	consoleWidth, err := e.Env.TerminalWidth()
	if err != nil || consoleWidth == 0 {
		return str
	}

	gap := consoleWidth - length - rightLength
	if gap < 0 {
		return str
	}

	if len(padText) == 0 {
		padText = strings.Repeat(" ", gap)
	}

	return str + terminal.SaveCursorPosition() + padText + rightStr + terminal.RestoreCursorPosition()
}

// renderRightTemplate renders the transient prompt's right-aligned template.
// Only zsh (via RPROMPT) and pwsh (via cursor save/restore) can display it.
func (e *Engine) renderRightTemplate(prompt *config.Segment, background, foreground color.Ansi) (string, int) {
	if len(prompt.RightTemplate) == 0 {
		return "", 0
	}

	switch e.Env.Shell() {
	case shell.ZSH, shell.PWSH:
	default:
		return "", 0
	}

	text, err := template.Render(prompt.RightTemplate, nil)
	if err != nil {
		text = err.Error()
	}

	terminal.Write(background, foreground, text)
	return terminal.String()
}
