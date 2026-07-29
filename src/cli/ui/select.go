package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrCancelled is what a caller gets when the user backs out rather than choosing. It is not a
// failure: the caller should exit quietly, the way pressing escape used to.
var ErrCancelled = errors.New("cancelled")

// Select prints a numbered list and reads a choice.
//
// Numbered rather than arrow-driven on purpose. Arrow keys need raw mode, and raw mode is the one
// genuinely risky thing in this whole package: it means termios on POSIX, SetConsoleMode with
// ENABLE_VIRTUAL_TERMINAL_INPUT on Windows, restoring both if the process panics or is killed,
// and a fallback for the MSYS and Cygwin ptys where stdin is a pipe and none of it applies. A
// number and Enter needs none of that and works identically everywhere, including when stdin is
// redirected - which is how anyone scripts an install.
func Select(in io.Reader, out io.Writer, prompt string, items []string) (int, error) {
	if len(items) == 0 {
		return 0, errors.New("nothing to choose from")
	}

	// The widest index decides the column, so the labels line up whether there are 9 entries or
	// 900. Counting digits rather than measuring text: an index is always ASCII, so no cell-width
	// question arises - which is the whole reason this package needs no Unicode width table.
	width := len(strconv.Itoa(len(items)))

	fmt.Fprintf(out, "%s\n\n", prompt)

	for i, item := range items {
		fmt.Fprintf(out, "  %*d  %s\n", width, i+1, item)
	}

	fmt.Fprintf(out, "\nEnter a number between 1 and %d, or press Enter to cancel: ", len(items))

	reader := bufio.NewReader(in)

	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		// EOF with nothing typed is the same intent as an empty line: the user, or the script,
		// declined to choose.
		return 0, ErrCancelled
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return 0, ErrCancelled
	}

	choice, err := strconv.Atoi(line)
	if err != nil || choice < 1 || choice > len(items) {
		return 0, fmt.Errorf("%q is not a number between 1 and %d", line, len(items))
	}

	return choice - 1, nil
}
