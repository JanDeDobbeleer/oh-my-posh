//go:build !windows

package runtime

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/shirou/gopsutil/v3/host"
	mem "github.com/shirou/gopsutil/v3/mem"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
	"golang.org/x/sys/unix"
)

func (term *Terminal) Root() bool {
	defer term.Trace(time.Now())
	return os.Geteuid() == 0
}

func (term *Terminal) Home() string {
	return os.Getenv("HOME")
}

func (term *Terminal) QueryWindowTitles(_, _ string) (string, error) {
	return "", &NotImplemented{}
}

func (term *Terminal) IsWsl() bool {
	defer term.Trace(time.Now())
	const key = "is_wsl"
	if val, found := term.Cache().Get(key); found {
		term.Debug(val)
		return val == "true"
	}

	var val bool
	defer func() {
		term.Cache().Set(key, strconv.FormatBool(val), cache.INFINITE)
	}()

	val = term.HasCommand("wslpath")
	term.Debug(strconv.FormatBool(val))

	return val
}

func (term *Terminal) IsWsl2() bool {
	defer term.Trace(time.Now())
	if !term.IsWsl() {
		return false
	}
	uname := term.FileContent("/proc/sys/kernel/osrelease")
	return strings.Contains(uname, "WSL2")
}

func (term *Terminal) IsCygwin() bool {
	defer term.Trace(time.Now())
	return false
}

func (term *Terminal) TerminalWidth() (int, error) {
	defer term.Trace(time.Now())

	if term.CmdFlags.TerminalWidth > 0 {
		term.DebugF("terminal width: %d", term.CmdFlags.TerminalWidth)
		return term.CmdFlags.TerminalWidth, nil
	}

	width, err := terminal.Width()
	if err != nil {
		term.Error(err)
	}

	// fetch width from the environment variable
	// in case the terminal width is not available
	if width == 0 {
		i, err := strconv.Atoi(term.Getenv("COLUMNS"))
		if err != nil {
			term.Error(err)
		}
		width = uint(i)
	}

	term.CmdFlags.TerminalWidth = int(width)
	term.DebugF("terminal width: %d", term.CmdFlags.TerminalWidth)
	return term.CmdFlags.TerminalWidth, err
}

func (term *Terminal) Platform() string {
	const key = "environment_platform"
	if val, found := term.Cache().Get(key); found {
		term.Debug(val)
		return val
	}

	var platform string
	defer func() {
		term.Cache().Set(key, platform, cache.INFINITE)
	}()

	if wsl := term.Getenv("WSL_DISTRO_NAME"); len(wsl) != 0 {
		platform = strings.Split(strings.ToLower(wsl), "-")[0]
		term.Debug(platform)
		return platform
	}

	platform, _, _, _ = host.PlatformInformation()
	if platform == "arch" {
		// validate for Manjaro
		lsbInfo := term.FileContent("/etc/lsb-release")
		if strings.Contains(strings.ToLower(lsbInfo), "manjaro") {
			platform = "manjaro"
		}
	}

	term.Debug(platform)
	return platform
}

func (term *Terminal) WindowsRegistryKeyValue(_ string) (*WindowsRegistryValue, error) {
	return nil, &NotImplemented{}
}

func (term *Terminal) InWSLSharedDrive() bool {
	if !term.IsWsl2() {
		return false
	}
	windowsPath := term.ConvertToWindowsPath(term.Pwd())
	return !strings.HasPrefix(windowsPath, `//wsl.localhost/`) && !strings.HasPrefix(windowsPath, `//wsl$/`)
}

func (term *Terminal) ConvertToWindowsPath(path string) string {
	windowsPath, err := term.RunCommand("wslpath", "-m", path)
	if err == nil {
		return windowsPath
	}
	return path
}

func (term *Terminal) ConvertToLinuxPath(path string) string {
	if linuxPath, err := term.RunCommand("wslpath", "-u", path); err == nil {
		return linuxPath
	}
	return path
}

func (term *Terminal) LookPath(command string) (string, error) {
	return exec.LookPath(command)
}

func (term *Terminal) DirIsWritable(path string) bool {
	defer term.Trace(time.Now(), path)
	return unix.Access(path, unix.W_OK) == nil
}

func (term *Terminal) Connection(_ ConnectionType) (*Connection, error) {
	// added to disable the linting error, we can implement this later
	if len(term.networks) == 0 {
		return nil, &NotImplemented{}
	}
	return nil, &NotImplemented{}
}

func (term *Terminal) Memory() (*Memory, error) {
	m := &Memory{}
	memStat, err := mem.VirtualMemory()
	if err != nil {
		term.Error(err)
		return nil, err
	}
	m.PhysicalTotalMemory = memStat.Total
	m.PhysicalAvailableMemory = memStat.Available
	m.PhysicalFreeMemory = memStat.Free
	m.PhysicalPercentUsed = memStat.UsedPercent
	swapStat, err := mem.SwapMemory()
	if err != nil {
		term.Error(err)
	}
	m.SwapTotalMemory = swapStat.Total
	m.SwapFreeMemory = swapStat.Free
	m.SwapPercentUsed = swapStat.UsedPercent
	return m, nil
}
