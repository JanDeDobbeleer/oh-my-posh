package maps

// Simple is a generic map type that can be specialized for different value types
type Simple[V any] map[string]V

func (m Simple[V]) ToConcurrent() *Concurrent[V] {
	cm := NewConcurrent[V]()
	for k, v := range m {
		cm.Set(k, v)
	}

	return cm
}
