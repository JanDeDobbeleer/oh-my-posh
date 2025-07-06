package generics

import "sync"

type Pool[T any] struct {
	pool sync.Pool
	new  func() T
}

func NewPool[T any](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any { return newFunc() },
		},
		new: newFunc,
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(item T) {
	p.pool.Put(item)
}
