package runtime

// reduceTerminalWidth narrows width by less columns, e.g. when the host only
// exposes part of the terminal to the prompt. It leaves width untouched when
// the reduction would not leave at least one column.
func reduceTerminalWidth(width, less int) int {
	if less <= 0 || less >= width {
		return width
	}

	return width - less
}
