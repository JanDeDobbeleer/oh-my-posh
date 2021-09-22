package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

const (
	cachePath = "/.omp.cache"
)

type cacheObject struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	TTL       int64  `json:"ttl"`
}

type fileCache struct {
	cache *concurrentMap
	home  string
}

func (fc *fileCache) init(home string) {
	fc.cache = newConcurrentMap()
	fc.home = home
	content, err := ioutil.ReadFile(fc.home + cachePath)
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
	fc.set("hello", "world", 200)
	if dump, err := json.Marshal(fc.cache.list()); err == nil {
		_ = ioutil.WriteFile(fc.home+cachePath, dump, 0644)
	}
}

func (fc *fileCache) get(key string) (string, bool) {
	val, found := fc.cache.get(key)
	if !found {
		return "", false
	}
	co, ok := val.(*cacheObject)
	if !ok {
		return "", false
	}
	expired := time.Now().Unix() >= (co.Timestamp + co.TTL*60)
	if expired {
		fc.cache.remove(key)
		return "", false
	}
	return co.Value, true
}

func (fc *fileCache) set(key, value string, ttl int64) {
	fc.cache.set(key, &cacheObject{
		Value:     value,
		Timestamp: time.Now().Unix(),
		TTL:       ttl,
	})
}
