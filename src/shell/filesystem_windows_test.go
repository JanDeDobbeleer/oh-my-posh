package shell

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openWithoutShareDelete opens path the way the C runtime's fopen does:
// shared for read and write, but not for deletion. This is how Lua (clink),
// PowerShell and most other script hosts hold the init script while sourcing
// it, and it makes replacing the file via rename fail with
// ERROR_ACCESS_DENIED.
func openWithoutShareDelete(t *testing.T, path string) {
	t.Helper()

	p, err := syscall.UTF16PtrFromString(path)
	require.NoError(t, err)

	h, err := syscall.CreateFile(p,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = syscall.CloseHandle(h)
	})
}

func TestWriteFileTargetHeldOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "init.123.lua")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))

	openWithoutShareDelete(t, path)

	err := writeFile(path, []byte("new"), 0o644)
	require.NoError(t, err)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "new", string(got))
}

func TestWriteFileAtomicTargetHeldOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "init.123.lua")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))

	openWithoutShareDelete(t, path)

	err := writeFileAtomic(path, []byte("new"), 0o644)
	assert.True(t, canRetryWrite(err), "expected a retryable error, got %v", err)
}
