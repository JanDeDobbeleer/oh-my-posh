//go:build !windows

package platform

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	mem "github.com/shirou/gopsutil/v3/mem"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
	"golang.org/x/sys/unix"
)

func (env *Shell) Root() bool {
	defer env.Trace(time.Now())
	return os.Geteuid() == 0
}

func (env *Shell) Home() string {
	return os.Getenv("HOME")
}

func (env *Shell) QueryWindowTitles(_, _ string) (string, error) {
	return "", &NotImplemented{}
}

func (env *Shell) IsWsl() bool {
	defer env.Trace(time.Now())
	const key = "is_wsl"
	if val, found := env.Cache().Get(key); found {
		env.Debug(val)
		return val == "true"
	}
	var val bool
	defer func() {
		env.Cache().Set(key, strconv.FormatBool(val), -1)
	}()
	val = env.HasCommand("wslpath")
	env.Debug(strconv.FormatBool(val))
	return val
}

func (env *Shell) IsWsl2() bool {
	defer env.Trace(time.Now())
	if !env.IsWsl() {
		return false
	}
	uname := env.FileContent("/proc/sys/kernel/osrelease")
	return strings.Contains(uname, "WSL2")
}

func (env *Shell) TerminalWidth() (int, error) {
	defer env.Trace(time.Now())
	if env.CmdFlags.TerminalWidth != 0 {
		return env.CmdFlags.TerminalWidth, nil
	}
	width, err := terminal.Width()
	if err != nil {
		env.Error(err)
	}
	return int(width), err
}

func (env *Shell) Platform() string {
	const key = "environment_platform"
	if val, found := env.Cache().Get(key); found {
		env.Debug(val)
		return val
	}
	var platform string
	defer func() {
		env.Cache().Set(key, platform, -1)
	}()
	if wsl := env.Getenv("WSL_DISTRO_NAME"); len(wsl) != 0 {
		platform = strings.Split(strings.ToLower(wsl), "-")[0]
		env.Debug(platform)
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
	env.Debug(platform)
	return platform
}

func (env *Shell) CachePath() string {
	defer env.Trace(time.Now())
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

func (env *Shell) WindowsRegistryKeyValue(_ string) (*WindowsRegistryValue, error) {
	return nil, &NotImplemented{}
}

func (env *Shell) InWSLSharedDrive() bool {
	if !env.IsWsl2() {
		return false
	}
	windowsPath := env.ConvertToWindowsPath(env.Pwd())
	return !strings.HasPrefix(windowsPath, `//wsl.localhost/`) && !strings.HasPrefix(windowsPath, `//wsl$/`)
}

func (env *Shell) ConvertToWindowsPath(path string) string {
	windowsPath, err := env.RunCommand("wslpath", "-m", path)
	if err == nil {
		return windowsPath
	}
	return path
}

func (env *Shell) ConvertToLinuxPath(path string) string {
	if linuxPath, err := env.RunCommand("wslpath", "-u", path); err == nil {
		return linuxPath
	}
	return path
}

func (env *Shell) LookWinAppPath(_ string) (string, error) {
	return "", errors.New("not relevant")
}

func (env *Shell) DirIsWritable(path string) bool {
	defer env.Trace(time.Now(), path)
	return unix.Access(path, unix.W_OK) == nil
}

func (env *Shell) Connection(_ ConnectionType) (*Connection, error) {
	// added to disable the linting error, we can implement this later
	if len(env.networks) == 0 {
		return nil, &NotImplemented{}
	}
	return nil, &NotImplemented{}
}

func (env *Shell) Memory() (*Memory, error) {
	m := &Memory{}
	memStat, err := mem.VirtualMemory()
	if err != nil {
		env.Error(err)
		return nil, err
	}
	m.PhysicalTotalMemory = memStat.Total
	m.PhysicalAvailableMemory = memStat.Available
	m.PhysicalFreeMemory = memStat.Free
	m.PhysicalPercentUsed = memStat.UsedPercent
	swapStat, err := mem.SwapMemory()
	if err != nil {
		env.Error(err)
	}
	m.SwapTotalMemory = swapStat.Total
	m.SwapFreeMemory = swapStat.Free
	m.SwapPercentUsed = swapStat.UsedPercent
	return m, nil
}
