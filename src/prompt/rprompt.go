package prompt

import (
	"strconv"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

const (
	RPromptKey       = "rprompt"
	RPromptLengthKey = "rprompt_length"
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

	text, length := e.writeBlockSegments(rprompt)

	// do not print anything when we don't have any text
	if length == 0 {
		return ""
	}

	e.rpromptLength = length

	if e.Env.Shell() == shell.ELVISH && e.Env.GOOS() != runtime.WINDOWS {
		// Workaround to align with a right-aligned block on non-Windows systems.
		text += " "
	}

	if !e.Config.ToolTipsAction.IsDefault() {
		e.Env.Session().Set(RPromptKey, text, cache.INFINITE)
		e.Env.Session().Set(RPromptLengthKey, strconv.Itoa(e.rpromptLength), cache.INFINITE)
	}

	return text
}
