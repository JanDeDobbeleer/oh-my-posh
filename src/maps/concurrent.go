package maps

import (
	"fmt"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func NewConcurrent[V any]() *Concurrent[V] {
	return &Concurrent[V]{}
}

// Concurrent is a generic type-safe concurrent map
type Concurrent[V any] struct {
	m sync.Map
}

func (cm *Concurrent[V]) Set(key string, value V) {
	cm.m.Store(key, value)
}

func (cm *Concurrent[V]) Get(key string) (V, bool) {
	val, ok := cm.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}

	return val.(V), true
}

func (cm *Concurrent[V]) MustGet(key string) V {
	val, ok := cm.m.Load(key)
	if !ok {
		log.Error(fmt.Errorf("key %s not found", key))
		var zero V
		return zero
	}

	return val.(V)
}

func (cm *Concurrent[V]) Delete(key string) {
	cm.m.Delete(key)
}

func (cm *Concurrent[V]) Contains(key string) bool {
	_, ok := cm.m.Load(key)
	return ok
}

func (cm *Concurrent[V]) ToSimple() Simple[V] {
	result := make(Simple[V])

	cm.m.Range(func(key, value any) bool {
		if value == nil {
			return true
		}
		result[key.(string)] = value.(V)
		return true
	})

	return result
}
