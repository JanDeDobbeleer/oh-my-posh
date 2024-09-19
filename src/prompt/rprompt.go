package prompt

import (
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

func (e *Engine) RPrompt() string {
	var rprompt *config.Block

	for _, block := range e.Config.Blocks {
		if block.Type != config.RPrompt {
			continue
		}

		rprompt = block
		break
	}

	if rprompt == nil {
		return ""
	}

	rprompt.Init(e.Env)

	if !rprompt.Enabled() {
		return ""
	}

	text, length := e.renderBlockSegments(rprompt)
	e.rpromptLength = length

	if e.Env.Shell() == shell.ELVISH && e.Env.GOOS() != runtime.WINDOWS {
		// Workaround to align with a right-aligned block on non-Windows systems.
		text += " "
	}

	return text
}
