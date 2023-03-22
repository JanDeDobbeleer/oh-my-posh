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

type Text string

func (t Text) Green() Text {
	if plain {
		return t
	}
	return "\x1b[38;2;191;207;240m" + t
}

func (t Text) Red() Text {
	if plain {
		return t
	}
	return "\x1b[38;2;253;122;140m" + t
}

func (t Text) Purple() Text {
	if plain {
		return t
	}
	return "\x1b[38;2;204;137;214m" + t
}

func (t Text) Yellow() Text {
	if plain {
		return t
	}
	return "\x1b[38;2;156;231;201m" + t
}

func (t Text) Bold() Text {
	if plain {
		return t
	}
	return "\x1b[1m" + t
}

func (t Text) Plain() Text {
	if plain {
		return t
	}
	return t + "\033[0m"
}

func (t Text) String() string {
	return string(t)
}

func printLn(lt logType, args ...string) {
	if len(args) == 0 {
		return
	}
	var str Text
	switch lt {
	case debug:
		str = Text("[DEBUG] ").Green()
	case bug:
		str = Text("[ERROR] ").Red()
	case trace:
		str = Text("[TRACE] ").Purple()
	}
	// timestamp 156, 231, 201
	str += Text(time.Now().Format("15:04:05.000") + " ").Yellow().Plain()
	str += Text(args[0])
	str += parseArgs(args...)
	log.WriteString(str.String())
}

func parseArgs(args ...string) Text {
	if len(args) == 1 {
		return "\n"
	}

	// display empty return values as NO DATA
	if len(args[1]) == 0 {
		text := Text(" \u2192").Yellow()
		text += Text(" NO DATA\n").Red().Plain()
		return text
	}

	// print a single line for single output
	splitted := strings.Split(args[1], "\n")
	if len(splitted) == 1 {
		text := Text(" \u2192").Yellow().Plain()
		return Text(fmt.Sprintf("%s %s\n", text, args[1]))
	}

	// indent multiline output with 4 spaces
	var str Text
	str += Text(" \u2193\n").Yellow().Plain()
	for _, line := range splitted {
		str += Text(fmt.Sprintf("    %s\n", line))
	}
	return str
}
