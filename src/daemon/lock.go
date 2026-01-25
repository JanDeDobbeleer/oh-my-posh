package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// LockFile represents an exclusive lock file to prevent multiple daemon instances.
type LockFile struct {
	path string
}

// NewLockFile creates and acquires an exclusive lock file.
// Returns error if lock is already held by another process.
func NewLockFile() (*LockFile, error) {
	stateDir := statePath()

	// Ensure state directory exists
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	path := filepath.Join(stateDir, "daemon.lock")

	// Try to acquire lock
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		// Lock file exists - check if process is still alive
		if os.IsExist(err) {
			pid, pidErr := ReadPID(path)
			if pidErr != nil {
				// Can't read PID - remove stale lock and retry
				log.Debug("failed to read PID from lock file, removing stale lock")
				_ = os.Remove(path)
				return NewLockFile()
			}

			if !IsProcessRunning(pid) {
				// Process is dead - remove stale lock and retry
				log.Debugf("daemon process %d is not running, removing stale lock", pid)
				_ = os.Remove(path)
				return NewLockFile()
			}

			return nil, fmt.Errorf("daemon already running with PID %d", pid)
		}

		return nil, err
	}

	lf := &LockFile{
		path: path,
	}

	// Write our PID to the lock file
	if err := lf.WritePID(file); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return nil, err
	}
	_ = file.Close()

	return lf, nil
}

// WritePID writes the daemon PID to the lock file.
func (lf *LockFile) WritePID(file *os.File) error {
	_, err := fmt.Fprintf(file, "%d\n", os.Getpid())
	if err != nil {
		return err
	}
	return file.Sync()
}

// ReadPID reads the PID from an existing lock file.
func ReadPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID in lock file: %s", pidStr)
	}

	return pid, nil
}

// Release removes the lock file.
func (lf *LockFile) Release() error {
	return os.Remove(lf.path)
}

// CleanupLock removes a stale lock file (for crash recovery).
func CleanupLock() error {
	path := filepath.Join(statePath(), "daemon.lock")
	return os.Remove(path)
}

// KillDaemon checks for an existing daemon and kills it if running.
// It also removes the lock file.
func KillDaemon() error {
	path := filepath.Join(statePath(), "daemon.lock")

	pid, err := ReadPID(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		// Corrupt lock file, force remove
		return os.Remove(path)
	}

	if IsProcessRunning(pid) {
		proc, err := os.FindProcess(pid)
		if err == nil {
			_ = proc.Kill()
			_, _ = proc.Wait()
		}
	}

	return os.Remove(path)
}
