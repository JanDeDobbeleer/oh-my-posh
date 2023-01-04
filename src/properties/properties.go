package properties

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/ansi"
	"github.com/jandedobbeleer/oh-my-posh/regex"
)

type Properties interface {
	GetColor(property Property, defaultColor string) string
	GetBool(property Property, defaultValue bool) bool
	GetString(property Property, defaultValue string) string
	GetFloat64(property Property, defaultValue float64) float64
	GetInt(property Property, defaultValue int) int
	GetKeyValueMap(property Property, defaultValue map[string]string) map[string]string
	GetStringArray(property Property, defaultValue []string) []string
}

// Property defines one property of a segment for context
type Property string

// general Properties used across Segments
const (
	// Style indicates the style to use
	Style Property = "style"
	// IncludeFolders indicates folders to be included for the segment logic
	IncludeFolders Property = "include_folders"
	// ExcludeFolders indicates folders to be excluded for the segment logic
	ExcludeFolders Property = "exclude_folders"
	// IgnoreFolders is a duplicate of ExcludeFolders
	IgnoreFolders Property = "ignore_folders"
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
	// DefaultCacheTimeout default timeout used when caching data
	DefaultCacheTimeout = 10
	// CacheTimeout cache timeout
	CacheTimeout Property = "cache_timeout"
)

type Map map[Property]interface{}

func (m Map) GetString(property Property, defaultValue string) string {
	val, found := m[property]
	if !found {
		return defaultValue
	}
	return fmt.Sprint(val)
}

func (m Map) GetColor(property Property, defaultValue string) string {
	val, found := m[property]
	if !found {
		return defaultValue
	}
	colorString := fmt.Sprint(val)
	if ansi.IsAnsiColorName(colorString) {
		return colorString
	}
	values := regex.FindNamedRegexMatch(`(?P<color>#[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3}|p:.*)`, colorString)
	if values != nil && values["color"] != "" {
		return values["color"]
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

	if floatValue, ok := val.(float64); ok {
		return floatValue
	}

	// config parser parses an int
	intValue, ok := val.(int)
	if !ok {
		return defaultValue
	}

	return float64(intValue)
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

func ParseStringArray(param interface{}) []string {
	switch v := param.(type) {
	default:
		return []string{}
	case []interface{}:
		list := make([]string, len(v))
		for i, v := range v {
			list[i] = fmt.Sprint(v)
		}
		return list
	case []string:
		return v
	}
}

func parseKeyValueArray(param interface{}) map[string]string {
	switch v := param.(type) {
	default:
		return map[string]string{}
	case map[interface{}]interface{}:
		keyValueArray := make(map[string]string)
		for key, value := range v {
			val := value.(string)
			keyString := fmt.Sprintf("%v", key)
			keyValueArray[keyString] = val
		}
		return keyValueArray
	case map[string]interface{}:
		keyValueArray := make(map[string]string)
		for key, value := range v {
			val := value.(string)
			keyValueArray[key] = val
		}
		return keyValueArray
	case []interface{}:
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
	case map[string]string:
		return v
	}
}
