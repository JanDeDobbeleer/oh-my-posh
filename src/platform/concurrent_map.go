package platform

import "sync"

type ConcurrentMap struct {
	values map[string]interface{}
	sync.RWMutex
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		values: make(map[string]interface{}),
	}
}

func (c *ConcurrentMap) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	c.values[key] = value
}

func (c *ConcurrentMap) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()
	if val, ok := c.values[key]; ok {
		return val, true
	}
	return "", false
}

func (c *ConcurrentMap) Delete(key string) {
	c.RLock()
	defer c.RUnlock()
	delete(c.values, key)
}

func (c *ConcurrentMap) List() map[string]interface{} {
	return c.values
}
