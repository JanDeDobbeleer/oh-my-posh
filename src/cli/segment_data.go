package cli

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
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
//
// The methods a struct field carries in its own right are the exception, and they are why this
// returns two trees rather than one. battery.State is an int, and every battery theme switches on
// `.State.String`; a map holding the number it marshals to has nothing under that name. Its method
// results have to go somewhere, and they cannot go where the number is - a file whose "State" is
// an object no longer unmarshals into the writer, which is the other half of what a data file is
// for. So the number stays in data, the methods go in a parallel methods tree shaped like it, and
// whoever renders without a writer merges the second over the first (config.MergeRecordedMethods).
func recordSegmentData(writer any) (data, methods []byte, err error) {
	recorded, err := asDataMap(reflect.ValueOf(writer), 0)
	if err != nil {
		return nil, nil, err
	}

	data, err = json.Marshal(recorded.data)
	if err != nil {
		return nil, nil, err
	}

	if recorded.methods == nil {
		return data, nil, nil
	}

	methods, err = json.Marshal(recorded.methods)
	if err != nil {
		return nil, nil, err
	}

	return data, methods, nil
}

// recorded is one value's two trees: what it marshals to, and the method results that have no
// place in that marshalling. methods is nil for everything that needs no overlay, which is most of
// what a writer holds.
type recorded struct {
	data    any
	methods any
}

func asDataMap(value reflect.Value, depth int) (recorded, error) {
	if !value.IsValid() {
		return recorded{}, nil
	}

	structValue := reflect.Indirect(value)

	// Anything that is not a plain struct - a string, a number, a slice, or a type with its own
	// MarshalJSON such as time.Time - is recorded as it marshals. Walking its fields would either
	// find none or take apart something that knows how to represent itself.
	if structValue.Kind() != reflect.Struct || marshalsItself(value) {
		raw, err := json.Marshal(value.Interface())
		if err != nil {
			return recorded{}, err
		}

		// A type with its own MarshalJSON gets no overlay: that representation is the one templates
		// consume, and time.Time's methods would bury the timestamp `date` parses.
		var methods any
		if overlay := methodResults(value, depth); len(overlay) != 0 {
			methods = overlay
		}

		return recorded{data: json.RawMessage(raw), methods: methods}, nil
	}

	fields := make(map[string]any)
	overlay := make(map[string]any)

	if depth < maxMethodDepth {
		addStructFields(structValue, fields, overlay, depth)
		addMethodResults(value, fields, overlay, depth)
	}

	if len(overlay) == 0 {
		return recorded{data: fields}, nil
	}

	return recorded{data: fields, methods: overlay}, nil
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

// addStructFields records a struct's exported fields under their Go names, and again under a json
// tag that renames them.
//
// The Go name is what a template reads. `{{ .Name }}` resolves against a field called Name; the az
// segment's json tag for it is "name", and wakatime's CumulativeTotal is tagged
// "cumulative_total". A struct bridges the two because encoding/json knows the tag - a map has no
// tags, so a recorded file keyed by tags leaves every such template unable to find anything.
//
// The tag is what a writer reads. encoding/json matches a tagged field by its tag and by nothing
// else, so a file keyed only by Go names restores every plainly-named field and silently skips the
// renamed ones - terraform's Version, tagged "terraform_version", came back nil and took the
// version out of the prompt. Both keys carry the same value, so whichever side reads the file
// finds what it is looking for under the name it knows.
func addStructFields(structValue reflect.Value, fields, overlay map[string]any, depth int) {
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
				addStructFields(embedded, fields, overlay, depth)
				continue
			}
		}

		nested, err := asDataMap(structValue.Field(i), depth+1)
		if err != nil {
			continue
		}

		fields[field.Name] = nested.data

		if name := taggedName(&field); name != "" {
			fields[name] = nested.data
		}

		// The overlay is read only where there is no writer, so it needs the template's name and
		// not the writer's.
		if nested.methods != nil {
			overlay[field.Name] = nested.methods
		}
	}
}

// taggedName is the name encoding/json will look for when it differs from the field's own, and ""
// when the two agree or the field carries no tag at all.
func taggedName(field *reflect.StructField) string {
	tag, tagged := field.Tag.Lookup("json")
	if !tagged {
		return ""
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == field.Name {
		return ""
	}

	return name
}

// methodResults records a value's methods on their own, for the non-struct case that has no
// fields to merge them with. What comes back is the whole of what replaces the value, so each
// method's own overlay is folded in here rather than left for the merge at restore to find - the
// merge stops descending the moment one side is not a map, and a scalar's data is never one.
func methodResults(value reflect.Value, depth int) map[string]any {
	if depth >= maxMethodDepth || marshalsItself(value) {
		return nil
	}

	fields := make(map[string]any)
	overlay := make(map[string]any)

	addMethodResults(value, fields, overlay, depth)

	for name, methods := range overlay {
		fields[name] = config.MergeRecordedMethods(fields[name], methods)
	}

	return fields
}

func addMethodResults(value reflect.Value, fields, overlay map[string]any, depth int) {
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

		fields[name] = nested.data

		if nested.methods != nil {
			overlay[name] = nested.methods
		}
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
