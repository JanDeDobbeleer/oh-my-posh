//go:build !windows

package environment

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func (env *ShellEnvironment) Root() bool {
	defer env.trace(time.Now(), "Root")
	return os.Geteuid() == 0
}

func (env *ShellEnvironment) Home() string {
	return os.Getenv("HOME")
}

func (env *ShellEnvironment) WindowTitle(imageName, windowTitleRegex string) (string, error) {
	return "", errors.New("not implemented")
}

func (env *ShellEnvironment) IsWsl() bool {
	defer env.trace(time.Now(), "IsWsl")
	// one way to check
	// version := env.FileContent("/proc/version")
	// return strings.Contains(version, "microsoft")
	// using env variable
	return env.Getenv("WSL_DISTRO_NAME") != ""
}

func (env *ShellEnvironment) IsWsl2() bool {
	defer env.trace(time.Now(), "IsWsl2")
	if !env.IsWsl() {
		return false
	}
	uname := env.FileContent("/proc/sys/kernel/osrelease")
	return strings.Contains(uname, "WSL2")
}

func (env *ShellEnvironment) TerminalWidth() (int, error) {
	defer env.trace(time.Now(), "TerminalWidth")
	if *env.args.TerminalWidth != 0 {
		return *env.args.TerminalWidth, nil
	}
	width, err := terminal.Width()
	if err != nil {
		env.log(Error, "TerminalWidth", err.Error())
	}
	return int(width), err
}

func (env *ShellEnvironment) Platform() string {
	const key = "environment_platform"
	if val, found := env.Cache().Get(key); found {
		return val
	}
	var platform string
	defer func() {
		env.Cache().Set(key, platform, -1)
	}()
	if wsl := env.Getenv("WSL_DISTRO_NAME"); len(wsl) != 0 {
		platform = strings.ToLower(wsl)
		return strings.Split(platform, "-")[0]
	}
	platform, _, _, _ = host.PlatformInformation()
	if platform == "arch" {
		// validate for Manjaro
		lsbInfo := env.FileContent("/etc/lsb-release")
		if strings.Contains(strings.ToLower(lsbInfo), "manjaro") {
			platform = "manjaro"
		}
	}
	return platform
}

func (env *ShellEnvironment) CachePath() string {
	defer env.trace(time.Now(), "CachePath")
	// get XDG_CACHE_HOME if present
	if cachePath := returnOrBuildCachePath(env.Getenv("XDG_CACHE_HOME")); len(cachePath) != 0 {
		return cachePath
	}
	// HOME cache folder
	if cachePath := returnOrBuildCachePath(env.Home() + "/.cache"); len(cachePath) != 0 {
		return cachePath
	}
	return env.Home()
}

func (env *ShellEnvironment) WindowsRegistryKeyValue(path string) (*WindowsRegistryValue, error) {
	return nil, errors.New("not implemented")
}

func (env *ShellEnvironment) InWSLSharedDrive() bool {
	return env.IsWsl() && strings.HasPrefix(env.Pwd(), "/mnt/")
}

func (env *ShellEnvironment) ConvertToWindowsPath(path string) string {
	windowsPath, err := env.RunCommand("wslpath", "-w", path)
	if err == nil {
		return windowsPath
	}
	return path
}

func (env *ShellEnvironment) ConvertToLinuxPath(path string) string {
	linuxPath, err := env.RunCommand("wslpath", "-u", path)
	if err == nil {
		return linuxPath
	}
	return path
}

func (env *ShellEnvironment) WifiNetwork() (*WifiInfo, error) {
	return nil, errors.New("not implemented")
}

func (env *ShellEnvironment) LookWinAppPath(file string) (string, error) {
	return "", errors.New("not relevant")
}
