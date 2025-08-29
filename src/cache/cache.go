package cache

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Cache interface {
	Init(filePath string, persist bool)
	Close()
	// Gets the value for a given key.
	// Returns the value and a boolean indicating if the key was found.
	// In case the duration expired, the function returns false.
	Get(key string) (string, bool)
	// Sets a value for a given key.
	// The duration indicates how many minutes to cache the value.
	Set(key, value string, duration Duration)
	// Deletes a key from the cache.
	Delete(key string)
}

const (
	FileName = "omp.cache"
)

var (
	sessionID string
	once      sync.Once
)

func SessionID() string {
	once.Do(func() {
		sessionID = os.Getenv("POSH_SESSION_ID")
		if sessionID == "" {
			sessionID = uuid.NewString()
		}
	})

	return sessionID
}

func SessionFileName() string {
	return fmt.Sprintf("%s.%s", FileName, SessionID())
}

const (
	TEMPLATECACHE    = "template_cache"
	TOGGLECACHE      = "toggle_cache"
	PROMPTCOUNTCACHE = "prompt_count_cache"
	ENGINECACHE      = "engine_cache"
	FONTLISTCACHE    = "font_list_cache"
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

	return time.Now().Unix() >= (c.Timestamp + int64(c.TTL))
}
