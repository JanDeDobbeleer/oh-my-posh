// +build !windows

package main

import (
	"errors"
	"os"
)

func (env *environment) isRunningAsRoot() bool {
	return os.Geteuid() == 0
}

func (env *environment) homeDir() string {
	return os.Getenv("HOME")
}

func (env *environment) getWindowTitle(imageName string, windowTitleRegex string) (string, error) {
	return "", errors.New("not implemented")
}
