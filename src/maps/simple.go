package maps

type Simple map[string]any

func (m Simple) ToConcurrent() *Concurrent {
	return &Concurrent{
		data: m,
	}
}
