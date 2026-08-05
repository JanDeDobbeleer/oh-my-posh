package cache

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
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
	// locked is set when the on-disk cache file could not be opened because
	// another process holds it (e.g. Windows sharing violation that
	// persisted past the retry window). When true, the store operates
	// purely in-memory for the lifetime of this process: close() must not
	// attempt to (re)create the file, since doing so would truncate the
	// other process's data.
	locked bool
	// mtime is the on-disk file's modification time as of the last load or
	// refresh. A long-lived process (serve) uses it to detect writes made by
	// other processes (e.g. toggle) between render cycles - see Refresh.
	mtime time.Time
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

func (s Store) get() *store {
	switch s {
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

func (s Store) init(filePath string, persist bool) {
	defer log.Trace(time.Now(), string(s), filePath)

	store := s.get()
	store.cache = maps.NewConcurrent[*Entry[any]]()
	store.filePath = filepath.Join(Path(), filePath)
	store.persist = persist
	store.dirty = false
	store.locked = false
	store.mtime = time.Time{}

	reader, err := openFile(store.filePath)
	if err != nil {
		if errors.Is(err, ErrLocked) {
			// Another process holds the file. Leave it alone: run this
			// session purely in-memory, don't mark dirty, don't recreate
			// on close.
			log.Debugf("(%s) cache file is locked, running in-memory only", string(s))
			store.locked = true
			return
		}

		// set to dirty so we create it on close
		log.Error(err)
		store.dirty = true
		return
	}

	defer reader.Close()

	if info, err := os.Stat(store.filePath); err == nil {
		store.mtime = info.ModTime()
	}

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

// Updates the modification time if older than 1 hour, preventing stale
// session cache files from being cleaned up while reducing steady-state
// overhead.
func touchSessionFile(filePath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		return
	}

	if time.Since(info.ModTime()) <= time.Hour {
		return
	}

	if err := os.Chtimes(filePath, time.Now(), time.Now()); err != nil {
		log.Error(err)
	}
}

// Refresh re-syncs the in-memory store with the on-disk file if it has
// changed since the last load or refresh, merging entries by Timestamp (the
// newer value wins). A short-lived, one-shot invocation never needs this -
// init() already loads the current file once. It exists for a long-lived
// process (serve) that keeps its cache in memory for the session: without
// it, a write from another process (e.g. `toggle`, `enable`/`disable`)
// would stay invisible until the daemon exits.
func Refresh(s Store) {
	defer log.Trace(time.Now(), string(s))

	store := s.get()
	if store == nil || store.locked || store.filePath == "" {
		return
	}

	info, err := os.Stat(store.filePath)
	if err != nil || !info.ModTime().After(store.mtime) {
		return
	}

	reader, err := openFile(store.filePath)
	if err != nil {
		return
	}

	defer reader.Close()

	var list maps.Simple[*Entry[any]]

	dec := gob.NewDecoder(reader)
	if err := dec.Decode(&list); err != nil {
		log.Error(err)
		return
	}

	for key, diskEntry := range list {
		if diskEntry.Expired() {
			continue
		}

		if current, found := store.cache.Get(key); found && current.Timestamp >= diskEntry.Timestamp {
			continue
		}

		log.Debugf("(%s) refreshing %s from disk", string(s), key)
		store.cache.Set(key, diskEntry)
	}

	store.mtime = info.ModTime()
}

func (s Store) close() {
	defer log.Trace(time.Now(), string(s))

	store := s.get()

	// Pick up any write from another process one last time before a dirty
	// store overwrites the file, so a change made after the last Refresh
	// (e.g. right before the shell exits) isn't clobbered by this store's
	// own, possibly stale, in-memory copy.
	if store != nil && !store.locked && store.persist && store.dirty {
		Refresh(s)
	}

	if store == nil || store.locked || !store.persist || !store.dirty {
		if s == Session && store != nil && !store.locked && store.filePath != "" {
			touchSessionFile(store.filePath)
		}

		log.Debugf("(%s) not persisting", string(s))
		return
	}

	cache := store.cache.ToSimple()

	file, err := openFileForWrite(store.filePath)
	if err != nil {
		if errors.Is(err, ErrLocked) {
			// Became locked between init and close (e.g. another process
			// started writing meanwhile) — do not recreate/truncate.
			log.Debugf("(%s) cache file locked on close, skipping persist", string(s))
			return
		}

		log.Error(err)
		return
	}

	enc := gob.NewEncoder(file)
	if err := enc.Encode(cache); err != nil {
		log.Error(err)
	}

	if err := file.Close(); err != nil {
		log.Error(err)
	}

	// On Windows, the mmap-backed write path doesn't reliably update the
	// file's on-disk last-write-time (per Microsoft's docs). For the session
	// store that can lead to an actively-used cache being mistaken for stale
	// and swept up by cache.Clear(); for either store, a long-lived Refresh
	// reader (serve) needs a trustworthy mtime to notice this write at all.
	// Explicitly bump it now that the file is closed (and the mmap
	// unmap/flush on Windows has happened).
	if err := os.Chtimes(store.filePath, time.Now(), time.Now()); err != nil {
		log.Error(err)
	}
}

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

	// gob.Register keys its decode-side type map on the exact type it was given,
	// while encoding an interface value uses the base type. Registering a type
	// as a pointer (see segment_registry.go) therefore makes every gob round-trip
	// of a value stored under that type come back as *T instead of T. Accept that
	// shape here rather than dropping it as a cache miss.
	if ptr, ok := entry.Value.(*T); ok && ptr != nil {
		log.Debugf("(%s) found entry: %s - %v", string(s), key, *ptr)
		return *ptr, true
	}

	log.Error(fmt.Errorf("(%s) type mismatch for key: %s. Got %T, expected %T", string(s), key, entry.Value, zero))
	return zero, false
}

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
