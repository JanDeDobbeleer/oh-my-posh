package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	universion = "1.3.307"
	uni        = "*.uni"
	corn       = "*.corn"
)

type languageArgs struct {
	version            string
	displayVersion     bool
	displayMode        string
	extensions         []string
	enabledExtensions  []string
	commands           []*cmd
	enabledCommands    []string
	missingCommandText string
	versionURLTemplate string
	enableHyperlink    bool
	expectedError      error
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
		env.On("hasCommand", command.executable).Return(args.hasvalue(command.executable, args.enabledCommands))
		env.On("runCommand", command.executable, command.args).Return(args.version, args.expectedError)
	}
	for _, extension := range args.extensions {
		env.On("hasFiles", extension).Return(args.hasvalue(extension, args.enabledExtensions))
	}
	props := &properties{
		values: map[Property]interface{}{
			DisplayVersion:  args.displayVersion,
			DisplayMode:     args.displayMode,
			EnableHyperlink: args.enableHyperlink,
		},
	}
	if args.missingCommandText != "" {
		props.values[MissingCommandTextProperty] = args.missingCommandText
	}
	l := &language{
		props:              props,
		env:                env,
		extensions:         args.extensions,
		commands:           args.commands,
		versionURLTemplate: args.versionURLTemplate,
	}
	return l
}

func TestLanguageFilesFoundButNoCommandAndVersionAndDisplayVersion(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, MissingCommandText, lang.string(), "unicorn is not available")
}

func TestLanguageFilesFoundButNoCommandAndVersionAndDontDisplayVersion(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		displayVersion:    false,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled(), "unicorn is not available")
}

func TestLanguageFilesFoundButNoCommandAndNoVersion(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled(), "unicorn is not available")
}

func TestLanguageDisabledNoFiles(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{},
		enabledCommands:   []string{"unicorn"},
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.False(t, lang.enabled(), "no files in the current directory")
}

func TestLanguageEnabledOneExtensionFound(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and uni files are found")
}

func TestLanguageEnabledSecondExtensionFound(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and corn files are found")
}

func TestLanguageEnabledSecondCommand(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{corn},
		enabledCommands:   []string{"corn"},
		version:           universion,
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and corn files are found")
}

func TestLanguageEnabledAllExtensionsFound(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		displayVersion:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.string(), "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledNoVersion(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		displayVersion:    false,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.string(), "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommand(t *testing.T) {
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		displayVersion:    false,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.string(), "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommandCustomText(t *testing.T) {
	args := &languageArgs{
		commands:           []*cmd{},
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{uni, corn},
		enabledCommands:    []string{"unicorn"},
		version:            universion,
		missingCommandText: "missing",
		displayVersion:     true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, args.missingCommandText, lang.string(), "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledCommandExitCode(t *testing.T) {
	expected := 200
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"uni"},
		version:           universion,
		displayVersion:    true,
		expectedError:     &commandError{exitCode: expected},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "err executing uni with [--version]", lang.string())
	assert.Equal(t, expected, lang.exitCode)
}

func TestLanguageHyperlinkEnabled(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "[%s](https://unicor.org/doc/%s.%s.%s)",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		displayVersion:     true,
		enableHyperlink:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "[1.3.307](https://unicor.org/doc/1.3.307)", lang.string())
}

func TestLanguageHyperlinkEnabledWrongRegex(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `wrong`,
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      `wrong`,
			},
		},
		versionURLTemplate: "[%s](https://unicor.org/doc/%s.%s.%s)",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		displayVersion:     true,
		enableHyperlink:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "err parsing info from corn with 1.3.307", lang.string())
}

func TestLanguageHyperlinkEnabledLessParamInTemplate(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "[%s](https://unicor.org/doc/%s)",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		displayVersion:     true,
		enableHyperlink:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "[1.3.307](https://unicor.org/doc/1)", lang.string())
}

func TestLanguageHyperlinkEnabledMoreParamInTemplate(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "[%s](https://unicor.org/doc/%s.%s.%s.%s)",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		displayVersion:     true,
		enableHyperlink:    true,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "1.3.307", lang.string())
}
