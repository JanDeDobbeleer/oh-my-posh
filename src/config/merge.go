package config

import (
	"errors"
	"reflect"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// presenceAware is implemented by config types (Config, Block, Segment) that
// record which of their fields were present in the source file. merge() uses
// it to tell "the source explicitly set this scalar to its zero value" apart
// from "the source never mentioned this field", a distinction bool/int/uint/
// float fields otherwise lose once decoded into their Go zero value.
type presenceAware interface {
	fieldPresent(name string) bool
}

type matcher interface {
	key() any
}

type matchMap[T matcher] map[any]T

func (mm *matchMap[T]) hasMatch(index int, m T) (T, bool) {
	for _, item := range *mm {
		if item.key() == index {
			return item, true
		}
	}

	match, OK := (*mm)[m.key()]
	return match, OK
}

func (mm *matchMap[T]) add(m T) {
	if *mm == nil {
		*mm = make(matchMap[T])
	}

	(*mm)[m.key()] = m
}

func (mm *matchMap[T]) remove(m T) {
	delete(*mm, m.key())
}

func createMatchMap[T matcher](items []T) matchMap[T] {
	mm := make(matchMap[T])
	for _, item := range items {
		if any(item) != nil {
			mm.add(item)
		}
	}
	return mm
}

func (cfg *Config) merge(override *Config) error {
	if cfg == nil || override == nil {
		return errors.New("configs cannot be nil")
	}

	nextExtends := cfg.Extends

	err := merge(override, cfg, "Blocks", "Source", "Format")
	if err != nil {
		return err
	}

	overrideBlockMap := createMatchMap(override.Blocks)

	for i := range cfg.Blocks {
		overrideBlock, exists := overrideBlockMap.hasMatch(i, cfg.Blocks[i])
		if !exists {
			continue
		}

		// remove the block from the override map so we don't match it again
		overrideBlockMap.remove(overrideBlock)

		err = merge(overrideBlock, cfg.Blocks[i], "Segments")
		if err != nil {
			return err
		}

		overrideSegmentMap := createMatchMap(overrideBlock.Segments)

		for k := range cfg.Blocks[i].Segments {
			overrideSegment, exists := overrideSegmentMap.hasMatch(k, cfg.Blocks[i].Segments[k])
			if !exists {
				log.Debugf("No matching segment found for %s in block %s", cfg.Blocks[i].Segments[k].Type, cfg.Blocks[i].Type)
				continue
			}

			// remove the block from the override map so we don't match it again
			overrideSegmentMap.remove(overrideSegment)

			baseSegment := cfg.Blocks[i].Segments[k]

			if baseSegment.Type != overrideSegment.Type {
				log.Debugf("Replacing segment %s with %s in block %s", baseSegment.Type, overrideSegment.Type, cfg.Blocks[i].Type)
				cfg.Blocks[i].Segments[k] = overrideSegment
				continue
			}

			err = merge(overrideSegment, baseSegment)
			if err != nil {
				return err
			}
		}

		// add any remaining segments that were not matched
		for _, segment := range overrideSegmentMap {
			log.Debugf("Adding segment %s to block %s", segment.Type, cfg.Blocks[i].Type)
			cfg.Blocks[i].Segments = append(cfg.Blocks[i].Segments, segment)
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

		// Skip fields that can't be set
		if !baseField.CanSet() {
			continue
		}

		if skipField(override, overrideField, &field) {
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

		baseField.Set(overrideField)
	}

	return nil
}

// skipField decides whether overrideField should be left alone during merge.
// For scalar kinds (bool/int/uint/float) whose zero value is ambiguous between
// "explicitly set" and "absent from the source", it consults override's
// field-presence data (when available) instead of isZeroValue. All other
// kinds keep the original isZeroValue behavior unchanged.
func skipField(override any, overrideField reflect.Value, field *reflect.StructField) bool {
	if !isScalarKind(overrideField.Kind()) {
		return isZeroValue(overrideField)
	}

	jsonKey := jsonFieldName(field)
	if jsonKey == "" || jsonKey == "-" {
		// No source key to look up presence for (internal/runtime-only field
		// that is never decoded from a config file) - fall back to legacy
		// behavior.
		return isZeroValue(overrideField)
	}

	pa, OK := override.(presenceAware)
	if !OK {
		return isZeroValue(overrideField)
	}

	return !pa.fieldPresent(jsonKey)
}

func isScalarKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// jsonFieldName returns the source key name a struct field decodes from,
// derived from its json tag (e.g. "final_space,omitempty" -> "final_space").
func jsonFieldName(field *reflect.StructField) string {
	tag := field.Tag.Get("json")
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Pointer:
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
