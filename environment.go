package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/distatus/battery"
)

type environmentInfo interface {
	getenv(key string) string
	getwd() (string, error)
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
}

type environment struct {
	args *args
}

func (env *environment) getenv(key string) string {
	return os.Getenv(key)
}

func (env *environment) getwd() (string, error) {
	return os.Getwd()
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

func cleanHostName(hostName string) string {
	garbage := []string{
		".lan",
		".local",
	}
	for _, g := range garbage {
		hostName = strings.Replace(hostName, g, "", 1)
	}
	return hostName
}
