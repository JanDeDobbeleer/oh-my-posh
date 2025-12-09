package options

import (
	"encoding/gob"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/generics"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

func init() {
	gob.Register([]any{})
	gob.Register(map[string]any{})
	gob.Register(map[any]any{})
	gob.Register([]string{})
	gob.Register(map[string]string{})
	gob.Register([]int{})
	gob.Register([]float64{})
	gob.Register([]bool{})
	gob.Register(int64(0))
	gob.Register(uint64(0))
	gob.Register(float32(0))
	gob.Register(Map{})
	gob.Register((*Option)(nil))
	gob.Register(map[Option]any{})
}

type Provider interface {
	Color(option Option, defaultValue color.Ansi) color.Ansi
	Bool(option Option, defaultValue bool) bool
	String(option Option, defaultValue string) string
	Float64(option Option, defaultValue float64) float64
	Int(option Option, defaultValue int) int
	KeyValueMap(option Option, defaultValue map[string]string) map[string]string
	StringArray(option Option, defaultValue []string) []string
	Any(option Option, defaultValue any) any
}

// Option defines one property of a segment for context
type Option string

// general options used across Segments
const (
	// Style indicates the style to use
	Style Option = "style"
	// FetchVersion decides whether to fetch the version number or not
	FetchVersion Option = "fetch_version"
	// AlwaysEnabled decides whether or not to always display the info
	AlwaysEnabled Option = "always_enabled"
	// VersionURLTemplate is the template to use when building language segment hyperlink
	VersionURLTemplate Option = "version_url_template"
	// DisplayError decides whether to display when an error occurs or not
	DisplayError Option = "display_error"
	// DisplayDefault hides or shows the default
	DisplayDefault Option = "display_default"
	// AccessToken is the access token to use for an API
	AccessToken Option = "access_token"
	// RefreshToken is the refresh token to use for an API
	RefreshToken Option = "refresh_token"
	// HTTPTimeout timeout used when executing http request
	HTTPTimeout Option = "http_timeout"
	// DefaultHTTPTimeout default timeout used when executing http request
	DefaultHTTPTimeout = 20
	// Files to trigger the segment on
	Files Option = "files"
	// Duration of the cache
	CacheDuration Option = "cache_duration"
)

type Map map[Option]any

func (m Map) String(option Option, defaultValue string) string {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %s", option, defaultValue))
		return defaultValue
	}
	value := fmt.Sprint(val)
	log.Debug(fmt.Sprintf("%s: %s", option, value))
	return value
}

func (m Map) Color(option Option, defaultValue color.Ansi) color.Ansi {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %s", option, defaultValue))
		return defaultValue
	}

	colorString := color.Ansi(fmt.Sprint(val))
	if color.IsAnsiColorName(colorString) {
		log.Debug(fmt.Sprintf("%s: %s", option, colorString))
		return colorString
	}

	values := regex.FindNamedRegexMatch(`(?P<color>#[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3}|p:.*)`, colorString.String())
	if values != nil && values["color"] != "" {
		value := color.Ansi(values["color"])
		log.Debug(fmt.Sprintf("%s: %s", option, value))
		return value
	}

	log.Debug(fmt.Sprintf("%s: %s", option, defaultValue))
	return defaultValue
}

func (m Map) Bool(option Option, defaultValue bool) bool {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %t", option, defaultValue))
		return defaultValue
	}
	boolValue, ok := val.(bool)
	if !ok {
		log.Debug(fmt.Sprintf("%s: %t", option, defaultValue))
		return defaultValue
	}
	log.Debug(fmt.Sprintf("%s: %t", option, boolValue))
	return boolValue
}

func (m Map) Float64(option Option, defaultValue float64) float64 {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %f", option, defaultValue))
		return defaultValue
	}

	// Direct type conversions for common numeric types
	switch v := val.(type) {
	case float64:
		log.Debug(fmt.Sprintf("%s: %f", option, v))
		return v
	case int:
		value := float64(v)
		log.Debug(fmt.Sprintf("%s: %f", option, value))
		return value
	case int64:
		value := float64(v)
		log.Debug(fmt.Sprintf("%s: %f", option, value))
		return value
	case uint64:
		value := float64(v)
		log.Debug(fmt.Sprintf("%s: %f", option, value))
		return value
	default:
		log.Debug(fmt.Sprintf("%s: %f", option, defaultValue))
		return defaultValue
	}
}

func (m Map) Int(option Option, defaultValue int) int {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %d", option, defaultValue))
		return defaultValue
	}

	// Direct type conversions for common numeric types
	switch v := val.(type) {
	case int:
		log.Debug(fmt.Sprintf("%s: %d", option, v))
		return v
	case int64:
		value := int(v)
		log.Debug(fmt.Sprintf("%s: %d", option, value))
		return value
	case uint64:
		value := int(v)
		log.Debug(fmt.Sprintf("%s: %d", option, value))
		return value
	case float64:
		value := int(v)
		log.Debug(fmt.Sprintf("%s: %d", option, value))
		return value
	default:
		log.Debug(fmt.Sprintf("%s: %d", option, defaultValue))
		return defaultValue
	}
}

func (m Map) KeyValueMap(option Option, defaultValue map[string]string) map[string]string {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %v", option, defaultValue))
		return defaultValue
	}

	keyValues := parseKeyValueArray(val)
	log.Debug(fmt.Sprintf("%s: %v", option, keyValues))
	return keyValues
}

func (m Map) StringArray(option Option, defaultValue []string) []string {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %v", option, defaultValue))
		return defaultValue
	}

	keyValues := ParseStringArray(val)
	log.Debug(fmt.Sprintf("%s: %v", option, keyValues))
	return keyValues
}

func (m Map) Any(option Option, defaultValue any) any {
	val, found := m[option]
	if !found {
		log.Debug(fmt.Sprintf("%s: %v", option, defaultValue))
		return defaultValue
	}

	log.Debug(fmt.Sprintf("%s: %v", option, val))
	return val
}

func ParseStringArray(param any) []string {
	return generics.ParseStringSlice(param)
}

func parseKeyValueArray(param any) map[string]string {
	switch v := param.(type) {
	default:
		return map[string]string{}
	case map[any]any:
		keyValueArray := make(map[string]string)
		for key, value := range v {
			val := value.(string)
			keyString := fmt.Sprintf("%v", key)
			keyValueArray[keyString] = val
		}
		return keyValueArray
	case map[string]any:
		keyValueArray := make(map[string]string)
		for key, value := range v {
			val := value.(string)
			keyValueArray[key] = val
		}
		return keyValueArray
	case []any:
		keyValueArray := make(map[string]string)
		for _, s := range v {
			l := ParseStringArray(s)
			if len(l) == 2 {
				key := l[0]
				val := l[1]
				keyValueArray[key] = val
			}
		}
		return keyValueArray
	case Map:
		keyValueArray := make(map[string]string)
		for key, value := range v {
			val := value.(string)
			keyString := fmt.Sprintf("%v", key)
			keyValueArray[keyString] = val
		}
		return keyValueArray
	case map[string]string:
		return v
	}
}

// Generic functions

type Value interface {
	string | int | []string | float64 | bool
}

func OneOf[T Value](options Provider, defaultValue T, props ...Option) T {
	for _, prop := range props {
		// get value on a generic get, then see if we can cast to T?
		val := options.Any(prop, nil)
		if val == nil {
			continue
		}

		if v, ok := val.(T); ok {
			return v
		}
	}

	return defaultValue
}
