package maps

import "sync"

func NewConcurrent() *Concurrent {
	var cm Concurrent
	return &cm
}

type Concurrent sync.Map

func (cm *Concurrent) Set(key string, value any) {
	(*sync.Map)(cm).Store(key, value)
}

func (cm *Concurrent) Get(key string) (any, bool) {
	return (*sync.Map)(cm).Load(key)
}

func (cm *Concurrent) Delete(key string) {
	(*sync.Map)(cm).Delete(key)
}

func (cm *Concurrent) Contains(key string) bool {
	_, ok := (*sync.Map)(cm).Load(key)
	return ok
}

func (cm *Concurrent) ToSimple() Simple {
	list := make(map[string]any)
	(*sync.Map)(cm).Range(func(key, value any) bool {
		list[key.(string)] = value
		return true
	})
	return list
}
