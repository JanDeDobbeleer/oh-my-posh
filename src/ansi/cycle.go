package ansi

type Cycle []*Colors

func (c Cycle) Loop() (*Colors, Cycle) {
	if len(c) == 0 {
		return nil, c
	}
	return c[0], append(c[1:], c[0])
}
