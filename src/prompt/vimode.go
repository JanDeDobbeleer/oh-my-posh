package prompt

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
)

// primaryVariants renders the primary prompt for every requested vi mode and returns the prompts
// as a JSON object keyed by mode. Shells that cannot report a mode change (Nushell) swap between
// the variants themselves, so a mode switch costs no render at all. Only the vimode segments run
// again per mode; everything else is rendered from the results of the first pass.
func (e *Engine) primaryVariants(modes []string, fromCache bool) string {
	previous := e.Env.Getenv(segments.VIModeEnv)
	defer e.Env.Setenv(segments.VIModeEnv, previous)

	prompts := make(map[string]string, len(modes))

	for i, mode := range modes {
		e.Env.Setenv(segments.VIModeEnv, mode)

		if i == 0 && !fromCache {
			prompts[mode] = e.primaryPrompt(false)
			continue
		}

		e.refreshVIModeSegments()
		e.resetPromptBuilder()
		prompts[mode] = e.primaryPrompt(true)
	}

	var out bytes.Buffer

	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(prompts); err != nil {
		log.Error(err)
		return ""
	}

	return strings.TrimSuffix(out.String(), "\n")
}

func (e *Engine) refreshVIModeSegments() {
	for _, block := range e.Config.Blocks {
		for _, segment := range block.Segments {
			if segment.Type != config.VIMODE {
				continue
			}

			segment.Refresh(e.Env)
		}
	}
}
