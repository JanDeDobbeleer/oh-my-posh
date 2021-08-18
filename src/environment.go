package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	darwinPlatform  = "darwin"
	linuxPlatform   = "linux"
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
	doGet(url string, timeout int) ([]byte, error)
	hasParentFilePath(path string) (fileInfo *fileInfo, err error)
	isWsl() bool
	stackCount() int
	getTerminalWidth() (int, error)
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

type tracer struct {
	file  *os.File
	debug bool
}

func (t *tracer) init(home string) {
	if !t.debug {
		return
	}
	var err error
	fileName := home + "/oh-my-posh.log"
	t.file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(t.file)
	log.Println("#### start oh-my-posh run ####")
}

func (t *tracer) close() {
	if !t.debug {
		return
	}
	log.Println("#### end oh-my-posh run ####")
	_ = t.file.Close()
}

func (t *tracer) trace(start time.Time, function string, args ...string) {
	if !t.debug {
		return
	}
	elapsed := time.Since(start)
	trace := fmt.Sprintf("%s duration: %s, args: %s", function, elapsed, strings.Trim(fmt.Sprint(args), "[]"))
	log.Println(trace)
}

type environment struct {
	args     *args
	cwd      string
	cmdCache *commandCache
	tracer   *tracer
}

func (env *environment) init(args *args) {
	env.args = args
	cmdCache := &commandCache{
		commands: make(map[string]string),
		lock:     sync.RWMutex{},
	}
	env.cmdCache = cmdCache
	tracer := &tracer{
		debug: *args.Debug,
	}
	tracer.init(env.homeDir())
	env.tracer = tracer
}

func (env *environment) getenv(key string) string {
	defer env.tracer.trace(time.Now(), "getenv", key)
	return os.Getenv(key)
}

func (env *environment) getcwd() string {
	defer env.tracer.trace(time.Now(), "getcwd")
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
	defer env.tracer.trace(time.Now(), "hasFiles", pattern)
	cwd := env.getcwd()
	pattern = cwd + env.getPathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func (env *environment) hasFilesInDir(dir, pattern string) bool {
	defer env.tracer.trace(time.Now(), "hasFilesInDir", pattern)
	pattern = dir + env.getPathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func (env *environment) hasFolder(folder string) bool {
	defer env.tracer.trace(time.Now(), "hasFolder", folder)
	_, err := os.Stat(folder)
	return !os.IsNotExist(err)
}

func (env *environment) getFileContent(file string) string {
	defer env.tracer.trace(time.Now(), "getFileContent", file)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(content)
}

func (env *environment) getPathSeperator() string {
	defer env.tracer.trace(time.Now(), "getPathSeperator")
	return string(os.PathSeparator)
}

func (env *environment) getCurrentUser() string {
	defer env.tracer.trace(time.Now(), "getCurrentUser")
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	return user
}

func (env *environment) getHostName() (string, error) {
	defer env.tracer.trace(time.Now(), "getHostName")
	hostName, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return cleanHostName(hostName), nil
}

func (env *environment) getRuntimeGOOS() string {
	defer env.tracer.trace(time.Now(), "getRuntimeGOOS")
	return runtime.GOOS
}

func (env *environment) getPlatform() string {
	defer env.tracer.trace(time.Now(), "getPlatform")
	if runtime.GOOS == windowsPlatform {
		return windowsPlatform
	}
	p, _, _, _ := host.PlatformInformation()

	return p
}

func (env *environment) runCommand(command string, args ...string) (string, error) {
	defer env.tracer.trace(time.Now(), "runCommand", append([]string{command}, args...)...)
	if cmd, ok := env.cmdCache.get(command); ok {
		command = cmd
	}
	copyAndCapture := func(r io.Reader) ([]byte, error) {
		var out []byte
		buf := make([]byte, 1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				d := buf[:n]
				out = append(out, d...)
			}
			if err == nil {
				continue
			}
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
	normalizeOutput := func(out []byte) string {
		return strings.TrimSuffix(string(out), "\n")
	}
	cmd := exec.Command(command, args...)
	var stdout, stderr []byte
	var stdoutErr, stderrErr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		errorStr := fmt.Sprintf("cmd.Start() failed with '%s'", err)
		return "", errors.New(errorStr)
	}
	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, stdoutErr = copyAndCapture(stdoutIn)
		wg.Done()
	}()
	stderr, stderrErr = copyAndCapture(stderrIn)
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", &commandError{
				err:      exitErr.Error(),
				exitCode: exitErr.ExitCode(),
			}
		}
	}
	if stdoutErr != nil || stderrErr != nil {
		return "", errors.New("failed to capture stdout or stderr")
	}
	stderrStr := normalizeOutput(stderr)
	if len(stderrStr) > 0 {
		return stderrStr, nil
	}
	return normalizeOutput(stdout), nil
}

func (env *environment) runShellCommand(shell, command string) string {
	defer env.tracer.trace(time.Now(), "runShellCommand", shell, command)
	out, _ := env.runCommand(shell, "-c", command)
	return out
}

func (env *environment) hasCommand(command string) bool {
	defer env.tracer.trace(time.Now(), "hasCommand", command)
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
	defer env.tracer.trace(time.Now(), "lastErrorCode")
	return *env.args.ErrorCode
}

func (env *environment) executionTime() float64 {
	defer env.tracer.trace(time.Now(), "executionTime")
	if *env.args.ExecutionTime < 0 {
		return 0
	}
	return *env.args.ExecutionTime
}

func (env *environment) getArgs() *args {
	defer env.tracer.trace(time.Now(), "getArgs")
	return env.args
}

func (env *environment) getBatteryInfo() ([]*battery.Battery, error) {
	defer env.tracer.trace(time.Now(), "getBatteryInfo")
	return battery.GetAll()
}

func (env *environment) getShellName() string {
	defer env.tracer.trace(time.Now(), "getShellName")
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

func (env *environment) doGet(url string, timeout int) ([]byte, error) {
	defer env.tracer.trace(time.Now(), "doGet", url)
	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
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
	defer env.tracer.trace(time.Now(), "hasParentFilePath", path)
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
	defer env.tracer.trace(time.Now(), "stackCount")
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
