package prompt

import (
	"maps"

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
	e.renderedSegments = nil
	if e.needsTooltipContext() {
		rendered, _ := cache.Session.Get[map[string]uint8](renderedSegmentsKey)
		e.renderedSegments = maps.Clone(rendered)
		if e.renderedSegments == nil {
			e.renderedSegments = make(map[string]uint8)
		}
		for name, visibility := range e.renderedSegments {
			visibility &^= renderedRPrompt
			if visibility == 0 {
				delete(e.renderedSegments, name)
				continue
			}
			e.renderedSegments[name] = visibility
		}
		defer e.cacheRenderedSegments()
	}
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
	e.recordRenderedSegments(rprompt)

	if e.Env.Shell() == shell.ELVISH && e.Env.GOOS() != runtime.WINDOWS {
		// Workaround to align with a right-aligned block on non-Windows systems.
		text += " "
	}

	if e.shouldCacheRPrompt() {
		cache.Session.Set(RPromptKey, text, cache.INFINITE)
		cache.Session.Set(RPromptLengthKey, e.rpromptLength, cache.INFINITE)
	}

	return text
}
