package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

const (
	fileName = "/omp.cache"
)

type cacheObject struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	TTL       int    `json:"ttl"`
}

type fileCache struct {
	cache     *concurrentMap
	cachePath string
}

func (fc *fileCache) init(cachePath string) {
	fc.cache = newConcurrentMap()
	fc.cachePath = cachePath
	content, err := ioutil.ReadFile(fc.cachePath + fileName)
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

func (fc *fileCache) close() {
	cache := fc.cache.list()
	if len(cache) == 0 {
		return
	}
	if dump, err := json.MarshalIndent(cache, "", "    "); err == nil {
		_ = ioutil.WriteFile(fc.cachePath+fileName, dump, 0644)
	}
}

// returns the value for the given key as long as
// the TTL (minutes) is not expired
func (fc *fileCache) get(key string) (string, bool) {
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
func (fc *fileCache) set(key, value string, ttl int) {
	fc.cache.set(key, &cacheObject{
		Value:     value,
		Timestamp: time.Now().Unix(),
		TTL:       ttl,
	})
}
