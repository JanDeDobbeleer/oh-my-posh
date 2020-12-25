package main

type language struct {
	props        *properties
	env          environmentInfo
	extensions   []string
	commands     []string
	versionParam string
	versionRegex string
	version      string
}

const (
	// DisplayModeProperty sets the display mode (always, when_in_context, never)
	DisplayModeProperty Property = "display_mode"
	// DisplayModeAlways displays the segement always
	DisplayModeAlways string = "always"
	// DisplayModeContext displays the segment when the current folder contains certain extensions
	DisplayModeContext string = "context"
	// DisplayModeNever hides the segment
	DisplayModeNever string = "never"
)

func (l *language) string() string {
	if l.props.getBool(DisplayVersion, true) {
		return l.version
	}
	return ""
}

func (l *language) enabled() bool {
	displayMode := l.props.getString(DisplayModeProperty, DisplayModeContext)
	displayVersion := l.props.getBool(DisplayVersion, true)
	hasVersion := l.getVersion()

	switch displayMode {
	case DisplayModeAlways:
		return (hasVersion || !displayVersion)
	case DisplayModeNever:
		return false
	case DisplayModeContext:
		fallthrough
	default:
		return l.isInContext() && (hasVersion || !displayVersion)
	}
}

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

func (l *language) getVersion() bool {
	var executable string
	for i, command := range l.commands {
		if l.env.hasCommand(command) {
			executable = command
			break
		}
		if i == len(l.commands)-1 {
			return false
		}
	}
	versionInfo, _ := l.env.runCommand(executable, l.versionParam)
	values := findNamedRegexMatch(l.versionRegex, versionInfo)
	l.version = values["version"]

	return true
}
