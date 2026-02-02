//go:build linux

package daemon

import (
	"context"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"golang.org/x/sys/unix"
)

// waitForProcessExit blocks until the process with the given PID exits.
// Uses pidfd_open + poll to get notified when the process exits.
// Requires kernel 5.3+. If pidfd is not available, the function returns
// and the session will only be cleaned up via explicit unregister.
func waitForProcessExit(ctx context.Context, pid int) {
	// Try to use pidfd (requires kernel 5.3+)
	pidfd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		log.Debugf("pidfd_open failed for PID %d: %v (kernel 5.3+ required)", pid, err)
		if !IsProcessRunning(pid) {
			return
		}
		pollForProcessExit(ctx, pid)
		return
	}
	defer unix.Close(pidfd)

	// Wait for process exit in a goroutine so we can watch for context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Poll the pidfd - it becomes readable when process exits
		fds := []unix.PollFd{
			{Fd: int32(pidfd), Events: unix.POLLIN},
		}
		// Block indefinitely (-1 timeout) until process exits
		_, err := unix.Poll(fds, -1)
		if err != nil {
			log.Debugf("Poll error for PID %d: %v", pid, err)
		}
	}()

	// Wait for either process exit or context cancellation
	select {
	case <-done:
		log.Debugf("Process %d exit detected via pidfd", pid)
	case <-ctx.Done():
		log.Debugf("Context cancelled while watching PID %d", pid)
	}
}
