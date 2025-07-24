package config

import (
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"
)

func TestAsyncTimeoutConfiguration(t *testing.T) {
	segment := &Segment{
		Type:         "git",
		AsyncTimeout: 100 * time.Millisecond,
	}
	
	assert.Equal(t, 100*time.Millisecond, segment.AsyncTimeout)
	assert.Equal(t, "git", string(segment.Type))
}

func TestAsyncCacheKey(t *testing.T) {
	env := &mock.Environment{}
	env.On("Pwd").Return("/home/user/test")
	env.On("Getenv", "GIT_DIR").Return("")
	
	segment := &Segment{
		Type: "git",
		env:  env,
	}
	
	key := segment.generateAsyncCacheKey()
	expected := "git_/home/user/test"
	assert.Equal(t, expected, key)
}

func TestAsyncCache(t *testing.T) {
	cacheFile := &cache.File{}
	cacheFile.Init("/tmp/test_cache", false)
	defer cacheFile.Close()
	
	asyncCache := cache.NewAsyncSegmentCache(cacheFile)
	
	// Test setting and getting async data
	data := &cache.AsyncSegmentData{
		Text:      "test output",
		Enabled:   true,
		Timestamp: time.Now(),
		Duration:  cache.Duration("5m"),
	}
	
	asyncCache.SetSegmentData("git", "test_key", data)
	
	retrieved, found := asyncCache.GetSegmentData("git", "test_key")
	assert.True(t, found)
	assert.Equal(t, "test output", retrieved.Text)
	assert.True(t, retrieved.Enabled)
}

func TestAsyncProcessMarker(t *testing.T) {
	cacheFile := &cache.File{}
	cacheFile.Init("/tmp/test_cache", false)
	defer cacheFile.Close()
	
	asyncCache := cache.NewAsyncSegmentCache(cacheFile)
	
	// Test process marker
	assert.False(t, asyncCache.IsAsyncProcessRunning("git", "test_key"))
	
	asyncCache.SetAsyncProcessRunning("git", "test_key")
	assert.True(t, asyncCache.IsAsyncProcessRunning("git", "test_key"))
	
	asyncCache.ClearAsyncProcessRunning("git", "test_key")
	assert.False(t, asyncCache.IsAsyncProcessRunning("git", "test_key"))
}