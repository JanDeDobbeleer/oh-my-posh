package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	httplib "net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/cmd"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"

	disk "github.com/shirou/gopsutil/v3/disk"
	load "github.com/shirou/gopsutil/v3/load"
	process "github.com/shirou/gopsutil/v3/process"
)

type Terminal struct {
	CmdFlags     *Flags
	Var          maps.Simple
	cmdCache     *cache.Command
	deviceCache  *cache.File
	sessionCache *cache.File
	tmplCache    *cache.Template
	lsDirMap     maps.Concurrent
	cwd          string
	host         string
	networks     []*Connection
	sync.RWMutex
}

func (term *Terminal) Init() {
	defer term.Trace(time.Now())
	if term.CmdFlags == nil {
		term.CmdFlags = &Flags{}
	}

	if term.CmdFlags.Debug {
		log.Enable()
		log.Debug("debug mode enabled")
	}

	if term.CmdFlags.Plain {
		log.Plain()
		log.Debug("plain mode enabled")
	}

	initCache := func(fileName string) *cache.File {
		cache := &cache.File{}
		cache.Init(filepath.Join(term.CachePath(), fileName))
		return cache
	}

	term.deviceCache = initCache(cache.FileName)
	term.sessionCache = initCache(cache.SessionFileName)

	term.resolveConfigPath()

	term.cmdCache = &cache.Command{
		Commands: maps.NewConcurrent(),
	}

	term.tmplCache = &cache.Template{}

	if !term.CmdFlags.Cached {
		term.SetPromptCount()
	}
}

func (term *Terminal) resolveConfigPath() {
	defer term.Trace(time.Now())

	// if the config flag is set, we'll use that over POSH_THEME
	// in our internal shell logic, we'll always use the POSH_THEME
	// due to not using --config to set the configuration
	hasConfigFlag := len(term.CmdFlags.Config) > 0

	if poshTheme := term.Getenv("POSH_THEME"); len(poshTheme) > 0 && !hasConfigFlag {
		term.DebugF("config set using POSH_THEME: %s", poshTheme)
		term.CmdFlags.Config = poshTheme
		return
	}

	if len(term.CmdFlags.Config) == 0 {
		term.Debug("no config set, fallback to default config")
		return
	}

	if strings.HasPrefix(term.CmdFlags.Config, "https://") {
		filePath, err := config.Download(term.CachePath(), term.CmdFlags.Config)
		if err != nil {
			term.Error(err)
			term.CmdFlags.Config = ""
			return
		}

		term.CmdFlags.Config = filePath
		return
	}

	isCygwin := func() bool {
		return term.Platform() == WINDOWS && len(term.Getenv("OSTYPE")) > 0
	}

	// Cygwin path always needs the full path as we're on Windows but not really.
	// Doing filepath actions will convert it to a Windows path and break the init script.
	if isCygwin() {
		term.Debug("cygwin detected, using full path for config")
		return
	}

	configFile := term.CmdFlags.Config
	if strings.HasPrefix(configFile, "~") {
		configFile = strings.TrimPrefix(configFile, "~")
		configFile = filepath.Join(term.Home(), configFile)
	}

	abs, err := filepath.Abs(configFile)
	if err != nil {
		term.Error(err)
		term.CmdFlags.Config = filepath.Clean(configFile)
		return
	}

	term.CmdFlags.Config = abs
}

func (term *Terminal) Trace(start time.Time, args ...string) {
	log.Trace(start, args...)
}

func (term *Terminal) Debug(message string) {
	log.Debug(message)
}

func (term *Terminal) DebugF(format string, a ...any) {
	if !term.CmdFlags.Debug {
		return
	}
	message := fmt.Sprintf(format, a...)
	log.Debug(message)
}

func (term *Terminal) Error(err error) {
	log.Error(err)
}

func (term *Terminal) Getenv(key string) string {
	defer term.Trace(time.Now(), key)
	val := os.Getenv(key)
	term.Debug(val)
	return val
}

func (term *Terminal) Pwd() string {
	term.Lock()
	defer term.Trace(time.Now())
	defer term.Unlock()
	if term.cwd != "" {
		return term.cwd
	}
	correctPath := func(pwd string) string {
		if term.GOOS() != WINDOWS {
			return pwd
		}
		// on Windows, and being case sensitive and not consistent and all, this gives silly issues
		driveLetter := regex.GetCompiledRegex(`^[a-z]:`)
		return driveLetter.ReplaceAllStringFunc(pwd, strings.ToUpper)
	}
	if term.CmdFlags != nil && term.CmdFlags.PWD != "" {
		term.cwd = correctPath(term.CmdFlags.PWD)
		term.Debug(term.cwd)
		return term.cwd
	}
	dir, err := os.Getwd()
	if err != nil {
		term.Error(err)
		return ""
	}
	term.cwd = correctPath(dir)
	term.Debug(term.cwd)
	return term.cwd
}

func (term *Terminal) HasFiles(pattern string) bool {
	return term.HasFilesInDir(term.Pwd(), pattern)
}

func (term *Terminal) HasFilesInDir(dir, pattern string) bool {
	defer term.Trace(time.Now(), pattern)

	fileSystem := os.DirFS(dir)
	var dirEntries []fs.DirEntry

	if files, OK := term.lsDirMap.Get(dir); OK {
		dirEntries, _ = files.([]fs.DirEntry)
	}

	if len(dirEntries) == 0 {
		var err error
		dirEntries, err = fs.ReadDir(fileSystem, ".")
		if err != nil {
			term.Error(err)
			term.Debug("false")
			return false
		}

		term.lsDirMap.Set(dir, dirEntries)
	}

	pattern = strings.ToLower(pattern)

	term.RWMutex.RLock()
	defer term.RWMutex.RUnlock()

	for _, match := range dirEntries {
		if match.IsDir() {
			continue
		}

		matchFileName, err := filepath.Match(pattern, strings.ToLower(match.Name()))
		if err != nil {
			term.Error(err)
			term.Debug("false")
			return false
		}

		if matchFileName {
			term.Debug("true")
			return true
		}
	}

	term.Debug("false")
	return false
}

func (term *Terminal) HasFileInParentDirs(pattern string, depth uint) bool {
	defer term.Trace(time.Now(), pattern, fmt.Sprint(depth))
	currentFolder := term.Pwd()

	for c := 0; c < int(depth); c++ {
		if term.HasFilesInDir(currentFolder, pattern) {
			term.Debug("true")
			return true
		}

		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
		} else {
			term.Debug("false")
			return false
		}
	}
	term.Debug("false")
	return false
}

func (term *Terminal) HasFolder(folder string) bool {
	defer term.Trace(time.Now(), folder)
	f, err := os.Stat(folder)
	if err != nil {
		term.Debug("false")
		return false
	}
	isDir := f.IsDir()
	term.DebugF("%t", isDir)
	return isDir
}

func (term *Terminal) ResolveSymlink(path string) (string, error) {
	defer term.Trace(time.Now(), path)
	link, err := filepath.EvalSymlinks(path)
	if err != nil {
		term.Error(err)
		return "", err
	}
	term.Debug(link)
	return link, nil
}

func (term *Terminal) FileContent(file string) string {
	defer term.Trace(time.Now(), file)
	if !filepath.IsAbs(file) {
		file = filepath.Join(term.Pwd(), file)
	}
	content, err := os.ReadFile(file)
	if err != nil {
		term.Error(err)
		return ""
	}
	fileContent := string(content)
	term.Debug(fileContent)
	return fileContent
}

func (term *Terminal) LsDir(path string) []fs.DirEntry {
	defer term.Trace(time.Now(), path)
	entries, err := os.ReadDir(path)
	if err != nil {
		term.Error(err)
		return nil
	}
	term.DebugF("%v", entries)
	return entries
}

func (term *Terminal) PathSeparator() string {
	defer term.Trace(time.Now())
	return string(os.PathSeparator)
}

func (term *Terminal) User() string {
	defer term.Trace(time.Now())
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	term.Debug(user)
	return user
}

func (term *Terminal) Host() (string, error) {
	defer term.Trace(time.Now())
	if len(term.host) != 0 {
		return term.host, nil
	}

	hostName, err := os.Hostname()
	if err != nil {
		term.Error(err)
		return "", err
	}

	hostName = cleanHostName(hostName)
	term.Debug(hostName)
	term.host = hostName

	return hostName, nil
}

func (term *Terminal) GOOS() string {
	defer term.Trace(time.Now())
	return runtime.GOOS
}

func (term *Terminal) RunCommand(command string, args ...string) (string, error) {
	defer term.Trace(time.Now(), append([]string{command}, args...)...)
	if cacheCommand, ok := term.cmdCache.Get(command); ok {
		command = cacheCommand
	}
	output, err := cmd.Run(command, args...)
	if err != nil {
		term.Error(err)
	}
	term.Debug(output)
	return output, err
}

func (term *Terminal) RunShellCommand(shell, command string) string {
	defer term.Trace(time.Now())
	if out, err := term.RunCommand(shell, "-c", command); err == nil {
		return out
	}
	return ""
}

func (term *Terminal) CommandPath(command string) string {
	defer term.Trace(time.Now(), command)
	if path, ok := term.cmdCache.Get(command); ok {
		term.Debug(path)
		return path
	}

	path, err := term.LookPath(command)
	if err == nil {
		term.cmdCache.Set(command, path)
		term.Debug(path)
		return path
	}

	term.Error(err)
	return ""
}

func (term *Terminal) HasCommand(command string) bool {
	defer term.Trace(time.Now(), command)
	if path := term.CommandPath(command); path != "" {
		return true
	}
	return false
}

func (term *Terminal) StatusCodes() (int, string) {
	defer term.Trace(time.Now())

	if term.CmdFlags.Shell != CMD || !term.CmdFlags.NoExitCode {
		return term.CmdFlags.ErrorCode, term.CmdFlags.PipeStatus
	}

	errorCode := term.Getenv("=ExitCode")
	term.Debug(errorCode)
	term.CmdFlags.ErrorCode, _ = strconv.Atoi(errorCode)

	return term.CmdFlags.ErrorCode, term.CmdFlags.PipeStatus
}

func (term *Terminal) ExecutionTime() float64 {
	defer term.Trace(time.Now())
	if term.CmdFlags.ExecutionTime < 0 {
		return 0
	}
	return term.CmdFlags.ExecutionTime
}

func (term *Terminal) Flags() *Flags {
	defer term.Trace(time.Now())
	return term.CmdFlags
}

func (term *Terminal) Shell() string {
	defer term.Trace(time.Now())
	if len(term.CmdFlags.Shell) != 0 {
		return term.CmdFlags.Shell
	}
	term.Debug("no shell name provided in flags, trying to detect it")
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		term.Error(err)
		return UNKNOWN
	}
	term.Debug("process name: " + name)
	// this is used for when scoop creates a shim, see
	// https://github.com/jandedobbeleer/oh-my-posh/issues/2806
	executable, _ := os.Executable()
	if name == executable {
		p, _ = p.Parent()
		name, err = p.Name()
		term.Debug("parent process name: " + name)
	}
	if err != nil {
		term.Error(err)
		return UNKNOWN
	}
	// Cache the shell value to speed things up.
	term.CmdFlags.Shell = strings.Trim(strings.TrimSuffix(name, ".exe"), " ")
	return term.CmdFlags.Shell
}

func (term *Terminal) unWrapError(err error) error {
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

func (term *Terminal) HTTPRequest(targetURL string, body io.Reader, timeout int, requestModifiers ...http.RequestModifier) ([]byte, error) {
	defer term.Trace(time.Now(), targetURL)

	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	defer cncl()

	request, err := httplib.NewRequestWithContext(ctx, httplib.MethodGet, targetURL, body)
	if err != nil {
		return nil, err
	}

	for _, modifier := range requestModifiers {
		modifier(request)
	}

	if term.CmdFlags.Debug {
		dump, _ := httputil.DumpRequestOut(request, true)
		term.Debug(string(dump))
	}

	response, err := http.HTTPClient.Do(request)
	if err != nil {
		term.Error(err)
		return nil, term.unWrapError(err)
	}

	// anything inside the range [200, 299] is considered a success
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "HTTP status code " + strconv.Itoa(response.StatusCode)
		err := errors.New(message)
		term.Error(err)
		return nil, err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		term.Error(err)
		return nil, err
	}

	term.Debug(string(responseBody))

	return responseBody, nil
}

func (term *Terminal) HasParentFilePath(parent string, followSymlinks bool) (*FileInfo, error) {
	defer term.Trace(time.Now(), parent)

	path := term.Pwd()
	if followSymlinks {
		if actual, err := term.ResolveSymlink(path); err == nil {
			path = actual
		}
	}

	for {
		fileSystem := os.DirFS(path)
		info, err := fs.Stat(fileSystem, parent)
		if err == nil {
			return &FileInfo{
				ParentFolder: path,
				Path:         filepath.Join(path, parent),
				IsDir:        info.IsDir(),
			}, nil
		}

		if !os.IsNotExist(err) {
			return nil, err
		}

		if dir := filepath.Dir(path); dir != path {
			path = dir
			continue
		}

		term.Error(err)
		return nil, errors.New("no match at root level")
	}
}

func (term *Terminal) StackCount() int {
	defer term.Trace(time.Now())
	if term.CmdFlags.StackCount < 0 {
		return 0
	}
	return term.CmdFlags.StackCount
}

func (term *Terminal) Cache() cache.Cache {
	return term.deviceCache
}

func (term *Terminal) Session() cache.Cache {
	return term.sessionCache
}

func (term *Terminal) Close() {
	defer term.Trace(time.Now())
	term.saveTemplateCache()
	term.deviceCache.Close()
	term.sessionCache.Close()
}

func (term *Terminal) saveTemplateCache() {
	// only store this when in a primary prompt
	// and when we have a transient prompt in the config
	canSave := term.CmdFlags.Primary && term.CmdFlags.HasTransient
	if !canSave {
		return
	}

	tmplCache := term.TemplateCache()
	tmplCache.SegmentsCache = tmplCache.Segments.ToSimple()

	templateCache, err := json.Marshal(tmplCache)
	if err == nil {
		term.sessionCache.Set(cache.TEMPLATECACHE, string(templateCache), "1day")
	}
}

func (term *Terminal) LoadTemplateCache() {
	defer term.Trace(time.Now())

	val, OK := term.sessionCache.Get(cache.TEMPLATECACHE)
	if !OK {
		return
	}

	var tmplCache cache.Template

	err := json.Unmarshal([]byte(val), &tmplCache)
	if err != nil {
		term.Error(err)
		return
	}

	tmplCache.Segments = tmplCache.SegmentsCache.ToConcurrent()
	tmplCache.Initialized = true

	term.tmplCache = &tmplCache
}

func (term *Terminal) Logs() string {
	return log.String()
}

func (term *Terminal) TemplateCache() *cache.Template {
	defer term.Trace(time.Now())
	tmplCache := term.tmplCache

	if tmplCache.Initialized {
		return tmplCache
	}

	tmplCache.Root = term.Root()
	tmplCache.Shell = term.Shell()
	tmplCache.ShellVersion = term.CmdFlags.ShellVersion
	tmplCache.Code, _ = term.StatusCodes()
	tmplCache.WSL = term.IsWsl()
	tmplCache.Segments = maps.NewConcurrent()
	tmplCache.PromptCount = term.CmdFlags.PromptCount
	tmplCache.Var = make(map[string]any)
	tmplCache.Jobs = term.CmdFlags.JobCount

	if term.Var != nil {
		tmplCache.Var = term.Var
	}

	pwd := term.Pwd()
	tmplCache.PWD = ReplaceHomeDirPrefixWithTilde(term, pwd)

	tmplCache.AbsolutePWD = pwd
	if term.IsWsl() {
		tmplCache.AbsolutePWD, _ = term.RunCommand("wslpath", "-m", pwd)
	}

	tmplCache.PSWD = term.CmdFlags.PSWD

	tmplCache.Folder = Base(term, pwd)
	if term.GOOS() == WINDOWS && strings.HasSuffix(tmplCache.Folder, ":") {
		tmplCache.Folder += `\`
	}

	tmplCache.UserName = term.User()
	if host, err := term.Host(); err == nil {
		tmplCache.HostName = host
	}

	goos := term.GOOS()
	tmplCache.OS = goos
	if goos == LINUX {
		tmplCache.OS = term.Platform()
	}

	val := term.Getenv("SHLVL")
	if shlvl, err := strconv.Atoi(val); err == nil {
		tmplCache.SHLVL = shlvl
	}

	tmplCache.Initialized = true
	return tmplCache
}

func (term *Terminal) DirMatchesOneOf(dir string, regexes []string) (match bool) {
	// sometimes the function panics inside golang, we want to silence that error
	// and assume that there's no match. Not perfect, but better than crashing
	// for the time being until we figure out what the actual root cause is
	defer func() {
		if err := recover(); err != nil {
			term.Error(errors.New("panic"))
			match = false
		}
	}()
	match = dirMatchesOneOf(dir, term.Home(), term.GOOS(), regexes)
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

func (term *Terminal) SetPromptCount() {
	defer term.Trace(time.Now())

	countStr := os.Getenv("POSH_PROMPT_COUNT")
	if len(countStr) > 0 {
		// this counter is incremented by the shell
		count, err := strconv.Atoi(countStr)
		if err == nil {
			term.CmdFlags.PromptCount = count
			return
		}
	}

	var count int
	if val, found := term.Session().Get(cache.PROMPTCOUNTCACHE); found {
		count, _ = strconv.Atoi(val)
	}

	// only write to cache if we're the primary prompt
	if term.CmdFlags.Primary {
		count++
		term.Session().Set(cache.PROMPTCOUNTCACHE, strconv.Itoa(count), "1day")
	}

	term.CmdFlags.PromptCount = count
}

func (term *Terminal) CursorPosition() (row, col int) {
	if number, err := strconv.Atoi(term.Getenv("POSH_CURSOR_LINE")); err == nil {
		row = number
	}

	if number, err := strconv.Atoi(term.Getenv("POSH_CURSOR_COLUMN")); err != nil {
		col = number
	}

	return
}

func (term *Terminal) SystemInfo() (*SystemInfo, error) {
	s := &SystemInfo{}

	mem, err := term.Memory()
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
	if err := os.Mkdir(cachePath, 0o755); err != nil {
		return ""
	}
	return cachePath
}
