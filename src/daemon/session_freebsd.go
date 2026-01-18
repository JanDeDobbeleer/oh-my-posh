//go:build freebsd

package daemon

import (
	"context"
	"syscall"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// waitForProcessExit blocks until the process with the given PID exits.
// Uses kqueue with EVFILT_PROC to get notified when the process exits.
// Returns immediately if context is cancelled or process doesn't exist.
func waitForProcessExit(ctx context.Context, pid int) {
	// Check if process exists first
	if !IsProcessRunning(pid) {
		log.Debugf("Process %d not running, returning immediately", pid)
		return
	}

	// Create kqueue
	kq, err := syscall.Kqueue()
	if err != nil {
		log.Debugf("Failed to create kqueue for PID %d: %v", pid, err)
		pollForProcessExit(ctx, pid)
		return
	}
	defer syscall.Close(kq)

	// Register for process exit notification
	event := syscall.Kevent_t{
		Filter: syscall.EVFILT_PROC,
		Flags:  syscall.EV_ADD | syscall.EV_ONESHOT,
		Fflags: syscall.NOTE_EXIT,
	}
	setIdent(&event, pid)

	// Register the event
	_, err = syscall.Kevent(kq, []syscall.Kevent_t{event}, nil, nil)
	if err != nil {
		log.Debugf("Failed to register kevent for PID %d: %v", pid, err)
		pollForProcessExit(ctx, pid)
		return
	}

	// Wait for event in a goroutine so we can also watch for context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		events := make([]syscall.Kevent_t, 1)
		// Block until process exits (no timeout)
		_, err := syscall.Kevent(kq, nil, events, nil)
		if err != nil {
			log.Debugf("Kevent wait error for PID %d: %v", pid, err)
		}
	}()

	// Wait for either process exit or context cancellation
	select {
	case <-done:
		log.Debugf("Process %d exit detected via kqueue", pid)
	case <-ctx.Done():
		log.Debugf("Context cancelled while watching PID %d", pid)
	}
}
