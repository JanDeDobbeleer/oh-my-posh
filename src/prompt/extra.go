package prompt

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
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
	var segment *config.Segment

	switch promptType {
	case Debug:
		segment = e.Config.DebugPrompt
	case Transient:
		segment = e.Config.TransientPrompt
	case Valid:
		segment = e.Config.ValidLine
	case Error:
		segment = e.Config.ErrorLine
	case Secondary:
		segment = e.Config.SecondaryPrompt
	}

	if segment == nil {
		segment = &config.Segment{}
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

	promptText, err := template.Render(getTemplate(segment.Template), nil)
	if err != nil {
		promptText = err.Error()
	}

	if promptType == Transient && segment.Newline {
		promptText = fmt.Sprintf("%s%s", e.getNewline(), promptText)
	}

	if promptType == Transient && e.Config.ShellIntegration {
		exitCode, _ := e.Env.StatusCodes()
		e.write(terminal.CommandFinished(exitCode, e.Env.Flags().NoExitCode))
		e.write(terminal.PromptStart())
	}

	foreground := color.Ansi(segment.ForegroundTemplates.FirstMatch(nil, string(segment.Foreground)))
	background := color.Ansi(segment.BackgroundTemplates.FirstMatch(nil, string(segment.Background)))
	terminal.SetColors(background, foreground)
	terminal.Write(background, foreground, promptText)

	str, length := terminal.String()

	if promptType == Transient && len(segment.Filler) != 0 {
		consoleWidth, err := e.Env.TerminalWidth()
		if err == nil || consoleWidth != 0 {
			if padText, OK := e.shouldFill(segment.Filler, consoleWidth-length); OK {
				str += padText
			}
		}
	}

	switch e.Env.Shell() {
	case shell.ZSH:
		if !e.Env.Flags().Eval {
			break
		}

		if promptType == Transient {
			prompt := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(str))
			// empty RPROMPT
			prompt += "\nRPROMPT=''"
			return prompt
		}

		if promptType == Secondary {
			prompt := fmt.Sprintf("_omp_secondary_prompt=%s", shell.QuotePosixStr(str))
			plain := regex.ReplaceAllString(terminal.AnsiRegex, str, "")
			prompt += fmt.Sprintf("\n_omp_secondary_prompt_plain=%s", shell.QuotePosixStr(plain))
			prompt += fmt.Sprintf("\nPOSH_MULTILINE_KEEPPROMPT=%t", segment.MultilineKeepPrompt)
			return prompt
		}
	case shell.PWSH:
		if promptType == Transient {
			// clear the line afterwards to prevent text from being written on the same line
			// see https://github.com/JanDeDobbeleer/oh-my-posh/issues/3628
			return str + terminal.ClearAfter()
		}
	}

	return str
}
