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
	projectFiles       []string
	folders            []string
	extensions         []string
	exitCode           int
	homeEnabled        bool
	Mismatch           bool
}

const (
	// DisplayMode sets the display mode (always, when_in_context, never)
	DisplayMode options.Option = "display_mode"
	// DisplayModeAlways displays the segment always
	DisplayModeAlways string = "always"
	// DisplayModeFiles displays the segment when the current folder contains certain extensions
	DisplayModeFiles string = "files"
	// DisplayModeEnvironment displays the segment when the environment has a language's context
	DisplayModeEnvironment string = "environment"
	// DisplayModeContext displays the segment when the environment or files is active
	DisplayModeContext string = "context"
	// MissingCommandText sets the text to display when the command is not present in the system
	MissingCommandText options.Option = "missing_command_text"
	// HomeEnabled displays the segment in the HOME folder or not
	HomeEnabled options.Option = "home_enabled"
	// LanguageExtensions the list of extensions to validate
	LanguageExtensions options.Option = "extensions"
	// LanguageFolders the list of folders to validate
	LanguageFolders options.Option = "folders"
)

func (l *Language) getName() string {
	_, file, _, _ := runtime_.Caller(2)
	base := filepath.Base(file)
	return base[:len(base)-3]
}

func (l *Language) Enabled() bool {
	l.name = l.getName()
	// override default extensions if needed
	l.extensions = l.options.StringArray(LanguageExtensions, l.extensions)
	l.folders = l.options.StringArray(LanguageFolders, l.folders)
	inHomeDir := func() bool {
		return l.env.Pwd() == l.env.Home()
	}

	var enabled bool

	homeEnabled := l.options.Bool(HomeEnabled, l.homeEnabled)
	if inHomeDir() && !homeEnabled {
		return false
	}

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
			enabled = l.hasLanguageFiles() || l.hasLanguageFolders()
		case DisplayModeContext:
			fallthrough
		default:
			enabled = l.hasLanguageFiles() || l.hasLanguageFolders() || l.inLanguageContext() || l.hasProjectFiles()
		}
	}

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

func (l *Language) hasLanguageFolders() bool {
	return slices.ContainsFunc(l.folders, l.env.HasFolder)
}

// setVersion parses the version string returned by the command
func (l *Language) setVersion() error {
	var lastError error

	cacheKey := fmt.Sprintf("version_%s", l.name)

	if versionCache, OK := cache.Get[Version](cache.Device, cacheKey); OK {
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
		cache.Set(cache.Device, cacheKey, l.Version, cache.Duration(duration))

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

		versionStr, err := l.env.RunCommand(command.executable, command.args...)
		if exitErr, ok := err.(*runtime.CommandError); ok {
			l.exitCode = exitErr.ExitCode
			return "", fmt.Errorf("err executing %s with %s", command.executable, command.args)
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

	url, err := template.Render(versionURLTemplate, l.Version)
	if err != nil {
		return
	}

	l.URL = url
}

func (l *Language) hasNodePackage(name string) bool {
	packageJSON := l.env.FileContent("package.json")

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

	const fileName string = "package.json"
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
