package maps

type Simple map[string]any

func (m Simple) ToConcurrent() *Concurrent {
	var cm Concurrent
	for k, v := range m {
		cm.Set(k, v)
	}
	return &cm
}
