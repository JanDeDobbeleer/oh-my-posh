package prompt

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/stretchr/testify/assert"
)

func TestGetTooltipCacheKey(t *testing.T) {
	engine := &Engine{}
	
	cases := []struct {
		Tip      string
		Expected string
	}{
		{Tip: "git", Expected: "tooltip_cache_git"},
		{Tip: "npm", Expected: "tooltip_cache_npm"},
		{Tip: "eslint", Expected: "tooltip_cache_eslint"},
	}
	
	for _, tc := range cases {
		result := engine.getTooltipCacheKey(tc.Tip)
		assert.Equal(t, tc.Expected, result, "Cache key should match expected format")
	}
}

func TestTooltipCacheStorage(t *testing.T) {
	// Initialize a fresh session cache
	cache.Init(shell.GENERIC)
	
	engine := &Engine{}
	tip := "test-tip"
	output := "test output"
	duration := cache.Duration("1h")
	
	// Store in cache
	cacheKey := engine.getTooltipCacheKey(tip)
	cache.Set(cache.Session, cacheKey, output, duration)
	
	// Retrieve from cache
	cachedOutput, cached := cache.Get[string](cache.Session, cacheKey)
	assert.True(t, cached, "Output should be cached")
	assert.Equal(t, output, cachedOutput, "Cached output should match")
}
