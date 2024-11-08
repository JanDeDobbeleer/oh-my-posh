package maps

import (
	"sync"
)

func NewConcurrent() *Concurrent {
	return &Concurrent{
		data: make(map[string]any),
	}
}

type Concurrent struct {
	data map[string]any
	sync.RWMutex
}

func (cm *Concurrent) Set(key string, value any) {
	cm.Lock()
	defer cm.Unlock()

	if cm.data == nil {
		cm.data = make(map[string]any)
	}

	cm.data[key] = value
}

func (cm *Concurrent) Get(key string) (any, bool) {
	cm.RLock()
	defer cm.RUnlock()

	if cm.data == nil {
		return nil, false
	}

	value, ok := cm.data[key]
	return value, ok
}

func (cm *Concurrent) Delete(key string) {
	cm.Lock()
	defer cm.Unlock()

	delete(cm.data, key)
}

func (cm *Concurrent) Contains(key string) bool {
	_, ok := cm.Get(key)
	return ok
}

func (cm *Concurrent) ToSimple() Simple {
	cm.RLock()
	defer cm.RUnlock()

	return cm.data
}
