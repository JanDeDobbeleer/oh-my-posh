package engine

import "oh-my-posh/platform"

func shouldHideForWidth(env platform.Environment, minWidth, maxWidth int) bool {
	if maxWidth == 0 && minWidth == 0 {
		return false
	}
	width, err := env.TerminalWidth()
	if err != nil {
		return false
	}
	if minWidth > 0 && maxWidth > 0 {
		return width < minWidth || width > maxWidth
	}
	if maxWidth > 0 && width > maxWidth {
		return true
	}
	if minWidth > 0 && width < minWidth {
		return true
	}
	return false
}
