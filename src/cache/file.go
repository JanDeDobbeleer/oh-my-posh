package cache

import (
	"encoding/json"
	"os"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type File struct {
	cache         *maps.Concurrent
	cacheFilePath string
	dirty         bool
	persist       bool
}

func (fc *File) Init(cacheFilePath string, persist bool) {
	defer log.Trace(time.Now(), cacheFilePath)

	fc.cache = maps.NewConcurrent()
	fc.cacheFilePath = cacheFilePath
	fc.persist = persist

	log.Debug("loading cache file:", fc.cacheFilePath)

	content, err := os.ReadFile(fc.cacheFilePath)
	if err != nil {
		// set to dirty so we create it on close
		fc.dirty = true
		log.Error(err)
		return
	}

	var list map[string]*Entry
	if err := json.Unmarshal(content, &list); err != nil {
		return
	}

	for key, co := range list {
		if co.Expired() {
			log.Debug("skipping expired cache key:", key)
			continue
		}

		log.Debug("loading cache key:", key)
		fc.cache.Set(key, co)
	}
}

func (fc *File) Close() {
	if !fc.persist || !fc.dirty {
		log.Debug("not persisting cache")
		return
	}

	cache := fc.cache.ToSimple()

	if dump, err := json.MarshalIndent(cache, "", "    "); err == nil {
		_ = os.WriteFile(fc.cacheFilePath, dump, 0o644)
	}
}

// returns the value for the given key as long as
// the duration is not expired
func (fc *File) Get(key string) (string, bool) {
	val, found := fc.cache.Get(key)
	if !found {
		log.Debug("cache key not found:", key)
		return "", false
	}

	if co, ok := val.(*Entry); ok {
		log.Debug("getting cache key:", key, "with value:", co.Value)
		return co.Value, true
	}

	log.Debug("unable to parse cache key:", key)
	return "", false
}

// sets the value for the given key with a duration
func (fc *File) Set(key, value string, duration Duration) {
	seconds := duration.Seconds()

	if seconds == 0 {
		return
	}

	log.Debug("setting cache key:", key, "with duration:", string(duration))

	fc.cache.Set(key, &Entry{
		Value:     value,
		Timestamp: time.Now().Unix(),
		TTL:       seconds,
	})

	fc.dirty = true
}

// delete the key from the cache
func (fc *File) Delete(key string) {
	log.Debug("deleting cache key:", key)
	fc.cache.Delete(key)
	fc.dirty = true
}
