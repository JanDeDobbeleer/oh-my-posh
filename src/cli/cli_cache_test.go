package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLICacheRelated_Combined tests all CLI commands that use the cache package.
// We combine them into a single test function because the 'cache' package
// uses global variables (cachePath, sessionID) that are initialized once
// per process, making independent tests with different environments difficult.
//
// This includes tests for: toggle, cache clear, cache ttl, cache show, cache path
func TestCLICacheRelated_Combined(t *testing.T) {
	// Setup shared environment for ALL cache-related tests
	tmpDir, err := os.MkdirTemp("", "omp-test-cli-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Ensure we are in CLI mode (no socket)
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(tmpDir, "runtime"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmpDir, "state"))
	_ = os.MkdirAll(filepath.Join(tmpDir, "runtime"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "state"), 0755)

	t.Setenv("OMP_CACHE_DIR", tmpDir)
	t.Setenv("POSH_SHELL", "test-shell")
	t.Setenv("POSH_SESSION_ID", "test-session-combined")

	// Helper to discard output
	discardOutput := func() func() {
		oldOut := RootCmd.OutOrStdout()
		oldErr := RootCmd.ErrOrStderr()
		RootCmd.SetOut(io.Discard)
		RootCmd.SetErr(io.Discard)
		return func() {
			RootCmd.SetOut(oldOut)
			RootCmd.SetErr(oldErr)
		}
	}

	// ========== TOGGLE TESTS ==========

	t.Run("Toggle_EmptyArgs", func(t *testing.T) {
		restore := discardOutput()
		defer restore()

		RootCmd.SetArgs([]string{"toggle"})
		err := RootCmd.Execute()
		// Expect error due to missing args
		assert.Error(t, err)
	})

	t.Run("Toggle_SingleToggle", func(t *testing.T) {
		RootCmd.SetArgs([]string{"toggle", "git", "node"})
		err := RootCmd.Execute()
		require.NoError(t, err)

		// Verify
		cache.Init("test-shell", cache.Persist)
		toggles, ok := cache.Get[map[string]bool](cache.Session, cache.TOGGLECACHE)
		cache.Close()

		require.True(t, ok, "toggle_cache should exist")
		assert.True(t, toggles["git"])
		assert.True(t, toggles["node"])
	})

	t.Run("Toggle_ToggleOff", func(t *testing.T) {
		RootCmd.SetArgs([]string{"toggle", "git"})
		err := RootCmd.Execute()
		require.NoError(t, err)

		// Verify
		cache.Init("test-shell", cache.Persist)
		toggles, ok := cache.Get[map[string]bool](cache.Session, cache.TOGGLECACHE)
		cache.Close()

		require.True(t, ok)
		assert.False(t, toggles["git"], "git should be untoggled")
		assert.True(t, toggles["node"], "node should remain toggled")
	})

	t.Run("Toggle_MultipleInvocations", func(t *testing.T) {
		// Toggle "aws"
		RootCmd.SetArgs([]string{"toggle", "aws"})
		err := RootCmd.Execute()
		require.NoError(t, err)

		// Toggle "azure"
		RootCmd.SetArgs([]string{"toggle", "azure"})
		err = RootCmd.Execute()
		require.NoError(t, err)

		// Verify all accumulated state (node, aws, azure should be on; git off)
		cache.Init("test-shell", cache.Persist)
		toggles, ok := cache.Get[map[string]bool](cache.Session, cache.TOGGLECACHE)
		cache.Close()

		require.True(t, ok)
		assert.True(t, toggles["aws"])
		assert.True(t, toggles["azure"])
		assert.True(t, toggles["node"])
		assert.False(t, toggles["git"])
	})

	// ========== CACHE COMMAND TESTS ==========

	t.Run("Cache_Path", func(t *testing.T) {
		RootCmd.SetArgs([]string{"cache", "path"})
		err := RootCmd.Execute()
		require.NoError(t, err)
	})

	t.Run("Cache_Clear_CLIMode", func(t *testing.T) {
		// First set some cache data
		cache.Init("test-shell", cache.Persist)
		cache.Set(cache.Device, "test_key_clear", "test_value", cache.INFINITE)
		cache.Close()

		RootCmd.SetArgs([]string{"cache", "clear"})
		err := RootCmd.Execute()
		require.NoError(t, err)

		// Verify cache was cleared
		cache.Init("test-shell")
		_, ok := cache.Get[string](cache.Device, "test_key_clear")
		cache.Close()
		assert.False(t, ok, "Cache should be cleared")

		// Recreate the cache directory for subsequent tests
		// (cache.Clear with force=true removes the entire directory)
		_ = os.MkdirAll(filepath.Join(tmpDir, "oh-my-posh"), 0755)
	})

	t.Run("Cache_TTL_Get_NoDaemon", func(t *testing.T) {
		RootCmd.SetArgs([]string{"cache", "ttl"})
		err := RootCmd.Execute()
		// Should succeed and show default TTL
		require.NoError(t, err)
	})

	t.Run("Cache_TTL_Set_CLIMode", func(t *testing.T) {
		RootCmd.SetArgs([]string{"cache", "ttl", "14"})
		err := RootCmd.Execute()
		require.NoError(t, err)

		// Verify TTL was set in cache (need Persist flag to read from disk)
		cache.Init("test-shell", cache.Persist)
		ttl, ok := cache.Get[int](cache.Device, cache.TTL)
		cache.Close()
		assert.True(t, ok, "TTL should be set")
		assert.Equal(t, 14, ttl)
	})

	t.Run("Cache_Show", func(t *testing.T) {
		// Set some data first
		cache.Init("test-shell", cache.Persist)
		cache.Set(cache.Device, "display_key", "display_value", cache.INFINITE)
		cache.Close()

		RootCmd.SetArgs([]string{"cache", "show"})
		err := RootCmd.Execute()
		require.NoError(t, err)
	})

	// ========== DAEMON HELPER TESTS ==========

	t.Run("DaemonHelpers_NoDaemon", func(t *testing.T) {
		// clearDaemonCache should return false when daemon not running
		result := clearDaemonCache()
		assert.False(t, result, "clearDaemonCache should return false when daemon is not running")

		// getDaemonTTL should return false when daemon not running
		days, ok := getDaemonTTL()
		assert.False(t, ok, "getDaemonTTL should return false when daemon is not running")
		assert.Equal(t, 0, days)

		// setDaemonTTL should return false when daemon not running
		result = setDaemonTTL(14)
		assert.False(t, result, "setDaemonTTL should return false when daemon is not running")
	})
}
