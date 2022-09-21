//go:build !windows

package environment

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func (env *ShellEnvironment) Root() bool {
	defer env.Trace(time.Now(), "Root")
	return os.Geteuid() == 0
}

func (env *ShellEnvironment) Home() string {
	return os.Getenv("HOME")
}

func (env *ShellEnvironment) QueryWindowTitles(processName, windowTitleRegex string) (string, error) {
	return "", errors.New("not implemented")
}

func (env *ShellEnvironment) IsWsl() bool {
	defer env.Trace(time.Now(), "IsWsl")
	// one way to check
	// version := env.FileContent("/proc/version")
	// return strings.Contains(version, "microsoft")
	// using env variable
	return env.Getenv("WSL_DISTRO_NAME") != ""
}

func (env *ShellEnvironment) IsWsl2() bool {
	defer env.Trace(time.Now(), "IsWsl2")
	if !env.IsWsl() {
		return false
	}
	uname := env.FileContent("/proc/sys/kernel/osrelease")
	return strings.Contains(uname, "WSL2")
}

func (env *ShellEnvironment) TerminalWidth() (int, error) {
	defer env.Trace(time.Now(), "TerminalWidth")
	if env.CmdFlags.TerminalWidth != 0 {
		return env.CmdFlags.TerminalWidth, nil
	}
	width, err := terminal.Width()
	if err != nil {
		env.Log(Error, "TerminalWidth", err.Error())
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
		platform = strings.Split(strings.ToLower(wsl), "-")[0]
		return platform
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
	defer env.Trace(time.Now(), "CachePath")
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
	if !env.IsWsl() {
		return false
	}
	windowsPath := env.ConvertToWindowsPath(env.Pwd())
	return !strings.HasPrefix(windowsPath, `//wsl.localhost`) && !strings.HasPrefix(windowsPath, `//wsl$`)
}

func (env *ShellEnvironment) ConvertToWindowsPath(path string) string {
	windowsPath, err := env.RunCommand("wslpath", "-m", path)
	if err == nil {
		return windowsPath
	}
	return path
}

func (env *ShellEnvironment) ConvertToLinuxPath(path string) string {
	if linuxPath, err := env.RunCommand("wslpath", "-u", path); err == nil {
		return linuxPath
	}
	return path
}

func (env *ShellEnvironment) LookWinAppPath(file string) (string, error) {
	return "", errors.New("not relevant")
}

func (env *ShellEnvironment) DirIsWritable(path string) bool {
	defer env.Trace(time.Now(), "DirIsWritable")
	info, err := os.Stat(path)
	if err != nil {
		env.Log(Error, "DirIsWritable", err.Error())
		return false
	}

	if !info.IsDir() {
		env.Log(Error, "DirIsWritable", "Path isn't a directory")
		return false
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		env.Log(Error, "DirIsWritable", "Write permission bit is not set on this file for user")
		return false
	}

	var stat syscall.Stat_t
	if err = syscall.Stat(path, &stat); err != nil {
		env.Log(Error, "DirIsWritable", err.Error())
		return false
	}

	if uint32(os.Geteuid()) != stat.Uid {
		env.Log(Error, "DirIsWritable", "User doesn't have permission to write to this directory")
		return false
	}

	return true
}

func (env *ShellEnvironment) GetAllNetworkInterfaces() (*[]NetworkInfo, error) {
	return nil , errors.New("not implemented")
}
