package properties

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

type Properties interface {
	GetColor(property Property, defaultValue color.Ansi) color.Ansi
	GetBool(property Property, defaultValue bool) bool
	GetString(property Property, defaultValue string) string
	GetFloat64(property Property, defaultValue float64) float64
	GetInt(property Property, defaultValue int) int
	GetKeyValueMap(property Property, defaultValue map[string]string) map[string]string
	GetStringArray(property Property, defaultValue []string) []string
	Get(property Property, defaultValue any) any
}

// Property defines one property of a segment for context
type Property string

// general Properties used across Segments
const (
	// Style indicates the style to use
	Style Property = "style"
	// FetchVersion decides whether to fetch the version number or not
	FetchVersion Property = "fetch_version"
	// AlwaysEnabled decides whether or not to always display the info
	AlwaysEnabled Property = "always_enabled"
	// VersionURLTemplate is the template to use when building language segment hyperlink
	VersionURLTemplate Property = "version_url_template"
	// DisplayError decides whether to display when an error occurs or not
	DisplayError Property = "display_error"
	// DisplayDefault hides or shows the default
	DisplayDefault Property = "display_default"
	// AccessToken is the access token to use for an API
	AccessToken Property = "access_token"
	// RefreshToken is the refresh token to use for an API
	RefreshToken Property = "refresh_token"
	// HTTPTimeout timeout used when executing http request
	HTTPTimeout Property = "http_timeout"
	// DefaultHTTPTimeout default timeout used when executing http request
	DefaultHTTPTimeout = 20
	// Files to trigger the segment on
	Files Property = "files"
	// Duration of the cache
	CacheDuration Property = "cache_duration"
)

type Map map[Property]any

func (m Map) GetString(property Property, defaultValue string) string {
	val, found := m[property]
	if !found {
		return defaultValue
	}
	return fmt.Sprint(val)
}

func (m Map) GetColor(property Property, defaultValue color.Ansi) color.Ansi {
	val, found := m[property]
	if !found {
		return defaultValue
	}

	colorString := color.Ansi(fmt.Sprint(val))
	if color.IsAnsiColorName(colorString) {
		return colorString
	}

	values := regex.FindNamedRegexMatch(`(?P<color>#[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3}|p:.*)`, colorString.String())
	if values != nil && values["color"] != "" {
		return color.Ansi(values["color"])
	}

	return defaultValue
}

func (m Map) GetBool(property Property, defaultValue bool) bool {
	val, found := m[property]
	if !found {
		return defaultValue
	}
	boolValue, ok := val.(bool)
	if !ok {
		return defaultValue
	}
	return boolValue
}

func (m Map) GetFloat64(property Property, defaultValue float64) float64 {
	val, found := m[property]
	if !found {
		return defaultValue
	}

	switch v := val.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	case float64:
		return v
	default:
		return defaultValue
	}
}

func (m Map) GetInt(property Property, defaultValue int) int {
	val, found := m[property]
	if !found {
		return defaultValue
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case uint64:
		return int(v)
	case float64:
		intValue, ok := val.(float64)
		if !ok {
			return defaultValue
		}
		return int(intValue)
	default:
		return defaultValue
	}
}

func (m Map) GetKeyValueMap(property Property, defaultValue map[string]string) map[string]string {
	val, found := m[property]
	if !found {
		return defaultValue
	}

	keyValues := parseKeyValueArray(val)

	return keyValues
}

func (m Map) GetStringArray(property Property, defaultValue []string) []string {
	val, found := m[property]
	if !found {
		return defaultValue
	}

	keyValues := ParseStringArray(val)

	return keyValues
}

func (m Map) Get(property Property, defaultValue any) any {
	val, found := m[property]
	if !found {
		return defaultValue
	}

	return val
}

func ParseStringArray(param any) []string {
	switch v := param.(type) {
	default:
		return []string{}
	case []any:
		list := make([]string, len(v))
		for i, v := range v {
			list[i] = fmt.Sprint(v)
		}
		return list
	case []string:
		return v
	}
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

func OneOf[T Value](properties Properties, defaultValue T, props ...Property) T {
	for _, prop := range props {
		// get value on a generic get, then see if we can cast to T?
		val := properties.Get(prop, nil)
		if val == nil {
			continue
		}

		if v, ok := val.(T); ok {
			return v
		}
	}

	return defaultValue
}
