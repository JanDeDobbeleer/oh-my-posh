//go:build windows

package daemon

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// statePath returns the state directory path for Windows.
// Uses %LOCALAPPDATA% which is the appropriate location for local application state.
func statePath() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "oh-my-posh")
}

// IsProcessRunning checks if a process with the given PID is running.
// Uses OpenProcess to check if the process exists without terminating it.
func IsProcessRunning(pid int) bool {
	const processQueryLimitedInformation = 0x1000
	h, err := windows.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = windows.CloseHandle(h)
	return true
}
