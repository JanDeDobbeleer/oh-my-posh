package cache

import (
	"encoding/gob"
	"os"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

func init() {
	gob.Register(&Entry{})
}

type File struct {
	cache         *maps.Concurrent[*Entry]
	cacheFilePath string
	dirty         bool
	persist       bool
}

func (fc *File) Init(cacheFilePath string, persist bool) {
	defer log.Trace(time.Now(), cacheFilePath)

	fc.cache = maps.NewConcurrent[*Entry]()
	fc.cacheFilePath = cacheFilePath
	fc.persist = persist

	log.Debug("loading cache file:", fc.cacheFilePath)

	file, err := os.Open(fc.cacheFilePath)
	if err != nil {
		// set to dirty so we create it on close
		log.Error(err)
		fc.dirty = true
		return
	}

	defer file.Close()

	var list maps.Simple[*Entry]

	dec := gob.NewDecoder(file)
	if err := dec.Decode(&list); err != nil {
		log.Error(err)
		// If gob decoding fails, the cache file might be from the old JSON format
		// Set dirty to true so we recreate it in gob format
		fc.dirty = true
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
	if fc == nil || !fc.persist || !fc.dirty {
		log.Debug("not persisting cache")
		return
	}

	defer log.Trace(time.Now(), fc.cacheFilePath)

	cache := fc.cache.ToSimple()

	file, err := os.Create(fc.cacheFilePath)
	if err != nil {
		log.Error(err)
		return
	}

	defer file.Close()

	enc := gob.NewEncoder(file)
	if err := enc.Encode(cache); err != nil {
		log.Error(err)
	}
}

// returns the value for the given key as long as
// the duration is not expired
func (fc *File) Get(key string) (string, bool) {
	if fc == nil {
		log.Debug("cache is nil, returning empty value for key:", key)
		return "", false
	}

	entry, found := fc.cache.Get(key)
	if !found {
		log.Debug("cache key not found:", key)
		return "", false
	}

	log.Debug("found cache key:", key)
	return entry.Value, true
}

// sets the value for the given key with a duration
func (fc *File) Set(key, value string, duration Duration) {
	if fc == nil {
		log.Debug("cache is nil, not setting value for key:", key)
		return
	}

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
	if fc == nil {
		log.Debug("cache is nil, not deleting key:", key)
		return
	}

	log.Debug("deleting cache key:", key)
	fc.cache.Delete(key)
	fc.dirty = true
}
