package ansi

type Cycle []*Colors

func PopColors(array []*Colors) (*Colors, Cycle) {
	if len(array) == 0 {
		return nil, array
	}
	return array[0], array[1:]
}
