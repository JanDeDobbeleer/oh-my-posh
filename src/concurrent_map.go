package main

import "sync"

type concurrentMap struct {
	values map[string]string
	lock   sync.RWMutex
}

func newConcurrentMap() *concurrentMap {
	return &concurrentMap{
		values: make(map[string]string),
		lock:   sync.RWMutex{},
	}
}

func (c *concurrentMap) set(key, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.values[key] = value
}

func (c *concurrentMap) get(key string) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if val, ok := c.values[key]; ok {
		return val, true
	}
	return "", false
}
