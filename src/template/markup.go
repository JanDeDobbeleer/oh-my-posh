package template

import (
	"encoding/gob"
	"encoding/json"
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

// markupJSONKey tags a Markup in JSON: {"$markup": "<red>text</>"}. A bare
// string would decode as plain data when a data file lands in a map (the
// website build has no segment writers), and every recorded anchor would
// render as literal text.
const markupJSONKey = "$markup"

func (m Markup) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{markupJSONKey: string(m)})
}

// UnmarshalJSON accepts the tagged form and, for hand-written data files and
// fixtures recorded before the tag existed, a bare string.
func (m *Markup) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}

		*m = Markup(s)

		return nil
	}

	var tagged map[string]string
	if err := json.Unmarshal(b, &tagged); err != nil {
		return err
	}

	text, ok := tagged[markupJSONKey]
	if !ok || len(tagged) != 1 {
		return fmt.Errorf("markup JSON must be an object with only the %q key", markupJSONKey)
	}

	*m = Markup(text)

	return nil
}

// ReviveMarkup walks a generically decoded JSON value and turns every tagged
// markup object (see markupJSONKey) back into a Markup, in place for maps and
// slices. It is what lets a map[string]any carry the same trust as the writer
// struct it was recorded from.
func ReviveMarkup(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		if text, ok := taggedMarkup(typed); ok {
			return text
		}

		for key, nested := range typed {
			typed[key] = ReviveMarkup(nested)
		}

		return typed
	case []any:
		for i, nested := range typed {
			typed[i] = ReviveMarkup(nested)
		}

		return typed
	default:
		return value
	}
}

func taggedMarkup(object map[string]any) (Markup, bool) {
	if len(object) != 1 {
		return "", false
	}

	text, ok := object[markupJSONKey].(string)
	if !ok {
		return "", false
	}

	return Markup(text), true
}

const noValue = "<no value>"

var markupType = reflect.TypeFor[Markup]()

// noPromote lists the functions whose output does not come from their text
// arguments (file contents, command output, decoded bytes, the environment).
// A Markup argument must not turn that output into trusted markup.
var noPromote = map[string]bool{
	"cmd":           true,
	"readFile":      true,
	"stat":          true,
	"glob":          true,
	"env":           true,
	"expandenv":     true,
	"getHostByName": true,
	"b64dec":        true,
	"b32dec":        true,
}

// markupAware adapts a template function so Markup values can pass through it.
// text/template refuses a Markup where a function expects a string, which
// would make every string function (contains, replace, trimSuffix, ...) fail
// on a field of that type. The wrapper hands string parameters the markup's
// text instead.
//
// A string result is promoted back to Markup when the call had a Markup
// argument and every other argument was a plain string, number or bool. The
// plain strings are escaped first, so no data can end up inside the trusted
// result unescaped; a template literal such as a printf format counts as a
// plain string here. Any other argument (a slice, a struct, an error) may hold
// data the wrapper cannot escape, so such a call keeps a plain result, which
// the renderer escapes on output.
func markupAware(name string, fn any) any {
	if fast, ok := markupAwareTyped(name, fn); ok {
		return fast
	}

	fv := reflect.ValueOf(fn)
	ft := fv.Type()

	if ft.Kind() != reflect.Func || !acceptsMarkup(ft) {
		return fn
	}

	numIn := ft.NumIn()
	variadic := ft.IsVariadic()
	mayPromote := ft.Out(0).Kind() == reflect.String && ft.Out(0) != markupType && !noPromote[name]
	returnsError := ft.NumOut() == 2

	paramType := func(i int) reflect.Type {
		if variadic && i >= numIn-1 {
			return ft.In(numIn - 1).Elem()
		}

		return ft.In(i)
	}

	stringParam := func(i int) bool {
		param := paramType(i)
		return param.Kind() == reflect.String && param != markupType
	}

	return func(args ...any) (any, error) {
		if (variadic && len(args) < numIn-1) || (!variadic && len(args) != numIn) {
			return nil, fmt.Errorf("wrong number of args for %s: got %d", name, len(args))
		}

		promote, err := markupArgs(args, stringParam, mayPromote)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			value, err := argValue(arg, paramType(i))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}

			in[i] = value
		}

		out := fv.Call(in)

		if returnsError && !out[1].IsNil() {
			return nil, out[1].Interface().(error)
		}

		if promote {
			return Markup(out[0].String()), nil
		}

		return out[0].Interface(), nil
	}
}

// markupAwareTyped covers the signatures theme templates call most (upper,
// trunc, replace, contains, trimSuffix, printf, date) without reflection,
// which costs several allocations per call. The rules are those of
// markupAware.
func markupAwareTyped(name string, fn any) (any, bool) {
	allStrings := func(int) bool { return true }
	promotes := !noPromote[name]

	switch f := fn.(type) {
	case func(string) string:
		return func(a any) (any, error) {
			args := [1]any{a}

			promote, err := markupArgs(args[:], allStrings, promotes)
			if err != nil {
				return nil, err
			}

			strs, err := stringArgs(args[:])
			if err != nil {
				return nil, err
			}

			return markupResult(f(strs[0]), promote), nil
		}, true
	case func(string, string) string:
		return func(a, b any) (any, error) {
			args := [2]any{a, b}

			promote, err := markupArgs(args[:], allStrings, promotes)
			if err != nil {
				return nil, err
			}

			strs, err := stringArgs(args[:])
			if err != nil {
				return nil, err
			}

			return markupResult(f(strs[0], strs[1]), promote), nil
		}, true
	case func(string, string, string) string:
		return func(a, b, c any) (any, error) {
			args := [3]any{a, b, c}

			promote, err := markupArgs(args[:], allStrings, promotes)
			if err != nil {
				return nil, err
			}

			strs, err := stringArgs(args[:])
			if err != nil {
				return nil, err
			}

			return markupResult(f(strs[0], strs[1], strs[2]), promote), nil
		}, true
	case func(string, string) bool:
		return func(a, b any) (any, error) {
			args := [2]any{a, b}

			if _, err := markupArgs(args[:], allStrings, false); err != nil {
				return nil, err
			}

			strs, err := stringArgs(args[:])
			if err != nil {
				return nil, err
			}

			return f(strs[0], strs[1]), nil
		}, true
	case func(any, string) string:
		return func(a, b any) (any, error) {
			args := [2]any{a, b}

			promote, err := markupArgs(args[:], func(i int) bool { return i == 1 }, promotes)
			if err != nil {
				return nil, err
			}

			s, err := stringArg(args[1])
			if err != nil {
				return nil, err
			}

			return markupResult(f(args[0], s), promote), nil
		}, true
	case func(string, any) Markup:
		return func(a, b any) (any, error) {
			args := [1]any{a}

			if _, err := markupArgs(args[:], allStrings, false); err != nil {
				return nil, err
			}

			s, err := stringArg(args[0])
			if err != nil {
				return nil, err
			}

			return f(s, b), nil
		}, true
	case func(string, ...any) string:
		return func(format any, values ...any) (any, error) {
			args := make([]any, 0, len(values)+1)
			args = append(args, format)
			args = append(args, values...)

			promote, err := markupArgs(args, func(i int) bool { return i == 0 }, promotes)
			if err != nil {
				return nil, err
			}

			s, err := stringArg(args[0])
			if err != nil {
				return nil, err
			}

			return markupResult(f(s, args[1:]...), promote), nil
		}, true
	case func(...any) string:
		return func(values ...any) (any, error) {
			promote, err := markupArgs(values, func(int) bool { return false }, promotes)
			if err != nil {
				return nil, err
			}

			return markupResult(f(values...), promote), nil
		}, true
	default:
		return nil, false
	}
}

// acceptsMarkup reports whether a Markup could ever reach the function: only
// string and interface parameters can carry one.
func acceptsMarkup(ft reflect.Type) bool {
	for i := range ft.NumIn() {
		param := ft.In(i)
		if ft.IsVariadic() && i == ft.NumIn()-1 {
			param = param.Elem()
		}

		if param.Kind() == reflect.Interface || (param.Kind() == reflect.String && param != markupType) {
			return true
		}
	}

	return false
}

// markupArgs rewrites the arguments of one call in place and reports whether
// its result must become Markup. A Markup argument becomes its text where the
// parameter is a string; a *string is dereferenced the way text/template
// does. When the result is promoted, every plain string argument is escaped.
func markupArgs(args []any, stringParam func(int) bool, mayPromote bool) (bool, error) {
	markup, opaque := false, false

	for i, arg := range args {
		if p, ok := arg.(*string); ok {
			if p == nil {
				return false, errors.New("nil pointer evaluating string")
			}

			arg = *p
			args[i] = arg
		}

		switch {
		case isMarkup(arg):
			markup = true
		case !isScalar(arg):
			opaque = true
		}
	}

	promote := mayPromote && markup && !opaque

	for i, arg := range args {
		switch v := arg.(type) {
		case Markup:
			if stringParam(i) {
				args[i] = string(v)
			}
		case string:
			if promote {
				args[i] = EscapeText(v)
			}
		}
	}

	return promote, nil
}

func isMarkup(v any) bool {
	_, ok := v.(Markup)
	return ok
}

func isScalar(v any) bool {
	switch reflect.ValueOf(v).Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func stringArg(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", v)
	}

	return s, nil
}

func stringArgs(args []any) ([3]string, error) {
	var strs [3]string

	for i, arg := range args {
		s, err := stringArg(arg)
		if err != nil {
			return strs, err
		}

		strs[i] = s
	}

	return strs, nil
}

func markupResult(s string, promote bool) any {
	if promote {
		return Markup(s)
	}

	return s
}

// argValue checks and converts an argument the way text/template would
// against the real parameter type; the wrapper accepts anything, so that work
// happens here.
func argValue(arg any, param reflect.Type) (reflect.Value, error) {
	if arg == nil {
		switch param.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
			return reflect.Zero(param), nil
		default:
			return reflect.Value{}, fmt.Errorf("invalid value; expected %s", param)
		}
	}

	value := reflect.ValueOf(arg)

	switch {
	case value.Type().AssignableTo(param):
		return value, nil
	case value.Kind() == reflect.Pointer && value.Elem().Type().AssignableTo(param):
		return value.Elem(), nil
	case value.CanInt() && isUint(param.Kind()):
		if value.Int() < 0 {
			return reflect.Value{}, fmt.Errorf("negative value for unsigned %s", param)
		}

		return value.Convert(param), nil
	case value.CanInt() && (isInt(param.Kind()) || isFloat(param.Kind())):
		// a template literal reaches the wrapper as int, whatever the
		// parameter's exact numeric type
		return value.Convert(param), nil
	case value.CanFloat() && isFloat(param.Kind()):
		return value.Convert(param), nil
	default:
		return reflect.Value{}, fmt.Errorf("wrong type for value; expected %s; got %s", param, value.Type())
	}
}

func isInt(kind reflect.Kind) bool {
	return kind >= reflect.Int && kind <= reflect.Int64
}

func isUint(kind reflect.Kind) bool {
	return kind >= reflect.Uint && kind <= reflect.Uint64
}

func isFloat(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
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
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() || (rv.Kind() == reflect.Pointer && rv.IsNil()) {
		return noValue, nil
	}

	// text/template prints through String and Error when a value has them,
	// including on pointer receivers, before it follows the pointer.
	switch m := v.(type) {
	case fmt.Stringer:
		return EscapeText(m.String()), nil
	case error:
		return EscapeText(m.Error()), nil
	}

	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Chan || rv.Kind() == reflect.Func {
		return "", errors.New("unprintable value")
	}

	return EscapeText(fmt.Sprint(rv.Interface())), nil
}

// escapeUntrustedActionValue is the untrusted renderer's output escape. An
// untrusted template (folder names are one, see the path segment) must not be
// able to produce trusted markup, so a Markup result is escaped like data.
func escapeUntrustedActionValue(v any) (string, error) {
	if m, ok := v.(Markup); ok {
		return EscapeText(string(m)), nil
	}

	return escapeActionValue(v)
}
