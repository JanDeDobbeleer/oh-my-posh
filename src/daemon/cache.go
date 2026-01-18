package daemon

import (
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
)

// DefaultTTL is the default cache TTL (7 days) for non-session-based entries.
const DefaultTTL = 7 * 24 * time.Hour

// CacheStrategy is internal to the daemon - not exposed to user config.
// This extends the user-facing strategies (session, folder) with daemon-specific behavior.
type CacheStrategy string

const (
	// StrategySession caches at session level (cleaned when session ends).
	StrategySession CacheStrategy = "session"
	// StrategyFolder caches based on folder (uses writer's CacheKey or pwd).
	StrategyFolder CacheStrategy = "folder"
	// StrategyAsyncRendering is the internal daemon default when no config is provided.
	// Always recomputes the segment, uses cache only for pending rendering display.
	StrategyAsyncRendering CacheStrategy = "async_rendering"
)

// ToDaemonStrategy converts a user-facing config.Strategy to a CacheStrategy.
// Returns the appropriate daemon strategy, defaulting to AsyncRendering when no strategy is set.
func ToDaemonStrategy(s config.Strategy) CacheStrategy {
	switch s {
	case config.Session:
		return StrategySession
	case config.Folder, config.Device:
		return StrategyFolder
	default:
		// When no strategy is configured, use AsyncRendering (daemon internal default)
		return StrategyAsyncRendering
	}
}

// CacheEntry represents a single cached value with metadata.
type CacheEntry struct {
	Value     any
	CreatedAt time.Time     // When entry was created
	ExpiresAt time.Time     // When entry expires
	Strategy  CacheStrategy // Strategy used for this entry
}

// Expired returns true if the cache entry has expired.
func (e *CacheEntry) Expired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Age returns how long ago this entry was created.
func (e *CacheEntry) Age() time.Duration {
	return time.Since(e.CreatedAt)
}

// MemoryCache provides an in-memory cache organized by terminal session.
// Each session has its own namespace to prevent cross-session interference.
// This is the L1 cache that lives in the daemon process.
type MemoryCache struct {
	// sessions maps session IDs to their cache entries
	// Using sync.Map for concurrent access since sessions may be
	// created/accessed from different goroutines
	sessions sync.Map // map[string]*sessionCache

	// defaultTTL is the default TTL for non-session-based entries.
	// Protected by ttlMu for thread-safe access.
	defaultTTL time.Duration
	ttlMu      sync.RWMutex
}

// sessionCache holds cache entries for a single terminal session.
type sessionCache struct {
	entries sync.Map // map[string]*CacheEntry
}

// NewMemoryCache creates a new empty memory cache with default TTL of 7 days.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		defaultTTL: DefaultTTL,
	}
}

// GetDefaultTTL returns the current default TTL for non-session-based entries.
func (mc *MemoryCache) GetDefaultTTL() time.Duration {
	mc.ttlMu.RLock()
	defer mc.ttlMu.RUnlock()
	return mc.defaultTTL
}

// SetDefaultTTL sets the default TTL for non-session-based entries.
func (mc *MemoryCache) SetDefaultTTL(ttl time.Duration) {
	mc.ttlMu.Lock()
	defer mc.ttlMu.Unlock()
	mc.defaultTTL = ttl
}

// Get retrieves a value from the cache for the given session and key.
// Returns the value and true if found and not expired, otherwise nil and false.
func (mc *MemoryCache) Get(sessionID, key string) (any, bool) {
	session, ok := mc.sessions.Load(sessionID)
	if !ok {
		return nil, false
	}

	sc := session.(*sessionCache)
	entry, ok := sc.entries.Load(key)
	if !ok {
		return nil, false
	}

	ce := entry.(*CacheEntry)
	if ce.Expired() {
		// Clean up expired entry
		sc.entries.Delete(key)
		return nil, false
	}

	return ce.Value, true
}

// GetWithMetadata retrieves a cache entry with full metadata for the given session and key.
// Returns the entry and true if found and not expired, otherwise nil and false.
// Use this when you need access to CreatedAt for age-based validation.
func (mc *MemoryCache) GetWithMetadata(sessionID, key string) (*CacheEntry, bool) {
	session, ok := mc.sessions.Load(sessionID)
	if !ok {
		return nil, false
	}

	sc := session.(*sessionCache)
	entry, ok := sc.entries.Load(key)
	if !ok {
		return nil, false
	}

	ce := entry.(*CacheEntry)
	if ce.Expired() {
		// Clean up expired entry
		sc.entries.Delete(key)
		return nil, false
	}

	return ce, true
}

// Set stores a value in the cache for the given session and key with a TTL.
// If TTL is 0 or negative, the entry never expires.
// Uses StrategyAsyncRendering as the default strategy.
func (mc *MemoryCache) Set(sessionID, key string, value any, ttl time.Duration) {
	mc.SetWithStrategy(sessionID, key, value, StrategyAsyncRendering, ttl)
}

// SetWithStrategy stores a value with explicit strategy and TTL.
// If TTL is 0 or negative, the entry never expires (within session lifetime).
func (mc *MemoryCache) SetWithStrategy(sessionID, key string, value any, strategy CacheStrategy, ttl time.Duration) {
	// Get or create session cache
	session, _ := mc.sessions.LoadOrStore(sessionID, &sessionCache{})
	sc := session.(*sessionCache)

	now := time.Now()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	} else {
		// Use a far future time for "never expires"
		expiresAt = now.Add(100 * 365 * 24 * time.Hour)
	}

	sc.entries.Store(key, &CacheEntry{
		Value:     value,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Strategy:  strategy,
	})
}

// Delete removes a specific key from a session's cache.
func (mc *MemoryCache) Delete(sessionID, key string) {
	session, ok := mc.sessions.Load(sessionID)
	if !ok {
		return
	}

	sc := session.(*sessionCache)
	sc.entries.Delete(key)
}

// CleanSession removes all cache entries for a specific session.
// Call this when a terminal session ends.
func (mc *MemoryCache) CleanSession(sessionID string) {
	mc.sessions.Delete(sessionID)
}

// ClearAll removes all cache entries from all sessions.
// Used by the "omp cache clear" command.
func (mc *MemoryCache) ClearAll() {
	mc.sessions.Range(func(key, _ any) bool {
		mc.sessions.Delete(key)
		return true
	})
}

// EvictExpired removes all expired entries from all sessions.
// This should be called periodically to prevent memory bloat.
func (mc *MemoryCache) EvictExpired() {
	mc.sessions.Range(func(_, sessionValue any) bool {
		sc := sessionValue.(*sessionCache)

		sc.entries.Range(func(entryKey, entryValue any) bool {
			ce := entryValue.(*CacheEntry)
			if ce.Expired() {
				sc.entries.Delete(entryKey)
			}
			return true
		})

		return true
	})
}

// SessionCount returns the number of active sessions in the cache.
// Useful for monitoring and debugging.
func (mc *MemoryCache) SessionCount() int {
	count := 0
	mc.sessions.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// EntryCount returns the total number of entries across all sessions.
// Useful for monitoring and debugging.
func (mc *MemoryCache) EntryCount() int {
	count := 0
	mc.sessions.Range(func(_, sessionValue any) bool {
		sc := sessionValue.(*sessionCache)
		sc.entries.Range(func(_, _ any) bool {
			count++
			return true
		})
		return true
	})
	return count
}
