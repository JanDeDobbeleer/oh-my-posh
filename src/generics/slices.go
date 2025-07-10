package generics

import "fmt"

// ParseStringSlice converts any slice to a string slice
func ParseStringSlice(param any) []string {
	return parseSlice(param, func(v any) string { return fmt.Sprint(v) })
}

// parseSlice converts any slice type to a typed slice using a converter function
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
