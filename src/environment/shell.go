package environment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"oh-my-posh/environment/battery"
	"oh-my-posh/environment/cmd"
	"oh-my-posh/regex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	process "github.com/shirou/gopsutil/v3/process"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	UNKNOWN = "unknown"
	WINDOWS = "windows"
	DARWIN  = "darwin"
	LINUX   = "linux"
)

var (
	lock          = sync.RWMutex{}
	TEMPLATECACHE = fmt.Sprintf("template_cache_%d", os.Getppid())
)

type Flags struct {
	ErrorCode     int
	Config        string
	Shell         string
	ShellVersion  string
	PWD           string
	PSWD          string
	ExecutionTime float64
	Eval          bool
	StackCount    int
	Migrate       bool
	TerminalWidth int
	Strict        bool
	Debug         bool
}

type CommandError struct {
	Err      string
	ExitCode int
}

func (e *CommandError) Error() string {
	return e.Err
}

type FileInfo struct {
	ParentFolder string
	Path         string
	IsDir        bool
}

type Cache interface {
	Init(home string)
	Close()
	Get(key string) (string, bool)
	// ttl in minutes
	Set(key, value string, ttl int)
}

type HTTPRequestModifier func(request *http.Request)

type WindowsRegistryValueType string

const (
	DWORD  = "DWORD"
	QWORD  = "QWORD"
	BINARY = "BINARY"
	STRING = "STRING"
)

type WindowsRegistryValue struct {
	ValueType WindowsRegistryValueType
	DWord     uint64
	QWord     uint64
	String    string
}

type IFTYPE string
type NDIS_MEDIUM string
type NDIS_PHYSICAL_MEDIUM string
type IF_OPER_STATUS string
type NET_IF_ADMIN_STATUS string
type NET_IF_MEDIA_CONNECT_STATE string

type ConnectedNetworks struct {
	Networks []NetworkInfo
}

type NetworkInfo struct {
	Alias                 string
	Interface             string
	InterfaceType         IFTYPE
	NDISMediaType         NDIS_MEDIUM
	NDISPhysicalMeidaType NDIS_PHYSICAL_MEDIUM
	TransmitLinkSpeed     uint64
	ReceiveLinkSpeed      uint64
	SSID                  string // Wi-Fi only
}

type WifiType string

type WifiInfo struct {
	SSID           string
	Interface      string
	RadioType      WifiType
	PhysType       WifiType
	Authentication WifiType
	Cipher         WifiType
	Channel        int
	ReceiveRate    int
	TransmitRate   int
	Signal         int
	Error          string
}

type TemplateCache struct {
	Root         bool
	PWD          string
	Folder       string
	Shell        string
	ShellVersion string
	UserName     string
	HostName     string
	Code         int
	Env          map[string]string
	OS           string
	WSL          bool
	Segments     map[string]interface{}
}

func (t *TemplateCache) AddSegmentData(key string, value interface{}) {
	lock.Lock()
	defer lock.Unlock()
	if t.Segments == nil {
		t.Segments = make(map[string]interface{})
	}
	key = cases.Title(language.English).String(key)
	t.Segments[key] = value
}

type Environment interface {
	Getenv(key string) string
	Pwd() string
	Home() string
	User() string
	Root() bool
	Host() (string, error)
	GOOS() string
	Shell() string
	Platform() string
	ErrorCode() int
	PathSeparator() string
	HasFiles(pattern string) bool
	HasFilesInDir(dir, pattern string) bool
	HasFolder(folder string) bool
	HasParentFilePath(path string) (fileInfo *FileInfo, err error)
	HasFileInParentDirs(pattern string, depth uint) bool
	ResolveSymlink(path string) (string, error)
	DirMatchesOneOf(dir string, regexes []string) bool
	DirIsWritable(path string) bool
	CommandPath(command string) string
	HasCommand(command string) bool
	FileContent(file string) string
	LsDir(path string) []fs.DirEntry
	RunCommand(command string, args ...string) (string, error)
	RunShellCommand(shell, command string) string
	ExecutionTime() float64
	Flags() *Flags
	BatteryState() (*battery.Info, error)
	QueryWindowTitles(processName, windowTitleRegex string) (string, error)
	WindowsRegistryKeyValue(path string) (*WindowsRegistryValue, error)
	HTTPRequest(url string, body io.Reader, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error)
	IsWsl() bool
	IsWsl2() bool
	StackCount() int
	TerminalWidth() (int, error)
	CachePath() string
	Cache() Cache
	Close()
	Logs() string
	InWSLSharedDrive() bool
	ConvertToLinuxPath(path string) string
	ConvertToWindowsPath(path string) string
	GetAllNetworkInterfaces() (*[]NetworkInfo, error)
	TemplateCache() *TemplateCache
	LoadTemplateCache()
	Log(logType LogType, funcName, message string)
	Trace(start time.Time, function string, args ...string)
}

type commandCache struct {
	commands *concurrentMap
}

func (c *commandCache) set(command, path string) {
	c.commands.set(command, path)
}

func (c *commandCache) get(command string) (string, bool) {
	cacheCommand, found := c.commands.get(command)
	if !found {
		return "", false
	}
	command, ok := cacheCommand.(string)
	return command, ok
}

type LogType string

const (
	Error LogType = "error"
	Debug LogType = "debug"
)

type ShellEnvironment struct {
	CmdFlags *Flags
	Version  string

	cwd        string
	cmdCache   *commandCache
	fileCache  *fileCache
	tmplCache  *TemplateCache
	logBuilder strings.Builder
}

func (env *ShellEnvironment) Init() {
	defer env.Trace(time.Now(), "Init")
	if env.CmdFlags == nil {
		env.CmdFlags = &Flags{}
	}
	if env.CmdFlags.Debug {
		log.SetOutput(&env.logBuilder)
	}
	env.fileCache = &fileCache{}
	env.fileCache.Init(env.CachePath())
	env.resolveConfigPath()
	env.cmdCache = &commandCache{
		commands: newConcurrentMap(),
	}
}

func (env *ShellEnvironment) resolveConfigPath() {
	defer env.Trace(time.Now(), "resolveConfigPath")
	if len(env.CmdFlags.Config) == 0 {
		env.CmdFlags.Config = env.Getenv("POSH_THEME")
	}
	if len(env.CmdFlags.Config) == 0 {
		env.CmdFlags.Config = fmt.Sprintf("https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/v%s/themes/default.omp.json", env.Version)
	}
	if strings.HasPrefix(env.CmdFlags.Config, "https://") {
		if err := env.downloadConfig(env.CmdFlags.Config); err != nil {
			// make it use default config when download fails
			env.CmdFlags.Config = ""
			return
		}
	}
	// Cygwin path always needs the full path as we're on Windows but not really.
	// Doing filepath actions will convert it to a Windows path and break the init script.
	if env.Platform() == WINDOWS && env.Shell() == "bash" {
		return
	}
	configFile := env.CmdFlags.Config
	if strings.HasPrefix(configFile, "~") {
		configFile = strings.TrimPrefix(configFile, "~")
		configFile = filepath.Join(env.Home(), configFile)
	}
	if !filepath.IsAbs(configFile) {
		if absConfigFile, err := filepath.Abs(configFile); err == nil {
			configFile = absConfigFile
		}
	}
	env.CmdFlags.Config = filepath.Clean(configFile)
}

func (env *ShellEnvironment) downloadConfig(location string) error {
	defer env.Trace(time.Now(), "downloadConfig", location)
	configPath := filepath.Join(env.CachePath(), "config.omp.json")
	cfg, err := env.HTTPRequest(location, nil, 5000)
	if err != nil {
		return err
	}
	out, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, bytes.NewReader(cfg))
	if err != nil {
		return err
	}
	env.CmdFlags.Config = configPath
	return nil
}

func (env *ShellEnvironment) Trace(start time.Time, function string, args ...string) {
	if env.CmdFlags == nil || !env.CmdFlags.Debug {
		return
	}
	elapsed := time.Since(start)
	trace := fmt.Sprintf("%s duration: %s, args: %s", function, elapsed, strings.Trim(fmt.Sprint(args), "[]"))
	log.Println(trace)
}

func (env *ShellEnvironment) Log(logType LogType, funcName, message string) {
	if !env.CmdFlags.Debug {
		return
	}
	trace := fmt.Sprintf("%s: %s\n%s", logType, funcName, message)
	log.Println(trace)
}

func (env *ShellEnvironment) debugF(function string, fn func() string) {
	if !env.CmdFlags.Debug {
		return
	}
	trace := fmt.Sprintf("%s: %s\n%s", Debug, function, fn())
	log.Println(trace)
}

func (env *ShellEnvironment) Getenv(key string) string {
	defer env.Trace(time.Now(), "Getenv", key)
	val := os.Getenv(key)
	env.Log(Debug, "Getenv", val)
	return val
}

func (env *ShellEnvironment) Pwd() string {
	defer env.Trace(time.Now(), "Pwd")
	lock.Lock()
	defer func() {
		lock.Unlock()
		env.Log(Debug, "Pwd", env.cwd)
	}()
	if env.cwd != "" {
		return env.cwd
	}
	correctPath := func(pwd string) string {
		// on Windows, and being case sensitive and not consistent and all, this gives silly issues
		driveLetter := regex.GetCompiledRegex(`^[a-z]:`)
		return driveLetter.ReplaceAllStringFunc(pwd, strings.ToUpper)
	}
	if env.CmdFlags != nil && env.CmdFlags.PWD != "" {
		env.cwd = correctPath(env.CmdFlags.PWD)
		return env.cwd
	}
	dir, err := os.Getwd()
	if err != nil {
		env.Log(Error, "Pwd", err.Error())
		return ""
	}
	env.cwd = correctPath(dir)
	return env.cwd
}

func (env *ShellEnvironment) HasFiles(pattern string) bool {
	defer env.Trace(time.Now(), "HasFiles", pattern)
	cwd := env.Pwd()
	pattern = cwd + env.PathSeparator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		env.Log(Error, "HasFiles", err.Error())
		return false
	}
	for _, match := range matches {
		f, _ := os.Stat(match)
		if f.IsDir() {
			continue
		}
		env.Log(Debug, "HasFiles", "true")
		return true
	}
	env.Log(Debug, "HasFiles", "false")
	return false
}

func (env *ShellEnvironment) HasFilesInDir(dir, pattern string) bool {
	defer env.Trace(time.Now(), "HasFilesInDir", pattern)
	pattern = dir + env.PathSeparator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		env.Log(Error, "HasFilesInDir", err.Error())
		return false
	}
	hasFilesInDir := len(matches) > 0
	env.debugF("HasFilesInDir", func() string { return strconv.FormatBool(hasFilesInDir) })
	return hasFilesInDir
}

func (env *ShellEnvironment) HasFileInParentDirs(pattern string, depth uint) bool {
	defer env.Trace(time.Now(), "HasFileInParent", pattern, fmt.Sprint(depth))
	currentFolder := env.Pwd()

	for c := 0; c < int(depth); c++ {
		if env.HasFilesInDir(currentFolder, pattern) {
			env.Log(Debug, "HasFileInParentDirs", "true")
			return true
		}

		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
		} else {
			env.Log(Debug, "HasFileInParentDirs", "false")
			return false
		}
	}
	env.Log(Debug, "HasFileInParentDirs", "false")
	return false
}

func (env *ShellEnvironment) HasFolder(folder string) bool {
	defer env.Trace(time.Now(), "HasFolder", folder)
	f, err := os.Stat(folder)
	if err != nil {
		env.Log(Debug, "HasFolder", "false")
		return false
	}
	env.debugF("HasFolder", func() string { return strconv.FormatBool(f.IsDir()) })
	return f.IsDir()
}

func (env *ShellEnvironment) ResolveSymlink(path string) (string, error) {
	defer env.Trace(time.Now(), "ResolveSymlink", path)
	link, err := filepath.EvalSymlinks(path)
	if err != nil {
		env.Log(Error, "ResolveSymlink", err.Error())
		return "", err
	}
	env.Log(Debug, "ResolveSymlink", link)
	return link, nil
}

func (env *ShellEnvironment) FileContent(file string) string {
	defer env.Trace(time.Now(), "FileContent", file)
	if !filepath.IsAbs(file) {
		file = filepath.Join(env.Pwd(), file)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		env.Log(Error, "FileContent", err.Error())
		return ""
	}
	fileContent := string(content)
	env.Log(Debug, "FileContent", fileContent)
	return fileContent
}

func (env *ShellEnvironment) LsDir(path string) []fs.DirEntry {
	defer env.Trace(time.Now(), "LsDir", path)
	entries, err := os.ReadDir(path)
	if err != nil {
		env.Log(Error, "LsDir", err.Error())
		return nil
	}
	env.debugF("LsDir", func() string {
		var entriesStr string
		for _, entry := range entries {
			entriesStr += entry.Name() + "\n"
		}
		return entriesStr
	})
	return entries
}

func (env *ShellEnvironment) PathSeparator() string {
	defer env.Trace(time.Now(), "PathSeparator")
	return string(os.PathSeparator)
}

func (env *ShellEnvironment) User() string {
	defer env.Trace(time.Now(), "User")
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	env.Log(Debug, "User", user)
	return user
}

func (env *ShellEnvironment) Host() (string, error) {
	defer env.Trace(time.Now(), "Host")
	hostName, err := os.Hostname()
	if err != nil {
		env.Log(Error, "Host", err.Error())
		return "", err
	}
	hostName = cleanHostName(hostName)
	env.Log(Debug, "Host", hostName)
	return hostName, nil
}

func (env *ShellEnvironment) GOOS() string {
	defer env.Trace(time.Now(), "GOOS")
	return runtime.GOOS
}

func (env *ShellEnvironment) RunCommand(command string, args ...string) (string, error) {
	defer env.Trace(time.Now(), "RunCommand", append([]string{command}, args...)...)
	if cacheCommand, ok := env.cmdCache.get(command); ok {
		command = cacheCommand
	}
	output, err := cmd.Run(command, args...)
	if err != nil {
		env.Log(Error, "RunCommand", "cmd.Run() failed")
	}
	env.Log(Debug, "RunCommand", output)
	return output, err
}

func (env *ShellEnvironment) RunShellCommand(shell, command string) string {
	defer env.Trace(time.Now(), "RunShellCommand")
	if out, err := env.RunCommand(shell, "-c", command); err == nil {
		return out
	}
	return ""
}

func (env *ShellEnvironment) CommandPath(command string) string {
	defer env.Trace(time.Now(), "CommandPath", command)
	if path, ok := env.cmdCache.get(command); ok {
		env.Log(Debug, "CommandPath", path)
		return path
	}
	path, err := exec.LookPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		env.Log(Debug, "CommandPath", path)
		return path
	}
	path, err = env.LookWinAppPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		env.Log(Debug, "CommandPath", path)
		return path
	}
	env.Log(Error, "CommandPath", err.Error())
	return ""
}

func (env *ShellEnvironment) HasCommand(command string) bool {
	defer env.Trace(time.Now(), "HasCommand", command)
	if path := env.CommandPath(command); path != "" {
		return true
	}
	return false
}

func (env *ShellEnvironment) ErrorCode() int {
	defer env.Trace(time.Now(), "ErrorCode")
	return env.CmdFlags.ErrorCode
}

func (env *ShellEnvironment) ExecutionTime() float64 {
	defer env.Trace(time.Now(), "ExecutionTime")
	if env.CmdFlags.ExecutionTime < 0 {
		return 0
	}
	return env.CmdFlags.ExecutionTime
}

func (env *ShellEnvironment) Flags() *Flags {
	defer env.Trace(time.Now(), "Flags")
	return env.CmdFlags
}

func (env *ShellEnvironment) Shell() string {
	defer env.Trace(time.Now(), "Shell")
	if env.CmdFlags.Shell != "" {
		return env.CmdFlags.Shell
	}
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		env.Log(Error, "Shell", err.Error())
		return UNKNOWN
	}
	if name == "cmd.exe" {
		p, _ = p.Parent()
		name, err = p.Name()
	}
	if err != nil {
		env.Log(Error, "Shell", err.Error())
		return UNKNOWN
	}
	// Cache the shell value to speed things up.
	env.CmdFlags.Shell = strings.Trim(strings.TrimSuffix(name, ".exe"), " ")
	return env.CmdFlags.Shell
}

func (env *ShellEnvironment) unWrapError(err error) error {
	cause := err
	for {
		type nested interface{ Unwrap() error }
		unwrap, ok := cause.(nested)
		if !ok {
			break
		}
		cause = unwrap.Unwrap()
	}
	return cause
}

func (env *ShellEnvironment) HTTPRequest(targetURL string, body io.Reader, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error) {
	defer env.Trace(time.Now(), "HTTPRequest", targetURL)
	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	defer cncl()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, body)
	if err != nil {
		return nil, err
	}
	for _, modifier := range requestModifiers {
		modifier(request)
	}
	if env.CmdFlags.Debug {
		dump, _ := httputil.DumpRequestOut(request, true)
		env.Log(Debug, "HTTPRequest", string(dump))
	}
	response, err := client.Do(request)
	if err != nil {
		env.Log(Error, "HTTPRequest", err.Error())
		return nil, env.unWrapError(err)
	}
	// anything inside the range [200, 299] is considered a success
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "HTTP status code " + strconv.Itoa(response.StatusCode)
		err := errors.New(message)
		env.Log(Error, "HTTPRequest", message)
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		env.Log(Error, "HTTPRequest", err.Error())
		return nil, err
	}
	env.Log(Debug, "HTTPRequest", string(responseBody))
	return responseBody, nil
}

func (env *ShellEnvironment) HasParentFilePath(path string) (*FileInfo, error) {
	defer env.Trace(time.Now(), "HasParentFilePath", path)
	currentFolder := env.Pwd()
	for {
		searchPath := filepath.Join(currentFolder, path)
		info, err := os.Stat(searchPath)
		if err == nil {
			return &FileInfo{
				ParentFolder: currentFolder,
				Path:         searchPath,
				IsDir:        info.IsDir(),
			}, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
			continue
		}
		env.Log(Error, "HasParentFilePath", err.Error())
		return nil, errors.New("no match at root level")
	}
}

func (env *ShellEnvironment) StackCount() int {
	defer env.Trace(time.Now(), "StackCount")
	if env.CmdFlags.StackCount < 0 {
		return 0
	}
	return env.CmdFlags.StackCount
}

func (env *ShellEnvironment) Cache() Cache {
	return env.fileCache
}

func (env *ShellEnvironment) Close() {
	defer env.Trace(time.Now(), "Close")
	templateCache, err := json.Marshal(env.TemplateCache())
	if err == nil {
		env.fileCache.Set(TEMPLATECACHE, string(templateCache), 1440)
	}
	env.fileCache.Close()
}

func (env *ShellEnvironment) LoadTemplateCache() {
	defer env.Trace(time.Now(), "LoadTemplateCache")
	val, OK := env.fileCache.Get(TEMPLATECACHE)
	if !OK {
		return
	}
	var templateCache TemplateCache
	err := json.Unmarshal([]byte(val), &templateCache)
	if err != nil {
		env.Log(Error, "LoadTemplateCache", err.Error())
		return
	}
	env.tmplCache = &templateCache
}

func (env *ShellEnvironment) Logs() string {
	return env.logBuilder.String()
}

func (env *ShellEnvironment) TemplateCache() *TemplateCache {
	defer env.Trace(time.Now(), "TemplateCache")
	if env.tmplCache != nil {
		return env.tmplCache
	}
	tmplCache := &TemplateCache{
		Root:         env.Root(),
		Shell:        env.Shell(),
		ShellVersion: env.CmdFlags.ShellVersion,
		Code:         env.ErrorCode(),
		WSL:          env.IsWsl(),
	}
	tmplCache.Env = make(map[string]string)
	const separator = "="
	values := os.Environ()
	for value := range values {
		splitted := strings.Split(values[value], separator)
		if len(splitted) != 2 {
			continue
		}
		key := splitted[0]
		val := splitted[1:]
		tmplCache.Env[key] = strings.Join(val, separator)
	}
	tmplCache.PWD = env.Pwd()
	tmplCache.Folder = Base(env, tmplCache.PWD)
	tmplCache.UserName = env.User()
	if host, err := env.Host(); err == nil {
		tmplCache.HostName = host
	}
	goos := env.GOOS()
	tmplCache.OS = goos
	if goos == LINUX {
		tmplCache.OS = env.Platform()
	}
	env.tmplCache = tmplCache
	return tmplCache
}

func (env *ShellEnvironment) DirMatchesOneOf(dir string, regexes []string) (match bool) {
	lock.Lock()
	defer lock.Unlock()
	// sometimes the function panics inside golang, we want to silence that error
	// and assume that there's no match. Not perfect, but better than crashing
	// for the time being until we figure out what the actual root cause is
	defer func() {
		if err := recover(); err != nil {
			message := fmt.Sprintf("%s", err)
			env.Log(Error, "DirMatchesOneOf", message)
			match = false
		}
	}()
	match = dirMatchesOneOf(dir, env.Home(), env.GOOS(), regexes)
	return
}

func dirMatchesOneOf(dir, home, goos string, regexes []string) bool {
	normalizedCwd := strings.ReplaceAll(dir, "\\", "/")
	normalizedHomeDir := strings.ReplaceAll(home, "\\", "/")

	for _, element := range regexes {
		normalizedElement := strings.ReplaceAll(element, "\\\\", "/")
		if strings.HasPrefix(normalizedElement, "~") {
			normalizedElement = strings.Replace(normalizedElement, "~", normalizedHomeDir, 1)
		}
		pattern := fmt.Sprintf("^%s$", normalizedElement)
		if goos == WINDOWS || goos == DARWIN {
			pattern = "(?i)" + pattern
		}
		matched := regex.MatchString(pattern, normalizedCwd)
		if matched {
			return true
		}
	}
	return false
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path consists entirely of separators, Base returns a single separator.
func Base(env Environment, path string) string {
	if path == "/" {
		return path
	}
	volumeName := filepath.VolumeName(path)
	// Strip trailing slashes.
	for len(path) > 0 && string(path[len(path)-1]) == env.PathSeparator() {
		path = path[0 : len(path)-1]
	}
	if volumeName == path {
		return path
	}
	// Throw away volume name
	path = path[len(filepath.VolumeName(path)):]
	// Find the last element
	i := len(path) - 1
	for i >= 0 && string(path[i]) != env.PathSeparator() {
		i--
	}
	if i >= 0 {
		path = path[i+1:]
	}
	// If empty now, it had only slashes.
	if path == "" {
		return env.PathSeparator()
	}
	return path
}

func cleanHostName(hostName string) string {
	garbage := []string{
		".lan",
		".local",
		".localdomain",
	}
	for _, g := range garbage {
		if strings.HasSuffix(hostName, g) {
			hostName = strings.Replace(hostName, g, "", 1)
		}
	}
	return hostName
}

func returnOrBuildCachePath(path string) string {
	// validate root path
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	// validate oh-my-posh folder, if non existent, create it
	cachePath := filepath.Join(path, "oh-my-posh")
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath
	}
	if err := os.Mkdir(cachePath, 0755); err != nil {
		return ""
	}
	return cachePath
}
