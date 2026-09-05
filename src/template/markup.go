package template

import (
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Markup carries terminal markup that may reach the prompt writer verbatim:
// the writer interprets <...> anchors (colors, styles, hyperlinks), so only
// text originating from user configuration or an intentional markup
// constructor may flow through this type. Everything a template action
// evaluates to is escaped by the renderer unless it is a Markup.
//
// A string kind, not a struct: text/template treats every struct as true, so
// a struct would break the `{{ if .Field }}` guards themes use on fields of
// this type, and `eq`/`ne` compare string kinds by value.
type Markup string

func (m Markup) String() string {
	return string(m)
}

// RawMarkup marks user-authored configuration text (a template render result,
// an icon option, a mapped value) as trusted markup. Never call it with data
// read from the filesystem, a VCS, or the network.
func RawMarkup(s string) Markup {
	return Markup(s)
}

// EscapeMarkup converts untrusted data into markup-safe text: chevrons are
// replaced so the writer prints them literally instead of parsing an anchor.
func EscapeMarkup(s string) Markup {
	return Markup(EscapeText(s))
}

// JoinMarkup concatenates markup fragments, preserving each fragment's trust.
func JoinMarkup(parts ...Markup) Markup {
	size := 0
	for _, part := range parts {
		size += len(part)
	}

	var sb strings.Builder
	sb.Grow(size)

	for _, part := range parts {
		sb.WriteString(string(part))
	}

	return Markup(sb.String())
}

// EscapeText replaces chevrons so the writer renders them literally. The
// escape sequences (<<>, <>>) are the writer's own quoted form of < and >.
func EscapeText(s string) string {
	return chevronReplacer.Replace(s)
}

var chevronReplacer = strings.NewReplacer("<", "<<>", ">", "<>>")

// The session cache stores segment data as map[string]any (interface values)
// and gob-encodes it per process; an unregistered concrete type inside an
// interface fails to encode and the whole segment entry is silently dropped.
func init() {
	gob.Register(Markup(""))
}

const noValue = "<no value>"

var (
	markupType = reflect.TypeFor[Markup]()
	errorType  = reflect.TypeFor[error]()
)

// markupAware adapts a template function so Markup values can pass through it.
// text/template refuses a Markup where a function expects a string, which
// would make every string function (contains, replace, trimSuffix, ...) fail
// on a field of that type. The wrapper feeds the string parameters the
// markup's text instead. When the function returns a string and any argument
// was Markup, the result is Markup again so the anchors survive; the plain
// string arguments of such a call are escaped first, since they would
// otherwise ride into the trusted result unchecked.
//
// Functions that neither take nor return strings are returned untouched.
func markupAware(name string, fn any) any {
	if fast, ok := markupAwareTyped(fn); ok {
		return fast
	}

	fv := reflect.ValueOf(fn)
	ft := fv.Type()

	if ft.Kind() != reflect.Func || !touchesStrings(ft) {
		return fn
	}

	numIn := ft.NumIn()
	variadic := ft.IsVariadic()
	returnsString := ft.NumOut() > 0 && ft.Out(0).Kind() == reflect.String && ft.Out(0) != markupType
	returnsError := ft.NumOut() == 2 && ft.Out(1) == errorType

	paramType := func(i int) reflect.Type {
		if variadic && i >= numIn-1 {
			return ft.In(numIn - 1).Elem()
		}

		return ft.In(i)
	}

	return func(args ...any) (any, error) {
		if (variadic && len(args) < numIn-1) || (!variadic && len(args) != numIn) {
			return nil, fmt.Errorf("wrong number of args for %s: want %d got %d", name, numIn, len(args))
		}

		fromMarkup := false
		for _, arg := range args {
			if _, ok := arg.(Markup); ok {
				fromMarkup = true
				break
			}
		}

		escapeStrings := fromMarkup && returnsString

		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			param := paramType(i)

			switch v := arg.(type) {
			case Markup:
				if param.Kind() == reflect.String && param != markupType {
					arg = string(v)
				}
			case string:
				if escapeStrings {
					arg = EscapeText(v)
				}
			}

			value, err := argValue(name, arg, param)
			if err != nil {
				return nil, err
			}

			in[i] = value
		}

		out := fv.Call(in)

		if returnsError && !out[1].IsNil() {
			return nil, out[1].Interface().(error)
		}

		if len(out) == 0 {
			return nil, nil
		}

		if returnsString && fromMarkup {
			return Markup(out[0].String()), nil
		}

		return out[0].Interface(), nil
	}
}

// markupAwareTyped covers the signatures most theme templates call (upper,
// trunc, replace, contains, trimSuffix, ...) without reflection: reflect.Call
// costs several allocations per invocation, which adds up in a prompt that
// pipes every segment through a function or two. The semantics are those of
// markupAware.
func markupAwareTyped(fn any) (any, bool) {
	switch f := fn.(type) {
	case func(string) string:
		return func(a any) (any, error) {
			args, markup, err := textArgs(a)
			if err != nil {
				return nil, err
			}

			return markupResult(f(args[0]), markup), nil
		}, true
	case func(string, string) string:
		return func(a, b any) (any, error) {
			args, markup, err := textArgs(a, b)
			if err != nil {
				return nil, err
			}

			return markupResult(f(args[0], args[1]), markup), nil
		}, true
	case func(string, string, string) string:
		return func(a, b, c any) (any, error) {
			args, markup, err := textArgs(a, b, c)
			if err != nil {
				return nil, err
			}

			return markupResult(f(args[0], args[1], args[2]), markup), nil
		}, true
	case func(string, string) bool:
		return func(a, b any) (any, error) {
			args, _, err := textArgs(a, b)
			if err != nil {
				return nil, err
			}

			return f(args[0], args[1]), nil
		}, true
	case func(any, string) string:
		return func(a, b any) (any, error) {
			args, markup, err := textArgs(b)
			if err != nil {
				return nil, err
			}

			return markupResult(f(a, args[0]), markup), nil
		}, true
	case func(string, ...any) string:
		return func(format any, values ...any) (any, error) {
			args, markup, err := textArgs(format)
			if err != nil {
				return nil, err
			}

			markup = escapeMixedValues(values, markup)

			return markupResult(f(args[0], values...), markup), nil
		}, true
	case func(...any) string:
		return func(values ...any) (any, error) {
			markup := escapeMixedValues(values, false)

			return markupResult(f(values...), markup), nil
		}, true
	default:
		return nil, false
	}
}

// escapeMixedValues applies the markupAware rule to a print-style argument
// list: when any value (or the seen flag) is Markup, the plain strings among
// them are escaped in place. Markup values are left as they are, fmt prints
// them through String.
func escapeMixedValues(values []any, seen bool) bool {
	markup := seen
	for _, v := range values {
		if _, ok := v.(Markup); ok {
			markup = true
			break
		}
	}

	if !markup {
		return false
	}

	for i, v := range values {
		if s, ok := v.(string); ok {
			values[i] = EscapeText(s)
		}
	}

	return true
}

// textArgs converts string parameters the way markupAware does: Markup
// arguments become their text and, when any argument was Markup, the plain
// ones are escaped. The returned flag reports whether a Markup was seen.
func textArgs(values ...any) ([3]string, bool, error) {
	var args [3]string
	var isMarkup [3]bool

	markup := false

	for i, v := range values {
		switch s := v.(type) {
		case Markup:
			args[i] = string(s)
			isMarkup[i] = true
			markup = true
		case string:
			args[i] = s
		default:
			return args, false, fmt.Errorf("expected string, got %T", v)
		}
	}

	if !markup {
		return args, false, nil
	}

	for i := range values {
		if !isMarkup[i] {
			args[i] = EscapeText(args[i])
		}
	}

	return args, true, nil
}

func markupResult(s string, markup bool) any {
	if markup {
		return Markup(s)
	}

	return s
}

func touchesStrings(ft reflect.Type) bool {
	if ft.NumOut() > 0 && ft.Out(0).Kind() == reflect.String {
		return true
	}

	for i := range ft.NumIn() {
		param := ft.In(i)
		if ft.IsVariadic() && i == ft.NumIn()-1 {
			param = param.Elem()
		}

		if param.Kind() == reflect.String {
			return true
		}
	}

	return false
}

// argValue mirrors the coercions text/template applies to a typed parameter
// before the wrapper hid the real signature behind a variadic any one.
func argValue(name string, arg any, param reflect.Type) (reflect.Value, error) {
	if arg == nil {
		switch param.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
			return reflect.Zero(param), nil
		default:
			return reflect.Value{}, fmt.Errorf("invalid value; expected %s for %s", param, name)
		}
	}

	value := reflect.ValueOf(arg)

	switch {
	case value.Type().AssignableTo(param):
		return value, nil
	case value.Kind() == reflect.Pointer && value.Elem().Type().AssignableTo(param):
		return value.Elem(), nil
	case isNumber(value.Kind()) && isNumber(param.Kind()):
		// a template literal reaches the wrapper as int or float64, whatever
		// the parameter's exact numeric type
		return value.Convert(param), nil
	default:
		return reflect.Value{}, fmt.Errorf("wrong type for value; expected %s; got %s for %s", param, value.Type(), name)
	}
}

func isNumber(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// escapeActionValue is appended to every print action's pipeline after parsing
// (see parsedTemplate), making action output safe for the writer: literal
// template text may carry markup, action results may not unless typed Markup.
func escapeActionValue(v any) (string, error) {
	// plain strings dominate segment output, so skip the reflection below
	switch m := v.(type) {
	case Markup:
		return string(m), nil
	case string:
		return EscapeText(m), nil
	case *Markup:
		if m == nil {
			return noValue, nil
		}

		return string(*m), nil
	}

	// text/template indirects pointers and interfaces before printing
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return noValue, nil
		}

		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return noValue, nil
	}

	if rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
		return "", errors.New("unprintable value")
	}

	return EscapeText(fmt.Sprint(rv.Interface())), nil
}
