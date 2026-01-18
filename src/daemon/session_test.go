package daemon

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionManager(t *testing.T) {
	called := false
	sm := NewSessionManager(func(_ int) { called = true }, nil)

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.sessions)
	assert.Equal(t, 0, sm.Count())
	assert.False(t, called)
}

func TestSessionManagerRegister(t *testing.T) {
	sm := NewSessionManager(nil, nil)

	// Register a session with current process PID (known to exist)
	pid := os.Getpid()
	sm.Register(pid, "uuid", "shell")

	assert.Equal(t, 1, sm.Count())

	// Registering same PID again should be a no-op
	sm.Register(pid, "uuid", "shell")
	assert.Equal(t, 1, sm.Count())
}

func TestSessionManagerUnregister(t *testing.T) {
	var callCount atomic.Int32
	sm := NewSessionManager(func(_ int) { callCount.Add(1) }, nil)

	pid := os.Getpid()
	sm.Register(pid, "uuid", "shell")
	assert.Equal(t, 1, sm.Count())

	sm.Unregister(pid)
	assert.Equal(t, 0, sm.Count())

	// onEmpty should be called when last session ends
	assert.Equal(t, int32(1), callCount.Load())

	// Unregistering non-existent PID should be a no-op
	sm.Unregister(99999)
	assert.Equal(t, 0, sm.Count())
}

func TestSessionManagerMultipleSessions(t *testing.T) {
	var callCount atomic.Int32
	sm := NewSessionManager(nil, func() { callCount.Add(1) })

	pid1 := os.Getpid()
	pid2 := os.Getppid() // Parent process, should also exist

	sm.Register(pid1, "uuid1", "shell")
	sm.Register(pid2, "uuid2", "shell")
	assert.Equal(t, 2, sm.Count())

	// Unregistering first should not trigger onEmpty
	sm.Unregister(pid1)
	assert.Equal(t, 1, sm.Count())
	assert.Equal(t, int32(0), callCount.Load())

	// Unregistering second should trigger onEmpty
	sm.Unregister(pid2)
	assert.Equal(t, 0, sm.Count())
	assert.Equal(t, int32(1), callCount.Load())
}

func TestSessionManagerProcessExitDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process exit detection test in short mode")
	}

	var callCount atomic.Int32
	sm := NewSessionManager(nil, func() { callCount.Add(1) })

	// Start a short-lived subprocess
	process := startTestProcess(t)
	pid := process.Pid

	// Register the session
	sm.Register(pid, "uuid", "shell")
	assert.Equal(t, 1, sm.Count())

	// Wait for process to exit and be detected
	// The process watcher should detect the exit and unregister
	assert.Eventually(t, func() bool {
		return sm.Count() == 0 && callCount.Load() == 1
	}, time.Second, 10*time.Millisecond)
}

func TestSessionManagerNonExistentPID(t *testing.T) {
	var callCount atomic.Int32
	sm := NewSessionManager(nil, func() { callCount.Add(1) })

	// Register a PID that doesn't exist
	// Use a very high PID that's unlikely to exist
	fakePID := 999999999
	sm.Register(fakePID, "uuid", "shell")

	// The watcher should detect immediately that process doesn't exist
	// and trigger cleanup
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, sm.Count())
	assert.Equal(t, int32(1), callCount.Load())
}

// startTestProcess starts a process that exits after a short time.
// Returns the started process.
func startTestProcess(t *testing.T) *os.Process {
	t.Helper()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: use ping as a sleep alternative
		cmd = exec.CommandContext(context.Background(), "ping", "-n", "1", "127.0.0.1")
	} else {
		// Unix: use sleep
		cmd = exec.CommandContext(context.Background(), "sleep", "0.1")
	}

	err := cmd.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	return cmd.Process
}
