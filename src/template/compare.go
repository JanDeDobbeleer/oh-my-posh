package template

import "github.com/jandedobbeleer/oh-my-posh/src/generics"

func toIntOrZero(e any) int {
	if value, err := generics.TryParseInt[int](e); err == nil {
		return value
	}

	return 0
}

func toInt(integer any) (int, error) {
	return generics.TryParseInt[int](integer)
}

func toFloat64(e any) float64 {
	if val, err := generics.TryParseFloat[float64](e); err == nil {
		return val
	}
	return 0
}

func gt(e1, e2 any) bool {
	if val, OK := e1.(int); OK {
		return val > toIntOrZero(e2)
	}
	if val, OK := e1.(int64); OK {
		return val > int64(toIntOrZero(e2))
	}
	if val, OK := e1.(float64); OK {
		return val > toFloat64(e2)
	}
	return false
}

func lt(e1, e2 any) bool {
	return gt(e2, e1)
}
