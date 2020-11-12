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
	r := regexp.MustCompile(`(?P<color>#[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})`)
	match := r.FindStringSubmatch(colorString)
	if match != nil && match[0] != "" {
		return match[0]
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
