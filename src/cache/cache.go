package cache

import (
	"encoding/gob"
	"errors"
	"time"
)

// ErrLocked is returned by openFile when the cache file is held exclusively
// by another process (e.g. a Windows sharing violation that persisted past
// the retry window). Callers must treat this as "leave the file alone":
// operate purely in-memory for this run and do not recreate/truncate the
// file on close.
var ErrLocked = errors.New("cache file is locked by another process")

func init() {
	gob.Register(&Entry[any]{})
	gob.Register(Template{})
	gob.Register(SimpleTemplate{})
	gob.Register((*Duration)(nil))
	gob.Register(map[string]bool{})
	gob.Register(commandPathEntry{})
}

const (
	DeviceStore = "omp.cache"
)

const (
	TEMPLATECACHE    = "template_cache"
	TOGGLECACHE      = "toggle_cache"
	PROMPTCOUNTCACHE = "prompt_count_cache"
	ENGINECACHE      = "engine_cache"
	FONTLISTCACHE    = "font_list_cache"
	CLAUDECACHE      = "claude_cache"
	COPILOTCLICACHE  = "copilot_cli_cache"
)

type Entry[T any] struct {
	Value     T
	Timestamp int64
	TTL       int
}

func (c *Entry[T]) Expired() bool {
	if c.TTL < 0 {
		return false
	}

	return time.Now().Unix() >= (c.Timestamp + int64(c.TTL))
}
