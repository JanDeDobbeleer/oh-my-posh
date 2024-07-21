package prompt

import (
	"github.com/jandedobbeleer/oh-my-posh/src/config"
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

	return text
}
