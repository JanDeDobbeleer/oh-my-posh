package dsc

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type State interface {
	Apply(c cache.Cache) error
	Resolve()
	Test(state State) error
	String() string
	CacheKey() string
	New() State
	Schema() string
}

func Get[T State](c cache.Cache) T {
	var state T
	state = state.New().(T)

	cached, ok := c.Get(state.CacheKey())
	if !ok {
		return state.New().(T)
	}

	err := json.Unmarshal([]byte(cached), state)
	if err == nil {
		return state
	}

	return state.New().(T)
}

func Save[T State](c cache.Cache, state T) {
	data, err := json.Marshal(state)
	if err != nil {
		log.Error(err)
		return
	}

	c.Set(state.CacheKey(), string(data), cache.INFINITE)
}

func export[T State](c cache.Cache) string {
	state := Get[T](c)
	state.Resolve()
	return state.String()
}

func set[T State](c cache.Cache, schema string) error {
	var state T

	if err := json.Unmarshal([]byte(schema), state); err != nil {
		return newError(err.Error())
	}

	if err := state.Apply(c); err != nil {
		return newError(err.Error())
	}

	Save(c, state)

	return nil
}

func test[T State](c cache.Cache, state string) error {
	actual := Get[T](c)

	var expected T
	if err := json.Unmarshal([]byte(state), expected); err != nil {
		return newError(err.Error())
	}

	return actual.Test(expected)
}
