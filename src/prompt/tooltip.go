package prompt

import (
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func (e *Engine) Tooltip(tip string) string {
	supportedShells := []string{
		shell.ZSH,
		shell.CMD,
		shell.FISH,
		shell.PWSH,
		shell.PWSH5,
		shell.GENERIC,
	}
	if !slices.Contains(supportedShells, e.Env.Shell()) {
		return ""
	}

	tip = strings.Trim(tip, " ")
	tooltips := make([]*config.Segment, 0, 1)

	for _, tooltip := range e.Config.Tooltips {
		if !e.shouldInvokeWithTip(tooltip, tip) {
			continue
		}

		tooltip.SetEnabled(e.Env)

		if !tooltip.Enabled {
			continue
		}

		tooltips = append(tooltips, tooltip)
	}

	if len(tooltips) == 0 {
		return ""
	}

	// little hack to reuse the current logic
	block := &config.Block{
		Alignment: config.Right,
		Segments:  tooltips,
	}
	block.Init(e.Env)
	if !block.Enabled() {
		return ""
	}
	text, length := e.renderBlockSegments(block)

	switch e.Env.Shell() {
	case shell.PWSH, shell.PWSH5:
		e.rprompt = text
		e.currentLineLength = e.Env.Flags().Column
		space, ok := e.canWriteRightBlock(length, true)
		if !ok {
			return ""
		}
		e.write(terminal.SaveCursorPosition())
		e.write(strings.Repeat(" ", space))
		e.write(text)
		e.write(terminal.RestoreCursorPosition())
		return e.string()
	default:
		return text
	}
}

func (e *Engine) shouldInvokeWithTip(segment *config.Segment, tip string) bool {
	for _, t := range segment.Tips {
		if t == tip {
			return true
		}
	}

	return false
}
