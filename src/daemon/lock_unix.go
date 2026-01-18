//go:build !windows

package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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

// IsProcessRunning checks if the daemon process with the given PID is running.
// Uses kill(pid, 0) to check existence, then verifies the process is actually
// oh-my-posh to avoid false positives from PID reuse.
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to check with kill(0)
	err = process.Signal(syscall.Signal(0))
	if err != nil && err != syscall.EPERM {
		return false
	}

	// If the PID is our own process, it's definitely running (used by tests and self-checks)
	if pid == os.Getpid() {
		return true
	}

	// Process exists, but verify it's actually oh-my-posh (PID reuse protection).
	exe, err := processExecutable(pid)
	if err != nil {
		// Can't determine — assume running to be safe
		return true
	}

	name := filepath.Base(exe)
	return strings.Contains(name, "oh-my-posh") || strings.HasSuffix(name, ".test")
}

// processExecutable returns the executable path for a given PID.
// Tries /proc/pid/exe (Linux) first, then falls back to ps (macOS).
func processExecutable(pid int) (string, error) {
	// Linux: readlink /proc/pid/exe
	if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
		return exe, nil
	}

	// macOS: ps -p PID -o comm=
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "ps", "-p", fmt.Sprint(pid), "-o", "comm=").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
