package environment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	CacheFile = "/omp.cache"
)

type cacheObject struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	TTL       int    `json:"ttl"`
}

func (c *cacheObject) expired() bool {
	if c.TTL < 0 {
		return false
	}
	return time.Now().Unix() >= (c.Timestamp + int64(c.TTL)*60)
}

type fileCache struct {
	cache     *concurrentMap
	cachePath string
	dirty     bool
}

func (fc *fileCache) Init(cachePath string) {
	fc.cache = newConcurrentMap()
	fc.cachePath = cachePath
	cacheFilePath := filepath.Join(fc.cachePath, CacheFile)
	content, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return
	}

	var list map[string]*cacheObject
	if err := json.Unmarshal(content, &list); err != nil {
		return
	}

	for key, co := range list {
		if co.expired() {
			continue
		}
		fc.cache.set(key, co)
	}
}

func (fc *fileCache) Close() {
	if !fc.dirty {
		return
	}
	cache := fc.cache.list()
	if dump, err := json.MarshalIndent(cache, "", "    "); err == nil {
		cacheFilePath := filepath.Join(fc.cachePath, CacheFile)
		_ = os.WriteFile(cacheFilePath, dump, 0644)
	}
}

// returns the value for the given key as long as
// the TTL (minutes) is not expired
func (fc *fileCache) Get(key string) (string, bool) {
	val, found := fc.cache.get(key)
	if !found {
		return "", false
	}
	if co, ok := val.(*cacheObject); ok {
		return co.Value, true
	}
	return "", false
}

// sets the value for the given key with a TTL (minutes)
func (fc *fileCache) Set(key, value string, ttl int) {
	fc.cache.set(key, &cacheObject{
		Value:     value,
		Timestamp: time.Now().Unix(),
		TTL:       ttl,
	})
	fc.dirty = true
}
