package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ApplyDefaults sets zero-valued struct fields from `default` struct tags.
// Supports string, bool, int, float64, time.Duration, slices, and maps.
func ApplyDefaults(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("ApplyDefaults requires a pointer, got %T", v)
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("ApplyDefaults requires a pointer to struct, got %T", v)
	}

	return applyDefaultsToStruct(rv)
}

func applyDefaultsToStruct(rv reflect.Value) error {
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle embedded structs
		if structField.Anonymous && field.Kind() == reflect.Struct {
			if err := applyDefaultsToStruct(field); err != nil {
				return err
			}
			continue
		}

		// Get default tag
		defaultTag := structField.Tag.Get("default")
		if defaultTag == "" {
			continue
		}

		// Only apply default if field is zero-valued
		if !isFieldZero(field) {
			continue
		}

		// Apply default based on field type
		if err := setDefault(field, defaultTag); err != nil {
			return fmt.Errorf("field %s: %w", structField.Name, err)
		}
	}

	return nil
}

func isFieldZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice:
		// A slice is zero if it's nil or has length 0
		return v.IsNil() || v.Len() == 0
	case reflect.Map:
		// A map is zero if it's nil or has length 0
		// Empty maps (non-nil with length 0) are considered zero-valued
		return v.IsNil() || v.Len() == 0
	default:
		return v.IsZero()
	}
}

func setDefault(field reflect.Value, defaultValue string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(defaultValue)
		return nil

	case reflect.Bool:
		val, err := strconv.ParseBool(defaultValue)
		if err != nil {
			return fmt.Errorf("invalid bool default %q: %w", defaultValue, err)
		}
		field.SetBool(val)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special handling for time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(defaultValue)
			if err != nil {
				return fmt.Errorf("invalid duration default %q: %w", defaultValue, err)
			}
			field.SetInt(int64(d))
			return nil
		}

		val, err := strconv.ParseInt(defaultValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int default %q: %w", defaultValue, err)
		}
		field.SetInt(val)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(defaultValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint default %q: %w", defaultValue, err)
		}
		field.SetUint(val)
		return nil

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(defaultValue, 64)
		if err != nil {
			return fmt.Errorf("invalid float default %q: %w", defaultValue, err)
		}
		field.SetFloat(val)
		return nil

	case reflect.Slice:
		return setSliceDefault(field, defaultValue)

	case reflect.Map:
		return setMapDefault(field, defaultValue)

	default:
		return fmt.Errorf("unsupported type %s", field.Type())
	}
}

func setSliceDefault(field reflect.Value, defaultValue string) error {
	// Handle empty slice literal "[]"
	if defaultValue == "[]" {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}

	// Try to parse as JSON array
	var result any
	if err := json.Unmarshal([]byte(defaultValue), &result); err == nil {
		// Validate that result is actually an array
		if arr, ok := result.([]any); ok {
			return setSliceFromJSON(field, arr)
		}
		return fmt.Errorf("expected JSON array for slice default, got %T", result)
	}

	// Fallback: comma-separated values for string slices
	if field.Type().Elem().Kind() == reflect.String {
		parts := strings.Split(defaultValue, ",")
		slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
		for i, part := range parts {
			slice.Index(i).SetString(strings.TrimSpace(part))
		}
		field.Set(slice)
		return nil
	}

	return fmt.Errorf("invalid slice default %q", defaultValue)
}

func setSliceFromJSON(field reflect.Value, data []any) error {
	slice := reflect.MakeSlice(field.Type(), len(data), len(data))
	elemType := field.Type().Elem()

	for i, item := range data {
		elem := slice.Index(i)
		if err := setValueFromAny(elem, elemType, item); err != nil {
			return fmt.Errorf("slice element %d: %w", i, err)
		}
	}

	field.Set(slice)
	return nil
}

func setMapDefault(field reflect.Value, defaultValue string) error {
	// Handle empty map literal "{}"
	if defaultValue == "{}" {
		field.Set(reflect.MakeMap(field.Type()))
		return nil
	}

	// Try to parse as JSON object
	var result map[string]any
	if err := json.Unmarshal([]byte(defaultValue), &result); err != nil {
		return fmt.Errorf("invalid map default %q: %w", defaultValue, err)
	}

	mapValue := reflect.MakeMap(field.Type())
	keyType := field.Type().Key()
	valueType := field.Type().Elem()

	// Note: Currently only string keys are supported in map defaults.
	// This is sufficient for common configuration use cases.
	// Extend this if other key types are needed.
	for k, v := range result {
		key := reflect.New(keyType).Elem()
		if keyType.Kind() == reflect.String {
			key.SetString(k)
		} else {
			return fmt.Errorf("unsupported map key type %s (only string keys supported)", keyType)
		}

		val := reflect.New(valueType).Elem()
		if err := setValueFromAny(val, valueType, v); err != nil {
			return fmt.Errorf("map value for key %q: %w", k, err)
		}

		mapValue.SetMapIndex(key, val)
	}

	field.Set(mapValue)
	return nil
}

func setValueFromAny(field reflect.Value, fieldType reflect.Type, value any) error {
	switch fieldType.Kind() {
	case reflect.String:
		if s, ok := value.(string); ok {
			field.SetString(s)
			return nil
		}
		field.SetString(fmt.Sprint(value))
		return nil

	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
			return nil
		}
		return fmt.Errorf("expected bool, got %T", value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case int:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		case float64:
			field.SetInt(int64(v))
		default:
			return fmt.Errorf("expected int, got %T", value)
		}
		return nil

	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case float64:
			field.SetFloat(v)
		case int:
			field.SetFloat(float64(v))
		case int64:
			field.SetFloat(float64(v))
		default:
			return fmt.Errorf("expected float, got %T", value)
		}
		return nil

	default:
		return fmt.Errorf("unsupported field type %s", fieldType)
	}
}
