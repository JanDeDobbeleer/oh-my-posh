package cli

import (
	"encoding/json"
	"reflect"
	"strings"
)

// interfaceMethods are the SegmentWriter methods every writer has to satisfy. They describe how a
// segment behaves, not what it renders, and none of them is reachable from a template, so
// recording their results would only bloat the file.
var interfaceMethods = map[string]bool{
	"Enabled":  true,
	"Template": true,
	"Init":     true,
	"CacheKey": true,
}

// maxMethodDepth bounds the walk below. Two levels covers every nested access any bundled theme
// or default template makes (`.Working.String`, `.Premium.Percent.Gauge` is the deepest at three
// names, which is two levels of nesting from the writer), and a bound is what keeps a writer that
// holds a reference back to itself from recursing forever.
const maxMethodDepth = 3

// recordSegmentData marshals a segment writer for a data file: its fields, as json.Marshal
// already gave them, plus the results of the exported zero-argument methods a template can call.
//
// A template reaches for both without distinguishing them - `{{ .Working.Added }}` reads a field
// and `{{ .Working.String }}` calls a method, and the template neither knows nor cares which.
// Replaying against the writer struct that difference stays invisible, because the methods come
// back with it. Replaying against plain data it does not: a map can carry a value but not a
// method, so a template calling one would find nothing where a terminal found a string. Recording
// what the method returned closes that gap - Go resolves a name against a map key exactly as it
// resolves it against a method, so `{{ .Working.String }}` reads the recorded result and renders
// identically.
//
// Fields win over methods on a name collision: the field is what json.Marshal chose to call the
// value, and a method shadowing it would change what replay sees.
func recordSegmentData(writer any) ([]byte, error) {
	data, err := asDataMap(reflect.ValueOf(writer), 0)
	if err != nil {
		return nil, err
	}

	return json.Marshal(data)
}

func asDataMap(value reflect.Value, depth int) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}

	raw, err := json.Marshal(value.Interface())
	if err != nil {
		return nil, err
	}

	var fields map[string]any
	// Anything that does not marshal as an object - a slice, a string, a number - has no methods
	// worth walking into and is returned as it stands.
	if err := json.Unmarshal(raw, &fields); err != nil {
		return json.RawMessage(raw), nil
	}

	if depth >= maxMethodDepth {
		return fields, nil
	}

	addMethodResults(value, fields, depth)
	recurseIntoFields(value, fields, depth)

	return fields, nil
}

func addMethodResults(value reflect.Value, fields map[string]any, depth int) {
	for i := range value.NumMethod() {
		name := value.Type().Method(i).Name

		if interfaceMethods[name] || !isExported(name) {
			continue
		}

		if _, taken := fields[name]; taken {
			continue
		}

		method := value.Method(i)
		methodType := method.Type()

		if methodType.NumIn() != 0 || methodType.NumOut() != 1 {
			continue
		}

		result, ok := callRecovered(method)
		if !ok {
			continue
		}

		nested, err := asDataMap(result, depth+1)
		if err != nil {
			continue
		}

		fields[name] = nested
	}
}

// recurseIntoFields walks the struct fields that are themselves structs, so a method one level
// down (`.Working.String`) is recorded too. The json key is what the map is keyed on, so the
// field's own tag decides where the result lands.
func recurseIntoFields(value reflect.Value, fields map[string]any, depth int) {
	structValue := reflect.Indirect(value)
	if structValue.Kind() != reflect.Struct {
		return
	}

	for i := range structValue.NumField() {
		field := structValue.Type().Field(i)

		if !field.IsExported() {
			continue
		}

		key := jsonKey(field)
		if key == "" {
			continue
		}

		if _, present := fields[key]; !present {
			continue
		}

		fieldValue := structValue.Field(i)
		if reflect.Indirect(fieldValue).Kind() != reflect.Struct {
			continue
		}

		if fieldValue.Kind() == reflect.Pointer && fieldValue.IsNil() {
			continue
		}

		nested, err := asDataMap(fieldValue, depth+1)
		if err != nil {
			continue
		}

		fields[key] = nested
	}
}

// callRecovered calls a zero-argument method, answering false if it panics. A writer's methods
// run against the live environment here, the same as they did while the prompt rendered, and one
// that panics on a machine where its own segment was disabled must not take the whole recording
// with it.
func callRecovered(method reflect.Value) (result reflect.Value, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	return method.Call(nil)[0], true
}

func jsonKey(field reflect.StructField) string {
	tag, tagged := field.Tag.Lookup("json")
	if !tagged {
		return field.Name
	}

	name, _, _ := strings.Cut(tag, ",")

	switch name {
	case "-":
		return ""
	case "":
		return field.Name
	default:
		return name
	}
}

func isExported(name string) bool {
	return name != "" && name[0] >= 'A' && name[0] <= 'Z'
}
