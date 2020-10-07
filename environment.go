package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/distatus/battery"
	ps "github.com/mitchellh/go-ps"
)

type environmentInfo interface {
	getenv(key string) string
	getcwd() string
	homeDir() string
	hasFiles(pattern string) bool
	getPathSeperator() string
	getCurrentUser() (*user.User, error)
	isRunningAsRoot() bool
	getHostName() (string, error)
	getRuntimeGOOS() string
	hasCommand(command string) bool
	runCommand(command string, args ...string) string
	runShellCommand(shell string, command string) string
	lastErrorCode() int
	getArgs() *args
	getBatteryInfo() (*battery.Battery, error)
	getShellName() string
}

type environment struct {
	args *args
}

func (env *environment) getenv(key string) string {
	return os.Getenv(key)
}

func (env *environment) getcwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	// on Windows, and being case sentisitive and not consistent and all, this gives silly issues
	return strings.Replace(dir, "c:", "C:", 1)
}

func (env *environment) homeDir() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	homeDir := usr.HomeDir
	// on Windows, and being case sentisitive and not consistent and all, this gives silly issues
	return strings.Replace(homeDir, "c:", "C:", 1)
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

func (env *environment) getPathSeperator() string {
	return string(os.PathSeparator)
}

func (env *environment) getCurrentUser() (*user.User, error) {
	return user.Current()
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

func (env *environment) runCommand(command string, args ...string) string {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (env *environment) runShellCommand(shell string, command string) string {
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

func (env *environment) getArgs() *args {
	return env.args
}

func (env *environment) getBatteryInfo() (*battery.Battery, error) {
	return battery.Get(0)
}

func (env *environment) getShellName() string {
	pid := os.Getppid()
	p, err := ps.FindProcess(pid)

	if err != nil {
		return "unknown"
	}
	shell := strings.Replace(p.Executable(), ".exe", "", 1)
	return strings.Trim(shell, " ")
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
