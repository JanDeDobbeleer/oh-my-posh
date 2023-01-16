package log

import (
	"fmt"
	"strings"
	"time"
)

type logType byte

const (
	debug logType = 1 << iota
	bug
	trace
)

func printLn(lt logType, args ...string) {
	if len(args) == 0 {
		return
	}
	var str string
	switch lt {
	case debug:
		str = "\x1b[38;2;191;207;240m[DEBUG] "
	case bug:
		str = "\x1b[38;2;253;122;140m[ERROR] "
	case trace:
		str = "\x1b[38;2;204;137;214m[TRACE] "
	}
	// timestamp 156, 231, 201
	str += fmt.Sprintf("\x1b[38;2;156;231;201m%s ", time.Now().Format("15:04:05.000"))
	str += "\033[0m"
	str += args[0]
	str += parseArgs(args...)
	log.WriteString(str)
}

func parseArgs(args ...string) string {
	if len(args) == 1 {
		return "\n"
	}

	// display empty return values as NO DATA
	if len(args[1]) == 0 {
		return " \x1b[38;2;156;231;201m\u2192\033[0m \x1b[38;2;253;122;140mNO DATA\033[0m\n"
	}

	// print a single line for single output
	splitted := strings.Split(args[1], "\n")
	if len(splitted) == 1 {
		return fmt.Sprintf(" \x1b[38;2;156;231;201m\u2192\033[0m %s\n", args[1])
	}

	// indent multiline output with 4 spaces
	var str string
	str += " \x1b[38;2;156;231;201m\u2193\033[0m\n"
	for _, line := range splitted {
		str += fmt.Sprintf("    %s\n", line)
	}
	return str
}
