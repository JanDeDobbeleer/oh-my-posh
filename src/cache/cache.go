package cache

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Cache interface {
	Init(home string)
	Close()
	// Gets the value for a given key.
	// Returns the value and a boolean indicating if the key was found.
	// In case the ttl expired, the function returns false.
	Get(key string) (string, bool)
	// Sets a value for a given key.
	// The ttl indicates how many minutes to cache the value.
	Set(key, value string, ttl int)
	// Deletes a key from the cache.
	Delete(key string)
}

func pid() string {
	pid := os.Getenv("POSH_PID")
	if len(pid) == 0 {
		pid = strconv.Itoa(os.Getppid())
	}
	return pid
}

var (
	TEMPLATECACHE    = fmt.Sprintf("template_cache_%s", pid())
	TOGGLECACHE      = fmt.Sprintf("toggle_cache_%s", pid())
	PROMPTCOUNTCACHE = fmt.Sprintf("prompt_count_cache_%s", pid())
)

type Entry struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	TTL       int    `json:"ttl"`
}

func (c *Entry) Expired() bool {
	if c.TTL < 0 {
		return false
	}
	return time.Now().Unix() >= (c.Timestamp + int64(c.TTL)*60)
}
