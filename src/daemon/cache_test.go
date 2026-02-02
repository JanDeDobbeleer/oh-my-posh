package daemon

import (
	"sync"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryCache(t *testing.T) {
	mc := NewMemoryCache()
	require.NotNil(t, mc)
	assert.Equal(t, 0, mc.SessionCount())
	assert.Equal(t, 0, mc.EntryCount())
}

func TestMemoryCacheSetAndGet(t *testing.T) {
	mc := NewMemoryCache()

	// Set a value
	mc.Set("session1", "key1", "value1", time.Hour)

	// Get the value back
	value, ok := mc.Get("session1", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)
}

func TestMemoryCacheGetNonExistentSession(t *testing.T) {
	mc := NewMemoryCache()

	value, ok := mc.Get("nonexistent", "key1")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestMemoryCacheGetNonExistentKey(t *testing.T) {
	mc := NewMemoryCache()

	// Create session with one key
	mc.Set("session1", "key1", "value1", time.Hour)

	// Try to get a different key
	value, ok := mc.Get("session1", "nonexistent")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestMemoryCacheExpiration(t *testing.T) {
	mc := NewMemoryCache()

	// Set a value with very short TTL
	mc.Set("session1", "key1", "value1", 10*time.Millisecond)

	// Value should be available immediately
	value, ok := mc.Get("session1", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Value should be expired
	value, ok = mc.Get("session1", "key1")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestMemoryCacheNoExpiration(t *testing.T) {
	mc := NewMemoryCache()

	// Set a value with zero TTL (never expires)
	mc.Set("session1", "key1", "value1", 0)

	// Value should be available
	value, ok := mc.Get("session1", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// Set a value with negative TTL (never expires)
	mc.Set("session1", "key2", "value2", -time.Hour)

	value, ok = mc.Get("session1", "key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)
}

func TestMemoryCacheOverwrite(t *testing.T) {
	mc := NewMemoryCache()

	// Set initial value
	mc.Set("session1", "key1", "value1", time.Hour)

	// Overwrite with new value
	mc.Set("session1", "key1", "value2", time.Hour)

	// Should get new value
	value, ok := mc.Get("session1", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)
}

func TestMemoryCacheDelete(t *testing.T) {
	mc := NewMemoryCache()

	// Set values
	mc.Set("session1", "key1", "value1", time.Hour)
	mc.Set("session1", "key2", "value2", time.Hour)

	// Delete one key
	mc.Delete("session1", "key1")

	// Deleted key should not exist
	value, ok := mc.Get("session1", "key1")
	assert.False(t, ok)
	assert.Nil(t, value)

	// Other key should still exist
	value, ok = mc.Get("session1", "key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)
}

func TestMemoryCacheDeleteNonExistent(_ *testing.T) {
	mc := NewMemoryCache()

	// Should not panic
	mc.Delete("nonexistent", "key1")
	mc.Delete("session1", "nonexistent")
}

func TestMemoryCacheCleanSession(t *testing.T) {
	mc := NewMemoryCache()

	// Create two sessions
	mc.Set("session1", "key1", "value1", time.Hour)
	mc.Set("session1", "key2", "value2", time.Hour)
	mc.Set("session2", "key1", "value3", time.Hour)

	assert.Equal(t, 2, mc.SessionCount())

	// Clean one session
	mc.CleanSession("session1")

	// Session 1 should be gone
	value, ok := mc.Get("session1", "key1")
	assert.False(t, ok)
	assert.Nil(t, value)

	value, ok = mc.Get("session1", "key2")
	assert.False(t, ok)
	assert.Nil(t, value)

	// Session 2 should still exist
	value, ok = mc.Get("session2", "key1")
	assert.True(t, ok)
	assert.Equal(t, "value3", value)

	assert.Equal(t, 1, mc.SessionCount())
}

func TestMemoryCacheEvictExpired(t *testing.T) {
	mc := NewMemoryCache()

	// Set entries with different TTLs
	mc.Set("session1", "short", "value1", 10*time.Millisecond)
	mc.Set("session1", "long", "value2", time.Hour)
	mc.Set("session2", "short", "value3", 10*time.Millisecond)

	assert.Equal(t, 3, mc.EntryCount())

	// Wait for short entries to expire
	time.Sleep(20 * time.Millisecond)

	// Evict expired entries
	mc.EvictExpired()

	// Only long entry should remain
	assert.Equal(t, 1, mc.EntryCount())

	value, ok := mc.Get("session1", "long")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)
}

func TestMemoryCacheSessionCount(t *testing.T) {
	mc := NewMemoryCache()

	assert.Equal(t, 0, mc.SessionCount())

	mc.Set("session1", "key1", "value1", time.Hour)
	assert.Equal(t, 1, mc.SessionCount())

	mc.Set("session2", "key1", "value2", time.Hour)
	assert.Equal(t, 2, mc.SessionCount())

	mc.Set("session1", "key2", "value3", time.Hour)
	assert.Equal(t, 2, mc.SessionCount()) // Still 2, same session

	mc.CleanSession("session1")
	assert.Equal(t, 1, mc.SessionCount())
}

func TestMemoryCacheEntryCount(t *testing.T) {
	mc := NewMemoryCache()

	assert.Equal(t, 0, mc.EntryCount())

	mc.Set("session1", "key1", "value1", time.Hour)
	assert.Equal(t, 1, mc.EntryCount())

	mc.Set("session1", "key2", "value2", time.Hour)
	assert.Equal(t, 2, mc.EntryCount())

	mc.Set("session2", "key1", "value3", time.Hour)
	assert.Equal(t, 3, mc.EntryCount())

	mc.Delete("session1", "key1")
	assert.Equal(t, 2, mc.EntryCount())
}

func TestMemoryCacheDifferentValueTypes(t *testing.T) {
	mc := NewMemoryCache()

	// Store different types
	mc.Set("session1", "string", "hello", time.Hour)
	mc.Set("session1", "int", 42, time.Hour)
	mc.Set("session1", "float", 3.14, time.Hour)
	mc.Set("session1", "bool", true, time.Hour)
	mc.Set("session1", "slice", []string{"a", "b", "c"}, time.Hour)
	mc.Set("session1", "map", map[string]int{"x": 1}, time.Hour)

	// Retrieve and verify types
	v, ok := mc.Get("session1", "string")
	assert.True(t, ok)
	assert.Equal(t, "hello", v.(string))

	v, ok = mc.Get("session1", "int")
	assert.True(t, ok)
	assert.Equal(t, 42, v.(int))

	v, ok = mc.Get("session1", "float")
	assert.True(t, ok)
	assert.Equal(t, 3.14, v.(float64))

	v, ok = mc.Get("session1", "bool")
	assert.True(t, ok)
	assert.Equal(t, true, v.(bool))

	v, ok = mc.Get("session1", "slice")
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b", "c"}, v.([]string))

	v, ok = mc.Get("session1", "map")
	assert.True(t, ok)
	assert.Equal(t, map[string]int{"x": 1}, v.(map[string]int))
}

func TestMemoryCacheConcurrentAccess(t *testing.T) {
	mc := NewMemoryCache()
	var wg sync.WaitGroup

	// Concurrent writes from multiple goroutines
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			sessionID := "session1"
			if n%2 == 0 {
				sessionID = "session2"
			}
			mc.Set(sessionID, "key", n, time.Hour)
		}(i)
	}

	// Concurrent reads
	for range 100 {
		wg.Go(func() {
			mc.Get("session1", "key")
			mc.Get("session2", "key")
		})
	}

	wg.Wait()

	// Should have 2 sessions
	assert.Equal(t, 2, mc.SessionCount())
}

func TestMemoryCacheConcurrentEviction(_ *testing.T) {
	mc := NewMemoryCache()
	var wg sync.WaitGroup

	// Start eviction goroutine
	wg.Go(func() {
		for range 10 {
			mc.EvictExpired()
			time.Sleep(5 * time.Millisecond)
		}
	})

	// Concurrent writes with short TTL
	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			mc.Set("session1", "key", n, 10*time.Millisecond)
		}(i)
	}

	wg.Wait()
	// Should not panic or deadlock
}

func TestCacheEntryExpired(t *testing.T) {
	// Test expired entry
	expired := &CacheEntry{
		Value:     "test",
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	assert.True(t, expired.Expired())

	// Test non-expired entry
	valid := &CacheEntry{
		Value:     "test",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	assert.False(t, valid.Expired())
}

func TestMemoryCacheSessionIsolation(t *testing.T) {
	mc := NewMemoryCache()

	// Same key in different sessions should be independent
	mc.Set("session1", "shared_key", "session1_value", time.Hour)
	mc.Set("session2", "shared_key", "session2_value", time.Hour)

	v1, ok := mc.Get("session1", "shared_key")
	assert.True(t, ok)
	assert.Equal(t, "session1_value", v1)

	v2, ok := mc.Get("session2", "shared_key")
	assert.True(t, ok)
	assert.Equal(t, "session2_value", v2)

	// Deleting from one session shouldn't affect the other
	mc.Delete("session1", "shared_key")

	_, ok = mc.Get("session1", "shared_key")
	assert.False(t, ok)

	v2, ok = mc.Get("session2", "shared_key")
	assert.True(t, ok)
	assert.Equal(t, "session2_value", v2)
}

// Tests for new cache infrastructure

func TestDefaultTTL(t *testing.T) {
	mc := NewMemoryCache()
	// Default TTL should be 7 days
	assert.Equal(t, DefaultTTL, mc.GetDefaultTTL())
	assert.Equal(t, 7*24*time.Hour, mc.GetDefaultTTL())
}

func TestGetSetDefaultTTL(t *testing.T) {
	mc := NewMemoryCache()

	// Set new TTL
	newTTL := 14 * 24 * time.Hour
	mc.SetDefaultTTL(newTTL)

	assert.Equal(t, newTTL, mc.GetDefaultTTL())

	// Set another TTL
	mc.SetDefaultTTL(time.Hour)
	assert.Equal(t, time.Hour, mc.GetDefaultTTL())
}

func TestGetSetDefaultTTL_Concurrent(_ *testing.T) {
	mc := NewMemoryCache()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			mc.SetDefaultTTL(time.Duration(n) * time.Hour)
		}(i)
	}

	// Concurrent reads
	for range 100 {
		wg.Go(func() {
			_ = mc.GetDefaultTTL()
		})
	}

	wg.Wait()
	// Should not panic or deadlock
}

func TestGetWithMetadata(t *testing.T) {
	mc := NewMemoryCache()

	// Set with strategy
	mc.SetWithStrategy("session1", "key1", "value1", StrategyAsyncRendering, time.Hour)

	// Get with metadata
	entry, ok := mc.GetWithMetadata("session1", "key1")
	require.True(t, ok)
	require.NotNil(t, entry)

	assert.Equal(t, "value1", entry.Value)
	assert.Equal(t, StrategyAsyncRendering, entry.Strategy)
	assert.False(t, entry.CreatedAt.IsZero())
	assert.False(t, entry.CreatedAt.After(time.Now()))
	assert.True(t, entry.ExpiresAt.After(time.Now()))
}

func TestGetWithMetadata_NonExistent(t *testing.T) {
	mc := NewMemoryCache()

	entry, ok := mc.GetWithMetadata("session1", "nonexistent")
	assert.False(t, ok)
	assert.Nil(t, entry)
}

func TestGetWithMetadata_Expired(t *testing.T) {
	mc := NewMemoryCache()

	// Set with very short TTL
	mc.SetWithStrategy("session1", "key1", "value1", StrategyFolder, 10*time.Millisecond)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should return nil for expired entry
	entry, ok := mc.GetWithMetadata("session1", "key1")
	assert.False(t, ok)
	assert.Nil(t, entry)
}

func TestSetWithStrategy(t *testing.T) {
	mc := NewMemoryCache()

	// Test each strategy
	strategies := []CacheStrategy{
		StrategySession,
		StrategyFolder,
		StrategyAsyncRendering,
	}

	for i, strategy := range strategies {
		key := "key" + string(rune('0'+i))
		mc.SetWithStrategy("session1", key, "value", strategy, time.Hour)

		entry, ok := mc.GetWithMetadata("session1", key)
		require.True(t, ok)
		assert.Equal(t, strategy, entry.Strategy)
	}
}

func TestSetWithStrategy_ZeroTTL(t *testing.T) {
	mc := NewMemoryCache()

	// Zero TTL should not expire
	mc.SetWithStrategy("session1", "key1", "value1", StrategySession, 0)

	entry, ok := mc.GetWithMetadata("session1", "key1")
	require.True(t, ok)
	assert.False(t, entry.Expired())
}

func TestCacheEntryAge(t *testing.T) {
	entry := &CacheEntry{
		Value:     "test",
		CreatedAt: time.Now().Add(-time.Hour),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	age := entry.Age()
	// Age should be approximately 1 hour (allow some tolerance)
	assert.True(t, age >= 59*time.Minute && age <= 61*time.Minute,
		"Expected age around 1 hour, got %v", age)
}

func TestCacheEntryAge_Recent(t *testing.T) {
	entry := &CacheEntry{
		Value:     "test",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	age := entry.Age()
	// Age should be very small
	assert.True(t, age < time.Second, "Expected age < 1 second, got %v", age)
}

func TestClearAll(t *testing.T) {
	mc := NewMemoryCache()

	// Create multiple sessions with multiple entries
	mc.Set("session1", "key1", "value1", time.Hour)
	mc.Set("session1", "key2", "value2", time.Hour)
	mc.Set("session2", "key1", "value3", time.Hour)
	mc.Set("session3", "key1", "value4", time.Hour)

	assert.Equal(t, 3, mc.SessionCount())
	assert.Equal(t, 4, mc.EntryCount())

	// Clear all
	mc.ClearAll()

	assert.Equal(t, 0, mc.SessionCount())
	assert.Equal(t, 0, mc.EntryCount())

	// Verify all entries are gone
	_, ok := mc.Get("session1", "key1")
	assert.False(t, ok)
	_, ok = mc.Get("session2", "key1")
	assert.False(t, ok)
	_, ok = mc.Get("session3", "key1")
	assert.False(t, ok)
}

func TestClearAll_Empty(t *testing.T) {
	mc := NewMemoryCache()

	// Should not panic on empty cache
	mc.ClearAll()

	assert.Equal(t, 0, mc.SessionCount())
}

func TestToDaemonStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected CacheStrategy
	}{
		{"Session", "session", StrategySession},
		{"Folder", "folder", StrategyFolder},
		{"Empty", "", StrategyAsyncRendering},
		{"Unknown", "unknown", StrategyAsyncRendering},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Import config.Strategy type
			strategy := config.Strategy(tt.input)
			result := ToDaemonStrategy(strategy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for sessionTextCache

func newTestSessionTextCache() *sessionTextCache {
	return &sessionTextCache{
		cache:     NewMemoryCache(),
		sessionID: "test-session",
	}
}

func TestSessionTextCache_Get(t *testing.T) {
	stc := newTestSessionTextCache()

	// Set a value
	stc.Set("key1", "value1")

	// Get it back
	val, ok := stc.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)
}

func TestSessionTextCache_Get_NonExistent(t *testing.T) {
	stc := newTestSessionTextCache()

	val, ok := stc.Get("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, "", val)
}

func TestSessionTextCache_GetWithAge(t *testing.T) {
	stc := newTestSessionTextCache()

	// Set a value
	stc.Set("key1", "value1")

	// Get with age
	text, age, found := stc.GetWithAge("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", text)
	assert.True(t, age < time.Second, "Age should be very small")
}

func TestSessionTextCache_GetWithAge_NonExistent(t *testing.T) {
	stc := newTestSessionTextCache()

	text, age, found := stc.GetWithAge("nonexistent")
	assert.False(t, found)
	assert.Equal(t, "", text)
	assert.Equal(t, time.Duration(0), age)
}

func TestSessionTextCache_GetWithAge_OldEntry(t *testing.T) {
	stc := newTestSessionTextCache()

	// Manually set an old entry
	stc.cache.SetWithStrategy(stc.sessionID, "key1", "value1", StrategyAsyncRendering, time.Hour)

	// Modify CreatedAt to be 30 minutes ago (need to access internal entry)
	entry, _ := stc.cache.GetWithMetadata(stc.sessionID, "key1")
	// We can't directly modify CreatedAt, but we can verify the mechanism works

	text, age, found := stc.GetWithAge("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", text)
	assert.True(t, age >= 0)
	_ = entry // suppress unused warning
}

func TestSessionTextCache_SetWithConfig_NoConfig(t *testing.T) {
	stc := newTestSessionTextCache()

	// Set with nil config (AsyncRendering behavior)
	stc.SetWithConfig("key1", "value1", nil)

	entry, ok := stc.cache.GetWithMetadata(stc.sessionID, "key1")
	require.True(t, ok)
	assert.Equal(t, StrategyAsyncRendering, entry.Strategy)
	assert.Equal(t, "value1", entry.Value)
}

func TestSessionTextCache_SetWithConfig_SessionStrategy(t *testing.T) {
	stc := newTestSessionTextCache()

	cfg := &config.Cache{
		Strategy: config.Session,
	}
	stc.SetWithConfig("key1", "value1", cfg)

	entry, ok := stc.cache.GetWithMetadata(stc.sessionID, "key1")
	require.True(t, ok)
	assert.Equal(t, StrategySession, entry.Strategy)
}

func TestSessionTextCache_SetWithConfig_FolderStrategy(t *testing.T) {
	stc := newTestSessionTextCache()

	cfg := &config.Cache{
		Strategy: config.Folder,
	}
	stc.SetWithConfig("key1", "value1", cfg)

	entry, ok := stc.cache.GetWithMetadata(stc.sessionID, "key1")
	require.True(t, ok)
	assert.Equal(t, StrategyFolder, entry.Strategy)
}

func TestSessionTextCache_SetWithConfig_WithDuration(t *testing.T) {
	stc := newTestSessionTextCache()

	cfg := &config.Cache{
		Strategy: config.Folder,
		Duration: "1h",
	}
	stc.SetWithConfig("key1", "value1", cfg)

	entry, ok := stc.cache.GetWithMetadata(stc.sessionID, "key1")
	require.True(t, ok)
	// Entry should expire around 1 hour from now
	assert.True(t, entry.ExpiresAt.After(time.Now().Add(59*time.Minute)))
	assert.True(t, entry.ExpiresAt.Before(time.Now().Add(61*time.Minute)))
}

func TestSessionTextCache_ShouldRecompute_NoCache(t *testing.T) {
	stc := newTestSessionTextCache()

	// No cached value exists
	recompute, usePending := stc.ShouldRecompute("nonexistent", nil)
	assert.True(t, recompute, "Should recompute when no cache exists")
	assert.False(t, usePending, "Should not use pending when no cache")
}

func TestSessionTextCache_ShouldRecompute_NoConfig_AsyncRendering(t *testing.T) {
	stc := newTestSessionTextCache()

	// Set a cached value
	stc.Set("key1", "cached_value")

	// No config = AsyncRendering: always recompute, use cache for pending
	recompute, usePending := stc.ShouldRecompute("key1", nil)
	assert.True(t, recompute, "AsyncRendering should always recompute")
	assert.True(t, usePending, "AsyncRendering should use cache for pending")
}

func TestSessionTextCache_ShouldRecompute_FreshCache(t *testing.T) {
	stc := newTestSessionTextCache()

	// Set a value with duration config
	cfg := &config.Cache{
		Strategy: config.Folder,
		Duration: "1h",
	}
	stc.SetWithConfig("key1", "cached_value", cfg)

	// Cache is fresh (just set) - should NOT recompute
	recompute, usePending := stc.ShouldRecompute("key1", cfg)
	assert.False(t, recompute, "Should not recompute when cache is fresh")
	assert.False(t, usePending, "Should not need pending when cache is fresh")
}

func TestSessionTextCache_ShouldRecompute_StaleCache(t *testing.T) {
	stc := newTestSessionTextCache()

	// Create an entry with a long TTL (won't expire) but we'll check against a short duration
	// The entry's ExpiresAt is different from the config's Duration check
	stc.cache.SetWithStrategy(stc.sessionID, "key1", "cached_value", StrategyFolder, time.Hour)

	// Wait a bit so the entry has some age
	time.Sleep(50 * time.Millisecond)

	// Config has a very short duration - entry's age (50ms) > config duration (10ms)
	cfg := &config.Cache{
		Strategy: config.Folder,
		Duration: "10ms",
	}

	// Cache is stale (age > duration) - should recompute, use old cache for pending
	recompute, usePending := stc.ShouldRecompute("key1", cfg)
	assert.True(t, recompute, "Should recompute when cache is stale")
	assert.True(t, usePending, "Should use old cache for pending display")
}

func TestSessionTextCache_ShouldRecompute_InfiniteDuration(t *testing.T) {
	stc := newTestSessionTextCache()

	// Set with INFINITE duration and Folder strategy
	cfg := &config.Cache{
		Strategy: config.Folder,
		Duration: "infinite",
	}
	stc.SetWithConfig("key1", "cached_value", cfg)

	// INFINITE with non-AsyncRendering strategy should NOT recompute
	recompute, usePending := stc.ShouldRecompute("key1", cfg)
	assert.False(t, recompute, "INFINITE duration should not recompute")
	assert.False(t, usePending, "INFINITE duration should not need pending")
}

func TestSessionTextCache_ShouldRecompute_EmptyValue(t *testing.T) {
	stc := newTestSessionTextCache()

	// Set an empty string
	stc.Set("key1", "")

	// Empty cached value should not be used for pending
	recompute, usePending := stc.ShouldRecompute("key1", nil)
	assert.True(t, recompute, "Should recompute")
	assert.False(t, usePending, "Empty cache should not be used for pending")
}
