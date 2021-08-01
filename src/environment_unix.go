// +build !windows

package main

import (
	"errors"
	"os"
	"time"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func (env *environment) isRunningAsRoot() bool {
	defer env.tracer.trace(time.Now(), "isRunningAsRoot")
	return os.Geteuid() == 0
}

func (env *environment) homeDir() string {
	return os.Getenv("HOME")
}

func (env *environment) getWindowTitle(imageName, windowTitleRegex string) (string, error) {
	return "", errors.New("not implemented")
}

func (env *environment) isWsl() bool {
	defer env.tracer.trace(time.Now(), "isWsl")
	// one way to check
	// version := env.getFileContent("/proc/version")
	// return strings.Contains(version, "microsoft")
	// using env variable
	return env.getenv("WSL_DISTRO_NAME") != ""
}

func (env *environment) getTerminalWidth() (int, error) {
	defer env.tracer.trace(time.Now(), "getTerminalWidth")
	width, err := terminal.Width()
	return int(width), err
}
