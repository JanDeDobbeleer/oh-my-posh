package prompt

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

func (e *Engine) Tooltip(tip string) string {
	tip = strings.Trim(tip, " ")

	// Check if we have any matching tooltips and if they have cache configured
	var matchingTooltips []*config.Segment
	var cacheableTooltips []*config.Segment

	for _, tooltip := range e.Config.Tooltips {
		if !slices.Contains(tooltip.Tips, tip) {
			continue
		}
		matchingTooltips = append(matchingTooltips, tooltip)
		if tooltip.Cache != nil && !tooltip.Cache.Duration.IsEmpty() {
			cacheableTooltips = append(cacheableTooltips, tooltip)
		}
	}

	// If we have cacheable tooltips, try to get from cache first
	if len(cacheableTooltips) > 0 {
		cacheKey := e.getTooltipCacheKey(tip)
		if cachedOutput, ok := cache.Get[string](cache.Session, cacheKey); ok {
			return cachedOutput
		}
	}

	tooltips := make([]*config.Segment, 0, len(matchingTooltips))

	for _, tooltip := range matchingTooltips {
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

	var finalOutput string

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
		finalOutput = e.string()
	default:
		finalOutput = text
	}

	// Cache the final output if any matching tooltip has cache configured
	if len(cacheableTooltips) > 0 && finalOutput != "" {
		cacheKey := e.getTooltipCacheKey(tip)
		// Use the shortest cache duration among all cacheable tooltips
		minDuration := cacheableTooltips[0].Cache.Duration
		for _, tooltip := range cacheableTooltips[1:] {
			if tooltip.Cache.Duration < minDuration {
				minDuration = tooltip.Cache.Duration
			}
		}
		cache.Set(cache.Session, cacheKey, finalOutput, minDuration)
	}

	return finalOutput
}

func (e *Engine) getTooltipCacheKey(tip string) string {
	return fmt.Sprintf("tooltip_cache_%s", tip)
}

func (e *Engine) handleToolTipAction(text string, length int) (string, int) {
	if e.Config.ToolTipsAction.IsDefault() {
		return text, length
	}

	rprompt, OK := cache.Get[string](cache.Session, RPromptKey)
	if !OK {
		return text, length
	}

	rpromptLength, OK := cache.Get[int](cache.Session, RPromptLengthKey)
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
