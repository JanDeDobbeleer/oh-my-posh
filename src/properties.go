package main

import (
	"fmt"
)

// Property defines one property of a segment for context
type Property string

// general Properties used across Segments
const (
	// Style indicates with style to use
	Style Property = "style"
	// Prefix adds a text prefix to the segment
	Prefix Property = "prefix"
	// Postfix adds a text postfix to the segment
	Postfix Property = "postfix"
	// ColorBackground color the background or foreground when a specific color is set
	ColorBackground Property = "color_background"
	// IncludeFolders folders to be included for the segment logic
	IncludeFolders Property = "include_folders"
	// ExcludeFolders folders to be excluded for the segment logic
	ExcludeFolders Property = "exclude_folders"
	// IgnoreFolders duplicate of ExcludeFolders
	IgnoreFolders Property = "ignore_folders"
	// DisplayVersion show the version number or not
	DisplayVersion Property = "display_version"
	// AlwaysEnabled decides whether or not to always display the info
	AlwaysEnabled Property = "always_enabled"
	// SegmentTemplate is the template to use to render the information
	SegmentTemplate Property = "template"
	// DisplayError to display when an error occurs or not
	DisplayError Property = "display_error"
	// DisplayDefault hides or shows the default
	DisplayDefault Property = "display_default"
)

type properties struct {
	values     map[Property]interface{}
	foreground string
	background string
}

func (p *properties) getString(property Property, defaultValue string) string {
	if p == nil || p.values == nil {
		return defaultValue
	}
	val, found := p.values[property]
	if !found {
		return defaultValue
	}
	return parseString(val, defaultValue)
}

func parseString(value interface{}, defaultValue string) string {
	stringValue, ok := value.(string)
	if !ok {
		return defaultValue
	}
	return stringValue
}

func (p *properties) getColor(property Property, defaultValue string) string {
	if p == nil || p.values == nil {
		return defaultValue
	}
	val, found := p.values[property]
	if !found {
		return defaultValue
	}
	colorString := parseString(val, defaultValue)
	_, err := getColorFromName(colorString, false)
	if err == nil {
		return colorString
	}
	values := findNamedRegexMatch(`(?P<color>#[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})`, colorString)
	if values != nil && values["color"] != "" {
		return values["color"]
	}
	return defaultValue
}

func (p *properties) getBool(property Property, defaultValue bool) bool {
	if p == nil || p.values == nil {
		return defaultValue
	}
	val, found := p.values[property]
	if !found {
		return defaultValue
	}
	boolValue, ok := val.(bool)
	if !ok {
		return defaultValue
	}
	return boolValue
}

func (p *properties) getFloat64(property Property, defaultValue float64) float64 {
	if p == nil || p.values == nil {
		return defaultValue
	}
	val, found := p.values[property]
	if !found {
		return defaultValue
	}

	floatValue, ok := val.(float64)
	if !ok {
		return defaultValue
	}

	return floatValue
}

func (p *properties) getKeyValueMap(property Property, defaultValue map[string]string) map[string]string {
	if p == nil || p.values == nil {
		return defaultValue
	}
	val, found := p.values[property]
	if !found {
		return defaultValue
	}

	keyValues := parseKeyValueArray(val)

	return keyValues
}

func parseStringArray(param interface{}) []string {
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
			l := parseStringArray(s)
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
