package cache

import (
	"encoding/gob"
	"path/filepath"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
)

type store struct {
	cache    *maps.Concurrent[*Entry[any]]
	filePath string
	dirty    bool
	persist  bool
}

var (
	session *store
	device  *store
	Persist = false
)

type Store string

const (
	Session Store = "session"
	Device  Store = "device"
)

func (s Store) new() *store {
	return &store{
		cache: maps.NewConcurrent[*Entry[any]](),
	}
}

// getStore returns the appropriate store based on the Store identifier
func (s Store) get() *store {
	switch s { //nolint:exhaustive
	case Device:
		if device == nil {
			device = s.new()
		}

		return device
	default:
		if session == nil {
			session = s.new()
		}

		return session
	}
}

// Init initializes a store with the given file path
func (s Store) init(filePath string, persist bool) {
	defer log.Trace(time.Now(), string(s), filePath)

	store := s.get()
	store.cache = maps.NewConcurrent[*Entry[any]]()
	store.filePath = filepath.Join(Path(), filePath)
	store.persist = persist

	reader, err := openFile(store.filePath)
	if err != nil {
		// set to dirty so we create it on close
		log.Error(err)
		store.dirty = true
		return
	}

	defer reader.Close()

	var list maps.Simple[*Entry[any]]

	dec := gob.NewDecoder(reader)
	if err := dec.Decode(&list); err != nil {
		log.Error(err)
		// If gob decoding fails, the cache file might be from the old format
		// Set dirty to true so we recreate it in gob format
		store.dirty = true
		return
	}

	for key, entry := range list {
		if entry.Expired() {
			log.Debug("skipping expired cache key:", key, "in store:", string(s))
			continue
		}

		log.Debug("loading cache key:", key, "in store:", string(s))
		store.cache.Set(key, entry)
	}
}

func (s Store) close() {
	defer log.Trace(time.Now(), string(s))

	store := s.get()
	if store == nil || !store.persist || !store.dirty {
		log.Debug("not persisting cache for store:", string(s))
		return
	}

	cache := store.cache.ToSimple()

	file, err := openFile(store.filePath)
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

// Get retrieves a typed value from the specified store
func Get[T any](s Store, key string) (T, bool) {
	var zero T
	defer log.Trace(time.Now(), string(s), key)

	store := s.get()
	if store == nil {
		log.Debug("store is nil for:", string(s))
		return zero, false
	}

	entry, found := store.cache.Get(key)
	if !found {
		log.Debug("cache key not found:", key, "in store:", string(s))
		return zero, false
	}

	if entry.Expired() {
		log.Debug("cache key expired:", key, "in store:", string(s))
		store.cache.Delete(key)
		store.dirty = true
		return zero, false
	}

	log.Debug("found cache key:", key, "in store:", string(s))

	// Type assertion to get the typed value
	if typed, ok := entry.Value.(T); ok {
		return typed, true
	}

	log.Debug("type mismatch for cache key:", key, "in store:", string(s))
	return zero, false
}

// Set stores a typed value in the specified store
func Set[T any](s Store, key string, value T, duration Duration) {
	defer log.Trace(time.Now(), string(s), key)

	store := s.get()
	if store == nil {
		log.Debug("store is nil for:", string(s))
		return
	}

	seconds := duration.Seconds()
	if seconds == 0 {
		return
	}

	log.Debug("setting cache key:", key, "in store:", string(s), "with duration:", string(duration))

	store.cache.Set(key, &Entry[any]{
		Value:     value,
		Timestamp: time.Now().Unix(),
		TTL:       seconds,
	})

	store.dirty = true
}

// Delete removes a key from the specified store
func Delete(s Store, key string) {
	defer log.Trace(time.Now(), string(s), key)

	store := s.get()
	if store == nil {
		log.Debug("store is nil for:", string(s))
		return
	}

	log.Debug("deleting cache key:", key, "from store:", string(s))
	store.cache.Delete(key)
	store.dirty = true
}

func DeleteAll(s Store) {
	defer log.Trace(time.Now(), string(s))

	store := s.get()
	if store == nil {
		log.Debug("store is nil for:", string(s))
		return
	}

	store.cache = maps.NewConcurrent[*Entry[any]]()
	store.dirty = true
}
