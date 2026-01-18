//go:build windows

package daemon

import (
	"context"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"golang.org/x/sys/windows"
)

// waitForProcessExit blocks until the process with the given PID exits.
// Uses OpenProcess + WaitForSingleObject to wait for process termination.
func waitForProcessExit(ctx context.Context, pid int) {
	// Open handle to the process
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		log.Debugf("OpenProcess failed for PID %d: %v", pid, err)
		if !IsProcessRunning(pid) {
			return
		}
		pollForProcessExit(ctx, pid)
		return
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	// Wait for process exit in a goroutine so we can watch for context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Wait indefinitely for process to exit
		_, err := windows.WaitForSingleObject(handle, windows.INFINITE)
		if err != nil {
			log.Debugf("WaitForSingleObject error for PID %d: %v", pid, err)
		}
	}()

	// Wait for either process exit or context cancellation
	select {
	case <-done:
		log.Debugf("Process %d exit detected via WaitForSingleObject", pid)
	case <-ctx.Done():
		log.Debugf("Context cancelled while watching PID %d", pid)
	}
}
