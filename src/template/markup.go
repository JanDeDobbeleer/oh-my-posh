package template

import (
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/sprig/v3"
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

// markupStringFunc adapts a unary string-transforming template function so
// Markup inputs stay Markup (their anchors survive the transformation) while
// plain string inputs behave exactly as before.
func markupStringFunc(fn func(string) string) func(any) (any, error) {
	return func(v any) (any, error) {
		switch s := v.(type) {
		case Markup:
			return Markup(fn(string(s))), nil
		case string:
			return fn(s), nil
		default:
			return nil, fmt.Errorf("expected string, got %T", v)
		}
	}
}

// markupStringFuncs returns Markup-preserving overrides for the unary sprig
// string functions that make sense on display text. Functions taking extra
// arguments (indent, replace, trimall, quote, ...) are not adapted: applied
// to a Markup value they error, which is visible rather than silently wrong.
func markupStringFuncs() map[string]any {
	wrapped := make(map[string]any)

	for _, name := range []string{
		"lower", "upper", "title", "untitle", "swapcase", "trim",
		"nospace", "snakecase", "kebabcase", "camelcase", "shuffle",
	} {
		fn, ok := sprig.TxtFuncMap()[name].(func(string) string)
		if !ok {
			continue
		}

		wrapped[name] = markupStringFunc(fn)
	}

	// trunc and TruncE are local (they take an any length); adapt them so a
	// Markup branch name or path survives truncation with anchors intact.
	wrapped["trunc"] = func(length any, v any) any {
		if m, ok := v.(Markup); ok {
			return Markup(trunc(length, string(m)))
		}

		return trunc(length, fmt.Sprint(v))
	}

	wrapped["truncE"] = func(length any, v any) any {
		if m, ok := v.(Markup); ok {
			return Markup(TruncE(length, string(m)))
		}

		return TruncE(length, fmt.Sprint(v))
	}

	return wrapped
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
