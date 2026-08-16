package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestLanguageActivation(t *testing.T) {
	cases := []struct {
		options           options.Map
		Case              string
		displayMode       string
		extensions        []string
		folders           []string
		projectFiles      []string
		expectedFileGlobs []string
		expectedAlways    bool
	}{
		{
			Case:              "files mode gates on extensions",
			extensions:        []string{uni, corn},
			expectedFileGlobs: []string{uni, corn},
		},
		{
			Case:           "display_mode always",
			extensions:     []string{uni},
			options:        options.Map{DisplayMode: DisplayModeAlways},
			expectedAlways: true,
		},
		{
			Case:           "display_mode environment",
			extensions:     []string{uni},
			options:        options.Map{DisplayMode: DisplayModeEnvironment},
			expectedAlways: true,
		},
		{
			Case:           "display_mode context",
			extensions:     []string{uni},
			options:        options.Map{DisplayMode: DisplayModeContext},
			expectedAlways: true,
		},
		{
			Case:           "preset display mode other than files",
			extensions:     []string{uni},
			displayMode:    DisplayModeEnvironment,
			expectedAlways: true,
		},
		{
			Case:           "folders cannot be expressed as file globs",
			extensions:     []string{uni},
			folders:        []string{"lua"},
			expectedAlways: true,
		},
		{
			Case:           "project files are searched in parent directories",
			extensions:     []string{uni},
			projectFiles:   []string{"build.zig"},
			expectedAlways: true,
		},
		{
			Case:           "no extensions means no gate",
			expectedAlways: true,
		},
		{
			Case:              "extensions overridden via options",
			extensions:        []string{uni},
			options:           options.Map{LanguageExtensions: []string{corn}},
			expectedFileGlobs: []string{corn},
		},
		{
			Case:           "folders added via options force Always",
			extensions:     []string{uni},
			options:        options.Map{LanguageFolders: []string{".venv"}},
			expectedAlways: true,
		},
		{
			Case:           "project files added via options force Always",
			extensions:     []string{uni},
			options:        options.Map{LanguageProjectFiles: []string{"quasar.config"}},
			expectedAlways: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		language := &Language{
			extensions:   tc.extensions,
			folders:      tc.folders,
			projectFiles: tc.projectFiles,
			displayMode:  tc.displayMode,
		}

		opts := tc.options
		if opts == nil {
			opts = options.Map{}
		}

		language.Init(opts, env)

		activation := language.activation()

		assert.Equal(t, tc.expectedAlways, activation.Always, tc.Case)
		if !tc.expectedAlways {
			assert.Equal(t, tc.expectedFileGlobs, activation.FileGlobs, tc.Case)
		}
	}
}

func TestNodeActivation(t *testing.T) {
	node := &Node{}
	node.Init(options.Map{}, new(mock.Environment))

	activation := node.Activation()

	assert.False(t, activation.Always)
	assert.Equal(t, []string{"*.js", "*.ts", fileName, ".nvmrc", "pnpm-workspace.yaml", ".pnpmfile.cjs", ".vue"}, activation.FileGlobs)
}

// Python defaults to the environment display mode (an active venv enables it
// without any file in the cwd), so it must never gate by default.
func TestPythonActivationDefaultsToAlways(t *testing.T) {
	python := &Python{}
	python.Init(options.Map{}, new(mock.Environment))

	assert.True(t, python.Activation().Always)
}

// With display_mode=files, Python still declares venv folders, which the file
// gate cannot express, so it stays Always.
func TestPythonActivationFilesModeStillAlwaysDueToFolders(t *testing.T) {
	python := &Python{}
	python.Init(options.Map{DisplayMode: DisplayModeFiles}, new(mock.Environment))

	assert.True(t, python.Activation().Always)
}

func TestConfiguredLanguageActivation(t *testing.T) {
	cases := []struct {
		Case              string
		name              string
		expectedFileGlobs []string
		expectedAlways    bool
	}{
		{
			Case:              "rust preset gates on its extensions",
			name:              "rust",
			expectedFileGlobs: []string{"*.rs", "Cargo.toml", "Cargo.lock"},
		},
		{
			Case:           "zig preset has project files",
			name:           "zig",
			expectedAlways: true,
		},
		{
			Case:           "lua preset has folders",
			name:           "lua",
			expectedAlways: true,
		},
		{
			Case:           "unknown preset has no extensions and no gate",
			name:           "unknown",
			expectedAlways: true,
		},
	}

	for _, tc := range cases {
		language := NewLanguage(tc.name)
		language.Init(options.Map{}, new(mock.Environment))

		activation := language.Activation()

		assert.Equal(t, tc.expectedAlways, activation.Always, tc.Case)
		if !tc.expectedAlways {
			assert.Equal(t, tc.expectedFileGlobs, activation.FileGlobs, tc.Case)
		}
	}
}

// Activation must leave the language in a state where Enabled() still works
// and reports the same answer, since the gate calls Activation() first and
// Enabled() after on a match.
func TestActivationThenEnabledStaysConsistent(t *testing.T) {
	env := new(mock.Environment)
	env.On("Pwd").Return("/usr/home/project")
	env.On("Home").Return("/usr/home")
	env.On("HasFiles", "bun.lockb").Return(true)
	env.On("HasFiles", "bun.lock").Return(false)

	bun := &Bun{}
	bun.Init(options.Map{options.FetchVersion: false}, env)

	activation := bun.Activation()
	assert.False(t, activation.Always)
	assert.Equal(t, []string{"bun.lockb", "bun.lock"}, activation.FileGlobs)

	assert.True(t, bun.Enabled())
}
