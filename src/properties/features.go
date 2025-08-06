package properties

import (
	"slices"
)

type Features []string

func (f *Features) Contains(feature string) bool {
	return slices.Contains(*f, feature)
}
