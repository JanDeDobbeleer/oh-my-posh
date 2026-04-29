package prompt

import (
	"fmt"
	"strings"

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

	promptText, err := template.Render(getTemplate(prompt.Template), e)
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
		if !e.Env.Flags().Eval {
			break
		}

		if promptType == Transient {
			evalOutput := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(str))
			// empty RPROMPT
			evalOutput += "\nRPROMPT=''"
			return evalOutput
		}

		if promptType == Secondary {
			evalOutput := fmt.Sprintf("_omp_secondary_prompt=%s", shell.QuotePosixStr(str))
			plain := regex.ReplaceAllString(terminal.AnsiRegex, str, "")
			plain = strings.ReplaceAll(plain, "%{", "")
			plain = strings.ReplaceAll(plain, "%}", "")
			evalOutput += fmt.Sprintf("\n_omp_secondary_prompt_plain=%s", shell.QuotePosixStr(plain))
			evalOutput += fmt.Sprintf("\nPOSH_MULTILINE_KEEPPROMPT=%t", prompt.MultilineKeepPrompt)
			return evalOutput
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
