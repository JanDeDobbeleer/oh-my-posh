//go:build darwin

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
	log.Debugf("waitForProcessExit(macos): starting for PID %d", pid)
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
		Ident:  uint64(pid),
		Filter: syscall.EVFILT_PROC,
		Flags:  syscall.EV_ADD | syscall.EV_ONESHOT,
		Fflags: syscall.NOTE_EXIT,
	}

	// Register the event
	log.Debugf("waitForProcessExit(macos): registering kevent for PID %d", pid)
	_, err = syscall.Kevent(kq, []syscall.Kevent_t{event}, nil, nil)
	if err != nil {
		log.Debugf("Failed to register kevent for PID %d: %v", pid, err)
		if !IsProcessRunning(pid) {
			log.Debugf("waitForProcessExit(macos): process %d already dead", pid)
			return
		}
		pollForProcessExit(ctx, pid)
		return
	}

	// Wait for event in a goroutine so we can also watch for context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		events := make([]syscall.Kevent_t, 1)
		// Block until process exits (no timeout)
		log.Debugf("waitForProcessExit(macos): waiting for kevent for PID %d", pid)
		n, err := syscall.Kevent(kq, nil, events, nil)
		if err != nil {
			log.Debugf("Kevent wait error for PID %d: %v", pid, err)
		}
		log.Debugf("waitForProcessExit(macos): kevent returned %d for PID %d", n, pid)
	}()

	// Wait for either process exit or context cancellation
	select {
	case <-done:
		log.Debugf("Process %d exit detected via kqueue", pid)
	case <-ctx.Done():
		log.Debugf("Context cancelled while watching PID %d", pid)
	}
}
