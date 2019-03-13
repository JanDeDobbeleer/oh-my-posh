package main

import "regexp"

//Property defines one property of a segment for context
type Property string

//general Properties used across Segments
const (
	//Style indicates with style to use
	Style Property = "style"
	//Prefix adds a text prefix to the segment
	Prefix Property = "prefix"
	//Postfix adds a text postfix to the segment
	Postfix Property = "postfix"
	//ColorBackground color the background or foreground when a specific color is set
	ColorBackground Property = "color_background"
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
