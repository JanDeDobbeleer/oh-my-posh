package cache

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Cache interface {
	Init(filePath string, persist bool)
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

const (
	FileName = "omp.cache"
)

var SessionFileName = fmt.Sprintf("%s.%s", FileName, sessionID())

func sessionID() string {
	pid := os.Getenv("POSH_SESSION_ID")
	if len(pid) == 0 {
		log.Debug("POSH_SESSION_ID not set, using PID")
		pid = strconv.Itoa(os.Getppid())
	}

	return pid
}

const (
	TEMPLATECACHE    = "template_cache"
	TOGGLECACHE      = "toggle_cache"
	PROMPTCOUNTCACHE = "prompt_count_cache"
	ENGINECACHE      = "engine_cache"
	FONTLISTCACHE    = "font_list_cache"

	ONEDAY   = 1440
	ONEWEEK  = 10080
	ONEMONTH = 43200
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
