package maps

import (
	"fmt"
	"sync"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

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

func (cm *Concurrent) MustGet(key string) any {
	val, OK := (*sync.Map)(cm).Load(key)
	if !OK {
		log.Error(fmt.Errorf("key %s not found", key))
	}

	return val
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
		if value == nil {
			return false
		}

		list[key.(string)] = value
		return true
	})

	return list
}
