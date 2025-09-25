package cache

import (
	"encoding/gob"
	"time"
)

func init() {
	gob.Register(&Entry[any]{})
	gob.Register(Template{})
	gob.Register(SimpleTemplate{})
	gob.Register((*Duration)(nil))
	gob.Register(map[string]bool{})
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
