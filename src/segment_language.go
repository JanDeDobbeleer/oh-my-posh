package main

import (
	"errors"
	"fmt"
	"strings"
)

type loadContext func()

type inContext func() bool

type version struct {
	full  string
	major string
	minor string
	patch string
}

type cmd struct {
	executable string
	args       []string
	regex      string
	version    *version
}

func (c *cmd) parse(versionInfo string) error {
	values := findNamedRegexMatch(c.regex, versionInfo)
	if len(values) == 0 {
		return errors.New("cannot parse version string")
	}
	c.version = &version{}
	c.version.full = values["version"]
	c.version.major = values["major"]
	c.version.minor = values["minor"]
	c.version.patch = values["patch"]
	return nil
}

func (c *cmd) buildVersionURL(template string) string {
	if template == "" {
		return c.version.full
	}
	truncatingSprintf := func(str string, args ...interface{}) (string, error) {
		n := strings.Count(str, "%s")
		if n > len(args) {
			return "", errors.New("Too many parameters")
		}
		if n == 0 {
			return fmt.Sprintf(str, args...), nil
		}
		return fmt.Sprintf(str, args[:n]...), nil
	}
	version, err := truncatingSprintf(template, c.version.full, c.version.major, c.version.minor, c.version.patch)
	if err != nil {
		return c.version.full
	}
	return version
}

type language struct {
	props              *properties
	env                environmentInfo
	extensions         []string
	commands           []*cmd
	versionURLTemplate string
	activeCommand      *cmd
	exitCode           int
	loadContext        loadContext
	inContext          inContext
}

const (
	// DisplayMode sets the display mode (always, when_in_context, never)
	DisplayMode Property = "display_mode"
	// DisplayModeAlways displays the segment always
	DisplayModeAlways string = "always"
	// DisplayModeFiles displays the segment when the current folder contains certain extensions
	DisplayModeFiles string = "files"
	// DisplayModeEnvironment displays the segment when the environment has a language's context
	DisplayModeEnvironment string = "environment"
	// DisplayModeContext displays the segment when the environment or files is active
	DisplayModeContext string = "context"
	// MissingCommandTextProperty sets the text to display when the command is not present in the system
	MissingCommandTextProperty Property = "missing_command_text"
	// MissingCommandText displays empty string by default
	MissingCommandText string = ""
)

func (l *language) string() string {
	if !l.props.getBool(DisplayVersion, true) {
		return ""
	}

	err := l.setVersion()
	if err != nil {
		return err.Error()
	}

	if l.props.getBool(EnableHyperlink, false) {
		return l.activeCommand.buildVersionURL(l.versionURLTemplate)
	}
	return l.activeCommand.version.full
}

func (l *language) enabled() bool {
	l.loadLanguageContext()
	displayMode := l.props.getString(DisplayMode, DisplayModeFiles)
	switch displayMode {
	case DisplayModeAlways:
		return true
	case DisplayModeEnvironment:
		return l.inLanguageContext()
	case DisplayModeFiles:
		return l.hasLanguageFiles()
	case DisplayModeContext:
		fallthrough
	default:
		return l.hasLanguageFiles() || l.inLanguageContext()
	}
}

// hasLanguageFiles will return true at least one file matching the extensions is found
func (l *language) hasLanguageFiles() bool {
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

// setVersion parses the version string returned by the command
func (l *language) setVersion() error {
	for _, command := range l.commands {
		if !l.env.hasCommand(command.executable) {
			continue
		}
		version, err := l.env.runCommand(command.executable, command.args...)
		if exitErr, ok := err.(*commandError); ok {
			l.exitCode = exitErr.exitCode
			return fmt.Errorf("err executing %s with %s", command.executable, command.args)
		}
		err = command.parse(version)
		if err != nil {
			return fmt.Errorf("err parsing info from %s with %s", command.executable, version)
		}
		l.activeCommand = command
		return nil
	}
	return errors.New(l.props.getString(MissingCommandTextProperty, MissingCommandText))
}

func (l *language) loadLanguageContext() {
	if l.loadContext == nil {
		return
	}
	l.loadContext()
}

func (l *language) inLanguageContext() bool {
	if l.inContext == nil {
		return false
	}
	return l.inContext()
}
