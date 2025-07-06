package generics

import (
	"errors"
	"strconv"
)

type Numeric interface {
	~int | ~int64 | ~uint64 | ~float64
}

func toNumeric[T Numeric](value any) (T, error) {
	switch v := value.(type) {
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return T(parsed), nil
		}
		return T(0), err
	case int:
		return T(v), nil
	case int64:
		return T(v), nil
	case uint64:
		return T(v), nil
	case float64:
		return T(v), nil
	default:
		return T(0), errors.New("invalid numeric type")
	}
}

func ToInt[T ~int | ~int64](value any) (T, error) {
	return toNumeric[T](value)
}

func ToFloat[T ~float64](value any) (T, error) {
	return toNumeric[T](value)
}
