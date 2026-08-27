package cache

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/maps"

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
				result := Session.Print()
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
				result := Session.Print()
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
				result := Session.Print()
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

	Device.Set("test_key", storeGobPointerType{Name: "value"}, Duration("1h"))
	Device.close()

	// Second process: fresh store, loaded from disk. This is the gob decode
	// that flips the interface-typed value from T to *T.
	device = nil
	Device.init(fileName, true)

	got, ok := Device.Get[storeGobPointerType]("test_key")
	require.True(t, ok, "expected a cache hit after a gob round-trip of a pointer-registered type")
	assert.Equal(t, "value", got.Name)
}

// Guards against the #7758 class of bug: a long-lived process (serve) that
// loaded the session cache once at startup must still see a write from a
// separate one-shot process (e.g. `oh-my-posh toggle`) that landed on disk
// afterwards, without losing values it has itself set more recently than
// what's on disk (e.g. prompt_count_cache).
func TestRefreshMergesExternalWriteByTimestamp(t *testing.T) {
	origSession := session
	t.Cleanup(func() { session = origSession })

	filePath := filepath.Join(t.TempDir(), "session.cache")

	// The daemon's in-memory state: it has its own, newer value for a key
	// the external writer also touches, and predates the external write.
	daemon := Session.new()
	daemon.filePath = filePath
	daemon.persist = true
	daemon.mtime = time.Now().Add(-time.Hour)
	daemon.cache.Set("shared_key", &Entry[any]{
		Value:     "daemon-value",
		Timestamp: time.Now().Unix(),
		TTL:       -1,
	})

	// Simulate `oh-my-posh toggle` running as a separate one-shot process:
	// it starts from a stale copy of shared_key and writes a brand new key.
	writer := Session.new()
	writer.filePath = filePath
	writer.persist = true
	writer.dirty = true
	writer.cache.Set("shared_key", &Entry[any]{
		Value:     "stale-external-copy",
		Timestamp: time.Now().Unix() - 100,
		TTL:       -1,
	})
	writer.cache.Set(TOGGLECACHE, &Entry[any]{
		Value:     map[string]bool{"shell": true},
		Timestamp: time.Now().Unix(),
		TTL:       -1,
	})
	session = writer
	Session.close()

	// Back to the daemon's perspective: refresh should pick up the new
	// toggle_cache key from disk...
	session = daemon
	Session.Refresh()

	toggled, found := Session.Get[map[string]bool](TOGGLECACHE)
	require.True(t, found, "toggle_cache written externally should be visible after Refresh")
	assert.True(t, toggled["shell"])

	// ...without clobbering the daemon's own newer value for a key both
	// sides touched.
	shared, found := Session.Get[string]("shared_key")
	require.True(t, found)
	assert.Equal(t, "daemon-value", shared, "a newer in-memory value must win over an older on-disk one")
}

// Guards against the daemon's own shutdown flush clobbering a write that
// landed after its last Refresh but before it exits.
func TestCloseRefreshesBeforePersistingToAvoidClobber(t *testing.T) {
	origSession := session
	t.Cleanup(func() { session = origSession })

	filePath := filepath.Join(t.TempDir(), "session.cache")

	daemon := Session.new()
	daemon.filePath = filePath
	daemon.persist = true
	daemon.dirty = true
	daemon.mtime = time.Now().Add(-time.Hour)
	daemon.cache.Set("prompt_count_cache", &Entry[any]{
		Value:     3,
		Timestamp: time.Now().Unix(),
		TTL:       -1,
	})

	// A toggle write lands on disk after the daemon's last refresh, right
	// before the shell (and daemon) exit.
	writer := Session.new()
	writer.filePath = filePath
	writer.persist = true
	writer.dirty = true
	writer.cache.Set(TOGGLECACHE, &Entry[any]{
		Value:     map[string]bool{"shell": true},
		Timestamp: time.Now().Unix(),
		TTL:       -1,
	})
	session = writer
	Session.close()

	// The daemon now shuts down. Its close() must merge the external write
	// in before overwriting the file, not blindly persist its stale copy.
	session = daemon
	Session.close()

	reader, err := openFile(filePath)
	require.NoError(t, err)
	defer reader.Close()

	var onDisk maps.Simple[*Entry[any]]
	require.NoError(t, gob.NewDecoder(reader).Decode(&onDisk))

	session = &store{cache: onDisk.ToConcurrent()}

	toggled, found := Session.Get[map[string]bool](TOGGLECACHE)
	require.True(t, found, "the external write must survive the daemon's own shutdown flush")
	assert.True(t, toggled["shell"])

	count, found := Session.Get[int]("prompt_count_cache")
	require.True(t, found)
	assert.Equal(t, 3, count)
}

// Guards against a variant of #7758: the serve daemon's Device store (which
// holds the RELOAD flag config.Get checks in prompt.New to bypass its own
// config cache) must also see a write from a separate `oh-my-posh enable
// reload` process, not just the Session store.
func TestRefreshPicksUpDeviceStoreWrite(t *testing.T) {
	origDevice := device
	t.Cleanup(func() { device = origDevice })

	filePath := filepath.Join(t.TempDir(), "omp.cache")

	daemon := Device.new()
	daemon.filePath = filePath
	daemon.persist = true
	daemon.mtime = time.Now().Add(-time.Hour)
	device = daemon

	writer := Device.new()
	writer.filePath = filePath
	writer.persist = true
	writer.dirty = true
	writer.cache.Set("reload", &Entry[any]{
		Value:     true,
		Timestamp: time.Now().Unix(),
		TTL:       -1,
	})
	device = writer
	Device.close()

	device = daemon
	Device.Refresh()

	reload, found := Device.Get[bool]("reload")
	require.True(t, found, "`enable reload` written externally should be visible after Refresh")
	assert.True(t, reload)
}

// Device's close() must bump the file's mtime as reliably as Session's does
// (the mmap-backed Windows write path doesn't do this on its own) - without
// it, Refresh would never notice an `enable`/`disable` write on Windows.
func TestStoreCloseTouchesDeviceFileMTime(t *testing.T) {
	origDevice := device
	t.Cleanup(func() { device = origDevice })

	filePath := filepath.Join(t.TempDir(), "omp.cache")

	testStore := Device.new()
	testStore.filePath = filePath
	testStore.persist = true
	testStore.dirty = true
	testStore.cache.Set("reload", &Entry[any]{
		Value:     true,
		Timestamp: time.Now().Unix(),
		TTL:       -1,
	})
	device = testStore

	before := time.Now()
	Device.close()

	info, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.False(t, info.ModTime().Before(before), "mtime should be bumped to the close time, not left stale")
}
