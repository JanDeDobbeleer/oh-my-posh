package template

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Text struct {
	Context  Data
	Template string
}

func (t *Text) Render() (string, error) {
	defer log.Trace(time.Now(), t.Template)

	if !strings.Contains(t.Template, "{{") || !strings.Contains(t.Template, "}}") {
		return t.Template, nil
	}

	t.patchTemplate()

	renderer := renderPool.Get().(*renderer)
	defer renderer.release()

	return renderer.execute(t)
}

func (t *Text) patchTemplate() {
	fields := &fields{}
	fields.init(t.Context)

	var result, property string
	var inProperty, inTemplate bool
	for i, char := range t.Template {
		// define start or end of template
		if !inTemplate && char == '{' {
			if i-1 >= 0 && rune(t.Template[i-1]) == '{' {
				inTemplate = true
			}
		} else if inTemplate && char == '}' {
			if i-1 >= 0 && rune(t.Template[i-1]) == '}' {
				inTemplate = false
			}
		}

		if !inTemplate {
			result += string(char)
			continue
		}

		switch char {
		case '.':
			var lastChar rune
			if len(result) > 0 {
				lastChar = rune(result[len(result)-1])
			}
			// only replace if we're in a valid property start
			// with a space, { or ( character
			switch lastChar {
			case ' ', '{', '(':
				property += string(char)
				inProperty = true
			default:
				result += string(char)
			}
		case ' ', '}', ')': // space or }
			if !inProperty {
				result += string(char)
				continue
			}

			switch {
			case strings.HasPrefix(property, ".Segments") && !strings.HasSuffix(property, ".Contains"):
				// as we can't provide a clean way to access the list
				// of segments, we need to replace the property with
				// the list of segments so they can be accessed directly
				parts := strings.Split(property, ".")
				if len(parts) > 3 {
					property = fmt.Sprintf(`(.Segments.MustGet "%s").%s`, parts[2], strings.Join(parts[3:], "."))
				} else {
					property = fmt.Sprintf(`(.Segments.MustGet "%s")`, parts[2])
				}
				result += property
				// property = strings.Replace(property, ".Segments", ".Segments.ToSimple", 1)
				// result += property
			case strings.HasPrefix(property, ".Env."):
				// we need to replace the property with the getEnv function
				// so we can access the environment variables directly
				property = strings.TrimPrefix(property, ".Env.")
				result += fmt.Sprintf(`(call .Getenv "%s")`, property)
			default:
				// check if we have the same property in Data
				// and replace it with the Data property so it
				// can take precedence
				if fields.hasField(property) {
					property = ".Data" + property
				}

				// remove the global reference so we can use it directly
				property = strings.TrimPrefix(property, globalRef)
				result += property
			}

			property = ""
			result += string(char)
			inProperty = false
		default:
			if inProperty {
				property += string(char)
				continue
			}
			result += string(char)
		}
	}

	// return the result and remaining unresolved property
	t.Template = result + property

	log.Debug(t.Template)
}

type fields struct {
	values map[string]bool
	sync.RWMutex
}

func (f *fields) init(data any) {
	if data == nil {
		return
	}

	if f.values == nil {
		f.values = make(map[string]bool)
	}

	val := reflect.TypeOf(data)
	switch val.Kind() { //nolint:exhaustive
	case reflect.Struct:
		name := val.Name()

		// check if we already know the fields of this struct
		if kf, OK := knownFields.Load(name); OK {
			f.append(kf)
			return
		}

		// Get struct fields and check embedded types
		fieldsNum := val.NumField()
		for i := range fieldsNum {
			field := val.Field(i)
			f.add(field.Name)

			// If this is an embedded field, get its methods too
			if !field.Anonymous {
				continue
			}

			embeddedType := field.Type

			// Recursively check if the embedded type is also a struct
			if embeddedType.Kind() == reflect.Struct {
				f.init(reflect.New(embeddedType).Elem().Interface())
			}
		}

		// Get pointer methods
		ptrType := reflect.PointerTo(val)
		methodsNum := ptrType.NumMethod()
		for i := range methodsNum {
			f.add(ptrType.Method(i).Name)
		}

		knownFields.Store(name, f)
	case reflect.Map:
		m, ok := data.(map[string]any)
		if !ok {
			return
		}
		for key := range m {
			f.add(key)
		}
	case reflect.Ptr:
		f.init(reflect.ValueOf(data).Elem().Interface())
	}
}

func (f *fields) append(values any) {
	if values == nil {
		return
	}

	fields, ok := values.(*fields)
	if !ok {
		return
	}

	f.Lock()
	fields.RLock()

	defer func() {
		f.Unlock()
		fields.RUnlock()
	}()

	for key := range fields.values {
		f.values[key] = true
	}
}

func (f *fields) add(field string) {
	if len(field) == 0 {
		return
	}

	r := []rune(field)[0]
	if !unicode.IsUpper(r) {
		return
	}

	f.Lock()
	defer f.Unlock()

	f.values[field] = true
}

func (f *fields) hasField(field string) bool {
	field = strings.TrimPrefix(field, ".")

	// get the first part of the field
	splitted := strings.Split(field, ".")
	field = splitted[0]

	f.RLock()
	defer f.RUnlock()

	_, ok := f.values[field]
	return ok
}
