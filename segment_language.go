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

func (l *language) string() string {
	if l.props.getBool(DisplayVersion, true) {
		return l.version
	}
	return ""
}

func (l *language) enabled() bool {
	for i, extension := range l.extensions {
		if l.env.hasFiles(extension) {
			break
		}
		if i == len(l.extensions)-1 {
			return false
		}
	}
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
