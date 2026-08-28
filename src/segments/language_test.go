package segments

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	universion = "1.3.307"
	uni        = "*.uni"
	corn       = "*.corn"
)

type languageArgs struct {
	expectedError      error
	options            options.Provider
	matchesVersionFile matchesVersionFile
	version            string
	versionURLTemplate string
	extensions         []string
	enabledExtensions  []string
	commands           []*cmd
	enabledCommands    []string
	inHome             bool
}

func (l *languageArgs) hasvalue(value string, list []string) bool {
	return slices.Contains(list, value)
}

// setVersionRefs seeds the derived version-fetch decision for a writer in
// tests: fetch mirrors an analyzable template referencing .Full, !fetch an
// analyzable template referencing no version field. Tests that seed nothing
// exercise the fail-open default (never-delivered set -> fetch).
func setVersionRefs(consumer interface{ SetReferencedFields(template.RefSet) }, fetch bool) {
	refs := template.RefSet{Analyzable: true}
	if fetch {
		refs.Fields = []string{"Full"}
	}

	consumer.SetReferencedFields(refs)
}

func bootStrapLanguageTest(args *languageArgs) *Language {
	env := new(mock.Environment)

	for _, command := range args.commands {
		env.On("HasCommand", command.executable).Return(args.hasvalue(command.executable, args.enabledCommands))
		env.On("RunCommandWithEnv", command.executable, command.envs, command.args).Return(args.version, args.expectedError)
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

	if args.options == nil {
		args.options = options.Map{}
	}

	l := &Language{
		extensions:         args.extensions,
		commands:           args.commands,
		versionURLTemplate: args.versionURLTemplate,
		matchesVersionFile: args.matchesVersionFile,
	}
	l.Init(args.options, env)

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
	assert.True(t, lang.Enabled())
	assert.Equal(t, noVersion, lang.Error, "unicorn is not available")
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
	}
	lang := bootStrapLanguageTest(args)
	// no version field referenced -> the version fetch is skipped
	setVersionRefs(lang, false)
	assert.True(t, lang.Enabled(), "unicorn is not available")
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
	assert.True(t, lang.Enabled(), "unicorn is not available")
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
	// Enabled on its own (as reached via the Force/pinned-data gate
	// bypasses) must stay standalone-correct, and the gate must agree
	assert.False(t, lang.Enabled(), "no files in the current directory")
	activation := lang.activation()
	assert.False(t, activation.Active(lang.env), "the gate agrees: no files, no activation")
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
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and uni files are found")
	assert.Equal(t, "unicorn", lang.Executable, "unicorn was used")
}

func TestLanguageEnabledMismatch(t *testing.T) {
	expectedVersion := "1.2.009"

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
		matchesVersionFile: func() (string, bool) {
			return expectedVersion, false
		},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, expectedVersion, lang.Expected, "the expected unicorn version is 1.2.009")
	assert.True(t, lang.Mismatch, "we require a different version of unicorn")
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
	assert.False(t, lang.Enabled())
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
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and corn files are found")
	assert.Equal(t, "unicorn", lang.Executable, "unicorn was used")
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
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and corn files are found")
	assert.Equal(t, "corn", lang.Executable, "corn was used")
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
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and uni and corn files are found")
	assert.Equal(t, "unicorn", lang.Executable, "unicorn was used")
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
	}
	lang := bootStrapLanguageTest(args)
	// no version field referenced -> the version fetch is skipped
	setVersionRefs(lang, false)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "", lang.Full, "unicorn is available and uni and corn files are found")
	assert.Equal(t, "", lang.Executable, "no version was found")
}

func TestLanguageEnabledMissingCommand(t *testing.T) {
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
	}
	lang := bootStrapLanguageTest(args)
	setVersionRefs(lang, false)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "", lang.Full, "unicorn is unavailable and uni and corn files are found")
	assert.Equal(t, "", lang.Executable, "no executable was found")
}

func TestLanguageEnabledNoVersionData(t *testing.T) {
	props := options.Map{}
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
		options:           props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "", lang.Full)
	assert.Equal(t, "", lang.Executable, "no version was found")
}

func TestLanguageEnabledMissingCommandCustomText(t *testing.T) {
	expected := "missing"
	props := options.Map{
		MissingCommandText: expected,
	}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		options:           props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, expected, lang.Error, "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommandCustomTextHideError(t *testing.T) {
	props := options.Map{MissingCommandText: "missing"}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		options:           props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
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
		expectedError:     &runtime.CommandError{ExitCode: expected},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
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
		options:            options.Map{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "https://unicor.org/doc/1.3.307", lang.URL)
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
		options:            options.Map{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
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
		props := options.Map{
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
			options:           props,
			inHome:            true,
		}
		lang := bootStrapLanguageTest(args)
		assert.Equal(t, tc.ExpectedEnabled, lang.Enabled(), tc.Case)
	}
}

func TestLanguageInnerHyperlink(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable:         "uni",
				args:               []string{"--version"},
				regex:              `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				versionURLTemplate: "https://uni.org/release/{{ .Full }}",
			},
			{
				executable:         "corn",
				args:               []string{"--version"},
				regex:              `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				versionURLTemplate: "https://unicor.org/doc/{{ .Full }}",
			},
		},
		versionURLTemplate: "This gets replaced with inner template",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		options:            options.Map{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "https://unicor.org/doc/1.3.307", lang.URL)
}

func TestLanguageHyperlinkTemplatePropertyTakesPriority(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable:         "uni",
				args:               []string{"--version"},
				regex:              `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				versionURLTemplate: "https://uni.org/release/{{ .Full }}",
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		enabledCommands:   []string{"uni"},
		version:           universion,
		options: options.Map{
			options.VersionURLTemplate: "https://custom/url/template/{{ .Major }}.{{ .Minor }}",
		},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "https://custom/url/template/1.3", lang.URL)
}

func TestLanguageTooling(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedFirst   string
		ExpectedVersion string
		ToolVersion     string
		DefaultVersion  string
		Tooling         []string
		DefaultTooling  []string
		EnabledTools    []string
	}{
		{
			Case:            "Custom tooling overrides default",
			Tooling:         []string{"mytool"},
			DefaultTooling:  []string{"unicorn"},
			EnabledTools:    []string{"mytool", "unicorn"},
			ExpectedFirst:   "mytool",
			ExpectedVersion: "2.0.0",
			ToolVersion:     "2.0.0",
			DefaultVersion:  "1.0.0",
		},
		{
			Case:            "Default tooling used when no override",
			Tooling:         nil,
			DefaultTooling:  []string{"unicorn"},
			EnabledTools:    []string{"mytool", "unicorn"},
			ExpectedFirst:   "unicorn",
			ExpectedVersion: "1.0.0",
			ToolVersion:     "2.0.0",
			DefaultVersion:  "1.0.0",
		},
		{
			Case:            "Tool not available falls back to next",
			Tooling:         []string{"mytool", "unicorn"},
			DefaultTooling:  []string{"unicorn"},
			EnabledTools:    []string{"unicorn"},
			ExpectedFirst:   "mytool",
			ExpectedVersion: "1.0.0",
			ToolVersion:     "",
			DefaultVersion:  "1.0.0",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("HasFiles", uni).Return(true)

		hasUnicorn := slices.Contains(tc.EnabledTools, "unicorn")
		env.On("HasCommand", "unicorn").Return(hasUnicorn)
		if hasUnicorn {
			env.On("RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"}).Return(tc.DefaultVersion, nil)
		}

		hasToolCommand := slices.Contains(tc.EnabledTools, "mytool")
		env.On("HasCommand", "mytool").Return(hasToolCommand)
		if hasToolCommand {
			env.On("RunCommandWithEnv", "mytool", []string(nil), []string{"--version"}).Return(tc.ToolVersion, nil)
		}

		props := options.Map{}
		if tc.Tooling != nil {
			props[Tooling] = tc.Tooling
		}

		l := &Language{
			extensions:     []string{uni},
			defaultTooling: tc.DefaultTooling,
			tooling: map[string]*cmd{
				"unicorn": {
					executable: "unicorn",
					args:       []string{"--version"},
					regex:      "(?P<version>.*)",
				},
				"mytool": {
					executable: "mytool",
					args:       []string{"--version"},
					regex:      "(?P<version>.*)",
				},
			},
		}
		l.Init(props, env)

		assert.True(t, l.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedFirst, l.commands[0].executable, tc.Case)
		assert.Equal(t, tc.ExpectedVersion, l.Full, tc.Case)
	}
}

type mockedLanguageParams struct {
	cmd           string
	versionParam  string
	versionOutput string
	extension     string
	envs          []string
}

func getMockedLanguageEnv(params *mockedLanguageParams) (*mock.Environment, options.Map) {
	// Each call sets up one test case, often sharing the same command name
	// (and therefore the same mocked path/stat, so the same cache key) with
	// other cases in the same table. Without this, a cache entry a later
	// case's mock intends to miss would instead be served stale from an
	// earlier case that happened to run first.
	cache.Device.DeleteAll()

	env := new(mock.Environment)
	env.On("HasCommand", params.cmd).Return(true)
	env.On("RunCommandWithEnv", params.cmd, params.envs, []string{params.versionParam}).Return(params.versionOutput, nil)
	env.On("HasFiles", params.extension).Return(true)
	env.On("Pwd").Return("/usr/home/project")
	env.On("Home").Return("/usr/home")

	// Only consulted for a cmd with versionCacheable set; harmless, and
	// registered with Maybe so callers that never reach that branch aren't
	// forced to satisfy it. CommandPath/StatFile are deliberately not
	// stubbed here (unlike Flags): python.go calls CommandPath itself for
	// venv detection, unrelated to version caching, and a blanket stub here
	// would shadow whatever value that test wants back. Segments that opted
	// a command into versionCacheable get their CommandPath/StatFile stubs
	// via mockVersionCacheable below instead.
	env.On("Flags").Return(&runtime.Flags{}).Maybe()

	// fetch_version no longer exists: the fail-open default (no refs
	// delivered) fetches, matching the option's old default of true
	props := options.Map{}

	return env, props
}

// mockVersionCacheable stubs CommandPath and StatFile for executable so a
// versionCacheable command resolves through Language's keyed cache path
// during a test. Tests for segments that did not opt any command into that
// cache never need this.
func mockVersionCacheable(env *mock.Environment, executable string) {
	path := "/usr/bin/" + executable
	env.On("CommandPath", executable).Return(path)
	env.On("StatFile", path).Return(runtime.FileStat{ModTime: 1700000000, Size: 12345}, nil)
}

// cacheableUnicornLanguage returns a Language wired with one versionCacheable
// command ("unicorn --version") plus the mocks needed to resolve it: env
// must additionally stub RunCommandWithEnv per test.
func cacheableUnicornLanguage(env *mock.Environment) (*Language, *cmd) {
	env.On("Pwd").Return("/usr/home/project")
	env.On("Home").Return("/usr/home")
	env.On("Flags").Return(&runtime.Flags{})
	env.On("HasCommand", "unicorn").Return(true)

	command := &cmd{
		executable:       "unicorn",
		args:             []string{"--version"},
		regex:            "(?P<version>.*)",
		versionCacheable: true,
	}

	l := &Language{}
	l.Init(options.Map{}, env)

	return l, command
}

// TestLanguageVersionCommandCachedOnFirstRun covers the first-call path: the
// command runs, and its output lands in the device cache under the key
// versionCacheKey computes from the resolved executable's identity.
func TestLanguageVersionCommandCachedOnFirstRun(t *testing.T) {
	cache.Device.DeleteAll()

	env := new(mock.Environment)
	l, command := cacheableUnicornLanguage(env)
	env.On("CommandPath", "unicorn").Return("/usr/bin/unicorn")
	env.On("StatFile", "/usr/bin/unicorn").Return(runtime.FileStat{ModTime: 100, Size: 10}, nil)
	env.On("RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"}).Return("unicorn 1.3.307", nil).Once()

	out, err := l.runCommand(command)
	require.NoError(t, err)
	assert.Equal(t, "unicorn 1.3.307", out)

	key, cacheable := l.versionCacheKey(command)
	require.True(t, cacheable)

	cached, found := cache.Device.Get[string](key)
	require.True(t, found, "the first successful run should have cached under versionCacheKey's key")
	assert.Equal(t, "unicorn 1.3.307", cached)

	env.AssertNumberOfCalls(t, "RunCommandWithEnv", 1)
}

// TestLanguageVersionCommandCacheHitSkipsSecondRun covers the second-call
// path: with the same resolved path/mtime/size/args, runCommand must serve
// the cached output instead of invoking RunCommandWithEnv again - enforced
// here by mocking it .Once(), which panics on a second call.
func TestLanguageVersionCommandCacheHitSkipsSecondRun(t *testing.T) {
	cache.Device.DeleteAll()

	env := new(mock.Environment)
	l, command := cacheableUnicornLanguage(env)
	env.On("CommandPath", "unicorn").Return("/usr/bin/unicorn")
	env.On("StatFile", "/usr/bin/unicorn").Return(runtime.FileStat{ModTime: 100, Size: 10}, nil)
	env.On("RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"}).Return("unicorn 1.3.307", nil).Once()

	first, err := l.runCommand(command)
	require.NoError(t, err)

	second, err := l.runCommand(command)
	require.NoError(t, err)

	assert.Equal(t, first, second)
	env.AssertNumberOfCalls(t, "RunCommandWithEnv", 1)
}

// TestLanguageVersionCommandCacheKeyInvalidation covers the miss side of
// versionCacheKey: any change to what the output actually depends on - the
// resolved path, the executable's mtime, its size, or the args - must
// produce a different key than a fixed baseline, so runCommand re-runs
// instead of serving a stale value.
func TestLanguageVersionCommandCacheKeyInvalidation(t *testing.T) {
	baseline := &cmd{executable: "unicorn", args: []string{"--version"}, versionCacheable: true}
	baselinePath := "/usr/bin/unicorn"
	baselineStat := runtime.FileStat{ModTime: 100, Size: 10}

	cases := []struct {
		Case    string
		Command *cmd
		Path    string
		Stat    runtime.FileStat
	}{
		{Case: "baseline", Command: baseline, Path: baselinePath, Stat: baselineStat},
		{Case: "changed mtime", Command: baseline, Path: baselinePath, Stat: runtime.FileStat{ModTime: 200, Size: 10}},
		{Case: "changed size", Command: baseline, Path: baselinePath, Stat: runtime.FileStat{ModTime: 100, Size: 20}},
		{Case: "changed path", Command: baseline, Path: "/usr/local/bin/unicorn", Stat: baselineStat},
		{
			Case:    "changed args",
			Command: &cmd{executable: "unicorn", args: []string{"--version", "--extra"}, versionCacheable: true},
			Path:    baselinePath,
			Stat:    baselineStat,
		},
		{
			Case:    "changed envs",
			Command: &cmd{executable: "unicorn", args: []string{"--version"}, envs: []string{"FOO=bar"}, versionCacheable: true},
			Path:    baselinePath,
			Stat:    baselineStat,
		},
	}

	env := new(mock.Environment)
	env.On("Flags").Return(&runtime.Flags{})

	l := &Language{}
	l.Init(options.Map{}, env)

	keys := make(map[string]bool, len(cases))

	for _, tc := range cases {
		env.On("CommandPath", tc.Command.executable).Return(tc.Path).Once()
		env.On("StatFile", tc.Path).Return(tc.Stat, nil).Once()

		key, cacheable := l.versionCacheKey(tc.Command)
		require.True(t, cacheable, tc.Case)

		assert.False(t, keys[key], "%s: expected a key not seen before, got a collision", tc.Case)
		keys[key] = true
	}
}

// TestLanguageVersionCommandFailureNotCached covers the failure path: a
// command that exits with an error must not populate the cache, so a
// transient failure cannot pin a bad (or empty) result for the cache's TTL.
func TestLanguageVersionCommandFailureNotCached(t *testing.T) {
	cache.Device.DeleteAll()

	env := new(mock.Environment)
	l, command := cacheableUnicornLanguage(env)
	env.On("CommandPath", "unicorn").Return("/usr/bin/unicorn")
	env.On("StatFile", "/usr/bin/unicorn").Return(runtime.FileStat{ModTime: 100, Size: 10}, nil)
	env.On("RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"}).Return("", &runtime.CommandError{ExitCode: 1})

	_, err := l.runCommand(command)
	require.Error(t, err)

	key, cacheable := l.versionCacheKey(command)
	require.True(t, cacheable)

	_, found := cache.Device.Get[string](key)
	assert.False(t, found, "a failed run must not populate the cache")
}

// TestLanguageVersionCommandNonCacheableNeverConsultsCache covers a cmd that
// did not opt in: runCommand must never call CommandPath, StatFile, or
// Flags for it - enforced by leaving those unstubbed, which panics if
// runCommand calls them anyway - and never write to the cache.
func TestLanguageVersionCommandNonCacheableNeverConsultsCache(t *testing.T) {
	cache.Device.DeleteAll()

	env := new(mock.Environment)
	env.On("Pwd").Return("/usr/home/project")
	env.On("Home").Return("/usr/home")
	env.On("HasCommand", "unicorn").Return(true)
	env.On("RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"}).Return("unicorn 1.3.307", nil)

	command := &cmd{executable: "unicorn", args: []string{"--version"}}

	l := &Language{}
	l.Init(options.Map{}, env)

	out, err := l.runCommand(command)
	require.NoError(t, err)
	assert.Equal(t, "unicorn 1.3.307", out)

	assert.Equal(t, "Store device is empty", cache.Device.Print())
}

// TestLanguageVersionCommandDataOnlyUntouched covers DataOnly: even for a
// versionCacheable command whose HasCommand still answers true (as it would
// not under a real DataOnly environment - see terminal.go - this isolates
// the language-layer guard from that separate guarantee), runCommand must
// skip the cache entirely while leaving the command's own execution
// unaffected.
func TestLanguageVersionCommandDataOnlyUntouched(t *testing.T) {
	cache.Device.DeleteAll()

	env := new(mock.Environment)
	env.On("Pwd").Return("/usr/home/project")
	env.On("Home").Return("/usr/home")
	env.On("Flags").Return(&runtime.Flags{DataOnly: true})
	env.On("HasCommand", "unicorn").Return(true)
	env.On("RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"}).Return("unicorn 1.3.307", nil)

	command := &cmd{executable: "unicorn", args: []string{"--version"}, versionCacheable: true}

	l := &Language{}
	l.Init(options.Map{}, env)

	out, err := l.runCommand(command)
	require.NoError(t, err)
	assert.Equal(t, "unicorn 1.3.307", out, "DataOnly must not change the command's own result")

	assert.Equal(t, "Store device is empty", cache.Device.Print())
}

func TestNodePackageVersion(t *testing.T) {
	cases := []struct {
		Case        string
		PackageJSON string
		Version     string
		ShouldFail  bool
		NoFiles     bool
	}{
		{Case: "14.1.5", Version: "14.1.5", PackageJSON: "{ \"name\": \"nx\",\"version\": \"14.1.5\"}"},
		{Case: "14.0.0", Version: "14.0.0", PackageJSON: "{ \"name\": \"nx\",\"version\": \"14.0.0\"}"},
		{Case: "no files", NoFiles: true, ShouldFail: true},
		{Case: "bad data", ShouldFail: true, PackageJSON: "bad data"},
	}

	for _, tc := range cases {
		var env = new(mock.Environment)
		env.On("Pwd").Return("posh")
		path := filepath.Join("posh", "node_modules", "nx")
		env.On("HasFilesInDir", path, "package.json").Return(!tc.NoFiles)
		env.On("FileContent", filepath.Join(path, "package.json")).Return(tc.PackageJSON)

		a := &Language{}
		a.Init(options.Map{}, env)
		got, err := a.nodePackageVersion("nx")

		if tc.ShouldFail {
			assert.Error(t, err, tc.Case)
			return
		}

		assert.Nil(t, err, tc.Case)
		assert.Equal(t, tc.Version, got, tc.Case)
	}
}

// TestVersionCacheWrapperOptOuts pins which tools stay out of the version
// cache: wrapper/shim executables whose own identity cannot key their output
// (versionCacheKey's first guard makes an unset flag skip the cache
// entirely). java_home stays cacheable because the resolved JDK path itself
// is the key there.
func TestVersionCacheWrapperOptOuts(t *testing.T) {
	env := new(mock.Environment)
	env.On("Getenv", "JAVA_HOME").Return("/opt/java")

	java := &Java{}
	java.Init(options.Map{}, env)
	java.loadSpec()

	flutter := &Flutter{}
	flutter.loadSpec()

	ui5 := &UI5Tooling{}
	ui5.Init(options.Map{}, env)
	ui5.loadSpec()

	cases := []struct {
		Command   *cmd
		Case      string
		Cacheable bool
	}{
		{Case: "flutter wrapper", Command: flutter.tooling[flutterToolName]},
		{Case: "fvm", Command: flutter.tooling[fvmToolName]},
		{Case: "swift xcrun shim", Command: languageDefinitions["swift"].tooling[swiftToolName]},
		{Case: "plain java stub", Command: java.tooling[javaToolName]},
		{Case: "java_home-resolved java", Command: java.tooling["java_home"], Cacheable: true},
		{Case: "ui5 project delegation", Command: ui5.tooling["ui5"]},
	}

	for _, tc := range cases {
		require.NotNil(t, tc.Command, tc.Case)
		assert.Equal(t, tc.Cacheable, tc.Command.versionCacheable, tc.Case)
	}
}

// TestLanguageVersionFetchDerived pins the derived version fetch: the unit
// runs iff a version field is referenced, and the unanalyzable/undelivered
// fallback fails OPEN, matching the removed fetch_version option's default
// of true.
func TestLanguageVersionFetchDerived(t *testing.T) {
	cases := []struct {
		Refs            *template.RefSet
		Case            string
		ExpectedVersion string
		ExpectFetch     bool
	}{
		{
			Case:            "referenced field fetches",
			Refs:            &template.RefSet{Fields: []string{"Full"}, Analyzable: true},
			ExpectedVersion: universion,
			ExpectFetch:     true,
		},
		{
			Case:            "embedded Version reference fetches",
			Refs:            &template.RefSet{Fields: []string{"Version"}, Analyzable: true},
			ExpectedVersion: universion,
			ExpectFetch:     true,
		},
		{
			Case: "unreferenced analyzable skips the command",
			Refs: &template.RefSet{Fields: []string{"Venv"}, Analyzable: true},
		},
		{
			Case:            "unanalyzable fails open",
			Refs:            &template.RefSet{Analyzable: false},
			ExpectedVersion: universion,
			ExpectFetch:     true,
		},
		{
			Case:            "never-delivered set fails open",
			ExpectedVersion: universion,
			ExpectFetch:     true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasCommand", "unicorn").Return(true)
		env.On("RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"}).Return(universion, nil)
		env.On("HasFiles", uni).Return(true)
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")

		lang := &Language{
			extensions: []string{uni},
			commands: []*cmd{
				{
					executable: "unicorn",
					args:       []string{"--version"},
					regex:      "(?P<version>.*)",
				},
			},
		}
		lang.Init(options.Map{}, env)

		if tc.Refs != nil {
			lang.SetReferencedFields(*tc.Refs)
		}

		assert.True(t, lang.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedVersion, lang.Full, tc.Case)

		if tc.ExpectFetch {
			env.AssertCalled(t, "RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"})
			continue
		}

		env.AssertNotCalled(t, "RunCommandWithEnv", "unicorn", []string(nil), []string{"--version"})
	}
}
