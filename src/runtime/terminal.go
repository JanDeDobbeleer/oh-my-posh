package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	httplib "net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/cmd"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"

	disk "github.com/shirou/gopsutil/v3/disk"
	load "github.com/shirou/gopsutil/v3/load"
	process "github.com/shirou/gopsutil/v3/process"
)

type Terminal struct {
	CmdFlags     *Flags
	cmdCache     *cache.Command
	deviceCache  *cache.File
	sessionCache *cache.File
	lsDirMap     maps.Concurrent
	cwd          string
	host         string
	networks     []*Connection
}

func (term *Terminal) Init(flags *Flags) {
	defer log.Trace(time.Now())

	term.CmdFlags = flags

	if term.CmdFlags == nil {
		term.CmdFlags = &Flags{}
	}

	if term.CmdFlags.Plain {
		log.Plain()
		log.Debug("plain mode enabled")
	}

	initCache := func(fileName string) *cache.File {
		fileCache := &cache.File{}
		fileCache.Init(filepath.Join(cache.Path(), fileName), term.CmdFlags.SaveCache)
		return fileCache
	}

	term.deviceCache = initCache(cache.FileName)
	term.sessionCache = initCache(cache.SessionFileName)
	term.setPromptCount()

	term.setPwd()

	term.cmdCache = &cache.Command{
		Commands: maps.NewConcurrent(),
	}
}

func (term *Terminal) Getenv(key string) string {
	defer log.Trace(time.Now(), key)
	val := os.Getenv(key)
	log.Debug(val)
	return val
}

func (term *Terminal) Pwd() string {
	return term.cwd
}

func (term *Terminal) setPwd() {
	defer log.Trace(time.Now())

	correctPath := func(pwd string) string {
		if term.GOOS() != WINDOWS {
			return pwd
		}

		// on Windows, and being case sensitive and not consistent and all, this gives silly issues
		driveLetter, err := regex.GetCompiledRegex(`^[a-z]:`)
		if err == nil {
			return driveLetter.ReplaceAllStringFunc(pwd, strings.ToUpper)
		}

		return pwd
	}

	if term.CmdFlags != nil && term.CmdFlags.PWD != "" {
		term.cwd = path.Clean(term.CmdFlags.PWD)
		log.Debug(term.cwd)
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return
	}

	term.cwd = correctPath(dir)
	log.Debug(term.cwd)
}

func (term *Terminal) HasFiles(pattern string) bool {
	return term.HasFilesInDir(term.Pwd(), pattern)
}

func (term *Terminal) HasFilesInDir(dir, pattern string) bool {
	defer log.Trace(time.Now(), pattern)

	fileSystem := os.DirFS(dir)
	var dirEntries []fs.DirEntry

	if files, OK := term.lsDirMap.Get(dir); OK {
		dirEntries, _ = files.([]fs.DirEntry)
	}

	if len(dirEntries) == 0 {
		var err error
		dirEntries, err = fs.ReadDir(fileSystem, ".")
		if err != nil {
			log.Error(err)
			log.Debug("false")
			return false
		}

		term.lsDirMap.Set(dir, dirEntries)
	}

	pattern = strings.ToLower(pattern)

	for _, match := range dirEntries {
		if match.IsDir() {
			continue
		}

		matchFileName, err := filepath.Match(pattern, strings.ToLower(match.Name()))
		if err != nil {
			log.Error(err)
			log.Debug("false")
			return false
		}

		if matchFileName {
			log.Debug("true")
			return true
		}
	}

	log.Debug("false")
	return false
}

func (term *Terminal) HasFileInParentDirs(pattern string, depth uint) bool {
	defer log.Trace(time.Now(), pattern, fmt.Sprint(depth))
	currentFolder := term.Pwd()

	for c := 0; c < int(depth); c++ {
		if term.HasFilesInDir(currentFolder, pattern) {
			log.Debug("true")
			return true
		}

		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
		} else {
			log.Debug("false")
			return false
		}
	}
	log.Debug("false")
	return false
}

func (term *Terminal) HasFolder(folder string) bool {
	defer log.Trace(time.Now(), folder)
	f, err := os.Stat(folder)
	if err != nil {
		log.Debug("false")
		return false
	}
	isDir := f.IsDir()
	log.Debugf("%t", isDir)
	return isDir
}

func (term *Terminal) ResolveSymlink(input string) (string, error) {
	defer log.Trace(time.Now(), input)
	link, err := filepath.EvalSymlinks(input)
	if err != nil {
		log.Error(err)
		return "", err
	}
	log.Debug(link)
	return link, nil
}

func (term *Terminal) FileContent(file string) string {
	defer log.Trace(time.Now(), file)
	if !filepath.IsAbs(file) {
		file = filepath.Join(term.Pwd(), file)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error(err)
		return ""
	}

	fileContent := string(content)
	log.Debug(fileContent)

	return fileContent
}

func (term *Terminal) LsDir(input string) []fs.DirEntry {
	defer log.Trace(time.Now(), input)

	entries, err := os.ReadDir(input)
	if err != nil {
		log.Error(err)
		return nil
	}

	log.Debugf("%v", entries)
	return entries
}

func (term *Terminal) User() string {
	defer log.Trace(time.Now())
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	log.Debug(user)
	return user
}

func (term *Terminal) Host() (string, error) {
	defer log.Trace(time.Now())
	if len(term.host) != 0 {
		return term.host, nil
	}

	hostName, err := os.Hostname()
	if err != nil {
		log.Error(err)
		return "", err
	}

	hostName = cleanHostName(hostName)
	log.Debug(hostName)
	term.host = hostName

	return hostName, nil
}

func (term *Terminal) GOOS() string {
	defer log.Trace(time.Now())
	return runtime.GOOS
}

func (term *Terminal) Home() string {
	return path.Home()
}

func (term *Terminal) RunCommand(command string, args ...string) (string, error) {
	defer log.Trace(time.Now(), append([]string{command}, args...)...)

	if cacheCommand, ok := term.cmdCache.Get(command); ok {
		command = cacheCommand
	}

	output, err := cmd.Run(command, args...)
	if err != nil {
		log.Error(err)
	}

	log.Debug(output)
	return output, err
}

func (term *Terminal) RunShellCommand(shell, command string) string {
	defer log.Trace(time.Now())

	if out, err := term.RunCommand(shell, "-c", command); err == nil {
		return out
	}

	return ""
}

func (term *Terminal) CommandPath(command string) string {
	defer log.Trace(time.Now(), command)
	if cmdPath, ok := term.cmdCache.Get(command); ok {
		log.Debug(cmdPath)
		return cmdPath
	}

	cmdPath, err := exec.LookPath(command)
	if err == nil {
		term.cmdCache.Set(command, cmdPath)
		log.Debug(cmdPath)
		return cmdPath
	}

	log.Error(err)
	return ""
}

func (term *Terminal) HasCommand(command string) bool {
	defer log.Trace(time.Now(), command)

	if cmdPath := term.CommandPath(command); cmdPath != "" {
		return true
	}

	return false
}

func (term *Terminal) StatusCodes() (int, string) {
	defer log.Trace(time.Now())

	if term.CmdFlags.Shell != CMD || !term.CmdFlags.NoExitCode {
		return term.CmdFlags.ErrorCode, term.CmdFlags.PipeStatus
	}

	errorCode := term.Getenv("=ExitCode")
	log.Debug(errorCode)
	term.CmdFlags.ErrorCode, _ = strconv.Atoi(errorCode)

	return term.CmdFlags.ErrorCode, term.CmdFlags.PipeStatus
}

func (term *Terminal) ExecutionTime() float64 {
	defer log.Trace(time.Now())
	if term.CmdFlags.ExecutionTime < 0 {
		return 0
	}
	return term.CmdFlags.ExecutionTime
}

func (term *Terminal) Flags() *Flags {
	defer log.Trace(time.Now())
	return term.CmdFlags
}

func (term *Terminal) Shell() string {
	defer log.Trace(time.Now())
	if len(term.CmdFlags.Shell) != 0 {
		return term.CmdFlags.Shell
	}

	log.Debug("no shell name provided in flags, trying to detect it")

	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))

	name, err := p.Name()
	if err != nil {
		log.Error(err)
		return UNKNOWN
	}

	log.Debug("process name: " + name)

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
	defer log.Trace(time.Now(), targetURL)

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
		log.Debug(string(dump))
	}

	response, err := http.HTTPClient.Do(request)
	if err != nil {
		log.Error(err)
		return nil, term.unWrapError(err)
	}

	// anything inside the range [200, 299] is considered a success
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "HTTP status code " + strconv.Itoa(response.StatusCode)
		err := errors.New(message)
		log.Error(err)
		return nil, err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Debug(string(responseBody))

	return responseBody, nil
}

func (term *Terminal) HasParentFilePath(parent string, followSymlinks bool) (*FileInfo, error) {
	defer log.Trace(time.Now(), parent)

	pwd := term.Pwd()
	if followSymlinks {
		if actual, err := term.ResolveSymlink(pwd); err == nil {
			pwd = actual
		}
	}

	for {
		fileSystem := os.DirFS(pwd)
		info, err := fs.Stat(fileSystem, parent)
		if err == nil {
			return &FileInfo{
				ParentFolder: pwd,
				Path:         filepath.Join(pwd, parent),
				IsDir:        info.IsDir(),
			}, nil
		}

		if !os.IsNotExist(err) {
			return nil, err
		}

		if dir := filepath.Dir(pwd); dir != pwd {
			pwd = dir
			continue
		}

		log.Error(err)
		return nil, errors.New("no match at root level")
	}
}

func (term *Terminal) StackCount() int {
	defer log.Trace(time.Now())

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
	defer log.Trace(time.Now())
	term.clearCacheFiles()
	term.deviceCache.Close()
	term.sessionCache.Close()
}

func (term *Terminal) clearCacheFiles() {
	if !term.CmdFlags.Init {
		return
	}

	deletedFiles, err := cache.Clear(cache.Path(), false)
	if err != nil {
		log.Error(err)
		return
	}

	for _, file := range deletedFiles {
		log.Debugf("removed cache file: %s", file)
	}
}

func (term *Terminal) Logs() string {
	return log.String()
}

func (term *Terminal) DirMatchesOneOf(dir string, regexes []string) (match bool) {
	// sometimes the function panics inside golang, we want to silence that error
	// and assume that there's no match. Not perfect, but better than crashing
	// for the time being until we figure out what the actual root cause is
	defer func() {
		if err := recover(); err != nil {
			log.Error(errors.New("panic"))
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
		normalized := strings.ReplaceAll(element, "\\\\", "/")
		if strings.HasPrefix(normalized, "~") {
			rem := normalized[1:]
			if len(rem) == 0 || rem[0] == '/' {
				normalized = home + rem
			}
		}
		pattern := fmt.Sprintf("^%s$", normalized)
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

func (term *Terminal) setPromptCount() {
	defer log.Trace(time.Now())

	var count int
	if val, found := term.Session().Get(cache.PROMPTCOUNTCACHE); found {
		count, _ = strconv.Atoi(val)
	}

	// Only update the count if we're generating a primary prompt.
	if term.CmdFlags.Type == PRIMARY {
		count++
		term.Session().Set(cache.PROMPTCOUNTCACHE, strconv.Itoa(count), cache.ONEDAY)
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
