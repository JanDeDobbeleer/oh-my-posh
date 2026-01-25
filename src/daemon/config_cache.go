package daemon

import (
	"strings"
	"sync"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
)

// remoteTTL is the cache expiration time for remote (https://) configs.
const remoteTTL = 5 * time.Minute

// CachedConfig holds a parsed config with metadata.
type CachedConfig struct {
	LoadedAt  time.Time
	ExpiresAt time.Time
	Config    *config.Config
	FilePaths []string
	Hash      uint64
	IsRemote  bool
}

// ConfigCache manages cached configs by path.
// Thread-safe for concurrent access.
type ConfigCache struct {
	configs map[string]*CachedConfig
	mu      sync.RWMutex
}

// NewConfigCache creates a new empty config cache.
func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		configs: make(map[string]*CachedConfig),
	}
}

// Get retrieves a cached config by path.
// Returns nil and false if not found or expired.
func (c *ConfigCache) Get(configPath string) (*CachedConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.configs[configPath]
	if !ok {
		return nil, false
	}

	// Check TTL for remote configs
	if cached.IsRemote && time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	return cached, true
}

// Set stores a config in the cache.
func (c *ConfigCache) Set(configPath string, cfg *config.Config, filePaths []string) *CachedConfig {
	isRemote := strings.HasPrefix(configPath, "https://") || strings.HasPrefix(configPath, "http://")

	cached := &CachedConfig{
		Config:    cfg,
		Hash:      cfg.Hash(),
		LoadedAt:  time.Now(),
		FilePaths: filePaths,
		IsRemote:  isRemote,
	}

	if isRemote {
		cached.ExpiresAt = time.Now().Add(remoteTTL)
	}

	c.mu.Lock()
	c.configs[configPath] = cached
	c.mu.Unlock()

	return cached
}

// Invalidate removes a config from the cache.
// Called when the config file changes.
func (c *ConfigCache) Invalidate(configPath string) {
	c.mu.Lock()
	delete(c.configs, configPath)
	c.mu.Unlock()
}

// InvalidateAll removes all configs from the cache.
func (c *ConfigCache) InvalidateAll() {
	c.mu.Lock()
	c.configs = make(map[string]*CachedConfig)
	c.mu.Unlock()
}

// Count returns the number of cached configs.
func (c *ConfigCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.configs)
}
