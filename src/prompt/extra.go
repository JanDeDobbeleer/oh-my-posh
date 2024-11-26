package prompt

import (
	"fmt"

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
	}

	promptText, err := tmpl.Render()
	if err != nil {
		promptText = err.Error()
	}

	if promptType == Transient && prompt.Newline {
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

	if promptType == Transient && len(prompt.Filler) != 0 {
		consoleWidth, err := e.Env.TerminalWidth()
		if err == nil || consoleWidth != 0 {
			if padText, OK := e.shouldFill(prompt.Filler, consoleWidth-length); OK {
				str += padText
			}
		}
	}

	switch e.Env.Shell() {
	case shell.ZSH:
		if promptType == Transient {
			if !e.Env.Flags().Eval {
				break
			}

			prompt := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(str))
			// empty RPROMPT
			prompt += "\nRPROMPT=''"
			return prompt
		}
	case shell.PWSH, shell.PWSH5:
		if promptType == Transient {
			// clear the line afterwards to prevent text from being written on the same line
			// see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3628
			return str + terminal.ClearAfter()
		}
	}

	return str
}
