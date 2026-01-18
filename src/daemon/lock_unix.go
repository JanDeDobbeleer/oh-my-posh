//go:build !windows

package daemon

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

// statePath returns the state directory path following XDG Base Directory Specification.
// Uses $XDG_STATE_HOME if set, otherwise defaults to ~/.local/state.
func statePath() string {
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, "oh-my-posh")
	}
	return filepath.Join(path.Home(), ".local", "state", "oh-my-posh")
}

// IsProcessRunning checks if a process with the given PID is running.
// Uses kill(pid, 0) which doesn't send a signal but checks if the process exists.
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to check with kill(0)
	err = process.Signal(syscall.Signal(0))
	return err == nil || err == syscall.EPERM
}
