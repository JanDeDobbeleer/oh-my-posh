package daemon

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinaryWatcher_ReplaceTriggersCallback(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "oh-my-posh")

	// Create initial binary
	require.NoError(t, os.WriteFile(binPath, []byte("v1"), 0o755))

	var called atomic.Bool
	bw, err := NewBinaryWatcher(binPath, func() { called.Store(true) })
	require.NoError(t, err)
	defer bw.Close()

	// Replace binary (atomic: write new, rename over)
	tmpPath := binPath + ".tmp"
	require.NoError(t, os.WriteFile(tmpPath, []byte("v2"), 0o755))
	require.NoError(t, os.Rename(tmpPath, binPath))

	// Wait for debounce (1s) + margin
	assert.Eventually(t, called.Load, 3*time.Second, 100*time.Millisecond,
		"callback should fire after binary replacement")
}

func TestBinaryWatcher_UnrelatedFileNoTrigger(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "oh-my-posh")

	require.NoError(t, os.WriteFile(binPath, []byte("v1"), 0o755))

	var called atomic.Bool
	bw, err := NewBinaryWatcher(binPath, func() { called.Store(true) })
	require.NoError(t, err)
	defer bw.Close()

	// Modify an unrelated file in the same directory
	require.NoError(t, os.WriteFile(filepath.Join(dir, "unrelated.txt"), []byte("hello"), 0o644))

	// Should not trigger within 2s
	time.Sleep(2 * time.Second)
	assert.False(t, called.Load(), "callback should not fire for unrelated file changes")
}

func TestBinaryWatcher_SymlinkTargetReplacement(t *testing.T) {
	dir := t.TempDir()
	realBin := filepath.Join(dir, "oh-my-posh-real")
	symlink := filepath.Join(dir, "oh-my-posh")

	require.NoError(t, os.WriteFile(realBin, []byte("v1"), 0o755))
	require.NoError(t, os.Symlink(realBin, symlink))

	var called atomic.Bool
	bw, err := NewBinaryWatcher(symlink, func() { called.Store(true) })
	require.NoError(t, err)
	defer bw.Close()

	// Replace the real binary (symlink target)
	tmpPath := realBin + ".tmp"
	require.NoError(t, os.WriteFile(tmpPath, []byte("v2"), 0o755))
	require.NoError(t, os.Rename(tmpPath, realBin))

	assert.Eventually(t, called.Load, 3*time.Second, 100*time.Millisecond,
		"callback should fire when symlink target is replaced")
}
