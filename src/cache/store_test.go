package cache

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	cases := []struct {
		setupFunc func() *store
		testFunc  func(t *testing.T)
		name      string
	}{
		{
			name: "Print store with data",
			setupFunc: func() *store {
				testStore := Session.new()
				testStore.cache.Set("test_key1", &Entry[any]{
					Value:     "test_value1",
					Timestamp: time.Now().Unix(),
					TTL:       3600, // 1 hour
				})
				testStore.cache.Set("test_key2", &Entry[any]{
					Value:     42,
					Timestamp: time.Now().Unix(),
					TTL:       -1, // never expires
				})
				testStore.cache.Set("expired_key", &Entry[any]{
					Value:     "expired_value",
					Timestamp: time.Now().Unix() - 7200, // 2 hours ago
					TTL:       3600,                     // 1 hour (should be expired)
				})
				session = testStore
				return testStore
			},
			testFunc: func(t *testing.T) {
				result := Print(Session)
				assert.Contains(t, result, "Key: test_key1")
				assert.Contains(t, result, `Value: "test_value1"`) // Note: quotes are included in output
				assert.Contains(t, result, "Type: string")
				assert.Contains(t, result, "Key: test_key2")
				assert.Contains(t, result, "Value: 42")
				assert.Contains(t, result, "Type: int")
				assert.Contains(t, result, "Key: expired_key [EXPIRED]")
				assert.Contains(t, result, "never expires")
				assert.Contains(t, result, "expires at")

				// Verify structure
				lines := strings.Split(result, "\n")
				assert.True(t, len(lines) > 10, "Output should have multiple lines")
			},
		},
		{
			name: "Print empty store",
			setupFunc: func() *store {
				testStore := Session.new()
				session = testStore
				return testStore
			},
			testFunc: func(t *testing.T) {
				result := Print(Session)
				assert.Contains(t, result, "Store session is empty")
			},
		},
		{
			name: "Print nil store check",
			setupFunc: func() *store {
				testStore := Session.new()
				session = testStore
				return testStore
			},
			testFunc: func(t *testing.T) {
				// Since get() always creates a store, we test empty store behavior
				result := Print(Session)
				assert.Contains(t, result, "Store session is empty")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFunc()
			tc.testFunc(t)
		})
	}
}

// Guards against #7340: on Windows the mmap-backed write path doesn't
// reliably update the file's last-write-time on its own, which could cause
// an actively-used session cache to look stale and get swept up by
// cache.Clear().
func TestStoreCloseTouchesSessionFileMTime(t *testing.T) {
	origSession := session
	t.Cleanup(func() { session = origSession })

	filePath := filepath.Join(t.TempDir(), "session.cache")

	testStore := Session.new()
	testStore.filePath = filePath
	testStore.persist = true
	testStore.dirty = true
	testStore.cache.Set("test_key", &Entry[any]{
		Value:     "test_value",
		Timestamp: time.Now().Unix(),
		TTL:       3600,
	})
	session = testStore

	before := time.Now()
	Session.close()

	info, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.False(t, info.ModTime().Before(before), "mtime should be bumped to the close time, not left stale")
	assert.WithinDuration(t, time.Now(), info.ModTime(), time.Minute)
}

type storeGobPointerType struct {
	Name string
}

// Guards against #7766: gob.Register keys its decode-side type map on the
// exact type registered, while encoding an interface value uses the base
// type. Registering a type as a pointer (as segment_registry.go does for
// segments.Version, ClaudeData and CopilotCLIData) means a value stored
// under that type comes back as *T, not T, on the next process's gob decode
// of the on-disk device cache. Get[T] must still return a hit in that case,
// otherwise cache_duration silently never survives a process restart.
func TestGetSurvivesGobPointerRegistration(t *testing.T) {
	gob.Register(&storeGobPointerType{})

	origDevice := device
	origCachePath := cachePath
	t.Cleanup(func() {
		device = origDevice
		cachePath = origCachePath
	})

	cachePath = t.TempDir()
	const fileName = "device.cache"

	// First process: set the value and persist it to disk.
	writer := Device.new()
	writer.filePath = filepath.Join(cachePath, fileName)
	writer.persist = true
	device = writer

	Set(Device, "test_key", storeGobPointerType{Name: "value"}, Duration("1h"))
	Device.close()

	// Second process: fresh store, loaded from disk. This is the gob decode
	// that flips the interface-typed value from T to *T.
	device = nil
	Device.init(fileName, true)

	got, ok := Get[storeGobPointerType](Device, "test_key")
	require.True(t, ok, "expected a cache hit after a gob round-trip of a pointer-registered type")
	assert.Equal(t, "value", got.Name)
}
