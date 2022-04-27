package environment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"oh-my-posh/regex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/distatus/battery"
	process "github.com/shirou/gopsutil/v3/process"
)

const (
	Unknown         = "unknown"
	WindowsPlatform = "windows"
	DarwinPlatform  = "darwin"
	LinuxPlatform   = "linux"
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
}

type CommandError struct {
	Err      string
	ExitCode int
}

func (e *CommandError) Error() string {
	return e.Err
}

type NoBatteryError struct{}

func (m *NoBatteryError) Error() string {
	return "no battery"
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

type WindowsRegistryValueType int

const (
	RegQword WindowsRegistryValueType = iota
	RegDword
	RegString
)

type WindowsRegistryValue struct {
	ValueType WindowsRegistryValueType
	Qword     uint64
	Dword     uint32
	Str       string
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
}

type BatteryInfo struct {
	Percentage int
	State      battery.State
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
	CommandPath(command string) string
	HasCommand(command string) bool
	FileContent(file string) string
	LsDir(path string) []fs.DirEntry
	RunCommand(command string, args ...string) (string, error)
	RunShellCommand(shell, command string) string
	ExecutionTime() float64
	Flags() *Flags
	BatteryState() (*BatteryInfo, error)
	QueryWindowTitles(processName, windowTitleRegex string) (string, error)
	WindowsRegistryKeyValue(path string) (*WindowsRegistryValue, error)
	HTTPRequest(url string, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error)
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
	WifiNetwork() (*WifiInfo, error)
	TemplateCache() *TemplateCache
}

type commandCache struct {
	commands *concurrentMap
}

func (c *commandCache) set(command, path string) {
	c.commands.set(command, path)
}

func (c *commandCache) get(command string) (string, bool) {
	cmd, found := c.commands.get(command)
	if !found {
		return "", false
	}
	command, ok := cmd.(string)
	return command, ok
}

type logType string

const (
	Error logType = "error"
	Debug logType = "debug"
)

type ShellEnvironment struct {
	CmdFlags *Flags
	Version  string

	cwd        string
	cmdCache   *commandCache
	fileCache  *fileCache
	tmplCache  *TemplateCache
	logBuilder strings.Builder
	debug      bool
	lock       sync.Mutex
}

func (env *ShellEnvironment) Init(debug bool) {
	if env.CmdFlags == nil {
		env.CmdFlags = &Flags{}
	}
	env.fileCache = &fileCache{}
	env.fileCache.Init(env.CachePath())
	env.resolveConfigPath()
	env.cmdCache = &commandCache{
		commands: newConcurrentMap(),
	}
	if debug {
		env.debug = true
		log.SetOutput(&env.logBuilder)
	}
}

func (env *ShellEnvironment) resolveConfigPath() {
	if len(env.CmdFlags.Config) == 0 {
		env.CmdFlags.Config = env.Getenv("POSH_THEME")
	}
	if len(env.CmdFlags.Config) == 0 {
		env.CmdFlags.Config = fmt.Sprintf("https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/v%s/themes/default.omp.json", env.Version)
	}
	if strings.HasPrefix(env.CmdFlags.Config, "https://") {
		if err := env.downloadConfig(env.CmdFlags.Config); err == nil {
			return
		}
	}
	// Cygwin path always needs the full path as we're on Windows but not really.
	// Doing filepath actions will convert it to a Windows path and break the init script.
	if env.Platform() == WindowsPlatform && env.Shell() == "constants.BASH" {
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
	configFileName := fmt.Sprintf("%s.omp.json", env.Version)
	configPath := filepath.Join(env.CachePath(), configFileName)
	if env.HasFilesInDir(env.CachePath(), configFileName) {
		env.CmdFlags.Config = configPath
		return nil
	}
	// clean old config files
	cleanCacheDir := func() {
		dir, err := ioutil.ReadDir(env.CachePath())
		if err != nil {
			return
		}
		for _, file := range dir {
			if strings.HasSuffix(file.Name(), ".omp.json") {
				os.Remove(filepath.Join(env.CachePath(), file.Name()))
			}
		}
	}
	cleanCacheDir()

	cfg, err := env.HTTPRequest(location, 5000)
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

func (env *ShellEnvironment) trace(start time.Time, function string, args ...string) {
	if !env.debug {
		return
	}
	elapsed := time.Since(start)
	trace := fmt.Sprintf("%s duration: %s, args: %s", function, elapsed, strings.Trim(fmt.Sprint(args), "[]"))
	log.Println(trace)
}

func (env *ShellEnvironment) log(lt logType, function, message string) {
	if !env.debug {
		return
	}
	trace := fmt.Sprintf("%s: %s\n%s", lt, function, message)
	log.Println(trace)
}

func (env *ShellEnvironment) debugF(function string, fn func() string) {
	if !env.debug {
		return
	}
	trace := fmt.Sprintf("%s: %s\n%s", Debug, function, fn())
	log.Println(trace)
}

func (env *ShellEnvironment) Getenv(key string) string {
	defer env.trace(time.Now(), "Getenv", key)
	val := os.Getenv(key)
	env.log(Debug, "Getenv", val)
	return val
}

func (env *ShellEnvironment) Pwd() string {
	defer env.trace(time.Now(), "Pwd")
	env.lock.Lock()
	defer func() {
		env.lock.Unlock()
		env.log(Debug, "Pwd", env.cwd)
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
		env.log(Error, "Pwd", err.Error())
		return ""
	}
	env.cwd = correctPath(dir)
	return env.cwd
}

func (env *ShellEnvironment) HasFiles(pattern string) bool {
	defer env.trace(time.Now(), "HasFiles", pattern)
	cwd := env.Pwd()
	pattern = cwd + env.PathSeparator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		env.log(Error, "HasFiles", err.Error())
		return false
	}
	for _, match := range matches {
		f, _ := os.Stat(match)
		if f.IsDir() {
			continue
		}
		env.log(Debug, "HasFiles", "true")
		return true
	}
	env.log(Debug, "HasFiles", "false")
	return false
}

func (env *ShellEnvironment) HasFilesInDir(dir, pattern string) bool {
	defer env.trace(time.Now(), "HasFilesInDir", pattern)
	pattern = dir + env.PathSeparator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		env.log(Error, "HasFilesInDir", err.Error())
		return false
	}
	hasFilesInDir := len(matches) > 0
	env.debugF("HasFilesInDir", func() string { return strconv.FormatBool(hasFilesInDir) })
	return hasFilesInDir
}

func (env *ShellEnvironment) HasFileInParentDirs(pattern string, depth uint) bool {
	defer env.trace(time.Now(), "HasFileInParent", pattern, fmt.Sprint(depth))
	currentFolder := env.Pwd()

	for c := 0; c < int(depth); c++ {
		if env.HasFilesInDir(currentFolder, pattern) {
			env.log(Debug, "HasFileInParentDirs", "true")
			return true
		}

		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
		} else {
			env.log(Debug, "HasFileInParentDirs", "false")
			return false
		}
	}
	env.log(Debug, "HasFileInParentDirs", "false")
	return false
}

func (env *ShellEnvironment) HasFolder(folder string) bool {
	defer env.trace(time.Now(), "HasFolder", folder)
	f, err := os.Stat(folder)
	if err != nil {
		env.log(Debug, "HasFolder", "false")
		return false
	}
	env.debugF("HasFolder", func() string { return strconv.FormatBool(f.IsDir()) })
	return f.IsDir()
}

func (env *ShellEnvironment) ResolveSymlink(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

func (env *ShellEnvironment) FileContent(file string) string {
	defer env.trace(time.Now(), "FileContent", file)
	if !filepath.IsAbs(file) {
		file = filepath.Join(env.Pwd(), file)
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		env.log(Error, "FileContent", err.Error())
		return ""
	}
	fileContent := string(content)
	env.log(Debug, "FileContent", fileContent)
	return fileContent
}

func (env *ShellEnvironment) LsDir(path string) []fs.DirEntry {
	defer env.trace(time.Now(), "LsDir", path)
	entries, err := os.ReadDir(path)
	if err != nil {
		env.log(Error, "LsDir", err.Error())
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
	defer env.trace(time.Now(), "PathSeparator")
	return string(os.PathSeparator)
}

func (env *ShellEnvironment) User() string {
	defer env.trace(time.Now(), "User")
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	env.log(Debug, "User", user)
	return user
}

func (env *ShellEnvironment) Host() (string, error) {
	defer env.trace(time.Now(), "Host")
	hostName, err := os.Hostname()
	if err != nil {
		env.log(Error, "Host", err.Error())
		return "", err
	}
	hostName = cleanHostName(hostName)
	env.log(Debug, "Host", hostName)
	return hostName, nil
}

func (env *ShellEnvironment) GOOS() string {
	defer env.trace(time.Now(), "GOOS")
	return runtime.GOOS
}

func (env *ShellEnvironment) RunCommand(command string, args ...string) (string, error) {
	defer env.trace(time.Now(), "RunCommand", append([]string{command}, args...)...)
	if cmd, ok := env.cmdCache.get(command); ok {
		command = cmd
	}
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	cmdErr := cmd.Run()
	if cmdErr != nil {
		output := err.String()
		errorStr := fmt.Sprintf("cmd.Start() failed with '%s'", output)
		env.log(Error, "RunCommand", errorStr)
		return output, cmdErr
	}
	// some silly commands return 0 and the output is in stderr instead of stdout
	result := out.String()
	if len(result) == 0 {
		result = err.String()
	}
	output := strings.TrimSpace(result)
	env.log(Debug, "RunCommand", output)
	return output, nil
}

func (env *ShellEnvironment) RunShellCommand(shell, command string) string {
	defer env.trace(time.Now(), "RunShellCommand", shell, command)
	if out, err := env.RunCommand(shell, "-c", command); err == nil {
		return out
	}
	return ""
}

func (env *ShellEnvironment) CommandPath(command string) string {
	defer env.trace(time.Now(), "HasCommand", command)
	if path, ok := env.cmdCache.get(command); ok {
		return path
	}
	path, err := exec.LookPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		return path
	}
	path, err = env.LookWinAppPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		return path
	}
	env.log(Error, "CommandPath", err.Error())
	return ""
}

func (env *ShellEnvironment) HasCommand(command string) bool {
	if path := env.CommandPath(command); path != "" {
		return true
	}
	return false
}

func (env *ShellEnvironment) ErrorCode() int {
	defer env.trace(time.Now(), "ErrorCode")
	return env.CmdFlags.ErrorCode
}

func (env *ShellEnvironment) ExecutionTime() float64 {
	defer env.trace(time.Now(), "ExecutionTime")
	if env.CmdFlags.ExecutionTime < 0 {
		return 0
	}
	return env.CmdFlags.ExecutionTime
}

func (env *ShellEnvironment) Flags() *Flags {
	defer env.trace(time.Now(), "Flags")
	return env.CmdFlags
}

func (env *ShellEnvironment) Shell() string {
	defer env.trace(time.Now(), "Shell")
	if env.CmdFlags.Shell != "" {
		return env.CmdFlags.Shell
	}
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		env.log(Error, "Shell", err.Error())
		return Unknown
	}
	if name == "cmd.exe" {
		p, _ = p.Parent()
		name, err = p.Name()
	}
	if err != nil {
		env.log(Error, "Shell", err.Error())
		return Unknown
	}
	// Cache the shell value to speed things up.
	env.CmdFlags.Shell = strings.Trim(strings.Replace(name, ".exe", "", 1), " ")
	return env.CmdFlags.Shell
}

func (env *ShellEnvironment) HTTPRequest(targetURL string, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error) {
	defer env.trace(time.Now(), "HTTPRequest", targetURL)
	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	defer cncl()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	for _, modifier := range requestModifiers {
		modifier(request)
	}
	response, err := client.Do(request)
	if err != nil {
		env.log(Error, "HTTPRequest", err.Error())
		return nil, err
	}
	// anything inside the range [200, 299] is considered a success
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "HTTP status code " + strconv.Itoa(response.StatusCode)
		err := errors.New(message)
		env.log(Error, "HTTPRequest", message)
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		env.log(Error, "HTTPRequest", err.Error())
		return nil, err
	}
	env.log(Debug, "HTTPRequest", string(body))
	return body, nil
}

func (env *ShellEnvironment) HasParentFilePath(path string) (*FileInfo, error) {
	defer env.trace(time.Now(), "HasParentFilePath", path)
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
		env.log(Error, "HasParentFilePath", err.Error())
		return nil, errors.New("no match at root level")
	}
}

func (env *ShellEnvironment) StackCount() int {
	defer env.trace(time.Now(), "StackCount")
	if env.CmdFlags.StackCount < 0 {
		return 0
	}
	return env.CmdFlags.StackCount
}

func (env *ShellEnvironment) Cache() Cache {
	return env.fileCache
}

func (env *ShellEnvironment) Close() {
	env.fileCache.Close()
}

func (env *ShellEnvironment) Logs() string {
	return env.logBuilder.String()
}

func (env *ShellEnvironment) TemplateCache() *TemplateCache {
	defer env.trace(time.Now(), "TemplateCache")
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
	pwd = strings.Replace(pwd, env.Home(), "~", 1)
	tmplCache.PWD = pwd
	tmplCache.Folder = Base(env, pwd)
	tmplCache.UserName = env.User()
	if host, err := env.Host(); err == nil {
		tmplCache.HostName = host
	}
	goos := env.GOOS()
	tmplCache.OS = goos
	if goos == LinuxPlatform {
		tmplCache.OS = env.Platform()
	}
	env.tmplCache = tmplCache
	return tmplCache
}

func (env *ShellEnvironment) DirMatchesOneOf(dir string, regexes []string) (match bool) {
	// sometimes the function panics inside golang, we want to silence that error
	// and assume that there's no match. Not perfect, but better than crashing
	// for the time being until we figure out what the actual root cause is
	defer func() {
		if err := recover(); err != nil {
			message := fmt.Sprintf("%s", err)
			env.log(Error, "DirMatchesOneOf", message)
			match = false
		}
	}()
	env.lock.Lock()
	defer env.lock.Unlock()
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
		if goos == WindowsPlatform || goos == DarwinPlatform {
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
