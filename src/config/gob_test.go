package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"

	"github.com/stretchr/testify/assert"
)

func writeTestConfig(t *testing.T) string {
	t.Helper()

	file := filepath.Join(t.TempDir(), "theme.omp.json")
	if err := os.WriteFile(file, []byte(`{"version": 3, "final_space": true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	return file
}

func TestGetRecoversFromDSCWhenSessionCacheLost(t *testing.T) {
	defer func() {
		cache.DeleteAll(cache.Device)
		cache.DeleteAll(cache.Session)
	}()

	file := writeTestConfig(t)

	// Simulate a prior init: the source is tracked in the device-level DSC, which
	// survives the loss of the session cache.
	resource := DSC()
	resource.Add(file)
	resource.Save()

	// A prompt render: no --config, and the session cache no longer holds the config.
	cfg := Get("", false)

	assert.NotNil(t, cfg)
	assert.Equal(t, file, cfg.Source, "should recover the configured theme instead of the default")

	// Recovery should self-heal the session cache so subsequent renders are fast.
	_, found := cache.Get[string](cache.Session, configKey)
	assert.True(t, found, "recovery should repopulate the session config cache")
}

func TestGetRecoversFromSessionSourceWhenConfigBlobMissing(t *testing.T) {
	defer func() {
		cache.DeleteAll(cache.Device)
		cache.DeleteAll(cache.Session)
	}()

	file := writeTestConfig(t)

	// The session source survives even when the config blob is corrupt or absent.
	cache.Set(cache.Session, SourceKey, file, cache.INFINITE)

	cfg := Get("", false)

	assert.NotNil(t, cfg)
	assert.Equal(t, file, cfg.Source)
}

func TestGetFallsBackToDefaultWithoutSource(t *testing.T) {
	defer func() {
		cache.DeleteAll(cache.Device)
		cache.DeleteAll(cache.Session)
	}()

	// No --config, no session cache, and no DSC: there is nothing to recover.
	cfg := Get("", false)

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Source, "should fall back to the default built-in config")
}
