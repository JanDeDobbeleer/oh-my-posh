package platform

import "sync"

func NewConcurrentMap() *ConcurrentMap {
	var cm ConcurrentMap
	return &cm
}

type ConcurrentMap sync.Map

func (cm *ConcurrentMap) Set(key string, value any) {
	(*sync.Map)(cm).Store(key, value)
}

func (cm *ConcurrentMap) Get(key string) (any, bool) {
	return (*sync.Map)(cm).Load(key)
}

func (cm *ConcurrentMap) Delete(key string) {
	(*sync.Map)(cm).Delete(key)
}

func (cm *ConcurrentMap) Contains(key string) bool {
	_, ok := (*sync.Map)(cm).Load(key)
	return ok
}

func (cm *ConcurrentMap) SimpleMap() SimpleMap {
	list := make(map[string]any)
	(*sync.Map)(cm).Range(func(key, value any) bool {
		list[key.(string)] = value
		return true
	})
	return list
}

type SimpleMap map[string]any

func (m SimpleMap) ConcurrentMap() *ConcurrentMap {
	var cm ConcurrentMap
	for k, v := range m {
		cm.Set(k, v)
	}
	return &cm
}
