package template

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Text struct {
	context  Data
	template string
}

// New returns a Text instance from the pool with the given template and context
func get(template string, context any) *Text {
	if textPool == nil {
		// Fallback if pool is not initialized yet
		return &Text{context: context, template: template}
	}

	text := textPool.Get()
	text.template = template
	text.context = context

	return text
}

func Render(template string, context any) (string, error) {
	t := get(template, context)
	defer t.release()

	if !strings.Contains(t.template, "{{") || !strings.Contains(t.template, "}}") {
		return t.template, nil
	}

	renderer := renderPool.Get()
	defer renderer.release()

	return renderer.execute(t)
}

// Release resets the Text instance and returns it to the pool
func (t *Text) release() {
	t.context = nil
	t.template = ""

	if textPool != nil {
		textPool.Put(t)
	}
}

func (t *Text) patchTemplate() {
	fields := &fields{}
	fields.init(t.context)

	var result, property strings.Builder
	var inProperty, inTemplate bool
	for i, char := range t.template {
		// define start or end of template
		if !inTemplate && char == '{' {
			if i-1 >= 0 && rune(t.template[i-1]) == '{' {
				inTemplate = true
			}
		} else if inTemplate && char == '}' {
			if i-1 >= 0 && rune(t.template[i-1]) == '}' {
				inTemplate = false
			}
		}

		if !inTemplate {
			result.WriteRune(char)
			continue
		}

		switch char {
		case '.':
			var lastChar rune
			rs := result.String()
			if len(rs) > 0 {
				lastChar = rune(rs[len(rs)-1])
			}
			// only replace if we're in a valid property start
			// with a space, { or ( character
			switch lastChar {
			case ' ', '{', '(':
				property.WriteRune(char)
				inProperty = true
			default:
				result.WriteRune(char)
			}
		case ' ', '}', ')': // space or }
			if !inProperty {
				result.WriteRune(char)
				continue
			}

			prop := property.String()
			switch {
			case strings.HasPrefix(prop, ".Segments") && !strings.HasSuffix(prop, ".Contains"):
				// as we can't provide a clean way to access the list
				// of segments, we need to replace the property with
				// the list of segments so they can be accessed directly
				parts := strings.Split(prop, ".")
				if len(parts) > 3 {
					fmt.Fprintf(&result, `(.Segments.MustGet "%s").%s`, parts[2], strings.Join(parts[3:], "."))
				} else {
					fmt.Fprintf(&result, `(.Segments.MustGet "%s")`, parts[2])
				}
				// property = strings.Replace(property, ".Segments", ".Segments.ToSimple", 1)
				// result += property
			case strings.HasPrefix(prop, ".Env."):
				// we need to replace the property with the getEnv function
				// so we can access the environment variables directly
				fmt.Fprintf(&result, `(call .Getenv "%s")`, strings.TrimPrefix(prop, ".Env."))
			default:
				// check if we have the same property in Data
				// and replace it with the Data property so it
				// can take precedence
				if fields.hasField(prop) {
					result.WriteString(".Data")
				}

				// remove the global reference so we can use it directly
				result.WriteString(strings.TrimPrefix(prop, globalRef))
			}

			property.Reset()
			result.WriteRune(char)
			inProperty = false
		default:
			if inProperty {
				property.WriteRune(char)
				continue
			}
			result.WriteRune(char)
		}
	}

	// return the result and remaining unresolved property
	t.template = result.String() + property.String()

	log.Debug(t.template)
}

// fieldSet is an immutable set of exported field/method names for a struct type.
// Once built and stored in knownFields it is never mutated.
type fieldSet map[string]bool

// fields holds the resolved set for the current render context.
// For struct types the set is shared from knownFields (immutable, no copy needed).
// For map types it is built locally and not shared.
type fields struct {
	values fieldSet
}

// initFromType recursively builds and caches the field set for a struct type.
// It may be called concurrently for the same type; LoadOrStore ensures only one
// result is shared.
func initFromType(typ reflect.Type) fieldSet {
	if cached, ok := knownFields.Load(typ); ok {
		return cached.(fieldSet)
	}

	set := make(fieldSet)

	// Get struct fields and check embedded types
	for field := range typ.Fields() {
		if r, _ := utf8.DecodeRuneInString(field.Name); unicode.IsUpper(r) {
			set[field.Name] = true
		}

		// If this is an embedded field, merge its fields recursively
		if !field.Anonymous {
			continue
		}

		embedded := field.Type
		if embedded.Kind() == reflect.Pointer {
			embedded = embedded.Elem()
		}

		if embedded.Kind() == reflect.Struct {
			for k := range initFromType(embedded) {
				set[k] = true
			}
		}
	}

	// Get pointer methods
	ptrType := reflect.PointerTo(typ)
	for method := range ptrType.Methods() {
		name := method.Name
		if r, _ := utf8.DecodeRuneInString(name); unicode.IsUpper(r) {
			set[name] = true
		}
	}

	// Store atomically; if another goroutine won the race, discard ours and use theirs.
	actual, _ := knownFields.LoadOrStore(typ, set)
	return actual.(fieldSet)
}

func (f *fields) init(data any) {
	if data == nil {
		return
	}

	val := reflect.TypeOf(data)
	switch val.Kind() {
	case reflect.Struct:
		// Shared immutable set — no copy needed.
		f.values = initFromType(val)
	case reflect.Map:
		m, ok := data.(map[string]any)
		if !ok {
			return
		}
		set := make(fieldSet, len(m))
		for key := range m {
			if r, _ := utf8.DecodeRuneInString(key); unicode.IsUpper(r) {
				set[key] = true
			}
		}
		f.values = set
	case reflect.Pointer:
		v := reflect.ValueOf(data)
		if v.IsNil() {
			return
		}

		f.init(v.Elem().Interface())
	default:
	}
}

func (f *fields) hasField(field string) bool {
	if f.values == nil {
		return false
	}

	field = strings.TrimPrefix(field, ".")

	// get the first part of the field
	field, _, _ = strings.Cut(field, ".")

	_, ok := f.values[field]
	return ok
}
