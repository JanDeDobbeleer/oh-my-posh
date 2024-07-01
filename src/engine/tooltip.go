package engine

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func (e *Engine) Tooltip(tip string) string {
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

	switch e.Env.Shell() {
	case shell.ZSH, shell.CMD, shell.FISH, shell.GENERIC:
		block.Init(e.Env)
		if !block.Enabled() {
			return ""
		}
		text, _ := e.renderBlockSegments(block)
		return text
	case shell.PWSH, shell.PWSH5:
		block.InitPlain(e.Env, e.Config)
		if !block.Enabled() {
			return ""
		}

		consoleWidth, err := e.Env.TerminalWidth()
		if err != nil || consoleWidth == 0 {
			return ""
		}

		text, length := e.renderBlockSegments(block)

		space := consoleWidth - e.Env.Flags().Column - length
		if space <= 0 {
			return ""
		}
		// clear from cursor to the end of the line in case a previous tooltip
		// is cut off and partially preserved, if the new one is shorter
		e.write(terminal.ClearAfter())
		e.write(strings.Repeat(" ", space))
		e.write(text)
		return e.string()
	}

	return ""
}

func (e *Engine) shouldInvokeWithTip(segment *config.Segment, tip string) bool {
	for _, t := range segment.Tips {
		if t == tip {
			return true
		}
	}

	return false
}
