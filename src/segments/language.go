package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	runtime_ "runtime"

	"slices"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

const (
	languageTemplate = " {{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }} "
	noVersion        = "NO VERSION"

	versionFlagArg      = "--version"
	versionFlagShortArg = "-version"
	versionArg          = "version"

	versionRegex         = `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`
	versionRegexPrefixed = `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`
	versionRegexSemver   = `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)(-(?P<prerelease>[a-z]+).(?P<buildmetadata>[0-9]+))?)))`

	fileName        = "package.json"
	pubspecFileName = "pubspec.yaml"

	asdfToolName   = "asdf"
	bunToolName    = "bun"
	dartToolName   = "dart"
	denoToolName   = "deno"
	dotnetToolName = "dotnet"
	fvmToolName    = "fvm"
	juliaToolName  = "julia"
	mojoToolName   = "mojo"
	nodeToolName   = "node"
	npmToolName    = "npm"
	phpToolName    = "php"
	pnpmToolName   = "pnpm"
	pythonToolName = "python"
	yarnToolName   = "yarn"
)

type loadContext func()

type inContext func() bool

type getVersion func() (string, error)
type matchesVersionFile func() (string, bool)

type Version struct {
	Full          string
	Major         string
	Minor         string
	Patch         string
	Prerelease    string
	BuildMetadata string
	URL           string
	Executable    string
	Expected      string
}

type cmd struct {
	getVersion         getVersion
	executable         string
	regex              string
	versionURLTemplate string
	args               []string
	envs               []string
}

func (c *cmd) parse(versionInfo string) (*Version, error) {
	values := regex.FindNamedRegexMatch(c.regex, versionInfo)
	if len(values) == 0 {
		return nil, errors.New("cannot parse version string")
	}

	version := &Version{
		Full:          values["version"],
		Major:         values["major"],
		Minor:         values["minor"],
		Patch:         values["patch"],
		Prerelease:    values["prerelease"],
		BuildMetadata: values["buildmetadata"],
	}
	return version, nil
}

type Language struct {
	Base

	projectRoot        *runtime.FileInfo
	loadContext        loadContext
	inContext          inContext
	matchesVersionFile matchesVersionFile
	Version
	displayMode        string
	Error              string
	versionURLTemplate string
	name               string
	commands           []*cmd
	tooling            map[string]*cmd
	defaultTooling     []string
	projectFiles       []string
	folders            []string
	extensions         []string
	// contextEnvVars declares the environment variables whose presence can
	// make inContext return true, so environment/context display modes can
	// gate on them instead of falling back to Always.
	contextEnvVars []string
	// contextFiles declares the cwd files inContext reads, same purpose.
	contextFiles []string
	exitCode     int
	homeEnabled  bool
	Mismatch     bool
}

const (
	// DisplayMode sets the display mode (always, when_in_context, never)
	DisplayMode            options.Option = "display_mode"
	DisplayModeAlways      string         = "always"
	DisplayModeFiles       string         = "files"
	DisplayModeEnvironment string         = "environment"
	DisplayModeContext     string         = "context"
	MissingCommandText     options.Option = "missing_command_text"
	HomeEnabled            options.Option = "home_enabled"
	LanguageExtensions     options.Option = "extensions"
	LanguageFolders        options.Option = "folders"
	LanguageProjectFiles   options.Option = "project_files"
	// Tooling allows enabling additional version fetching tools
	Tooling options.Option = "tooling"
	// Tools defines custom tools (executable, args, regex) for a configured language
	Tools options.Option = "tools"
	// LanguageName identifies a configured language segment; used as its cache key and preset lookup key
	LanguageName options.Option = "name"
)

func (l *Language) getName() string {
	_, file, _, _ := runtime_.Caller(2)
	base := filepath.Base(file)
	return base[:len(base)-3]
}

// Enabled decides whether the segment renders. It runs only when the
// activation gate (see activation) passed - or was bypassed by Force or
// pinned data - so the file/extension/folder presence checks the gate
// already proved are not repeated here; what remains are the decisions the
// gate cannot express (home directory, context callbacks) and the version
// fetching.
func (l *Language) Enabled() bool {
	if l.name == "" {
		l.name = l.getName()
	}
	// override default extensions if needed
	l.extensions = l.options.StringArray(LanguageExtensions, l.extensions)
	l.folders = l.options.StringArray(LanguageFolders, l.folders)
	l.projectFiles = l.options.StringArray(LanguageProjectFiles, l.projectFiles)
	inHomeDir := func() bool {
		return l.env.Pwd() == l.env.Home()
	}

	var enabled bool

	homeEnabled := l.options.Bool(HomeEnabled, l.homeEnabled)
	if inHomeDir() && !homeEnabled {
		return false
	}

	// The gate already established that a project file exists, but the
	// segment needs the project root itself (InProjectDir, quasar's
	// dependency fetching), so the search still runs; the runtime-level
	// HasParentFilePath cache makes it a cache hit.
	if len(l.projectFiles) != 0 && l.hasProjectFiles() {
		enabled = true
	}

	if !enabled {
		// set default mode when not set
		if l.displayMode == "" {
			l.displayMode = l.options.String(DisplayMode, DisplayModeFiles)
		}

		l.loadLanguageContext()

		switch l.displayMode {
		case DisplayModeAlways:
			enabled = true
		case DisplayModeEnvironment:
			enabled = l.inLanguageContext()
		case DisplayModeFiles:
			// the gate verified a matching file or folder in the cwd;
			// re-checking here would duplicate it. Under Force the gate is
			// skipped, but forced segments render regardless. A spec with
			// neither extensions nor folders carries no gate at all, so
			// nothing was verified: such a segment stays disabled, as it
			// always was in files mode.
			enabled = len(l.extensions) != 0 || len(l.folders) != 0
		case DisplayModeContext:
			fallthrough
		default:
			// the gate narrows this down but cannot decide it: a pass via a
			// declared context trigger (env var, context file) does not
			// guarantee the context callback agrees
			enabled = l.hasLanguageFiles() || l.hasLanguageFolders() || l.inLanguageContext()
		}
	}

	l.loadTooling()

	if !enabled || !l.options.Bool(options.FetchVersion, true) {
		return enabled
	}

	err := l.setVersion()
	if err != nil {
		l.Error = err.Error()
	}

	if l.matchesVersionFile != nil {
		expected, match := l.matchesVersionFile()
		if !match {
			l.Mismatch = true
			l.Expected = expected
		}
	}

	return enabled
}

// activation backs the Activation() implementation of the concrete language
// segments. It expresses everything Enabled()'s decision tree can react to
// as OR'd preconditions: the extension globs and folders of the file check,
// the project files (searched in parent directories), and the declared
// triggers of the context callback (contextEnvVars/contextFiles). Where the
// context callback is opaque - it exists but declares no triggers - the gate
// falls back to Always so the full Enabled() path decides, exactly as
// before.
//
// Deliberately unexported: promoting an exported Activation() from Language
// would gate every embedder, including one that only builds its spec inside
// Enabled() and would therefore be gated against an empty (or incomplete)
// spec. A concrete segment opts in by defining Activation() itself, building
// its spec first and then delegating here - the same way its Enabled()
// builds the spec before delegating to Language.Enabled().
func (l *Language) activation() Activation {
	displayMode := l.displayMode
	if displayMode == "" {
		displayMode = l.options.String(DisplayMode, DisplayModeFiles)
	}

	if displayMode == DisplayModeAlways {
		return Activation{Always: true}
	}

	activation := Activation{
		FileGlobs:    l.options.StringArray(LanguageExtensions, l.extensions),
		Folders:      l.options.StringArray(LanguageFolders, l.folders),
		ProjectFiles: l.options.StringArray(LanguageProjectFiles, l.projectFiles),
	}

	if displayMode == DisplayModeFiles {
		return activation
	}

	// environment and context modes can also enable through the context
	// callback; gate on its declared triggers when it has any
	activation.EnvVars = l.contextEnvVars

	if len(l.contextFiles) != 0 {
		globs := make([]string, 0, len(activation.FileGlobs)+len(l.contextFiles))
		globs = append(globs, activation.FileGlobs...)
		globs = append(globs, l.contextFiles...)
		activation.FileGlobs = globs
	}

	if l.inContext != nil && len(l.contextEnvVars) == 0 && len(l.contextFiles) == 0 {
		// opaque context callback: it can enable the segment through state
		// the gate cannot see, so there is no gate
		activation.Always = true
	}

	return activation
}

// Users can override the default tooling via the Tooling option (e.g. "uv" for Python to use the UV package manager).
func (l *Language) loadTooling() {
	enabledTools := l.options.StringArray(Tooling, l.defaultTooling)
	if len(enabledTools) == 0 {
		return
	}

	var commands []*cmd
	for _, tool := range enabledTools {
		if command, exists := l.tooling[tool]; exists {
			commands = append(commands, command)
		}
	}

	l.commands = commands
}

func (l *Language) hasLanguageFiles() bool {
	return slices.ContainsFunc(l.extensions, l.env.HasFiles)
}

func (l *Language) hasProjectFiles() bool {
	for _, extension := range l.projectFiles {
		if configPath, err := l.env.HasParentFilePath(extension, false); err == nil {
			l.projectRoot = configPath
			return true
		}
	}

	return false
}

func (l *Language) InProjectDir() bool {
	return l.projectRoot != nil
}

func (l *Language) hasLanguageFolders() bool {
	return slices.ContainsFunc(l.folders, l.env.HasFolder)
}

func (l *Language) setVersion() error {
	var lastError error

	cacheKey := fmt.Sprintf("version_%s", l.name)

	if versionCache, OK := cache.Device.Get[Version](cacheKey); OK {
		l.Version = versionCache
		return nil
	}

	for _, command := range l.commands {
		versionStr, err := l.runCommand(command)
		if err != nil {
			log.Error(err)
			lastError = err
			continue
		}

		version, err := command.parse(versionStr)
		if err != nil {
			log.Error(err)
			lastError = fmt.Errorf("err parsing info from %s with %s", command.executable, versionStr)
			continue
		}

		l.Version = *version
		if command.versionURLTemplate != "" {
			l.versionURLTemplate = command.versionURLTemplate
		}

		l.buildVersionURL()
		l.Executable = command.executable

		duration := l.options.String(options.CacheDuration, string(cache.NONE))
		cache.Device.Set(cacheKey, l.Version, cache.Duration(duration))

		return nil
	}

	if lastError != nil {
		return lastError
	}

	return errors.New(l.options.String(MissingCommandText, ""))
}

func (l *Language) runCommand(command *cmd) (string, error) {
	if command.getVersion == nil {
		if !l.env.HasCommand(command.executable) {
			return "", errors.New(noVersion)
		}

		versionStr, err := l.env.RunCommandWithEnv(command.executable, command.envs, command.args...)

		if exitErr, ok := err.(*runtime.CommandError); ok {
			l.exitCode = exitErr.ExitCode
			return "", fmt.Errorf("err executing %s with %v", command.executable, command.args)
		}

		return versionStr, nil
	}

	versionStr, err := command.getVersion()
	if err != nil {
		return "", err
	}

	if versionStr == "" {
		return "", errors.New("no version found")
	}

	return versionStr, nil
}

func (l *Language) loadLanguageContext() {
	if l.loadContext == nil {
		return
	}
	l.loadContext()
}

func (l *Language) inLanguageContext() bool {
	if l.inContext == nil {
		return false
	}
	return l.inContext()
}

func (l *Language) buildVersionURL() {
	versionURLTemplate := l.options.String(options.VersionURLTemplate, l.versionURLTemplate)
	if versionURLTemplate == "" {
		return
	}

	url, err := template.RenderTrusted(versionURLTemplate, l.Version)
	if err != nil {
		return
	}

	l.URL = url
}

func (l *Language) hasNodePackage(name string) bool {
	packageJSON := l.env.FileContent(fileName)

	var packageData map[string]any
	if err := json.Unmarshal([]byte(packageJSON), &packageData); err != nil {
		return false
	}

	dependencies, ok := packageData["dependencies"].(map[string]any)
	if !ok {
		return false
	}

	if _, exists := dependencies[name]; !exists {
		return false
	}

	return true
}

func (l *Language) nodePackageVersion(name string) (string, error) {
	folder := filepath.Join(l.env.Pwd(), "node_modules", name)

	if !l.env.HasFilesInDir(folder, fileName) {
		return "", fmt.Errorf("%s not found in %s", fileName, folder)
	}

	content := l.env.FileContent(filepath.Join(folder, fileName))
	var data ProjectData
	err := json.Unmarshal([]byte(content), &data)

	if err != nil {
		return "", err
	}

	return data.Version, nil
}
