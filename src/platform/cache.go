package platform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
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

const (
	CacheFile = "/omp.cache"
)

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

type cacheObject struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	TTL       int    `json:"ttl"`
}

func (c *cacheObject) expired() bool {
	if c.TTL < 0 {
		return false
	}
	return time.Now().Unix() >= (c.Timestamp + int64(c.TTL)*60)
}

type fileCache struct {
	cache     *ConcurrentMap
	cachePath string
	dirty     bool
}

func (fc *fileCache) Init(cachePath string) {
	fc.cache = NewConcurrentMap()
	fc.cachePath = cachePath
	cacheFilePath := filepath.Join(fc.cachePath, CacheFile)
	content, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return
	}

	var list map[string]*cacheObject
	if err := json.Unmarshal(content, &list); err != nil {
		return
	}

	for key, co := range list {
		if co.expired() {
			continue
		}
		fc.cache.Set(key, co)
	}
}

func (fc *fileCache) Close() {
	if !fc.dirty {
		return
	}
	cache := fc.cache.SimpleMap()
	if dump, err := json.MarshalIndent(cache, "", "    "); err == nil {
		cacheFilePath := filepath.Join(fc.cachePath, CacheFile)
		_ = os.WriteFile(cacheFilePath, dump, 0644)
	}
}

// returns the value for the given key as long as
// the TTL (minutes) is not expired
func (fc *fileCache) Get(key string) (string, bool) {
	val, found := fc.cache.Get(key)
	if !found {
		return "", false
	}
	if co, ok := val.(*cacheObject); ok {
		return co.Value, true
	}
	return "", false
}

// sets the value for the given key with a TTL (minutes)
func (fc *fileCache) Set(key, value string, ttl int) {
	fc.cache.Set(key, &cacheObject{
		Value:     value,
		Timestamp: time.Now().Unix(),
		TTL:       ttl,
	})
	fc.dirty = true
}

// delete the key from the cache
func (fc *fileCache) Delete(key string) {
	fc.cache.Delete(key)
	fc.dirty = true
}

type commandCache struct {
	commands *ConcurrentMap
}

func (c *commandCache) set(command, path string) {
	c.commands.Set(command, path)
}

func (c *commandCache) get(command string) (string, bool) {
	cacheCommand, found := c.commands.Get(command)
	if !found {
		return "", false
	}
	command, ok := cacheCommand.(string)
	return command, ok
}

type TemplateCache struct {
	Root          bool
	PWD           string
	Folder        string
	Shell         string
	ShellVersion  string
	UserName      string
	HostName      string
	Code          int
	Env           map[string]string
	Var           SimpleMap
	OS            string
	WSL           bool
	PromptCount   int
	SHLVL         int
	Segments      *ConcurrentMap
	SegmentsCache SimpleMap

	initialized bool
	sync.RWMutex
}

func (t *TemplateCache) AddSegmentData(key string, value any) {
	t.Segments.Set(key, value)
}

func (t *TemplateCache) RemoveSegmentData(key string) {
	t.Segments.Delete(key)
}
