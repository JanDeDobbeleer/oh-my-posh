package runtime

import (
	"strconv"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

const maxColumnsLessEnv = "POSH_MAX_COLUMNS_LESS"

// reducedWidth keeps CmdFlags.TerminalWidth raw: callers adjust it directly
// (see Engine.rectifyTerminalWidth), so the reduction must happen on every read.
// A reduction that would not leave one column is ignored.
func (term *Terminal) reducedWidth() int {
	width := term.CmdFlags.TerminalWidth
	less := term.CmdFlags.MaxColumnsLess

	if less > 0 && less < width {
		width -= less
	}

	log.Debugf("terminal width: %d", width)

	return width
}

// setMaxColumnsLess lets every shell and render path narrow the width without
// passing a flag on each call.
func (term *Terminal) setMaxColumnsLess() {
	if term.CmdFlags.MaxColumnsLess != 0 {
		return
	}

	less, err := strconv.Atoi(term.Getenv(maxColumnsLessEnv))
	if err != nil {
		return
	}

	term.CmdFlags.MaxColumnsLess = less
}
