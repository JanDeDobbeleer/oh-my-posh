package prompt

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func (e *Engine) RPrompt() string {
	var rprompt *config.Block
	var lineCount int

	for _, block := range e.Config.Blocks {
		if block.Type == config.RPrompt {
			rprompt = block
		}

		if block.Newline || block.Type == config.LineBreak {
			lineCount++
		}
	}

	if rprompt == nil {
		return ""
	}

	if e.Env.Shell() == shell.BASH {
		terminal.Init(shell.GENERIC)
	}

	rprompt.Init(e.Env)

	if !rprompt.Enabled() {
		return ""
	}

	text, length := e.renderBlockSegments(rprompt)
	e.rpromptLength = length

	if e.Env.Shell() != shell.BASH {
		return text
	}

	width, err := e.Env.TerminalWidth()
	if err != nil {
		log.Error(err)
		return ""
	}

	padding := width - e.rpromptLength
	if padding < 0 {
		padding = 0
	}

	text = fmt.Sprintf("%s%s\r", strings.Repeat(" ", padding), text)

	// bash prints this on the same line as the prompt so we need to move the cursor down
	// in case the prompt spans multiple lines
	if lineCount > 0 {
		return terminal.SaveCursorPosition() + strings.Repeat("\n", lineCount) + text + terminal.RestoreCursorPosition()
	}

	return text
}
