package engine

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/shell"
)

func (e *Engine) Tooltip(tip string) string {
	tip = strings.Trim(tip, " ")
	var tooltip *Segment
	for _, tp := range e.Config.Tooltips {
		if !tp.shouldInvokeWithTip(tip) {
			continue
		}
		tooltip = tp
	}

	if tooltip == nil {
		return ""
	}

	if err := tooltip.mapSegmentWithWriter(e.Env); err != nil {
		return ""
	}

	if !tooltip.writer.Enabled() {
		return ""
	}

	tooltip.Enabled = true

	// little hack to reuse the current logic
	block := &Block{
		Alignment: Right,
		Segments:  []*Segment{tooltip},
	}

	switch e.Env.Shell() {
	case shell.ZSH, shell.CMD, shell.FISH, shell.GENERIC:
		block.Init(e.Env, e.Writer)
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
		text, length := block.RenderSegments()
		// clear from cursor to the end of the line in case a previous tooltip
		// is cut off and partially preserved, if the new one is shorter
		e.write(e.Writer.ClearAfter())
		e.write(e.Writer.CarriageForward())
		e.write(e.Writer.GetCursorForRightWrite(length, 0))
		e.write(text)
		return e.string()
	}

	return ""
}
