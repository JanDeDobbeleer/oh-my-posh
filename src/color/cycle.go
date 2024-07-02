package color

type Cycle []*Set

func (c Cycle) Loop() (*Set, Cycle) {
	if len(c) == 0 {
		return nil, c
	}
	return c[0], append(c[1:], c[0])
}
