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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/platform/battery"
	"github.com/jandedobbeleer/oh-my-posh/src/platform/cmd"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"

	cpu "github.com/shirou/gopsutil/v3/cpu"
	disk "github.com/shirou/gopsutil/v3/disk"
	load "github.com/shirou/gopsutil/v3/load"
	process "github.com/shirou/gopsutil/v3/process"
)

const (
	UNKNOWN = "unknown"
	WINDOWS = "windows"
	DARWIN  = "darwin"
	LINUX   = "linux"
)

func pid() string {
	pid := os.Getenv("POSH_PID")
	if len(pid) == 0 {
		pid = strconv.Itoa(os.Getppid())
	}
	return pid
}

var (
	TEMPLATECACHE    = fmt.Sprintf("template_cache_%s", pid())
	TOGGLECACHE      = fmt.Sprintf("toggle_cache_%s", pid())
	PROMPTCOUNTCACHE = fmt.Sprintf("prompt_count_cache_%s", pid())
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
	Primary       bool
	PromptCount   int
	Cleared       bool
	Version       string
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
	// Gets the value for a given key.
	// Returns the value and a boolean indicating if the key was found.
	// In case the ttl expired, the function returns false.
	Get(key string) (string, bool)
	// Sets a value for a given key.
	// The ttl indicates how may minutes to cache the value.
	Set(key, value string, ttl int)
	// Deletes a key from the cache.
	Delete(key string)
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

type Memory struct {
	PhysicalTotalMemory     uint64
	PhysicalAvailableMemory uint64
	PhysicalFreeMemory      uint64
	PhysicalPercentUsed     float64
	SwapTotalMemory         uint64
	SwapFreeMemory          uint64
	SwapPercentUsed         float64
}

type SystemInfo struct {
	// mem
	Memory
	// cpu
	Times float64
	CPU   []cpu.InfoStat
	// load
	Load1  float64
	Load5  float64
	Load15 float64
	// disk
	Disks map[string]disk.IOCountersStat
}

type SegmentsCache map[string]interface{}

func (s *SegmentsCache) Contains(key string) bool {
	_, ok := (*s)[key]
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
	Var          map[string]interface{}
	OS           string
	WSL          bool
	PromptCount  int
	Segments     SegmentsCache

	sync.RWMutex
}

func (t *TemplateCache) AddSegmentData(key string, value interface{}) {
	t.Lock()
	t.Segments[key] = value
	t.Unlock()
}

func (t *TemplateCache) RemoveSegmentData(key string) {
	t.Lock()
	delete(t.Segments, key)
	t.Unlock()
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
	SetPromptCount()
	CursorPosition() (row, col int)
	SystemInfo() (*SystemInfo, error)
	Debug(message string)
	Error(err error)
	Trace(start time.Time, args ...string)
}

type commandCache struct {
	commands *ConcurrentMap
}

func (c *commandCache) set(command, path string) {
	c.commands.Set(command, path)
}

func (c *commandCache) get(command string) (string, bool) {
	cacheCommand, found := c.commands.Get(command)
	if !found {
		return "", false
	}
	command, ok := cacheCommand.(string)
	return command, ok
}

type Shell struct {
	CmdFlags *Flags
	Var      map[string]interface{}

	cwd       string
	cmdCache  *commandCache
	fileCache *fileCache
	tmplCache *TemplateCache
	networks  []*Connection

	sync.RWMutex
}

func (env *Shell) Init() {
	defer env.Trace(time.Now())
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
		commands: NewConcurrentMap(),
	}
	env.SetPromptCount()
}

func (env *Shell) resolveConfigPath() {
	defer env.Trace(time.Now())
	if len(env.CmdFlags.Config) == 0 {
		env.CmdFlags.Config = env.Getenv("POSH_THEME")
	}
	if len(env.CmdFlags.Config) == 0 {
		env.Debug("No config set, fallback to default config")
		return
	}
	if strings.HasPrefix(env.CmdFlags.Config, "https://") {
		if err := env.downloadConfig(env.CmdFlags.Config); err != nil {
			// make it use default config when download fails
			env.Error(err)
			env.CmdFlags.Config = ""
			return
		}
	}
	// Cygwin path always needs the full path as we're on Windows but not really.
	// Doing filepath actions will convert it to a Windows path and break the init script.
	if env.Platform() == WINDOWS && env.Shell() == "bash" {
		env.Debug("Cygwin detected, using full path for config")
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
	defer env.Trace(time.Now(), location)
	ext := filepath.Ext(location)
	configPath := filepath.Join(env.CachePath(), "config.omp"+ext)
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

func (env *Shell) Trace(start time.Time, args ...string) {
	log.Trace(start, args...)
}

func (env *Shell) Debug(message string) {
	log.Debug(message)
}

func (env *Shell) Error(err error) {
	log.Error(err)
}

func (env *Shell) debugF(fn func() string) {
	log.DebugF(fn)
}

func (env *Shell) Getenv(key string) string {
	defer env.Trace(time.Now(), key)
	val := os.Getenv(key)
	env.Debug(val)
	return val
}

func (env *Shell) Pwd() string {
	defer env.Trace(time.Now())
	defer env.Debug(env.cwd)
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
		env.Error(err)
		return ""
	}
	env.cwd = correctPath(dir)
	return env.cwd
}

func (env *Shell) HasFiles(pattern string) bool {
	defer env.Trace(time.Now(), pattern)
	cwd := env.Pwd()
	fileSystem := os.DirFS(cwd)
	matches, err := fs.Glob(fileSystem, pattern)
	if err != nil {
		env.Error(err)
		env.Debug("false")
		return false
	}
	for _, match := range matches {
		file, err := fs.Stat(fileSystem, match)
		if err != nil || file.IsDir() {
			continue
		}
		env.Debug("true")
		return true
	}
	env.Debug("false")
	return false
}

func (env *Shell) HasFilesInDir(dir, pattern string) bool {
	defer env.Trace(time.Now(), pattern)
	fileSystem := os.DirFS(dir)
	matches, err := fs.Glob(fileSystem, pattern)
	if err != nil {
		env.Error(err)
		env.Debug("false")
		return false
	}
	hasFilesInDir := len(matches) > 0
	env.debugF(func() string { return strconv.FormatBool(hasFilesInDir) })
	return hasFilesInDir
}

func (env *Shell) HasFileInParentDirs(pattern string, depth uint) bool {
	defer env.Trace(time.Now(), pattern, fmt.Sprint(depth))
	currentFolder := env.Pwd()

	for c := 0; c < int(depth); c++ {
		if env.HasFilesInDir(currentFolder, pattern) {
			env.Debug("true")
			return true
		}

		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
		} else {
			env.Debug("false")
			return false
		}
	}
	env.Debug("false")
	return false
}

func (env *Shell) HasFolder(folder string) bool {
	defer env.Trace(time.Now(), folder)
	f, err := os.Stat(folder)
	if err != nil {
		env.Debug("false")
		return false
	}
	env.debugF(func() string { return strconv.FormatBool(f.IsDir()) })
	return f.IsDir()
}

func (env *Shell) ResolveSymlink(path string) (string, error) {
	defer env.Trace(time.Now(), path)
	link, err := filepath.EvalSymlinks(path)
	if err != nil {
		env.Error(err)
		return "", err
	}
	env.Debug(link)
	return link, nil
}

func (env *Shell) FileContent(file string) string {
	defer env.Trace(time.Now(), file)
	if !filepath.IsAbs(file) {
		file = filepath.Join(env.Pwd(), file)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		env.Error(err)
		return ""
	}
	fileContent := string(content)
	env.Debug(fileContent)
	return fileContent
}

func (env *Shell) LsDir(path string) []fs.DirEntry {
	defer env.Trace(time.Now(), path)
	entries, err := os.ReadDir(path)
	if err != nil {
		env.Error(err)
		return nil
	}
	env.debugF(func() string {
		var entriesStr string
		for _, entry := range entries {
			entriesStr += entry.Name() + "\n"
		}
		return entriesStr
	})
	return entries
}

func (env *Shell) PathSeparator() string {
	defer env.Trace(time.Now())
	return string(os.PathSeparator)
}

func (env *Shell) User() string {
	defer env.Trace(time.Now())
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	env.Debug(user)
	return user
}

func (env *Shell) Host() (string, error) {
	defer env.Trace(time.Now())
	hostName, err := os.Hostname()
	if err != nil {
		env.Error(err)
		return "", err
	}
	hostName = cleanHostName(hostName)
	env.Debug(hostName)
	return hostName, nil
}

func (env *Shell) GOOS() string {
	defer env.Trace(time.Now())
	return runtime.GOOS
}

func (env *Shell) RunCommand(command string, args ...string) (string, error) {
	defer env.Trace(time.Now(), append([]string{command}, args...)...)
	if cacheCommand, ok := env.cmdCache.get(command); ok {
		command = cacheCommand
	}
	output, err := cmd.Run(command, args...)
	if err != nil {
		env.Error(err)
	}
	env.Debug(output)
	return output, err
}

func (env *Shell) RunShellCommand(shell, command string) string {
	defer env.Trace(time.Now())
	if out, err := env.RunCommand(shell, "-c", command); err == nil {
		return out
	}
	return ""
}

func (env *Shell) CommandPath(command string) string {
	defer env.Trace(time.Now(), command)
	if path, ok := env.cmdCache.get(command); ok {
		env.Debug(path)
		return path
	}
	path, err := exec.LookPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		env.Debug(path)
		return path
	}
	env.Error(err)
	return ""
}

func (env *Shell) HasCommand(command string) bool {
	defer env.Trace(time.Now(), command)
	if path := env.CommandPath(command); path != "" {
		return true
	}
	return false
}

func (env *Shell) ErrorCode() int {
	defer env.Trace(time.Now())
	return env.CmdFlags.ErrorCode
}

func (env *Shell) ExecutionTime() float64 {
	defer env.Trace(time.Now())
	if env.CmdFlags.ExecutionTime < 0 {
		return 0
	}
	return env.CmdFlags.ExecutionTime
}

func (env *Shell) Flags() *Flags {
	defer env.Trace(time.Now())
	return env.CmdFlags
}

func (env *Shell) Shell() string {
	defer env.Trace(time.Now())
	if env.CmdFlags.Shell != "" {
		return env.CmdFlags.Shell
	}
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		env.Error(err)
		return UNKNOWN
	}
	env.Debug("process name: " + name)
	// this is used for when scoop creates a shim, see
	// https://github.com/jandedobbeleer/oh-my-posh/issues/2806
	executable, _ := os.Executable()
	if name == "cmd.exe" || name == executable {
		p, _ = p.Parent()
		name, err = p.Name()
		env.Debug("parent process name: " + name)
	}
	if err != nil {
		env.Error(err)
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
	defer env.Trace(time.Now(), targetURL)
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
		env.Debug(string(dump))
	}
	response, err := client.Do(request)
	if err != nil {
		env.Error(err)
		return nil, env.unWrapError(err)
	}
	// anything inside the range [200, 299] is considered a success
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "HTTP status code " + strconv.Itoa(response.StatusCode)
		err := errors.New(message)
		env.Error(err)
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		env.Error(err)
		return nil, err
	}
	env.Debug(string(responseBody))
	return responseBody, nil
}

func (env *Shell) HasParentFilePath(path string) (*FileInfo, error) {
	defer env.Trace(time.Now(), path)
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
		env.Error(err)
		return nil, errors.New("no match at root level")
	}
}

func (env *Shell) StackCount() int {
	defer env.Trace(time.Now())
	if env.CmdFlags.StackCount < 0 {
		return 0
	}
	return env.CmdFlags.StackCount
}

func (env *Shell) Cache() Cache {
	return env.fileCache
}

func (env *Shell) Close() {
	defer env.Trace(time.Now())
	templateCache, err := json.Marshal(env.TemplateCache())
	if err == nil {
		env.fileCache.Set(TEMPLATECACHE, string(templateCache), 1440)
	}
	env.fileCache.Close()
}

func (env *Shell) LoadTemplateCache() {
	defer env.Trace(time.Now())
	val, OK := env.fileCache.Get(TEMPLATECACHE)
	if !OK {
		return
	}
	var templateCache TemplateCache
	err := json.Unmarshal([]byte(val), &templateCache)
	if err != nil {
		env.Error(err)
		return
	}
	env.tmplCache = &templateCache
}

func (env *Shell) Logs() string {
	return log.String()
}

func (env *Shell) TemplateCache() *TemplateCache {
	env.Lock()
	defer env.Trace(time.Now())
	defer env.Unlock()
	if env.tmplCache != nil {
		return env.tmplCache
	}

	tmplCache := &TemplateCache{
		Root:         env.Root(),
		Shell:        env.Shell(),
		ShellVersion: env.CmdFlags.ShellVersion,
		Code:         env.ErrorCode(),
		WSL:          env.IsWsl(),
		Segments:     make(map[string]interface{}),
		PromptCount:  env.CmdFlags.PromptCount,
	}
	tmplCache.Env = make(map[string]string)
	tmplCache.Var = make(map[string]interface{})

	if env.Var != nil {
		tmplCache.Var = env.Var
	}

	const separator = "="
	values := os.Environ()
	for value := range values {
		key, val, valid := strings.Cut(values[value], separator)
		if !valid {
			continue
		}
		tmplCache.Env[key] = val
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
	// sometimes the function panics inside golang, we want to silence that error
	// and assume that there's no match. Not perfect, but better than crashing
	// for the time being until we figure out what the actual root cause is
	defer func() {
		if err := recover(); err != nil {
			env.Error(errors.New("panic"))
			match = false
		}
	}()
	match = dirMatchesOneOf(dir, env.Home(), env.GOOS(), regexes)
	return
}

func dirMatchesOneOf(dir, home, goos string, regexes []string) bool {
	if len(regexes) == 0 {
		return false
	}

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

func (env *Shell) SetPromptCount() {
	countStr := os.Getenv("POSH_PROMPT_COUNT")
	if len(countStr) > 0 {
		// this counter is incremented by the shell
		count, err := strconv.Atoi(countStr)
		if err == nil {
			env.CmdFlags.PromptCount = count
			return
		}
	}
	var count int
	if val, found := env.Cache().Get(PROMPTCOUNTCACHE); found {
		count, _ = strconv.Atoi(val)
	}
	// only write to cache if we're the primary prompt
	if env.CmdFlags.Primary {
		count++
		env.Cache().Set(PROMPTCOUNTCACHE, strconv.Itoa(count), 1440)
	}
	env.CmdFlags.PromptCount = count
}

func (env *Shell) CursorPosition() (row, col int) {
	if number, err := strconv.Atoi(env.Getenv("POSH_CURSOR_LINE")); err == nil {
		row = number
	}
	if number, err := strconv.Atoi(env.Getenv("POSH_CURSOR_COLUMN")); err != nil {
		col = number
	}
	return
}

func (env *Shell) SystemInfo() (*SystemInfo, error) {
	s := &SystemInfo{}

	mem, err := env.Memory()
	if err != nil {
		return nil, err
	}
	s.Memory = *mem

	loadStat, err := load.Avg()
	if err == nil {
		s.Load1 = loadStat.Load1
		s.Load5 = loadStat.Load5
		s.Load15 = loadStat.Load15
	}

	processorTimes, err := cpu.Percent(0, false)
	if err == nil && len(processorTimes) > 0 {
		s.Times = processorTimes[0]
	}

	processors, err := cpu.Info()
	if err == nil {
		s.CPU = processors
	}

	diskIO, err := disk.IOCounters()
	if err == nil {
		s.Disks = diskIO
	}
	return s, nil
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
