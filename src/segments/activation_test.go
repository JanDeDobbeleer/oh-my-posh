package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestLanguageActivation(t *testing.T) {
	cases := []struct {
		options              options.Map
		Case                 string
		displayMode          string
		extensions           []string
		folders              []string
		projectFiles         []string
		contextEnvVars       []string
		contextFiles         []string
		expectedFileGlobs    []string
		expectedFolders      []string
		expectedProjectFiles []string
		expectedEnvVars      []string
		withContext          bool
		expectedAlways       bool
	}{
		{
			Case:              "files mode gates on extensions",
			extensions:        []string{uni, corn},
			expectedFileGlobs: []string{uni, corn},
		},
		{
			Case:              "files mode carries folders and project files",
			extensions:        []string{uni},
			folders:           []string{"lua"},
			projectFiles:      []string{"build.zig"},
			expectedFileGlobs: []string{uni},
			expectedFolders:   []string{"lua"},
			expectedProjectFiles: []string{
				"build.zig",
			},
		},
		{
			Case:           "display_mode always",
			extensions:     []string{uni},
			options:        options.Map{DisplayMode: DisplayModeAlways},
			expectedAlways: true,
		},
		{
			Case:              "environment mode with declared env vars",
			extensions:        []string{uni},
			contextEnvVars:    []string{"VIRTUAL_ENV"},
			withContext:       true,
			options:           options.Map{DisplayMode: DisplayModeEnvironment},
			expectedFileGlobs: []string{uni},
			expectedEnvVars:   []string{"VIRTUAL_ENV"},
		},
		{
			Case:           "environment mode with an opaque context callback",
			extensions:     []string{uni},
			withContext:    true,
			options:        options.Map{DisplayMode: DisplayModeEnvironment},
			expectedAlways: true,
		},
		{
			Case:              "environment mode without a context callback",
			extensions:        []string{uni},
			options:           options.Map{DisplayMode: DisplayModeEnvironment},
			expectedFileGlobs: []string{uni},
		},
		{
			Case:              "context mode with declared context files",
			extensions:        []string{uni},
			contextFiles:      []string{"package.json"},
			withContext:       true,
			options:           options.Map{DisplayMode: DisplayModeContext},
			expectedFileGlobs: []string{uni, "package.json"},
		},
		{
			Case:           "context mode with an opaque context callback",
			extensions:     []string{uni},
			withContext:    true,
			options:        options.Map{DisplayMode: DisplayModeContext},
			expectedAlways: true,
		},
		{
			Case:              "extensions overridden via options",
			extensions:        []string{uni},
			options:           options.Map{LanguageExtensions: []string{corn}},
			expectedFileGlobs: []string{corn},
		},
		{
			Case:            "folders overridden via options",
			extensions:      []string{uni},
			options:         options.Map{LanguageFolders: []string{".venv"}},
			expectedFolders: []string{".venv"},
			expectedFileGlobs: []string{
				uni,
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		language := &Language{
			extensions:     tc.extensions,
			folders:        tc.folders,
			projectFiles:   tc.projectFiles,
			contextEnvVars: tc.contextEnvVars,
			contextFiles:   tc.contextFiles,
			displayMode:    tc.displayMode,
		}

		if tc.withContext {
			language.inContext = func() bool { return false }
		}

		opts := tc.options
		if opts == nil {
			opts = options.Map{}
		}

		language.Init(opts, env)

		activation := language.activation()

		assert.Equal(t, tc.expectedAlways, activation.Always, tc.Case)
		if tc.expectedAlways {
			continue
		}

		assert.Equal(t, tc.expectedFileGlobs, activation.FileGlobs, tc.Case)
		assert.Equal(t, tc.expectedFolders, activation.Folders, tc.Case)
		assert.Equal(t, tc.expectedProjectFiles, activation.ProjectFiles, tc.Case)
		assert.Equal(t, tc.expectedEnvVars, activation.EnvVars, tc.Case)
	}
}

func TestNodeActivation(t *testing.T) {
	node := &Node{}
	node.Init(options.Map{}, new(mock.Environment))

	activation := node.Activation()

	assert.False(t, activation.Always)
	assert.Equal(t, []string{"*.js", "*.ts", fileName, ".nvmrc", "pnpm-workspace.yaml", ".pnpmfile.cjs", ".vue"}, activation.FileGlobs)
}

// Python defaults to the environment display mode; the gate carries its
// declared venv triggers (env vars) alongside the file spec instead of
// falling back to Always.
func TestPythonActivation(t *testing.T) {
	python := &Python{}
	python.Init(options.Map{}, new(mock.Environment))

	activation := python.Activation()

	assert.False(t, activation.Always)
	assert.Equal(t, []string{"*.py", "*.ipynb", "pyproject.toml", "venv.bak"}, activation.FileGlobs)
	assert.Equal(t, []string{".venv", "venv", "virtualenv", "venv-win", "pyenv-win"}, activation.Folders)
	assert.Equal(t, []string{"VIRTUAL_ENV", "CONDA_ENV_PATH", "CONDA_DEFAULT_ENV"}, activation.EnvVars)
}

func TestConfiguredLanguageActivation(t *testing.T) {
	cases := []struct {
		Case                 string
		name                 string
		expectedFileGlobs    []string
		expectedFolders      []string
		expectedProjectFiles []string
	}{
		{
			Case:              "rust preset gates on its extensions",
			name:              "rust",
			expectedFileGlobs: []string{"*.rs", "Cargo.toml", "Cargo.lock"},
		},
		{
			Case:                 "zig preset carries its project files",
			name:                 "zig",
			expectedFileGlobs:    []string{"*.zig", "*.zon"},
			expectedProjectFiles: []string{"build.zig"},
		},
		{
			Case:              "lua preset carries its folders",
			name:              "lua",
			expectedFileGlobs: []string{"*.lua", "*.rockspec"},
			expectedFolders:   []string{"lua"},
		},
	}

	for _, tc := range cases {
		language := NewLanguage(tc.name)
		language.Init(options.Map{}, new(mock.Environment))

		activation := language.Activation()

		assert.False(t, activation.Always, tc.Case)
		assert.Equal(t, tc.expectedFileGlobs, activation.FileGlobs, tc.Case)
		assert.Equal(t, tc.expectedFolders, activation.Folders, tc.Case)
		assert.Equal(t, tc.expectedProjectFiles, activation.ProjectFiles, tc.Case)
	}
}

// An unknown preset has no conditions: the zero-value activation carries no
// gate, so the segment executes (and Enabled decides, as before).
func TestConfiguredLanguageActivationUnknownPreset(t *testing.T) {
	language := NewLanguage("unknown")
	language.Init(options.Map{}, new(mock.Environment))

	activation := language.Activation()

	assert.False(t, activation.Always)
	assert.Empty(t, activation.FileGlobs)
	assert.Empty(t, activation.Folders)
	assert.Empty(t, activation.ProjectFiles)
	assert.Empty(t, activation.EnvVars)
	assert.True(t, activation.Active(new(mock.Environment)), "the zero value carries no gate")
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
	assert.True(t, activation.Active(env))

	assert.True(t, bun.Enabled())
}
