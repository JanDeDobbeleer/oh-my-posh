package engine

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func (e *Engine) Tooltip(tip string) string {
	tip = strings.Trim(tip, " ")
	tooltips := make([]*Segment, 0, 1)

	for _, tooltip := range e.Config.Tooltips {
		if !tooltip.shouldInvokeWithTip(tip) {
			continue
		}

		if err := tooltip.mapSegmentWithWriter(e.Env); err != nil {
			continue
		}

		if !tooltip.writer.Enabled() {
			continue
		}

		tooltips = append(tooltips, tooltip)
	}

	if len(tooltips) == 0 {
		return ""
	}

	// little hack to reuse the current logic
	block := &Block{
		Alignment: Right,
		Segments:  tooltips,
	}

	switch e.Env.Shell() {
	case shell.ZSH, shell.CMD, shell.FISH, shell.GENERIC:
		block.Init(e.Env)
		if !block.Enabled() {
			return ""
		}
		text, _ := block.RenderSegments()
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

		text, length := block.RenderSegments()

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
