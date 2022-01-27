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
	extensions         []string
	enabledExtensions  []string
	commands           []*cmd
	enabledCommands    []string
	versionURLTemplate string
	expectedError      error
	properties         Properties
	matchesVersionFile matchesVersionFile
	inHome             bool
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
		env.On("HasCommand", command.executable).Return(args.hasvalue(command.executable, args.enabledCommands))
		env.On("RunCommand", command.executable, command.args).Return(args.version, args.expectedError)
	}
	for _, extension := range args.extensions {
		env.On("HasFiles", extension).Return(args.hasvalue(extension, args.enabledExtensions))
	}
	home := "/usr/home"
	cwd := "/usr/home/project"
	if args.inHome {
		cwd = home
	}
	env.On("Pwd").Return(cwd)
	env.On("Home").Return(home)
	env.On("TemplateCache").Return(&TemplateCache{
		Env: make(map[string]string),
	})
	if args.properties == nil {
		args.properties = properties{}
	}
	l := &language{
		props:              args.properties,
		env:                env,
		extensions:         args.extensions,
		commands:           args.commands,
		versionURLTemplate: args.versionURLTemplate,
		matchesVersionFile: args.matchesVersionFile,
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
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.Error, "unicorn is not available")
}

func TestLanguageFilesFoundButNoCommandAndVersionAndDontDisplayVersion(t *testing.T) {
	props := properties{
		FetchVersion: false,
	}
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		properties:        props,
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
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and uni files are found")
}

func TestLanguageDisabledInHome(t *testing.T) {
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
		inHome:            true,
	}
	lang := bootStrapLanguageTest(args)
	assert.False(t, lang.enabled())
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
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and corn files are found")
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
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and corn files are found")
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
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledNoVersion(t *testing.T) {
	props := properties{
		FetchVersion: false,
	}
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
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.Full, "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommand(t *testing.T) {
	props := properties{
		FetchVersion: false,
	}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.Full, "unicorn is unavailable and uni and corn files are found")
}

func TestLanguageEnabledNoVersionData(t *testing.T) {
	props := properties{
		FetchVersion: true,
	}
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?:Python (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"uni"},
		version:           "",
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.Full)
}

func TestLanguageEnabledMissingCommandCustomText(t *testing.T) {
	expected := "missing"
	props := properties{
		MissingCommandText: expected,
	}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, expected, lang.Error, "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommandCustomTextHideError(t *testing.T) {
	props := properties{
		MissingCommandText: "missing",
		DisplayError:       false,
	}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "", lang.Full)
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
		expectedError:     &commandError{exitCode: expected},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "err executing uni with [--version]", lang.Error)
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
		versionURLTemplate: "https://unicor.org/doc/{{ .Full }}",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		properties:         properties{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "https://unicor.org/doc/1.3.307", lang.version.URL)
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
		versionURLTemplate: "https://unicor.org/doc/{{ .Full }}",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		properties:         properties{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.enabled())
	assert.Equal(t, "err parsing info from corn with 1.3.307", lang.Error)
}

func TestLanguageEnabledInHome(t *testing.T) {
	cases := []struct {
		Case            string
		HomeEnabled     bool
		ExpectedEnabled bool
	}{
		{Case: "Always enabled", HomeEnabled: true, ExpectedEnabled: true},
		{Case: "Context disabled", HomeEnabled: false, ExpectedEnabled: false},
	}
	for _, tc := range cases {
		props := properties{
			HomeEnabled: tc.HomeEnabled,
		}
		args := &languageArgs{
			commands: []*cmd{
				{
					executable: "uni",
					args:       []string{"--version"},
					regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				},
			},
			extensions:        []string{uni, corn},
			enabledExtensions: []string{corn},
			enabledCommands:   []string{"corn"},
			version:           universion,
			properties:        props,
			inHome:            true,
		}
		lang := bootStrapLanguageTest(args)
		assert.Equal(t, tc.ExpectedEnabled, lang.enabled(), tc.Case)
	}
}
