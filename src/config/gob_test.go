package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"

	"github.com/stretchr/testify/assert"
)

func writeTestConfig(t *testing.T, name string) string {
	t.Helper()

	file := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(file, []byte(`{"version": 3, "final_space": true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	return file
}

func resetCache(t *testing.T) {
	t.Helper()

	// pin the config source resolution: with OSTYPE set (as in Git Bash),
	// Windows keeps forward-slash paths instead of resolving native ones
	t.Setenv("OSTYPE", "")

	reset := func() {
		cache.DeleteAll(cache.Session)
		cache.DeleteAll(cache.Device)
	}

	reset()
	t.Cleanup(reset)
}

func TestGetRecoversFromEnvironmentWhenSessionCacheLost(t *testing.T) {
	resetCache(t)

	file := writeTestConfig(t, "theme.omp.json")

	// Simulate a prompt render after the session cache is lost: init pinned the
	// source in POSH_CONFIG and the root command resolved it into the config flag.
	t.Setenv(envKey, file)

	cfg := Get(file, false)

	assert.NotNil(t, cfg)
	assert.Equal(t, file, cfg.Source, "should render the configured theme instead of the default")

	// Recovery should self-heal the session cache so subsequent renders are fast.
	_, found := cache.Get[string](cache.Session, configKey)
	assert.True(t, found, "recovery should repopulate the session config cache")
}

func TestGetPrefersSessionCacheOverEnvironment(t *testing.T) {
	resetCache(t)

	cached := writeTestConfig(t, "cached.omp.json")
	other := writeTestConfig(t, "other.omp.json")

	Load(cached).Store()

	// POSH_CONFIG can not change the configuration mid-session:
	// a healthy session cache always wins.
	t.Setenv(envKey, other)

	cfg := Get(other, false)

	assert.NotNil(t, cfg)
	assert.Equal(t, cached, cfg.Source)
}

func TestGetFallsBackToDefaultWithoutRecoverySource(t *testing.T) {
	resetCache(t)

	// No --config, no session cache, and no environment: there is nothing to recover.
	t.Setenv(envKey, "")

	cfg := Get("", false)

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Source, "should fall back to the default built-in config")
}

func TestGetDropsCorruptedSessionCacheEntry(t *testing.T) {
	resetCache(t)

	// valid base64, but not a gob-encoded Config
	cache.Set(cache.Session, configKey, "Y29ycnVwdGVkIGVudHJ5", cache.INFINITE)

	file := filepath.Join(t.TempDir(), "missing.omp.json")
	t.Setenv(envKey, file)

	cfg := Get(file, false)

	assert.NotNil(t, cfg)

	_, found := cache.Get[string](cache.Session, configKey)
	assert.False(t, found, "a corrupted session config entry should be dropped after a failed restore")
}

func TestGetFallsBackToDefaultWhenRecoverySourceInvalid(t *testing.T) {
	resetCache(t)

	file := filepath.Join(t.TempDir(), "missing.omp.json")
	t.Setenv(envKey, file)

	cfg := Get(file, false)

	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Source, "should fall back to the default built-in config")

	// The source may only be temporarily unavailable: the default must not be
	// cached, otherwise the next render can no longer retry the recovery.
	_, found := cache.Get[string](cache.Session, configKey)
	assert.False(t, found, "a failed recovery should not pin the default in the session cache")
}
