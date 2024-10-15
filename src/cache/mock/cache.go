package mock

import (
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	mock "github.com/stretchr/testify/mock"
)

type Cache struct {
	mock.Mock
}

func (_m *Cache) Init(filePath string, persist bool) {
	_m.Called(filePath, persist)
}

func (_m *Cache) Close() {
	_m.Called()
}

func (_m *Cache) Get(key string) (string, bool) {
	ret := _m.Called(key)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// set provides a mock function with given fields: key, value, ttl
func (_m *Cache) Set(key, value string, duration cache.Duration) {
	_m.Called(key, value, duration)
}

func (_m *Cache) Delete(key string) {
	_m.Called(key)
}
