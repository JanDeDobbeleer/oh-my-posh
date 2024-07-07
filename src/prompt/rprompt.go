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
	filterRPromptBlock := func(blocks []*config.Block) *config.Block {
		for _, block := range blocks {
			if block.Type == config.RPrompt {
				return block
			}
		}
		return nil
	}

	block := filterRPromptBlock(e.Config.Blocks)
	if block == nil {
		return ""
	}

	if e.Env.Shell() == shell.BASH {
		terminal.Init(shell.GENERIC)
	}

	block.Init(e.Env)

	if !block.Enabled() {
		return ""
	}

	text, length := e.renderBlockSegments(block)
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

	return text
}
