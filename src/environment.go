package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/distatus/battery"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/process"
)

const (
	unknown         = "unknown"
	windowsPlatform = "windows"
)

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
	getBatteryInfo() (*battery.Battery, error)
	getShellName() string
	getWindowTitle(imageName, windowTitleRegex string) (string, error)
	doGet(url string) ([]byte, error)
}

type environment struct {
	args *args
	cwd  string
}

type commandError struct {
	exitCode int
}

func (e *commandError) Error() string {
	return fmt.Sprintf("%d", e.exitCode)
}

func (env *environment) getenv(key string) string {
	return os.Getenv(key)
}

func (env *environment) getcwd() string {
	if env.cwd != "" {
		return env.cwd
	}
	correctPath := func(pwd string) string {
		// on Windows, and being case sentisitive and not consistent and all, this gives silly issues
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
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()
	output := new(bytes.Buffer)
	defer output.Reset()
	buf := bufio.NewReader(stdout)
	multiline := false
	for {
		line, _, _ := buf.ReadLine()
		if line == nil {
			break
		}
		if multiline {
			output.WriteString("\n")
		}
		output.Write(line)
		multiline = true
	}
	return output.String(), nil
}

func (env *environment) runShellCommand(shell, command string) string {
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Println(err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (env *environment) hasCommand(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
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

func (env *environment) getBatteryInfo() (*battery.Battery, error) {
	return battery.Get(0)
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
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
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
