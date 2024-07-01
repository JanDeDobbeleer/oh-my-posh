package engine

import (
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
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
	block.Init(e.Env, e.Writer)
	if !block.Enabled() {
		return ""
	}
	text, length := block.RenderSegments()

	switch e.Env.Shell() {
	case shell.PWSH, shell.PWSH5:
		e.rprompt = text
		e.rpromptLength = length
		e.currentLineLength = e.Env.Flags().Column
		space, ok := e.canWriteRightBlock(true)
		if !ok {
			return ""
		}
		e.write(strings.Repeat(" ", space))
		// Workaround to avoid leftover when a previous tooltip is cut off by a shorter new one.
		e.Writer.ClearAfter()
		e.write(text)
		return e.string()
	default:
		return text
	}
}
