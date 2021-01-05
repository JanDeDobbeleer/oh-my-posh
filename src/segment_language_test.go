package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	universion = "1.3.3.7"
	uni        = "*.uni"
	corn       = "*.corn"
)

type languageArgs struct {
	version            string
	displayVersion     bool
	displayMode        string
	extensions         []string
	enabledExtensions  []string
	commands           []string
	enabledCommands    []string
	versionParam       string
	versionRegex       string
	missingCommandText string
}

func (l *languageArgs) hasvalue(value string, list []string) bool {
	for _, element := range list {
		if element == value {
			return true
		}
	}
	return false
}

func bootStrapLanguageTest(args *languageArgs) *language {
	env := new(MockedEnvironment)
	for _, command := range args.commands {
		env.On("hasCommand", command).Return(args.hasvalue(command, args.enabledCommands))
		env.On("runCommand", command, []string{args.versionParam}).Return(args.version, nil)
	}
	for _, extension := range args.extensions {
		env.On("hasFiles", extension).Return(args.hasvalue(extension, args.enabledExtensions))
	}
	props := &properties{
		values: map[Property]interface{}{
			DisplayVersion:      args.displayVersion,
			DisplayModeProperty: args.displayMode,
		},
	}
	if args.missingCommandText != "" {
		props.values[MissingCommandTextProperty] = args.missingCommandText
	}
	l := &language{
		props:        props,
		env:          env,
		extensions:   args.extensions,
		commands:     args.commands,
		versionParam: args.versionParam,
		versionRegex: args.versionRegex,
	}
	return l
}

func TestLanguageFilesFoundButNoCommandAndVersionAndDisplayVersion(t *testing.T) {
	args := &languageArgs{
		commands:          []string{"unicorn"},
		versionParam:      "--version",
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.False(t, lang.enabled(), "unicorn is not available")
}

func TestLanguageFilesFoundButNoCommandAndVersionAndDontDisplayVersion(t *testing.T) {
	args := &languageArgs{
		commands:          []string{"unicorn"},
		versionParam:      "--version",
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		displayVersion:    false,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled(), "unicorn is not available")
}

func TestLanguageFilesFoundButNoCommandAndNoVersion(t *testing.T) {
	args := &languageArgs{
		commands:          []string{"unicorn"},
		versionParam:      "--version",
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled(), "unicorn is not available")
}

func TestLanguageDisabledNoFiles(t *testing.T) {
	args := &languageArgs{
		versionParam:    "--version",
		commands:        []string{"unicorn"},
		enabledCommands: []string{"unicorn"},
		extensions:      []string{uni},
	}
	lang := bootStrapLanguageTest(args)
	assert.False(t, lang.enabled(), "no files in the current directory")
}

func TestLanguageEnabledOneExtensionFound(t *testing.T) {
	args := &languageArgs{
		versionParam:      "--version",
		commands:          []string{"unicorn"},
		enabledCommands:   []string{"unicorn"},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni},
		versionRegex:      "(?P<version>.*)",
		version:           universion,
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and uni files are found")
}

func TestLanguageEnabledSecondExtensionFound(t *testing.T) {
	args := &languageArgs{
		versionParam:      "--version",
		commands:          []string{"unicorn"},
		enabledCommands:   []string{"unicorn"},
		extensions:        []string{uni, corn},
		versionRegex:      "(?P<version>.*)",
		version:           universion,
		enabledExtensions: []string{corn},
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and corn files are found")
}

func TestLanguageEnabledSecondCommand(t *testing.T) {
	args := &languageArgs{
		versionParam:      "--version",
		commands:          []string{"uni", "corn"},
		enabledCommands:   []string{"corn"},
		extensions:        []string{uni, corn},
		versionRegex:      "(?P<version>.*)",
		version:           universion,
		enabledExtensions: []string{corn},
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and corn files are found")
}

func TestLanguageEnabledAllExtensionsFound(t *testing.T) {
	args := &languageArgs{
		versionParam:      "--version",
		commands:          []string{"unicorn"},
		enabledCommands:   []string{"unicorn"},
		extensions:        []string{uni, corn},
		versionRegex:      "(?P<version>.*)",
		version:           universion,
		enabledExtensions: []string{uni, corn},
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledNoVersion(t *testing.T) {
	args := &languageArgs{
		versionParam:      "--version",
		commands:          []string{"unicorn"},
		enabledCommands:   []string{"unicorn"},
		extensions:        []string{uni, corn},
		versionRegex:      "(?P<version>.*)",
		version:           universion,
		enabledExtensions: []string{uni, corn},
		displayVersion:    false,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.string(), "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommand(t *testing.T) {
	args := &languageArgs{
		versionParam:      "--version",
		commands:          []string{""},
		enabledCommands:   []string{"unicorn"},
		extensions:        []string{uni, corn},
		versionRegex:      "(?P<version>.*)",
		version:           universion,
		enabledExtensions: []string{uni, corn},
		displayVersion:    false,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.string(), "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommandCustomText(t *testing.T) {
	args := &languageArgs{
		versionParam:       "--version",
		commands:           []string{""},
		enabledCommands:    []string{"unicorn"},
		extensions:         []string{uni, corn},
		versionRegex:       "(?P<version>.*)",
		version:            universion,
		enabledExtensions:  []string{uni, corn},
		displayVersion:     false,
		missingCommandText: "missing",
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, args.missingCommandText, lang.string(), "unicorn is available and uni and corn files are found")
}
