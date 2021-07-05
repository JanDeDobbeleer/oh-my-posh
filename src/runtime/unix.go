// +build !windows

package runtime

import (
	"errors"
	"os"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func (env *Shell) IsRunningAsRoot() bool {
	return os.Geteuid() == 0
}

func (env *Shell) HomeDir() string {
	return os.Getenv("HOME")
}

func (env *Shell) GetWindowTitle(imageName, windowTitleRegex string) (string, error) {
	return "", errors.New("not implemented")
}

func (env *Shell) IsWsl() bool {
	// one way to check
	// version := env.GetFileContent("/proc/version")
	// return strings.Contains(version, "microsoft")
	// using env variable
	return env.Getenv("WSL_DISTRO_NAME") != ""
}

func (env *Shell) GetTerminalWidth() (int, error) {
	width, err := terminal.Width()
	return int(width), err
}
