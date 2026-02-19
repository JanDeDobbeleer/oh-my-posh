package prompt

import "github.com/jandedobbeleer/oh-my-posh/src/config"

// TmuxStatusLeft renders the tmux status-left section from the config's tmux.status_left blocks.
func (e *Engine) TmuxStatusLeft() string {
	if e.Config.Tmux == nil {
		return ""
	}

	return e.renderTmuxSection(e.Config.Tmux.StatusLeft.Blocks)
}

// TmuxStatusRight renders the tmux status-right section from the config's tmux.status_right blocks.
func (e *Engine) TmuxStatusRight() string {
	if e.Config.Tmux == nil {
		return ""
	}

	return e.renderTmuxSection(e.Config.Tmux.StatusRight.Blocks)
}

// renderTmuxSection renders a slice of blocks using the existing block rendering pipeline
// and returns the concatenated result. Shell integration sequences are intentionally
// skipped here since tmux status bars do not use OSC 133 marks.
func (e *Engine) renderTmuxSection(blocks []*config.Block) string {
	if len(blocks) == 0 {
		return ""
	}

	cycle = &e.Config.Cycle

	for _, block := range blocks {
		e.renderBlock(block, true)
	}

	return e.string()
}
