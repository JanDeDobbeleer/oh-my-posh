package prompt

import (
	"strconv"
	"strings"

	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func (e *Engine) Tooltip(tip string) string {
	tip = strings.Trim(tip, " ")
	tooltips := make([]*config.Segment, 0, 1)

	for _, tooltip := range e.Config.Tooltips {
		if !slices.Contains(tooltip.Tips, tip) {
			continue
		}

		tooltip.Execute(e.Env)

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

	text, length := e.writeBlockSegments(block)

	// do not print anything when we don't have any text
	if length == 0 {
		return ""
	}

	text, length = e.handleToolTipAction(text, length)

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

func (e *Engine) handleToolTipAction(text string, length int) (string, int) {
	if e.Config.ToolTipsAction.IsDefault() {
		return text, length
	}

	rprompt, OK := e.Env.Cache().Get(RPromptKey)
	if !OK {
		return text, length
	}

	rpromptLengthStr, OK := e.Env.Cache().Get(RPromptLengthKey)
	if !OK {
		return text, length
	}

	rpromptLength, err := strconv.Atoi(rpromptLengthStr)
	if err != nil {
		return text, length
	}

	length += rpromptLength

	switch e.Config.ToolTipsAction {
	case config.Extend:
		text = rprompt + text
	case config.Prepend:
		text += rprompt
	}

	return text, length
}
