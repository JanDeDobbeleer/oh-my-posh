package main

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/process"
)

const (
	unknown         = "unknown"
	windowsPlatform = "windows"
)

type commandError struct {
	err      string
	exitCode int
}

func (e *commandError) Error() string {
	return e.err
}

type noBatteryError struct{}

func (m *noBatteryError) Error() string {
	return "no battery"
}

type fileInfo struct {
	parentFolder string
	path         string
	isDir        bool
}

type environmentInfo interface {
	getenv(key string) string
	getcwd() string
	homeDir() string
	hasFiles(pattern string) bool
	hasFilesInDir(dir, pattern string) bool
	hasFolder(folder string) bool
	getFileContent(file string) string
	getPathSeperator() string
	getCurrentUser() string
	isRunningAsRoot() bool
	getHostName() (string, error)
	getRuntimeGOOS() string
	getPlatform() string
	hasCommand(command string) bool
	runCommand(command string, args ...string) (string, error)
	runShellCommand(shell, command string) string
	lastErrorCode() int
	executionTime() float64
	getArgs() *args
	getBatteryInfo() ([]*battery.Battery, error)
	getShellName() string
	getWindowTitle(imageName, windowTitleRegex string) (string, error)
	doGet(url string) ([]byte, error)
	hasParentFilePath(path string) (fileInfo *fileInfo, err error)
	isWsl() bool
	stackCount() int
}

type commandCache struct {
	commands map[string]string
	lock     sync.RWMutex
}

func (c *commandCache) set(command, path string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.commands[command] = path
}

func (c *commandCache) get(command string) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if cmd, ok := c.commands[command]; ok {
		command = cmd
		return command, true
	}
	return "", false
}

type environment struct {
	args     *args
	cwd      string
	cmdCache *commandCache
}

func (env *environment) init(args *args) {
	env.args = args
	cmdCache := &commandCache{
		commands: make(map[string]string),
		lock:     sync.RWMutex{},
	}
	env.cmdCache = cmdCache
}

func (env *environment) getenv(key string) string {
	return os.Getenv(key)
}

func (env *environment) getcwd() string {
	if env.cwd != "" {
		return env.cwd
	}
	correctPath := func(pwd string) string {
		// on Windows, and being case sensitive and not consistent and all, this gives silly issues
		return strings.Replace(pwd, "c:", "C:", 1)
	}
	if env.args != nil && *env.args.PWD != "" {
		env.cwd = correctPath(*env.args.PWD)
		return env.cwd
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	env.cwd = correctPath(dir)
	return env.cwd
}

func (env *environment) hasFiles(pattern string) bool {
	cwd := env.getcwd()
	pattern = cwd + env.getPathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func (env *environment) hasFilesInDir(dir, pattern string) bool {
	pattern = dir + env.getPathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func (env *environment) hasFolder(folder string) bool {
	_, err := os.Stat(folder)
	return !os.IsNotExist(err)
}

func (env *environment) getFileContent(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(content)
}

func (env *environment) getPathSeperator() string {
	return string(os.PathSeparator)
}

func (env *environment) getCurrentUser() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	return user
}

func (env *environment) getHostName() (string, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return cleanHostName(hostName), nil
}

func (env *environment) getRuntimeGOOS() string {
	return runtime.GOOS
}

func (env *environment) getPlatform() string {
	if runtime.GOOS == windowsPlatform {
		return windowsPlatform
	}
	p, _, _, _ := host.PlatformInformation()

	return p
}

func (env *environment) runCommand(command string, args ...string) (string, error) {
	if cmd, ok := env.cmdCache.get(command); ok {
		command = cmd
	}
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", &commandError{
				err:      exitErr.Error(),
				exitCode: exitErr.ExitCode(),
			}
		}
	}
	return strings.TrimSpace(string(out)), nil
}

func (env *environment) runShellCommand(shell, command string) string {
	out, _ := env.runCommand(shell, "-c", command)
	return out
}

func (env *environment) hasCommand(command string) bool {
	if _, ok := env.cmdCache.get(command); ok {
		return true
	}
	path, err := exec.LookPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		return true
	}
	return false
}

func (env *environment) lastErrorCode() int {
	return *env.args.ErrorCode
}

func (env *environment) executionTime() float64 {
	if *env.args.ExecutionTime < 0 {
		return 0
	}
	return *env.args.ExecutionTime
}

func (env *environment) getArgs() *args {
	return env.args
}

func (env *environment) getBatteryInfo() ([]*battery.Battery, error) {
	return battery.GetAll()
}

func (env *environment) getShellName() string {
	if *env.args.Shell != "" {
		return *env.args.Shell
	}
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		return unknown
	}
	if name == "cmd.exe" {
		p, _ = p.Parent()
		name, err = p.Name()
	}
	if err != nil {
		return unknown
	}
	// Cache the shell value to speed things up.
	*env.args.Shell = strings.Trim(strings.Replace(name, ".exe", "", 1), " ")
	return *env.args.Shell
}

func (env *environment) doGet(url string) ([]byte, error) {
	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*20)
	defer cncl()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (env *environment) hasParentFilePath(path string) (*fileInfo, error) {
	currentFolder := env.getcwd()
	for {
		searchPath := filepath.Join(currentFolder, path)
		info, err := os.Stat(searchPath)
		if err == nil {
			return &fileInfo{
				parentFolder: currentFolder,
				path:         searchPath,
				isDir:        info.IsDir(),
			}, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
			continue
		}
		return nil, errors.New("no match at root level")
	}
}

func (env *environment) stackCount() int {
	if *env.args.StackCount < 0 {
		return 0
	}
	return *env.args.StackCount
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
