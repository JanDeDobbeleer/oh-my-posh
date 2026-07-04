//go:build !windows

package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenFileForWriteAtomicShrink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "omp.cache")

	// Simulate a large pre-existing cache file.
	assert.NoError(t, os.WriteFile(path, []byte("this is a much longer previous payload"), 0o644))

	writer, err := openFileForWrite(path)
	assert.NoError(t, err)

	_, err = writer.Write([]byte("short"))
	assert.NoError(t, err)

	assert.NoError(t, writer.Close())

	data, err := os.ReadFile(path)
	assert.NoError(t, err)

	// No stale trailing bytes from the longer previous payload.
	assert.Equal(t, "short", string(data))
}

func TestOpenFileForWritePreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "omp.cache")

	assert.NoError(t, os.WriteFile(path, []byte("initial"), 0o600))

	writer, err := openFileForWrite(path)
	assert.NoError(t, err)

	_, err = writer.Write([]byte("updated"))
	assert.NoError(t, err)

	assert.NoError(t, writer.Close())

	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestOpenFileForWriteNewFileDefaultPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.cache")

	writer, err := openFileForWrite(path)
	assert.NoError(t, err)

	_, err = writer.Write([]byte("data"))
	assert.NoError(t, err)

	assert.NoError(t, writer.Close())

	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

func TestOpenFileForWriteNoTmpFileLeftBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "omp.cache")

	writer, err := openFileForWrite(path)
	assert.NoError(t, err)

	_, err = writer.Write([]byte("data"))
	assert.NoError(t, err)

	assert.NoError(t, writer.Close())

	matches, err := filepath.Glob(filepath.Join(dir, ".cache-*.tmp"))
	assert.NoError(t, err)
	assert.Empty(t, matches, "temp file should be renamed away, not left behind")
}
