package config

import (
	"errors"
	"reflect"
	"slices"
)

func (cfg *Config) merge(override *Config) error {
	if cfg == nil || override == nil {
		return errors.New("configs cannot be nil")
	}

	nextExtends := cfg.Extends

	err := merge(override, cfg, "Blocks", "Source", "Format")
	if err != nil {
		return err
	}

	overrideBlockMap := make(map[string]*Block)
	for _, block := range override.Blocks {
		if block != nil {
			overrideBlockMap[block.key()] = block
		}
	}

	for i := range cfg.Blocks {
		overrideBlock, exists := overrideBlockMap[cfg.Blocks[i].key()]
		if !exists {
			continue
		}

		// remove the block from the override map so we don't match it again
		delete(overrideBlockMap, overrideBlock.key())

		err = merge(overrideBlock, cfg.Blocks[i], "Segments")
		if err != nil {
			return err
		}

		overrideSegmentMap := make(map[any]*Segment)
		for _, segment := range overrideBlock.Segments {
			if segment == nil {
				continue
			}

			var key any

			switch {
			case segment.Index <= 0:
				key = segment.Name()
			default:
				key = segment.Index - 1 // Index is 1-based, so we subtract 1 to use it as a map key
			}

			overrideSegmentMap[key] = segment
		}

		for k := range cfg.Blocks[i].Segments {
			var overrideSegment *Segment

			overrideSegment = overrideSegmentMap[k]

			if overrideSegment == nil {
				overrideSegment = overrideSegmentMap[cfg.Blocks[i].Segments[k].Name()]
			}

			if overrideSegment == nil {
				continue
			}

			// remove the segment from the override map so we don't match it again
			delete(overrideSegmentMap, overrideSegment.Name())

			baseSegment := cfg.Blocks[i].Segments[k]

			if baseSegment.Type != overrideSegment.Type {
				cfg.Blocks[i].Segments[k] = overrideSegment
				continue
			}

			err = merge(overrideSegment, baseSegment)
			if err != nil {
				return err
			}
		}
	}

	cfg.Extends = nextExtends
	cfg.extended = true

	return nil
}

func merge(override, base any, skipFields ...string) error {
	if base == nil || override == nil {
		return errors.New("config to merge cannot be nil")
	}

	overrideValue := reflect.ValueOf(override).Elem()
	baseValue := reflect.ValueOf(base).Elem()
	overrideType := overrideValue.Type()

	for i := 0; i < overrideValue.NumField(); i++ {
		field := overrideType.Field(i)

		if !field.IsExported() {
			continue
		}

		overrideField := overrideValue.Field(i)
		baseField := baseValue.FieldByName(field.Name)

		// Skip unexported fields or fields that can't be set
		if isZeroValue(overrideField) || !baseField.CanSet() {
			continue
		}

		// Skip internal fields that shouldn't be merged
		if slices.Contains(skipFields, field.Name) {
			continue
		}

		// Special handling for slices - merge instead of replace
		if overrideField.Kind() == reflect.Slice {
			mergeSlices(overrideField, baseField)
			continue
		}

		// Special handling for maps - merge instead of replace
		if overrideField.Kind() == reflect.Map {
			mergeMaps(overrideField, baseField)
			continue
		}

		if baseField.CanSet() {
			baseField.Set(overrideField)
		}
	}

	return nil
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() { //nolint: exhaustive
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	case reflect.String:
		return v.String() == ""
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return false
	default:
		return v.IsZero()
	}
}

func mergeSlices(override, base reflect.Value) {
	if base.IsNil() && !override.IsNil() {
		base.Set(override)
		return
	}

	if !base.IsNil() && !override.IsNil() {
		newSlice := reflect.AppendSlice(base, override)
		base.Set(newSlice)
	}
}

func mergeMaps(override, base reflect.Value) {
	if base.IsNil() && !override.IsNil() {
		base.Set(override)
		return
	}

	if !base.IsNil() && !override.IsNil() {
		// Merge maps - cfg values override base values
		for _, key := range override.MapKeys() {
			base.SetMapIndex(key, override.MapIndex(key))
		}
	}

	if base.IsNil() {
		// Initialize empty map if both are nil but base has the type
		base.Set(reflect.MakeMap(base.Type()))
	}
}
