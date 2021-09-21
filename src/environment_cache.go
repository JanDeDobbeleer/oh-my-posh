package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	cachePath = "/.omp.cache"
)

type fileCache struct {
	cache *concurrentMap
	home  string
}

func (fc *fileCache) init(home string) {
	fc.cache = newConcurrentMap()
	fc.home = home
	content, err := ioutil.ReadFile(home + cachePath)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(content), "\n") {
		if len(line) == 0 || !strings.Contains(line, "=") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		fc.set(kv[0], kv[1])
	}
}

func (fc *fileCache) close() {
	var sb strings.Builder
	for key, value := range fc.cache.values {
		cacheEntry := fmt.Sprintf("%s=%s\n", key, value)
		sb.WriteString(cacheEntry)
	}
	_ = ioutil.WriteFile(fc.home+cachePath, []byte(sb.String()), 0644)
}

func (fc *fileCache) get(key string) (string, bool) {
	return fc.cache.get(key)
}

func (fc *fileCache) set(key, value string) {
	fc.cache.set(key, value)
}
