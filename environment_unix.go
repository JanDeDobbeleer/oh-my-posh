// +build !windows

package main

import (
	"os"
)

func (env *environment) isRunningAsRoot() bool {
	return os.Geteuid() == 0
}

func (env *environment) homeDir() string {
	return os.Getenv("HOME")
}
