package generics

import (
	"errors"
	"reflect"
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
	case bool:
		if v {
			return T(1), nil
		}
		return T(0), nil
	default:
		// Handle named types with numeric underlying types (e.g. type Percentage int)
		rv := reflect.ValueOf(value)
		switch rv.Kind() { //nolint:exhaustive
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return T(rv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return T(rv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return T(rv.Float()), nil
		}
		return T(0), errors.New("invalid numeric type")
	}
}

func TryParseInt[T ~int | ~int64](value any) (T, error) {
	return toNumeric[T](value)
}

func TryParseFloat[T ~float64](value any) (T, error) {
	return toNumeric[T](value)
}

func ToInt[T ~int | ~int64](value any) T {
	result, err := toNumeric[T](value)
	if err != nil {
		return T(0)
	}

	return result
}
