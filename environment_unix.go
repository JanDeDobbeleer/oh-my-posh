// +build !windows

package main

import (
	"os"
)

func (env *environment) isRunningAsRoot() bool {
	return os.Geteuid() == 0
}
