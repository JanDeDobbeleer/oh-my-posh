package cache

import (
	"encoding/json"
	"fmt"
	"time"
)

// AsyncSegmentData represents cached data for async segments
type AsyncSegmentData struct {
	Text      string    `json:"text"`
	Enabled   bool      `json:"enabled"`
	Timestamp time.Time `json:"timestamp"`
	Duration  Duration  `json:"duration"`
}

// AsyncSegmentCache manages async segment caching
type AsyncSegmentCache struct {
	cache Cache
}

// NewAsyncSegmentCache creates a new async segment cache
func NewAsyncSegmentCache(cache Cache) *AsyncSegmentCache {
	return &AsyncSegmentCache{
		cache: cache,
	}
}

// GetSegmentData retrieves cached segment data
func (a *AsyncSegmentCache) GetSegmentData(segmentName, cacheKey string) (*AsyncSegmentData, bool) {
	key := fmt.Sprintf("async_segment_%s_%s", segmentName, cacheKey)
	data, found := a.cache.Get(key)
	if !found {
		return nil, false
	}

	var segmentData AsyncSegmentData
	if err := json.Unmarshal([]byte(data), &segmentData); err != nil {
		return nil, false
	}

	return &segmentData, true
}

// SetSegmentData stores segment data in cache
func (a *AsyncSegmentCache) SetSegmentData(segmentName, cacheKey string, data *AsyncSegmentData) {
	key := fmt.Sprintf("async_segment_%s_%s", segmentName, cacheKey)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	a.cache.Set(key, string(jsonData), data.Duration)
}

// DeleteSegmentData removes cached segment data
func (a *AsyncSegmentCache) DeleteSegmentData(segmentName, cacheKey string) {
	key := fmt.Sprintf("async_segment_%s_%s", segmentName, cacheKey)
	a.cache.Delete(key)
}

// IsAsyncProcessRunning checks if an async process is currently running for a segment
func (a *AsyncSegmentCache) IsAsyncProcessRunning(segmentName, cacheKey string) bool {
	key := fmt.Sprintf("async_process_%s_%s", segmentName, cacheKey)
	_, found := a.cache.Get(key)
	return found
}

// SetAsyncProcessRunning marks an async process as running
func (a *AsyncSegmentCache) SetAsyncProcessRunning(segmentName, cacheKey string) {
	key := fmt.Sprintf("async_process_%s_%s", segmentName, cacheKey)
	// Set a short TTL to prevent stale process markers
	a.cache.Set(key, "running", Duration("5m"))
}

// ClearAsyncProcessRunning removes the async process marker
func (a *AsyncSegmentCache) ClearAsyncProcessRunning(segmentName, cacheKey string) {
	key := fmt.Sprintf("async_process_%s_%s", segmentName, cacheKey)
	a.cache.Delete(key)
}