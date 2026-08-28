package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
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
	// versionCacheable opts this command's raw output into the keyed device
	// cache (see Language.versionCacheKey): set it only when the output is a
	// pure function of the resolved executable - its path, mtime and size -
	// and the exact args, with no cwd file, environment variable, or project
	// config able to change what that same binary prints. A getVersion-backed
	// command never reaches the cache regardless of this flag, since the
	// cache sits in the executable-invocation branch of runCommand.
	versionCacheable bool
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

// languageVersionFields lists what the version fetch populates: the
// embedded Version struct's promoted fields (plus the embedded field name
// itself), the fetch-path errors, and the version-file mismatch state.
// setVersion and the mismatch check only run when one of these is
// referenced; see fetchUnitFailOpen for the fail-open polarity.
var languageVersionFields = []string{
	"Version", "Full", "Major", "Minor", "Patch", "Prerelease", "BuildMetadata",
	"URL", "Executable", "Expected", "Error", "Mismatch",
}

type Language struct {
	Base
	tooling            map[string]*cmd
	projectRoot        *runtime.FileInfo
	loadContext        loadContext
	inContext          inContext
	matchesVersionFile matchesVersionFile
	Version
	name               string
	Error              string
	versionURLTemplate string
	displayMode        string
	projectFiles       []string
	defaultTooling     []string
	commands           []*cmd
	folders            []string
	extensions         []string
	contextEnvVars     []string
	contextFiles       []string
	FieldRefs
	exitCode    int
	homeEnabled bool
	Mismatch    bool
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

// Enabled decides whether the segment renders, and is standalone-correct:
// it never assumes the activation gate (see activation) ran, because Force
// and pinned data bypass the gate entirely. The presence checks the gate
// also evaluates are re-verified here, which is near-free: the directory
// listing and parent-path searches are memoized per invocation, so the gate
// remains a pure skip-optimization.
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

	// Runs the search regardless of the gate: the segment needs the project
	// root itself (InProjectDir, quasar's dependency fetching), and after a
	// gate pass the runtime-level HasParentFilePath cache makes it a hit.
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
			// Re-verified even though a passing gate implies a match: the
			// gate is a skip-optimization, not a precondition Enabled may
			// rely on - Force and pinned data bypass it entirely, and the
			// gate's symlink-following project-file variant can pass where
			// this segment's own search misses. The re-check is a map hit:
			// the directory listing and parent-path searches are memoized
			// per invocation.
			enabled = l.hasLanguageFiles() || l.hasLanguageFolders()
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

	if !enabled || !l.fetchUnitFailOpen(languageVersionFields...) {
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
	// joined with Pwd, matching the activation gate's folder check, so a
	// caller whose PWD flag differs from the process cwd cannot get
	// gate/Enabled divergence
	return slices.ContainsFunc(l.folders, func(folder string) bool {
		return l.env.HasFolder(filepath.Join(l.env.Pwd(), folder))
	})
}

func (l *Language) setVersion() error {
	var lastError error

	// This is the legacy, opt-in cache: it only ever stores anything once a
	// user sets cache_duration (cache.Set is a no-op for the zero/NONE
	// duration this defaults to, below), and it caches the resolved Version
	// for the whole segment under a TTL the user picks, not one derived from
	// what the value depends on. Whether or not it fires, runCommand's keyed
	// cache below is still consulted per command - that one is the default,
	// always-on path for the commands that opted into it, and needs no
	// option to be effective.
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

		cacheKey, cacheable := l.versionCacheKey(command)
		if cacheable {
			if cached, found := cache.Device.Get[string](cacheKey); found {
				log.Debugf("using cached version output for %s", command.executable)
				return cached, nil
			}
		}

		versionStr, err := l.env.RunCommandWithEnv(command.executable, command.envs, command.args...)

		if exitErr, ok := err.(*runtime.CommandError); ok {
			l.exitCode = exitErr.ExitCode
			return "", fmt.Errorf("err executing %s with %v", command.executable, command.args)
		}

		// Success only: a failed run (a non-CommandError err, or one the
		// caller above already turned into an early return) never gets
		// cached, so a transient failure cannot pin a bad result for a week.
		if err == nil && cacheable {
			cache.Device.Set(cacheKey, versionStr, cache.ONEWEEK)
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

// versionCacheKey reports the device-cache key for command's raw output, and
// whether command is eligible for that cache at all. Eligibility requires:
// opting in via versionCacheable, an environment that can resolve a real
// executable identity (DataOnly cannot), and a resolvable, statable
// executable path.
//
// The key mixes the resolved absolute path with the executable's mtime and
// size, the exact arguments, and any extra environment the command runs
// with, so a PATH change, a reinstalled or upgraded binary, different args,
// or different envs all produce a different key and therefore a cache miss -
// invalidation follows from what the output depends on, not from a guessed
// TTL. The TTL passed to cache.Set (see runCommand) exists only so a device
// cache nobody prunes doesn't grow forever; it plays no part in correctness.
func (l *Language) versionCacheKey(command *cmd) (string, bool) {
	if !command.versionCacheable || l.env.Flags().DataOnly {
		return "", false
	}

	path := l.env.CommandPath(command.executable)
	if path == "" {
		return "", false
	}

	stat, err := l.env.StatFile(path)
	if err != nil {
		return "", false
	}

	h := fnv.New64a()
	fmt.Fprintf(h, "%s\x00%d\x00%d", path, stat.ModTime, stat.Size)

	for _, arg := range command.args {
		fmt.Fprintf(h, "\x00%s", arg)
	}

	// No cacheable cmd sets envs today, but the key must not silently
	// under-specify one that does: the output can depend on them.
	for _, env := range command.envs {
		fmt.Fprintf(h, "\x00env:%s", env)
	}

	return fmt.Sprintf("version_%x", h.Sum64()), true
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
