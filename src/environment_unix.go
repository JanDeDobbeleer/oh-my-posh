//go:build !windows

package main

import (
	"errors"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func (env *environment) isRunningAsRoot() bool {
	defer env.trace(time.Now(), "isRunningAsRoot")
	return os.Geteuid() == 0
}

func (env *environment) homeDir() string {
	return os.Getenv("HOME")
}

func (env *environment) getWindowTitle(imageName, windowTitleRegex string) (string, error) {
	return "", errors.New("not implemented")
}

func (env *environment) isWsl() bool {
	defer env.trace(time.Now(), "isWsl")
	// one way to check
	// version := env.getFileContent("/proc/version")
	// return strings.Contains(version, "microsoft")
	// using env variable
	return env.getenv("WSL_DISTRO_NAME") != ""
}

func (env *environment) getTerminalWidth() (int, error) {
	defer env.trace(time.Now(), "getTerminalWidth")
	width, err := terminal.Width()
	if err != nil {
		env.log(Error, "runCommand", err.Error())
	}
	return int(width), err
}

func (env *environment) getPlatform() string {
	p, _, _, _ := host.PlatformInformation()
	return p
}

func (env *environment) getCachePath() string {
	defer env.trace(time.Now(), "getCachePath")
	// get XDG_CACHE_HOME if present
	if cachePath := returnOrBuildCachePath(env.getenv("XDG_CACHE_HOME")); len(cachePath) != 0 {
		return cachePath
	}
	// HOME cache folder
	if cachePath := returnOrBuildCachePath(env.homeDir() + "/.cache"); len(cachePath) != 0 {
		return cachePath
	}
	return env.homeDir()
}

func (env *environment) getWindowsRegistryKeyValue(regPath, regKey string) (string, error) {
	return "", errors.New("not implemented")
}
