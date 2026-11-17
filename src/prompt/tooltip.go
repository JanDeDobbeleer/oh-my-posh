package prompt

import (
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

const renderedSegmentsKey = "rendered_segments"

const (
	renderedPrompt uint8 = 1 << iota
	renderedRPrompt
)

func (e *Engine) needsTooltipContext() bool {
	return slices.ContainsFunc(e.Config.Tooltips, func(tooltip *config.Segment) bool {
		return tooltip.HideIfSegment
	})
}

func (e *Engine) recordRenderedSegments(block *config.Block) {
	if e.renderedSegments == nil || block == nil {
		return
	}

	var visibility uint8
	switch block.Type {
	case config.Prompt:
		visibility = renderedPrompt
	case config.RPrompt:
		visibility = renderedRPrompt
	default:
		return
	}

	for _, segment := range block.Segments {
		if segment.Killed || e.segmentPending(segment) || e.segmentAbandoned(segment) {
			continue
		}

		if segment.Enabled {
			e.renderedSegments[segment.Name()] |= visibility
		}
	}
}

func (e *Engine) cacheRenderedSegments() {
	cache.Session.Set(renderedSegmentsKey, e.renderedSegments, cache.ONEDAY)
	e.renderedSegments = nil
}

func (e *Engine) shouldCacheRPrompt() bool {
	return !e.Config.ToolTipsAction.IsDefault() || e.renderedSegments != nil || e.needsTooltipContext()
}

func (e *Engine) tooltipFallback() string {
	rprompt, OK := cache.Session.Get[string](RPromptKey)
	if !OK {
		return ""
	}

	rpromptLength, OK := cache.Session.Get[int](RPromptLengthKey)
	if !OK {
		return rprompt
	}

	switch e.Env.Shell() {
	case shell.PWSH:
		e.rprompt = rprompt
		e.currentLineLength = e.Env.Flags().Column

		space, ok := e.canWriteRightBlock(rpromptLength, true)
		if !ok {
			return ""
		}

		e.write(terminal.SaveCursorPosition())
		e.write(strings.Repeat(" ", space))
		e.write(rprompt)
		e.write(terminal.RestoreCursorPosition())
		return e.string()
	default:
		return rprompt
	}
}

func (e *Engine) Tooltip(tip string) string {
	tip = strings.Trim(tip, " ")
	tooltips := make([]*config.Segment, 0, 1)
	var rendered map[string]uint8
	contextLoaded := false

	for _, tooltip := range e.Config.Tooltips {
		if !slices.Contains(tooltip.Tips, tip) {
			continue
		}

		if tooltip.HideIfSegment {
			if !contextLoaded {
				rendered, _ = cache.Session.Get[map[string]uint8](renderedSegmentsKey)
				contextLoaded = true
			}
			if rendered[tooltip.Name()] != 0 {
				continue
			}
		}

		tooltip.Execute(e.Env)

		if !tooltip.Enabled {
			continue
		}

		tooltips = append(tooltips, tooltip)
	}

	if len(tooltips) == 0 {
		return e.tooltipFallback()
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
	case shell.PWSH:
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

	rprompt, OK := cache.Session.Get[string](RPromptKey)
	if !OK {
		return text, length
	}

	rpromptLength, OK := cache.Session.Get[int](RPromptLengthKey)
	if !OK {
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
