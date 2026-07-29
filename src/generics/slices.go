package generics

import "fmt"

func ParseStringSlice(param any) []string {
	return parseSlice(param, func(v any) string { return fmt.Sprint(v) })
}

func parseSlice[T any](param any, converter func(any) T) []T {
	switch v := param.(type) {
	case []any:
		if len(v) == 0 {
			return []T{}
		}

		result := make([]T, len(v))
		for i, item := range v {
			result[i] = converter(item)
		}

		return result
	case []T:
		return v
	default:
		return []T{}
	}
}
