package main

import (
	"errors"
	"fmt"
)

type loadContext func()

type inContext func() bool

type getVersion func() (string, error)
type matchesVersionFile func() bool

type version struct {
	Full          string
	Major         string
	Minor         string
	Patch         string
	Prerelease    string
	BuildMetadata string
	URL           string
}

type cmd struct {
	executable string
	args       []string
	regex      string
	getVersion getVersion
}

func (c *cmd) parse(versionInfo string) (*version, error) {
	values := findNamedRegexMatch(c.regex, versionInfo)
	if len(values) == 0 {
		return nil, errors.New("cannot parse version string")
	}

	version := &version{
		Full:          values["version"],
		Major:         values["major"],
		Minor:         values["minor"],
		Patch:         values["patch"],
		Prerelease:    values["prerelease"],
		BuildMetadata: values["buildmetadata"],
	}
	return version, nil
}

type language struct {
	props              Properties
	env                Environment
	extensions         []string
	commands           []*cmd
	versionURLTemplate string
	exitCode           int
	loadContext        loadContext
	inContext          inContext
	matchesVersionFile matchesVersionFile
	homeEnabled        bool
	displayMode        string

	version
	Error    string
	Mismatch bool
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
	// MissingCommandText sets the text to display when the command is not present in the system
	MissingCommandText Property = "missing_command_text"
	// HomeEnabled displays the segment in the HOME folder or not
	HomeEnabled Property = "home_enabled"
	// LanguageExtensions the list of extensions to validate
	LanguageExtensions Property = "extensions"
)

func (l *language) renderTemplate(segmentTemplate string, context SegmentWriter) string {
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  context,
		Env:      l.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}

func (l *language) enabled() bool {
	// override default extensions if needed
	l.extensions = l.props.getStringArray(LanguageExtensions, l.extensions)
	inHomeDir := func() bool {
		return l.env.pwd() == l.env.homeDir()
	}
	var enabled bool
	homeEnabled := l.props.getBool(HomeEnabled, l.homeEnabled)
	if inHomeDir() && !homeEnabled {
		enabled = false
	} else {
		// set default mode when not set
		if len(l.displayMode) == 0 {
			l.displayMode = l.props.getString(DisplayMode, DisplayModeFiles)
		}
		l.loadLanguageContext()
		switch l.displayMode {
		case DisplayModeAlways:
			enabled = true
		case DisplayModeEnvironment:
			enabled = l.inLanguageContext()
		case DisplayModeFiles:
			enabled = l.hasLanguageFiles()
		case DisplayModeContext:
			fallthrough
		default:
			enabled = l.hasLanguageFiles() || l.inLanguageContext()
		}
	}
	if !enabled || !l.props.getOneOfBool(FetchVersion, DisplayVersion, true) {
		return enabled
	}
	err := l.setVersion()
	if err != nil {
		l.Error = err.Error()
	}
	return enabled
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
		var versionStr string
		var err error
		if command.getVersion == nil {
			if !l.env.hasCommand(command.executable) {
				continue
			}
			versionStr, err = l.env.runCommand(command.executable, command.args...)
			if exitErr, ok := err.(*commandError); ok {
				l.exitCode = exitErr.exitCode
				return fmt.Errorf("err executing %s with %s", command.executable, command.args)
			}
		} else {
			versionStr, err = command.getVersion()
			if err != nil {
				return err
			}
		}
		if versionStr == "" {
			continue
		}
		version, err := command.parse(versionStr)
		if err != nil {
			return fmt.Errorf("err parsing info from %s with %s", command.executable, versionStr)
		}
		l.version = *version
		l.buildVersionURL()
		return nil
	}
	return errors.New(l.props.getString(MissingCommandText, ""))
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

func (l *language) setVersionFileMismatch() {
	if l.matchesVersionFile == nil {
		return
	}
	l.Mismatch = !l.matchesVersionFile()
	if !l.Mismatch {
		return
	}
	l.colorMismatch()
}

func (l *language) buildVersionURL() {
	versionURLTemplate := l.props.getString(VersionURLTemplate, l.versionURLTemplate)
	if len(versionURLTemplate) == 0 {
		return
	}
	template := &textTemplate{
		Template: versionURLTemplate,
		Context:  l.version,
		Env:      l.env,
	}
	url, err := template.render()
	if err != nil {
		return
	}
	l.version.URL = url
}
