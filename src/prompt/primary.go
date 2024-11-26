package prompt

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func (e *Engine) Primary() string {
	needsPrimaryRightPrompt := e.needsPrimaryRightPrompt()

	e.writePrimaryPrompt(needsPrimaryRightPrompt)

	switch e.Env.Shell() {
	case shell.ZSH:
		if !e.Env.Flags().Eval {
			break
		}

		// Warp doesn't support RPROMPT so we need to write it manually
		if e.isWarp() {
			e.writePrimaryRightPrompt()
			prompt := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(e.string()))
			return prompt
		}

		prompt := fmt.Sprintf("PS1=%s", shell.QuotePosixStr(e.string()))
		prompt += fmt.Sprintf("\nRPROMPT=%s", shell.QuotePosixStr(e.rprompt))

		return prompt
	default:
		if !needsPrimaryRightPrompt {
			break
		}

		e.writePrimaryRightPrompt()
	}

	return e.string()
}

func (e *Engine) writePrimaryPrompt(needsPrimaryRPrompt bool) {
	if e.Config.ShellIntegration {
		exitCode, _ := e.Env.StatusCodes()
		e.write(terminal.CommandFinished(exitCode, e.Env.Flags().NoExitCode))
		e.write(terminal.PromptStart())
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

		if block.Type == config.RPrompt && !needsPrimaryRPrompt {
			continue
		}

		if e.renderBlock(block, cancelNewline) {
			didRender = true
		}
	}

	if len(e.Config.ConsoleTitleTemplate) > 0 && !e.Env.Flags().Plain {
		title := e.getTitleTemplateText()
		e.write(terminal.FormatTitle(title))
	}

	if e.Config.FinalSpace {
		e.write(" ")
		e.currentLineLength++
	}

	if e.Config.ITermFeatures != nil && e.isIterm() {
		host, _ := e.Env.Host()
		e.write(terminal.RenderItermFeatures(e.Config.ITermFeatures, e.Env.Shell(), e.Env.Pwd(), e.Env.User(), host))
	}

	if e.Config.ShellIntegration {
		e.write(terminal.CommandStart())
	}

	e.pwd()
}

func (e *Engine) needsPrimaryRightPrompt() bool {
	if e.Env.Flags().Debug {
		return true
	}

	switch e.Env.Shell() {
	case shell.PWSH, shell.PWSH5, shell.GENERIC, shell.ZSH:
		return true
	default:
		return false
	}
}

func (e *Engine) writePrimaryRightPrompt() {
	space, OK := e.canWriteRightBlock(e.rpromptLength, true)
	if !OK {
		return
	}

	e.write(terminal.SaveCursorPosition())
	e.write(strings.Repeat(" ", space))
	e.write(e.rprompt)
	e.write(terminal.RestoreCursorPosition())
}
