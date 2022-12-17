package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"oh-my-posh/log"
	"oh-my-posh/platform/battery"
	"oh-my-posh/platform/cmd"
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
)

const (
	UNKNOWN = "unknown"
	WINDOWS = "windows"
	DARWIN  = "darwin"
	LINUX   = "linux"
)

func getPID() string {
	pid := os.Getenv("POSH_PID")
	if len(pid) == 0 {
		pid = strconv.Itoa(os.Getppid())
	}
	return pid
}

var (
	lock          = sync.RWMutex{}
	TEMPLATECACHE = fmt.Sprintf("template_cache_%s", getPID())
	TOGGLECACHE   = fmt.Sprintf("toggle_cache_%s", getPID())
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
	Manual        bool
	Plain         bool
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

type NotImplemented struct{}

func (n *NotImplemented) Error() string {
	return "not implemented"
}

type ConnectionType string

const (
	ETHERNET  ConnectionType = "ethernet"
	WIFI      ConnectionType = "wifi"
	CELLULAR  ConnectionType = "cellular"
	BLUETOOTH ConnectionType = "bluetooth"
)

type Connection struct {
	Name         string
	Type         ConnectionType
	TransmitRate uint64
	ReceiveRate  uint64
	SSID         string // Wi-Fi only
}

type SegmentsCache map[string]interface{}

func (c *SegmentsCache) Contains(key string) bool {
	_, ok := (*c)[key]
	return ok
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
	Segments     SegmentsCache
}

func (t *TemplateCache) AddSegmentData(key string, value interface{}) {
	lock.Lock()
	defer lock.Unlock()
	if t.Segments == nil {
		t.Segments = make(map[string]interface{})
	}
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
	Connection(connectionType ConnectionType) (*Connection, error)
	TemplateCache() *TemplateCache
	LoadTemplateCache()
	Debug(funcName, message string)
	Error(funcName string, err error)
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

type Shell struct {
	CmdFlags *Flags
	Version  string

	cwd       string
	cmdCache  *commandCache
	fileCache *fileCache
	tmplCache *TemplateCache
	networks  []*Connection
}

func (env *Shell) Init() {
	defer env.Trace(time.Now(), "Init")
	if env.CmdFlags == nil {
		env.CmdFlags = &Flags{}
	}
	if env.CmdFlags.Debug {
		log.Enable()
	}
	env.fileCache = &fileCache{}
	env.fileCache.Init(env.CachePath())
	env.resolveConfigPath()
	env.cmdCache = &commandCache{
		commands: newConcurrentMap(),
	}
}

func (env *Shell) resolveConfigPath() {
	defer env.Trace(time.Now(), "resolveConfigPath")
	if len(env.CmdFlags.Config) == 0 {
		env.CmdFlags.Config = env.Getenv("POSH_THEME")
	}
	if len(env.CmdFlags.Config) == 0 {
		return
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
		configFile = filepath.Join(env.Pwd(), configFile)
	}
	env.CmdFlags.Config = filepath.Clean(configFile)
}

func (env *Shell) downloadConfig(location string) error {
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

func (env *Shell) Trace(start time.Time, function string, args ...string) {
	log.Trace(start, function, args...)
}

func (env *Shell) Debug(funcName, message string) {
	log.Debug(funcName, message)
}

func (env *Shell) Error(funcName string, err error) {
	log.Error(funcName, err)
}

func (env *Shell) debugF(function string, fn func() string) {
	log.DebugF(function, fn)
}

func (env *Shell) Getenv(key string) string {
	defer env.Trace(time.Now(), "Getenv", key)
	val := os.Getenv(key)
	env.Debug("Getenv", val)
	return val
}

func (env *Shell) Pwd() string {
	defer env.Trace(time.Now(), "Pwd")
	lock.Lock()
	defer func() {
		lock.Unlock()
		env.Debug("Pwd", env.cwd)
	}()
	if env.cwd != "" {
		return env.cwd
	}
	correctPath := func(pwd string) string {
		if env.GOOS() != WINDOWS {
			return pwd
		}
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
		env.Error("Pwd", err)
		return ""
	}
	env.cwd = correctPath(dir)
	return env.cwd
}

func (env *Shell) HasFiles(pattern string) bool {
	defer env.Trace(time.Now(), "HasFiles", pattern)
	cwd := env.Pwd()
	fileSystem := os.DirFS(cwd)
	matches, err := fs.Glob(fileSystem, pattern)
	if err != nil {
		env.Error("HasFiles", err)
		return false
	}
	for _, match := range matches {
		file, err := fs.Stat(fileSystem, match)
		if err != nil || file.IsDir() {
			continue
		}
		return true
	}
	return false
}

func (env *Shell) HasFilesInDir(dir, pattern string) bool {
	defer env.Trace(time.Now(), "HasFilesInDir", pattern)
	fileSystem := os.DirFS(dir)
	matches, err := fs.Glob(fileSystem, pattern)
	if err != nil {
		env.Error("HasFilesInDir", err)
		return false
	}
	hasFilesInDir := len(matches) > 0
	env.debugF("HasFilesInDir", func() string { return strconv.FormatBool(hasFilesInDir) })
	return hasFilesInDir
}

func (env *Shell) HasFileInParentDirs(pattern string, depth uint) bool {
	defer env.Trace(time.Now(), "HasFileInParent", pattern, fmt.Sprint(depth))
	currentFolder := env.Pwd()

	for c := 0; c < int(depth); c++ {
		if env.HasFilesInDir(currentFolder, pattern) {
			env.Debug("HasFileInParentDirs", "true")
			return true
		}

		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
		} else {
			env.Debug("HasFileInParentDirs", "false")
			return false
		}
	}
	env.Debug("HasFileInParentDirs", "false")
	return false
}

func (env *Shell) HasFolder(folder string) bool {
	defer env.Trace(time.Now(), "HasFolder", folder)
	f, err := os.Stat(folder)
	if err != nil {
		env.Debug("HasFolder", "false")
		return false
	}
	env.debugF("HasFolder", func() string { return strconv.FormatBool(f.IsDir()) })
	return f.IsDir()
}

func (env *Shell) ResolveSymlink(path string) (string, error) {
	defer env.Trace(time.Now(), "ResolveSymlink", path)
	link, err := filepath.EvalSymlinks(path)
	if err != nil {
		env.Error("ResolveSymlink", err)
		return "", err
	}
	env.Debug("ResolveSymlink", link)
	return link, nil
}

func (env *Shell) FileContent(file string) string {
	defer env.Trace(time.Now(), "FileContent", file)
	if !filepath.IsAbs(file) {
		file = filepath.Join(env.Pwd(), file)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		env.Error("FileContent", err)
		return ""
	}
	fileContent := string(content)
	env.Debug("FileContent", fileContent)
	return fileContent
}

func (env *Shell) LsDir(path string) []fs.DirEntry {
	defer env.Trace(time.Now(), "LsDir", path)
	entries, err := os.ReadDir(path)
	if err != nil {
		env.Error("LsDir", err)
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

func (env *Shell) PathSeparator() string {
	defer env.Trace(time.Now(), "PathSeparator")
	return string(os.PathSeparator)
}

func (env *Shell) User() string {
	defer env.Trace(time.Now(), "User")
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	env.Debug("User", user)
	return user
}

func (env *Shell) Host() (string, error) {
	defer env.Trace(time.Now(), "Host")
	hostName, err := os.Hostname()
	if err != nil {
		env.Error("Host", err)
		return "", err
	}
	hostName = cleanHostName(hostName)
	env.Debug("Host", hostName)
	return hostName, nil
}

func (env *Shell) GOOS() string {
	defer env.Trace(time.Now(), "GOOS")
	return runtime.GOOS
}

func (env *Shell) RunCommand(command string, args ...string) (string, error) {
	defer env.Trace(time.Now(), "RunCommand", append([]string{command}, args...)...)
	if cacheCommand, ok := env.cmdCache.get(command); ok {
		command = cacheCommand
	}
	output, err := cmd.Run(command, args...)
	if err != nil {
		env.Error("RunCommand", err)
	}
	env.Debug("RunCommand", output)
	return output, err
}

func (env *Shell) RunShellCommand(shell, command string) string {
	defer env.Trace(time.Now(), "RunShellCommand")
	if out, err := env.RunCommand(shell, "-c", command); err == nil {
		return out
	}
	return ""
}

func (env *Shell) CommandPath(command string) string {
	defer env.Trace(time.Now(), "CommandPath", command)
	if path, ok := env.cmdCache.get(command); ok {
		env.Debug("CommandPath", path)
		return path
	}
	path, err := exec.LookPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		env.Debug("CommandPath", path)
		return path
	}
	env.Error("CommandPath", err)
	return ""
}

func (env *Shell) HasCommand(command string) bool {
	defer env.Trace(time.Now(), "HasCommand", command)
	if path := env.CommandPath(command); path != "" {
		return true
	}
	return false
}

func (env *Shell) ErrorCode() int {
	defer env.Trace(time.Now(), "ErrorCode")
	return env.CmdFlags.ErrorCode
}

func (env *Shell) ExecutionTime() float64 {
	defer env.Trace(time.Now(), "ExecutionTime")
	if env.CmdFlags.ExecutionTime < 0 {
		return 0
	}
	return env.CmdFlags.ExecutionTime
}

func (env *Shell) Flags() *Flags {
	defer env.Trace(time.Now(), "Flags")
	return env.CmdFlags
}

func (env *Shell) Shell() string {
	defer env.Trace(time.Now(), "Shell")
	if env.CmdFlags.Shell != "" {
		return env.CmdFlags.Shell
	}
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		env.Error("Shell", err)
		return UNKNOWN
	}
	env.Debug("Shell", "process name: "+name)
	// this is used for when scoop creates a shim, see
	// https://github.com/JanDeDobbeleer/oh-my-posh/issues/2806
	executable, _ := os.Executable()
	if name == "cmd.exe" || name == executable {
		p, _ = p.Parent()
		name, err = p.Name()
		env.Debug("Shell", "parent process name: "+name)
	}
	if err != nil {
		env.Error("Shell", err)
		return UNKNOWN
	}
	// Cache the shell value to speed things up.
	env.CmdFlags.Shell = strings.Trim(strings.TrimSuffix(name, ".exe"), " ")
	return env.CmdFlags.Shell
}

func (env *Shell) unWrapError(err error) error {
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

func (env *Shell) HTTPRequest(targetURL string, body io.Reader, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error) {
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
		env.Debug("HTTPRequest", string(dump))
	}
	response, err := client.Do(request)
	if err != nil {
		env.Error("HTTPRequest", err)
		return nil, env.unWrapError(err)
	}
	// anything inside the range [200, 299] is considered a success
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "HTTP status code " + strconv.Itoa(response.StatusCode)
		err := errors.New(message)
		env.Error("HTTPRequest", err)
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		env.Error("HTTPRequest", err)
		return nil, err
	}
	env.Debug("HTTPRequest", string(responseBody))
	return responseBody, nil
}

func (env *Shell) HasParentFilePath(path string) (*FileInfo, error) {
	defer env.Trace(time.Now(), "HasParentFilePath", path)
	currentFolder := env.Pwd()
	for {
		fileSystem := os.DirFS(currentFolder)
		info, err := fs.Stat(fileSystem, path)
		if err == nil {
			return &FileInfo{
				ParentFolder: currentFolder,
				Path:         filepath.Join(currentFolder, path),
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
		env.Error("HasParentFilePath", err)
		return nil, errors.New("no match at root level")
	}
}

func (env *Shell) StackCount() int {
	defer env.Trace(time.Now(), "StackCount")
	if env.CmdFlags.StackCount < 0 {
		return 0
	}
	return env.CmdFlags.StackCount
}

func (env *Shell) Cache() Cache {
	return env.fileCache
}

func (env *Shell) Close() {
	defer env.Trace(time.Now(), "Close")
	templateCache, err := json.Marshal(env.TemplateCache())
	if err == nil {
		env.fileCache.Set(TEMPLATECACHE, string(templateCache), 1440)
	}
	env.fileCache.Close()
}

func (env *Shell) LoadTemplateCache() {
	defer env.Trace(time.Now(), "LoadTemplateCache")
	val, OK := env.fileCache.Get("template_cache_91508")
	if !OK {
		return
	}
	var templateCache TemplateCache
	err := json.Unmarshal([]byte(val), &templateCache)
	if err != nil {
		env.Error("LoadTemplateCache", err)
		return
	}
	env.tmplCache = &templateCache
}

func (env *Shell) Logs() string {
	return log.String()
}

func (env *Shell) TemplateCache() *TemplateCache {
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
	pwd := env.Pwd()
	tmplCache.PWD = ReplaceHomeDirPrefixWithTilde(env, pwd)
	tmplCache.Folder = Base(env, pwd)
	if env.GOOS() == WINDOWS && strings.HasSuffix(tmplCache.Folder, ":") {
		tmplCache.Folder += `\`
	}
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

func (env *Shell) DirMatchesOneOf(dir string, regexes []string) (match bool) {
	lock.Lock()
	defer lock.Unlock()
	// sometimes the function panics inside golang, we want to silence that error
	// and assume that there's no match. Not perfect, but better than crashing
	// for the time being until we figure out what the actual root cause is
	defer func() {
		if err := recover(); err != nil {
			env.Error("DirMatchesOneOf", errors.New("panic"))
			match = false
		}
	}()
	match = dirMatchesOneOf(dir, env.Home(), env.GOOS(), regexes)
	return
}

func dirMatchesOneOf(dir, home, goos string, regexes []string) bool {
	if goos == WINDOWS {
		dir = strings.ReplaceAll(dir, "\\", "/")
		home = strings.ReplaceAll(home, "\\", "/")
	}

	for _, element := range regexes {
		normalizedElement := strings.ReplaceAll(element, "\\\\", "/")
		if strings.HasPrefix(normalizedElement, "~") {
			normalizedElement = strings.Replace(normalizedElement, "~", home, 1)
		}
		pattern := fmt.Sprintf("^%s$", normalizedElement)
		if goos == WINDOWS || goos == DARWIN {
			pattern = "(?i)" + pattern
		}
		matched := regex.MatchString(pattern, dir)
		if matched {
			return true
		}
	}
	return false
}

func IsPathSeparator(env Environment, c uint8) bool {
	if c == '/' {
		return true
	}
	if env.GOOS() == WINDOWS && c == '\\' {
		return true
	}
	return false
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path consists entirely of separators, Base returns a single separator.
func Base(env Environment, path string) string {
	volumeName := filepath.VolumeName(path)
	// Strip trailing slashes.
	for len(path) > 0 && IsPathSeparator(env, path[len(path)-1]) {
		path = path[0 : len(path)-1]
	}
	if len(path) == 0 {
		return env.PathSeparator()
	}
	if volumeName == path {
		return path
	}
	// Throw away volume name
	path = path[len(filepath.VolumeName(path)):]
	// Find the last element
	i := len(path) - 1
	for i >= 0 && !IsPathSeparator(env, path[i]) {
		i--
	}
	if i >= 0 {
		path = path[i+1:]
	}
	// If empty now, it had only slashes.
	if len(path) == 0 {
		return env.PathSeparator()
	}
	return path
}

func ReplaceHomeDirPrefixWithTilde(env Environment, path string) string {
	home := env.Home()
	// match Home directory exactly
	if !strings.HasPrefix(path, home) {
		return path
	}
	rem := path[len(home):]
	if len(rem) == 0 || IsPathSeparator(env, rem[0]) {
		return "~" + rem
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
