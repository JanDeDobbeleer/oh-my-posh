// +build !windows

package main

import (
	"errors"
	"os"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func (env *environment) isRunningAsRoot() bool {
	return os.Geteuid() == 0
}

func (env *environment) homeDir() string {
	return os.Getenv("HOME")
}

func (env *environment) getWindowTitle(imageName, windowTitleRegex string) (string, error) {
	return "", errors.New("not implemented")
}

func (env *environment) isWsl() bool {
	// one way to check
	// version := env.getFileContent("/proc/version")
	// return strings.Contains(version, "microsoft")
	// using env variable
	return env.getenv("WSL_DISTRO_NAME") != ""
}

func (env *environment) getTerminalWidth() (int, error) {
	width, err := terminal.Width()
	return int(width), err
}
