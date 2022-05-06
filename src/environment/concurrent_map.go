package environment

type concurrentMap struct {
	values map[string]interface{}
}

func newConcurrentMap() *concurrentMap {
	return &concurrentMap{
		values: make(map[string]interface{}),
	}
}

func (c *concurrentMap) set(key string, value interface{}) {
	lock.Lock()
	defer lock.Unlock()
	c.values[key] = value
}

func (c *concurrentMap) get(key string) (interface{}, bool) {
	lock.RLock()
	defer lock.RUnlock()
	if val, ok := c.values[key]; ok {
		return val, true
	}
	return "", false
}

func (c *concurrentMap) remove(key string) {
	lock.RLock()
	defer lock.RUnlock()
	delete(c.values, key)
}

func (c *concurrentMap) list() map[string]interface{} {
	return c.values
}
