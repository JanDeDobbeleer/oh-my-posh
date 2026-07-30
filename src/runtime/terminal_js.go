//go:build js

package runtime

import (
	"strconv"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// QueryWindowTitles has no window manager to query inside a browser (or a Node
// test harness), unlike the real unix/windows implementations that shell out to
// (or link against) a windowing API.
func (term *Terminal) QueryWindowTitles(_, _ string) (string, error) {
	return "", &NotImplemented{}
}

// QueryMediaPlayer has no OS media session to query inside a browser, unlike the
// real unix/windows implementations.
func (term *Terminal) QueryMediaPlayer(_ string) (*MediaInfo, error) {
	return nil, &NotImplemented{}
}

// WindowsRegistryKeyValue has no Windows registry to query inside a browser.
func (term *Terminal) WindowsRegistryKeyValue(_ string) (*WindowsRegistryValue, error) {
	return nil, &NotImplemented{}
}

// Connection has no network interfaces to enumerate inside a browser sandbox -
// unlike the real unix/windows implementations, which read them off the host.
func (term *Terminal) Connection(_ ConnectionType) (*Connection, error) {
	return nil, &NotImplemented{}
}

// Memory has no OS memory stats to read inside a browser, unlike the real
// unix/windows implementations backed by gopsutil/mem.
func (term *Terminal) Memory() (*Memory, error) {
	return nil, &NotImplemented{}
}

// IsWsl always reports false on js/wasm: WSL is a Windows-specific concept and
// there is no wslpath (or any command at all) to probe for inside a browser.
func (term *Terminal) IsWsl() bool {
	return false
}

// IsWsl2 always reports false for the same reason as IsWsl.
func (term *Terminal) IsWsl2() bool {
	return false
}

// IsCygwin always reports false: there is no Cygwin environment inside a browser.
func (term *Terminal) IsCygwin() bool {
	return false
}

// InWSLSharedDrive always reports false: this is only ever meaningful once
// IsWsl2 is true, which it never is here.
func (term *Terminal) InWSLSharedDrive() bool {
	return false
}

// ConvertToWindowsPath returns input unchanged: there is no wslpath to shell out
// to, and a browser path never needs the Windows/WSL translation in the first
// place.
func (term *Terminal) ConvertToWindowsPath(input string) string {
	return input
}

// ConvertToLinuxPath returns input unchanged, for the same reason as
// ConvertToWindowsPath.
func (term *Terminal) ConvertToLinuxPath(input string) string {
	return input
}

// TerminalWidth answers from CmdFlags.TerminalWidth (set by the caller, e.g.
// render.Config's columns) or the COLUMNS env var fallback - never from the
// wayneashleyberry/terminal-dimensions package, which shells out to `stty` and
// has nothing to query inside a browser.
func (term *Terminal) TerminalWidth() (int, error) {
	defer log.Trace(time.Now())

	if term.CmdFlags.TerminalWidth > 0 {
		return term.CmdFlags.TerminalWidth, nil
	}

	width, err := strconv.Atoi(term.Getenv("COLUMNS"))
	if err != nil {
		return 0, err
	}

	term.CmdFlags.TerminalWidth = width

	return term.CmdFlags.TerminalWidth, nil
}

// Platform matches the current wasm behavior: host.PlatformInformation errors
// out via gopsutil's js fallback stub, and reading /etc/lsb-release is refused
// under DataOnly (the only mode this build ever runs in) - so this has always
// resolved to "", not a real value. Keeping that instead of switching to
// UNKNOWN preserves golden-render parity; revisit if that ever changes.
func (term *Terminal) Platform() string {
	return ""
}
