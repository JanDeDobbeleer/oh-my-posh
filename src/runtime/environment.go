package runtime

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

	"oh-my-posh/engine"
)

// Platform identifiers
const (
	Unknown = "unknown"
	Windows = "windows"
	Darwin  = "darwin"
	Linux   = "linux"
)

// Shell identifiers
const (
	Zsh        = "zsh"
	Bash       = "bash"
	PSCore     = "pwsh"
	Fish       = "fish"
	Powershell = "powershell"
	Plain      = "shell"
)

// CommandError is returned when a command fails
type CommandError struct {
	Err      string
	ExitCode int
}

func (e *CommandError) Error() string {
	return e.Err
}

// NoBatteryError is returned when no battery is found
type NoBatteryError struct{}

func (m *NoBatteryError) Error() string {
	return "no battery"
}

// File is a helper struct to get file information
type File struct {
	ParentFolder string
	Path         string
	IsDir        bool
}

// Environment contains information about the current environment
type Environment interface {
	Getenv(key string) string
	Getcwd() string
	HomeDir() string
	HasFiles(pattern string) bool
	HasFilesInDir(dir, pattern string) bool
	HasFolder(folder string) bool
	GetFileContent(file string) string
	GetPathSeperator() string
	GetCurrentUser() string
	IsRunningAsRoot() bool
	GetHostName() (string, error)
	GetRuntimeGOOS() string
	GetPlatform() string
	HasCommand(command string) bool
	RunCommand(command string, args ...string) (string, error)
	RunShellCommand(shell, command string) string
	LastErrorCode() int
	ExecutionTime() float64
	GetArgs() *engine.Args
	GetBatteryInfo() ([]*battery.Battery, error)
	GetShellName() string
	GetWindowTitle(imageName, windowTitleRegex string) (string, error)
	DoGet(url string) ([]byte, error)
	HasParentFilePath(path string) (fileInfo *File, err error)
	IsWsl() bool
	StackCount() int
	GetTerminalWidth() (int, error)
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

// Shell exposes an API for the platform environment
type Shell struct {
	args     *engine.Args
	cwd      string
	cmdCache *commandCache
}

func (sh *Shell) Init(args *engine.Args) {
	sh.args = args
	cmdCache := &commandCache{
		commands: make(map[string]string),
		lock:     sync.RWMutex{},
	}
	sh.cmdCache = cmdCache
}

func (sh *Shell) Getenv(key string) string {
	return os.Getenv(key)
}

func (sh *Shell) Getcwd() string {
	if sh.cwd != "" {
		return sh.cwd
	}
	correctPath := func(pwd string) string {
		// on Windows, and being case sensitive and not consistent and all, this gives silly issues
		return strings.Replace(pwd, "c:", "C:", 1)
	}
	if sh.args != nil && *sh.args.PWD != "" {
		sh.cwd = correctPath(*sh.args.PWD)
		return sh.cwd
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	sh.cwd = correctPath(dir)
	return sh.cwd
}

func (sh *Shell) HasFiles(pattern string) bool {
	cwd := sh.Getcwd()
	pattern = cwd + sh.GetPathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func (sh *Shell) HasFilesInDir(dir, pattern string) bool {
	pattern = dir + sh.GetPathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func (sh *Shell) HasFolder(folder string) bool {
	_, err := os.Stat(folder)
	return !os.IsNotExist(err)
}

func (sh *Shell) GetFileContent(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(content)
}

func (sh *Shell) GetPathSeperator() string {
	return string(os.PathSeparator)
}

func (sh *Shell) GetCurrentUser() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	return user
}

func (sh *Shell) GetHostName() (string, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return cleanHostName(hostName), nil
}

func (sh *Shell) GetRuntimeGOOS() string {
	return runtime.GOOS
}

func (sh *Shell) GetPlatform() string {
	if runtime.GOOS == Windows {
		return Windows
	}
	p, _, _, _ := host.PlatformInformation()

	return p
}

func (sh *Shell) RunCommand(command string, args ...string) (string, error) {
	if cmd, ok := sh.cmdCache.get(command); ok {
		command = cmd
	}
	out, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", &CommandError{
				Err:      exitErr.Error(),
				ExitCode: exitErr.ExitCode(),
			}
		}
	}
	return strings.TrimSpace(string(out)), nil
}

func (sh *Shell) RunShellCommand(shell, command string) string {
	out, _ := sh.RunCommand(shell, "-c", command)
	return out
}

func (sh *Shell) HasCommand(command string) bool {
	if _, ok := sh.cmdCache.get(command); ok {
		return true
	}
	path, err := exec.LookPath(command)
	if err == nil {
		sh.cmdCache.set(command, path)
		return true
	}
	return false
}

func (sh *Shell) LastErrorCode() int {
	return *sh.args.ErrorCode
}

func (sh *Shell) ExecutionTime() float64 {
	if *sh.args.ExecutionTime < 0 {
		return 0
	}
	return *sh.args.ExecutionTime
}

func (sh *Shell) GetArgs() *engine.Args {
	return sh.args
}

func (sh *Shell) GetBatteryInfo() ([]*battery.Battery, error) {
	return battery.GetAll()
}

func (sh *Shell) GetShellName() string {
	if *sh.args.Shell != "" {
		return *sh.args.Shell
	}
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		return Unknown
	}
	if name == "cmd.exe" {
		p, _ = p.Parent()
		name, err = p.Name()
	}
	if err != nil {
		return Unknown
	}
	// Cache the shell value to speed things up.
	*sh.args.Shell = strings.Trim(strings.Replace(name, ".exe", "", 1), " ")
	return *sh.args.Shell
}

func (sh *Shell) DoGet(url string) ([]byte, error) {
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

func (sh *Shell) HasParentFilePath(path string) (*File, error) {
	currentFolder := sh.Getcwd()
	for {
		searchPath := filepath.Join(currentFolder, path)
		info, err := os.Stat(searchPath)
		if err == nil {
			return &File{
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
		return nil, errors.New("no match at root level")
	}
}

func (sh *Shell) StackCount() int {
	if *sh.args.StackCount < 0 {
		return 0
	}
	return *sh.args.StackCount
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
