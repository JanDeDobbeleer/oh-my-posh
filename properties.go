package main

import (
	"fmt"
	"regexp"
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
	// IgnoreFolders folders to ignore and not run the segment logic
	IgnoreFolders Property = "ignore_folders"
	// DisplayVersion show the version number or not
	DisplayVersion Property = "display_version"
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
	r := regexp.MustCompile(`(?P<color>#[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})`)
	values := groupDict(r, colorString)
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

func parseStringArray(value interface{}) []string {
	expectedValue, ok := value.([]interface{})
	if !ok {
		return []string{}
	}
	list := make([]string, len(expectedValue))
	for i, v := range expectedValue {
		list[i] = fmt.Sprint(v)
	}
	return list
}

func parseKeyValueArray(value interface{}) map[string]string {
	locations, ok := value.([]interface{})
	if !ok {
		return map[string]string{}
	}
	keyValueArray := make(map[string]string)
	for _, s := range locations {
		l := parseStringArray(s)
		if len(l) == 2 {
			key := l[0]
			val := l[1]
			keyValueArray[key] = val
		}
	}
	return keyValueArray
}
