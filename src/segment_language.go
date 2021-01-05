package main

import "errors"

type language struct {
	props        *properties
	env          environmentInfo
	extensions   []string
	commands     []string
	executable   string
	versionParam string
	versionRegex string
	version      string
	exitCode     int
}

const (
	// DisplayModeProperty sets the display mode (always, when_in_context, never)
	DisplayModeProperty Property = "display_mode"
	// DisplayModeAlways displays the segment always
	DisplayModeAlways string = "always"
	// DisplayModeContext displays the segment when the current folder contains certain extensions
	DisplayModeContext string = "context"
	// MissingCommandTextProperty sets the text to display when the command is not present in the system
	MissingCommandTextProperty Property = "missing_command_text"
	// MissingCommandText displays empty string by default
	MissingCommandText string = ""
)

func (l *language) string() string {
	// check if one of the defined commands exists in the system
	if !l.hasCommand() {
		return l.props.getString(MissingCommandTextProperty, MissingCommandText)
	}

	// call getVersion if displayVersion set in config
	if l.props.getBool(DisplayVersion, true) && l.getVersion() {
		return l.version
	}
	return ""
}

func (l *language) enabled() bool {
	displayMode := l.props.getString(DisplayModeProperty, DisplayModeContext)
	displayVersion := l.props.getBool(DisplayVersion, true)

	switch displayMode {
	case DisplayModeAlways:
		return (!displayVersion || l.hasCommand())
	case DisplayModeContext:
		fallthrough
	default:
		return l.isInContext() && (!displayVersion || l.hasCommand())
	}
}

// isInContext will return true at least one file matching the extensions is found
func (l *language) isInContext() bool {
	for i, extension := range l.extensions {
		if l.env.hasFiles(extension) {
			break
		}
		if i == len(l.extensions)-1 {
			return false
		}
	}

	return true
}

// getVersion returns the version and exit code returned by the executable
func (l *language) getVersion() bool {
	versionInfo, err := l.env.runCommand(l.executable, l.versionParam)
	var exerr *commandError
	if err == nil {
		values := findNamedRegexMatch(l.versionRegex, versionInfo)
		l.exitCode = 0
		l.version = values["version"]
	} else {
		errors.As(err, &exerr)
		l.exitCode = exerr.exitCode
		l.version = ""
	}
	return true
}

// hasCommand checks if one of the commands exists and sets it as executable
func (l *language) hasCommand() bool {
	for i, command := range l.commands {
		commandPath, commandExists := l.env.hasCommand(command)
		if commandExists {
			l.executable = commandPath
			break
		}
		if i == len(l.commands)-1 {
			return false
		}
	}
	return true
}
