package daemon

import (
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/config"

	"github.com/stretchr/testify/assert"
)

func TestConfigCache(t *testing.T) {
	cache := NewConfigCache()
	cfg := &config.Config{}
	path := "/test/config.json"
	filePaths := []string{path}

	// Test Set
	cached := cache.Set(path, cfg, filePaths)
	assert.NotNil(t, cached)
	assert.Equal(t, cfg, cached.Config)
	assert.Equal(t, path, cached.FilePaths[0])
	assert.False(t, cached.IsRemote)

	// Test Get
	cached2, ok := cache.Get(path)
	assert.True(t, ok)
	assert.Equal(t, cached, cached2)

	// Test Count
	assert.Equal(t, 1, cache.Count())

	// Test Invalidate
	cache.Invalidate(path)
	cached3, ok := cache.Get(path)
	assert.False(t, ok)
	assert.Nil(t, cached3)
	assert.Equal(t, 0, cache.Count())
}

func TestConfigCacheRemote(t *testing.T) {
	cache := NewConfigCache()
	cfg := &config.Config{}
	path := "https://example.com/config.json"
	filePaths := []string{path}

	// Test Set Remote
	cached := cache.Set(path, cfg, filePaths)
	assert.True(t, cached.IsRemote)
	assert.True(t, cached.ExpiresAt.After(time.Now()))

	// Test Get before expiration
	_, ok := cache.Get(path)
	assert.True(t, ok)

	// Force expiration
	cached.ExpiresAt = time.Now().Add(-1 * time.Minute)

	// Test Get after expiration
	_, ok = cache.Get(path)
	assert.False(t, ok)
}

func TestConfigCacheInvalidateAll(t *testing.T) {
	cache := NewConfigCache()
	cfg := &config.Config{}

	cache.Set("/path1", cfg, []string{"/path1"})
	cache.Set("/path2", cfg, []string{"/path2"})
	assert.Equal(t, 2, cache.Count())

	cache.InvalidateAll()
	assert.Equal(t, 0, cache.Count())
}
