package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

const (
	CacheFile = "/omp.cache"
)

type File struct {
	cache     *maps.Concurrent
	cachePath string
	dirty     bool
}

func (fc *File) Init(cachePath string) {
	fc.cache = maps.NewConcurrent()
	fc.cachePath = cachePath
	cacheFilePath := filepath.Join(fc.cachePath, CacheFile)
	content, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return
	}

	var list map[string]*Entry
	if err := json.Unmarshal(content, &list); err != nil {
		return
	}

	for key, co := range list {
		if co.Expired() {
			continue
		}

		fc.cache.Set(key, co)
	}
}

func (fc *File) Close() {
	if !fc.dirty {
		return
	}

	cache := fc.cache.ToSimple()

	if dump, err := json.MarshalIndent(cache, "", "    "); err == nil {
		cacheFilePath := filepath.Join(fc.cachePath, CacheFile)
		_ = os.WriteFile(cacheFilePath, dump, 0644)
	}
}

// returns the value for the given key as long as
// the TTL (minutes) is not expired
func (fc *File) Get(key string) (string, bool) {
	val, found := fc.cache.Get(key)
	if !found {
		return "", false
	}
	if co, ok := val.(*Entry); ok {
		return co.Value, true
	}
	return "", false
}

// sets the value for the given key with a TTL (minutes)
func (fc *File) Set(key, value string, ttl int) {
	fc.cache.Set(key, &Entry{
		Value:     value,
		Timestamp: time.Now().Unix(),
		TTL:       ttl,
	})
	fc.dirty = true
}

// delete the key from the cache
func (fc *File) Delete(key string) {
	fc.cache.Delete(key)
	fc.dirty = true
}
