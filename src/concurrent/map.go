package concurrent

import "sync"

func NewMap() *Map {
	var cm Map
	return &cm
}

type Map sync.Map

func (cm *Map) Set(key string, value any) {
	(*sync.Map)(cm).Store(key, value)
}

func (cm *Map) Get(key string) (any, bool) {
	return (*sync.Map)(cm).Load(key)
}

func (cm *Map) Delete(key string) {
	(*sync.Map)(cm).Delete(key)
}

func (cm *Map) Contains(key string) bool {
	_, ok := (*sync.Map)(cm).Load(key)
	return ok
}

func (cm *Map) ToSimpleMap() SimpleMap {
	list := make(map[string]any)
	(*sync.Map)(cm).Range(func(key, value any) bool {
		list[key.(string)] = value
		return true
	})
	return list
}

type SimpleMap map[string]any

func (m SimpleMap) ConcurrentMap() *Map {
	var cm Map
	for k, v := range m {
		cm.Set(k, v)
	}
	return &cm
}
