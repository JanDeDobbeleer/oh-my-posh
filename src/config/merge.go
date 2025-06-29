package config

import (
	"errors"
	"reflect"
	"slices"
)

func (cfg *Config) merge(override *Config) error {
	nextExtends := cfg.Extends

	err := merge(override, cfg, "Blocks", "Source", "Format", "updated", "extended")
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

		overrideSegmentMap := make(map[string]*Segment)
		for _, segment := range overrideBlock.Segments {
			if segment != nil {
				overrideSegmentMap[segment.Name()] = segment
			}
		}

		for k := range cfg.Blocks[i].Segments {
			overrideSegment, exists := overrideSegmentMap[cfg.Blocks[i].Segments[k].Name()]
			if !exists {
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
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
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
		// For slices like Blocks and Tooltips, append base items that don't exist in base
		// This is a simple append - you might want more sophisticated merging logic
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
			if !base.MapIndex(key).IsValid() {
				base.SetMapIndex(key, override.MapIndex(key))
			}
		}
	}

	if base.IsNil() {
		// Initialize empty map if both are nil but base has the type
		base.Set(reflect.MakeMap(base.Type()))
	}
}
