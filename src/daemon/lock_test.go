package daemon

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newLockFileWithPath creates a lock file at the specified path for testing.
// This avoids dependency on cache.Path() which caches globally.
func newLockFileWithPath(path string) (*LockFile, error) {
	// Try to acquire lock
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		// Lock file exists - check if process is still alive
		if os.IsExist(err) {
			pid, pidErr := ReadPID(path)
			if pidErr != nil {
				// Can't read PID - remove stale lock and retry
				_ = os.Remove(path)
				return newLockFileWithPath(path)
			}

			if !IsProcessRunning(pid) {
				// Process is dead - remove stale lock and retry
				_ = os.Remove(path)
				return newLockFileWithPath(path)
			}

			return nil, err
		}

		return nil, err
	}

	lf := &LockFile{
		path: path,
	}

	// Write our PID to the lock file
	if err := lf.WritePID(file); err != nil {
		_ = file.Close()
		_ = lf.Release()
		return nil, err
	}
	_ = file.Close()

	return lf, nil
}

func TestLockFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	// Create lock - should succeed
	lock, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)
	require.NotNil(t, lock)
	defer func() { _ = lock.Release() }()

	// Verify lock file exists
	_, err = os.Stat(lockPath)
	assert.NoError(t, err, "lock file should exist")

	// Verify PID was written
	pid, err := ReadPID(lockPath)
	require.NoError(t, err)
	assert.Equal(t, os.Getpid(), pid)
}

func TestLockFileExclusivity(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	// Create first lock
	lock1, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)
	require.NotNil(t, lock1)
	defer func() { _ = lock1.Release() }()

	// Try to create second lock - should fail because current process is running
	lock2, err := newLockFileWithPath(lockPath)
	assert.Error(t, err)
	assert.Nil(t, lock2)
}

func TestLockFileRelease(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	// Create and release lock
	lock1, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)
	require.NotNil(t, lock1)

	err = lock1.Release()
	assert.NoError(t, err)

	// Verify lock file was removed
	_, err = os.Stat(lockPath)
	assert.True(t, os.IsNotExist(err), "lock file should be removed after release")

	// Should be able to create new lock after release
	lock2, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)
	require.NotNil(t, lock2)
	defer func() { _ = lock2.Release() }()
}

func TestStaleLockRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	// Create a stale lock file with a non-existent PID
	// Use PID 999999 which is unlikely to exist
	err := os.WriteFile(lockPath, []byte("999999\n"), 0o600)
	require.NoError(t, err)

	// newLockFileWithPath should recover from stale lock
	lock, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)
	require.NotNil(t, lock)
	defer func() { _ = lock.Release() }()

	// Verify our PID is now in the lock file
	pid, err := ReadPID(lockPath)
	require.NoError(t, err)
	assert.Equal(t, os.Getpid(), pid)
}

func TestStaleLockWithInvalidPID(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	// Create a lock file with invalid content
	err := os.WriteFile(lockPath, []byte("not-a-pid\n"), 0o600)
	require.NoError(t, err)

	// newLockFileWithPath should recover from invalid lock
	lock, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)
	require.NotNil(t, lock)
	defer func() { _ = lock.Release() }()
}

func TestReadPID(t *testing.T) {
	tmpDir := t.TempDir()

	cases := []struct {
		name        string
		content     string
		expectedPID int
		expectError bool
	}{
		{
			name:        "valid PID",
			content:     "12345\n",
			expectedPID: 12345,
			expectError: false,
		},
		{
			name:        "PID without newline",
			content:     "12345",
			expectedPID: 12345,
			expectError: false,
		},
		{
			name:        "PID with whitespace",
			content:     "  12345  \n",
			expectedPID: 12345,
			expectError: false,
		},
		{
			name:        "invalid PID",
			content:     "not-a-number",
			expectedPID: 0,
			expectError: true,
		},
		{
			name:        "empty file",
			content:     "",
			expectedPID: 0,
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lockPath := filepath.Join(tmpDir, tc.name+".lock")
			err := os.WriteFile(lockPath, []byte(tc.content), 0o600)
			require.NoError(t, err)

			pid, err := ReadPID(lockPath)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPID, pid)
			}
		})
	}
}

func TestReadPIDNonExistentFile(t *testing.T) {
	_, err := ReadPID("/non/existent/path/daemon.lock")
	assert.Error(t, err)
}

func TestWritePID(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	// Create lock file manually
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	require.NoError(t, err)

	lf := &LockFile{
		path: lockPath,
	}
	defer func() { _ = lf.Release() }()

	// Write PID
	err = lf.WritePID(file)
	require.NoError(t, err)
	_ = file.Close()

	// Read back and verify
	pid, err := ReadPID(lockPath)
	require.NoError(t, err)
	assert.Equal(t, os.Getpid(), pid)
}

func TestIsProcessRunning(t *testing.T) {
	// Current process should be running
	assert.True(t, IsProcessRunning(os.Getpid()))

	// Non-existent process should not be running
	// Using a very high PID that's unlikely to exist
	assert.False(t, IsProcessRunning(999999))

	// PID 0 is special and should return false for our purposes
	assert.False(t, IsProcessRunning(0))
}

func TestLockFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	lock, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)
	defer func() { _ = lock.Release() }()

	// Verify permissions are 0600 (owner read/write only)
	// Skip on Windows - it doesn't support Unix-style file permissions
	if runtime.GOOS != windowsOS {
		info, err := os.Stat(lockPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestLockFileMultipleRelease(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "daemon.lock")

	lock, err := newLockFileWithPath(lockPath)
	require.NoError(t, err)

	// First release should succeed
	err = lock.Release()
	assert.NoError(t, err)

	// Second release should fail (file already removed)
	err = lock.Release()
	assert.Error(t, err)
}

// TestNewLockFileIntegration tests the actual NewLockFile function
// This test depends on XDG_STATE_HOME being set correctly.
func TestNewLockFileIntegration(t *testing.T) {
	// Skip if running as part of a larger test suite
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Set up temp directory as state path
	tmpDir := t.TempDir()
	setTestEnv(t, tmpDir)

	// Create lock file using the real NewLockFile function
	lock, err := NewLockFile()
	require.NoError(t, err)
	require.NotNil(t, lock)
	defer func() { _ = lock.Release() }()

	// Verify lock file was created in the expected location
	expectedPath := filepath.Join(tmpDir, "oh-my-posh", "daemon.lock")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err, "lock file should exist at XDG_STATE_HOME/oh-my-posh/daemon.lock")
}
