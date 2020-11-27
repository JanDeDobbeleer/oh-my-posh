package main

import "fmt"

type exit struct {
	props *properties
	env   environmentInfo
}

const (
	// DisplayExitCode shows or hides the error code
	DisplayExitCode Property = "display_exit_code"
	// AlwaysEnabled decides whether or not to always display the exitcode info
	AlwaysEnabled Property = "always_enabled"
	// ErrorColor specify a different foreground color for the error text when using always_show = true
	ErrorColor Property = "error_color"
	// AlwaysNumeric shows error codes as numbers
	AlwaysNumeric Property = "always_numeric"
)

func (e *exit) enabled() bool {
	if e.props.getBool(AlwaysEnabled, false) {
		return true
	}
	return e.env.lastErrorCode() != 0
}

func (e *exit) string() string {
	return e.getFormattedText()
}

func (e *exit) init(props *properties, env environmentInfo) {
	e.props = props
	e.env = env
}

func (e *exit) getFormattedText() string {
	exitCode := e.getMeaningFromExitCode()
	colorBackground := e.props.getBool(ColorBackground, false)
	if e.env.lastErrorCode() != 0 && !colorBackground {
		e.props.foreground = e.props.getColor(ErrorColor, e.props.foreground)
	}
	if e.env.lastErrorCode() != 0 && colorBackground {
		e.props.background = e.props.getColor(ErrorColor, e.props.background)
	}
	return exitCode
}

func (e *exit) getMeaningFromExitCode() string {
	if !e.props.getBool(DisplayExitCode, true) {
		return ""
	}
	if e.props.getBool(AlwaysNumeric, false) {
		return fmt.Sprintf("%d", e.env.lastErrorCode())
	}
	switch e.env.lastErrorCode() {
	case 1:
		return "ERROR"
	case 2:
		return "USAGE"
	case 126:
		return "NOPERM"
	case 127:
		return "NOTFOUND"
	case 128 + 1:
		return "SIGHUP"
	case 128 + 2:
		return "SIGINT"
	case 128 + 3:
		return "SIGQUIT"
	case 128 + 4:
		return "SIGILL"
	case 128 + 5:
		return "SIGTRAP"
	case 128 + 6:
		return "SIGIOT"
	case 128 + 7:
		return "SIGBUS"
	case 128 + 8:
		return "SIGFPE"
	case 128 + 9:
		return "SIGKILL"
	case 128 + 10:
		return "SIGUSR1"
	case 128 + 11:
		return "SIGSEGV"
	case 128 + 12:
		return "SIGUSR2"
	case 128 + 13:
		return "SIGPIPE"
	case 128 + 14:
		return "SIGALRM"
	case 128 + 15:
		return "SIGTERM"
	case 128 + 16:
		return "SIGSTKFLT"
	case 128 + 17:
		return "SIGCHLD"
	case 128 + 18:
		return "SIGCONT"
	case 128 + 19:
		return "SIGSTOP"
	case 128 + 20:
		return "SIGTSTP"
	case 128 + 21:
		return "SIGTTIN"
	case 128 + 22:
		return "SIGTTOU"
	default:
		return fmt.Sprintf("%d", e.env.lastErrorCode())
	}
}
