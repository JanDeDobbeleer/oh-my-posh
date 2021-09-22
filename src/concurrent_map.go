package main

import "sync"

type concurrentMap struct {
	values map[string]interface{}
	lock   sync.RWMutex
}

func newConcurrentMap() *concurrentMap {
	return &concurrentMap{
		values: make(map[string]interface{}),
		lock:   sync.RWMutex{},
	}
}

func (c *concurrentMap) set(key string, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.values[key] = value
}

func (c *concurrentMap) get(key string) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if val, ok := c.values[key]; ok {
		return val, true
	}
	return "", false
}

func (c *concurrentMap) remove(key string) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	delete(c.values, key)
}

func (c *concurrentMap) list() map[string]interface{} {
	return c.values
}
