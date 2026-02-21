package cache

import (
	"encoding/gob"
	"fmt"
	"path/filepath"
	"strings"
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
)

type Store string

const (
	Session Store  = "session"
	Device  Store  = "device"
	TTL     string = "ttl"
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
			log.Debugf("(%s) skipping expired key: %s", string(s), key)
			continue
		}

		log.Debugf("(%s) loading %s", string(s), key)
		store.cache.Set(key, entry)
	}
}

func (s Store) close() {
	defer log.Trace(time.Now(), string(s))

	store := s.get()
	if store == nil || !store.persist || !store.dirty {
		log.Debugf("(%s) not persisting", string(s))
		return
	}

	cache := store.cache.ToSimple()

	file, err := openFile(store.filePath)
	if err != nil {
		log.Error(err)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Error(err)
		}
	}()

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
		log.Debugf("(%s) store is nil", string(s))
		return zero, false
	}

	entry, found := store.cache.Get(key)
	if !found {
		log.Debugf("(%s) key not found: %s", string(s), key)
		return zero, false
	}

	if entry.Expired() {
		log.Debugf("(%s) key expired: %s", string(s), key)
		store.cache.Delete(key)
		store.dirty = true
		return zero, false
	}

	// Type assertion to get the typed value
	if typed, ok := entry.Value.(T); ok {
		log.Debugf("(%s) found entry: %s - %v", string(s), key, typed)
		return typed, true
	}

	log.Error(fmt.Errorf("(%s) type mismatch for key: %s. Got %T, expected %T", string(s), key, entry.Value, zero))
	return zero, false
}

// Set stores a typed value in the specified store
func Set[T any](s Store, key string, value T, duration Duration) {
	defer log.Trace(time.Now(), string(s), key)

	store := s.get()
	if store == nil {
		log.Debugf("(%s) store is nil", string(s))
		return
	}

	seconds := duration.Seconds()
	if seconds == 0 {
		return
	}

	log.Debugf("(%s) setting entry: %s - %v with duration: %s", string(s), key, value, string(duration))

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
		log.Debugf("(%s) store is nil", string(s))
		return
	}

	log.Debugf("(%s) deleting key: %s", string(s), key)
	store.cache.Delete(key)
	store.dirty = true
}

func DeleteAll(s Store) {
	defer log.Trace(time.Now(), string(s))

	store := s.get()
	if store == nil {
		log.Debugf("(%s) store is nil", string(s))
		return
	}

	store.cache = maps.NewConcurrent[*Entry[any]]()
	store.dirty = true
}

func Print(s Store) string {
	defer log.Trace(time.Now(), string(s))

	store := s.get()
	if store == nil {
		return fmt.Sprintf("Store %s is nil", string(s))
	}

	cache := store.cache.ToSimple()
	if len(cache) == 0 {
		return fmt.Sprintf("Store %s is empty", string(s))
	}

	var builder strings.Builder

	for key, entry := range cache {
		builder.WriteString("\n")

		if entry.Expired() {
			fmt.Fprintf(&builder, "Key: %s [EXPIRED]\n", key)
			builder.WriteString("\n")
			continue
		}

		var ttlInfo string
		if entry.TTL < 0 {
			ttlInfo = "never expires"
		}
		if entry.TTL >= 0 {
			expiresAt := time.Unix(entry.Timestamp+int64(entry.TTL), 0)
			ttlInfo = fmt.Sprintf("expires at %s", expiresAt.Format("2006-01-02 15:04:05"))
		}

		fmt.Fprintf(&builder, "Key: %s\n", key)
		fmt.Fprintf(&builder, "  Value: %s\n", fmt.Sprintf("%#v", entry.Value))
		fmt.Fprintf(&builder, "  Type: %T\n", entry.Value)
		fmt.Fprintf(&builder, "  Created: %s\n", time.Unix(entry.Timestamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&builder, "  TTL: %s\n", ttlInfo)
	}

	return builder.String()
}
