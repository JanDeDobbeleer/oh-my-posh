package environment

import (
	"encoding/json"
	"io/ioutil"
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

type fileCache struct {
	cache     *concurrentMap
	cachePath string
	dirty     bool
}

func (fc *fileCache) Init(cachePath string) {
	fc.cache = newConcurrentMap()
	fc.cachePath = cachePath
	cacheFilePath := filepath.Join(fc.cachePath, CacheFile)
	content, err := ioutil.ReadFile(cacheFilePath)
	if err != nil {
		return
	}

	var list map[string]*cacheObject
	if err := json.Unmarshal(content, &list); err != nil {
		return
	}

	for key, co := range list {
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
		_ = ioutil.WriteFile(cacheFilePath, dump, 0644)
	}
}

// returns the value for the given key as long as
// the TTL (minutes) is not expired
func (fc *fileCache) Get(key string) (string, bool) {
	val, found := fc.cache.get(key)
	if !found {
		return "", false
	}
	co, ok := val.(*cacheObject)
	if !ok {
		return "", false
	}
	if co.TTL <= 0 {
		return co.Value, true
	}
	expired := time.Now().Unix() >= (co.Timestamp + int64(co.TTL)*60)
	if expired {
		fc.cache.remove(key)
		return "", false
	}
	return co.Value, true
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
