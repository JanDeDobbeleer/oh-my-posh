package template

import (
	"errors"
	"strconv"
)

func toIntOrZero(e any) int {
	if value, err := toInt(e); err == nil {
		return value
	}

	return 0
}

func toInt(integer any) (int, error) {
	switch seconds := integer.(type) {
	default:
		return 0, errors.New("invalid integer type")
	case string:
		return strconv.Atoi(seconds)
	case int:
		return seconds, nil
	case int64:
		return int(seconds), nil
	case uint64:
		return int(seconds), nil
	case float64:
		return int(seconds), nil
	}
}

func toFloat64(e any) float64 {
	if val, OK := e.(float64); OK {
		return val
	}
	if val, OK := e.(int); OK {
		return float64(val)
	}
	if val, OK := e.(int64); OK {
		return float64(val)
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
