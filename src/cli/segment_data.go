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

	structValue := reflect.Indirect(value)

	// Anything that is not a plain struct - a string, a number, a slice, or a type with its own
	// MarshalJSON such as time.Time - is recorded as it marshals. Walking its fields would either
	// find none or take apart something that knows how to represent itself.
	if structValue.Kind() != reflect.Struct || marshalsItself(value) {
		raw, err := json.Marshal(value.Interface())
		if err != nil {
			return nil, err
		}

		return json.RawMessage(raw), nil
	}

	fields := make(map[string]any)

	if depth < maxMethodDepth {
		addStructFields(structValue, fields, depth)
		addMethodResults(value, fields, depth)
	}

	return fields, nil
}

// marshalsItself reports whether a type defines its own JSON representation, in which case taking
// it apart field by field would produce something that does not round-trip - time.Time being the
// one that matters here.
func marshalsItself(value reflect.Value) bool {
	if !value.CanInterface() {
		return false
	}

	_, ok := value.Interface().(json.Marshaler)

	return ok
}

// addStructFields records a struct's exported fields keyed by their Go names rather than their
// json tags.
//
// This is what a template reads. `{{ .Name }}` resolves against a field called Name; the az
// segment's json tag for it is "name", and wakatime's CumulativeTotal is tagged
// "cumulative_total". A struct bridges the two because encoding/json knows the tag - a map has no
// tags, so a recorded file keyed by tags leaves every such template unable to find anything. Since
// the whole point of recording is to be replayed where no struct exists, the keys have to be the
// names the templates use.
func addStructFields(structValue reflect.Value, fields map[string]any, depth int) {
	for i := range structValue.NumField() {
		field := structValue.Type().Field(i)

		if !field.IsExported() {
			continue
		}

		// A field the writer explicitly keeps out of JSON stays out: those hold internal state
		// (buffers, handles) rather than anything a prompt renders.
		if tag, tagged := field.Tag.Lookup("json"); tagged && strings.HasPrefix(tag, "-") {
			continue
		}

		if field.Anonymous {
			// An embedded struct's fields belong to the outer one as far as a template is
			// concerned, so they are flattened rather than nested under the type name.
			embedded := reflect.Indirect(structValue.Field(i))
			if embedded.Kind() == reflect.Struct {
				addStructFields(embedded, fields, depth)
				continue
			}
		}

		nested, err := asDataMap(structValue.Field(i), depth+1)
		if err != nil {
			continue
		}

		fields[field.Name] = nested
	}
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

func isExported(name string) bool {
	return name != "" && name[0] >= 'A' && name[0] <= 'Z'
}
