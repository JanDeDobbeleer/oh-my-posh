package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	expected      = "expected"
	expectedColor = "#768954"

	Foo Property = "color"
)

func TestGetString(t *testing.T) {
	var properties properties = properties{Foo: expected}
	value := properties.getString(Foo, "err")
	assert.Equal(t, expected, value)
}

func TestGetStringNoEntry(t *testing.T) {
	var properties properties = properties{}
	value := properties.getString(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetStringNoTextEntry(t *testing.T) {
	var properties properties = properties{Foo: true}
	value := properties.getString(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetHexColor(t *testing.T) {
	expected := expectedColor
	var properties properties = properties{Foo: expected}
	value := properties.getColor(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestGetColor(t *testing.T) {
	expected := "yellow"
	var properties properties = properties{Foo: expected}
	value := properties.getColor(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithInvalidColorCode(t *testing.T) {
	expected := expectedColor
	var properties properties = properties{Foo: "invalid"}
	value := properties.getColor(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithUnavailableProperty(t *testing.T) {
	expected := expectedColor
	var properties properties = properties{}
	value := properties.getColor(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetPaletteColor(t *testing.T) {
	expected := "p:red"
	var properties properties = properties{Foo: expected}
	value := properties.getColor(Foo, "white")
	assert.Equal(t, expected, value)
}

func TestGetBool(t *testing.T) {
	expected := true
	var properties properties = properties{Foo: expected}
	value := properties.getBool(Foo, false)
	assert.True(t, value)
}

func TestGetBoolPropertyNotInMap(t *testing.T) {
	var properties properties = properties{}
	value := properties.getBool(Foo, false)
	assert.False(t, value)
}

func TestGetBoolInvalidProperty(t *testing.T) {
	var properties properties = properties{Foo: "borked"}
	value := properties.getBool(Foo, false)
	assert.False(t, value)
}

func TestGetFloat64(t *testing.T) {
	expected := float64(1337)
	var properties properties = properties{Foo: expected}
	value := properties.getFloat64(Foo, 9001)
	assert.Equal(t, expected, value)
}

func TestGetFloat64PropertyNotInMap(t *testing.T) {
	expected := float64(1337)
	var properties properties = properties{}
	value := properties.getFloat64(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidStringProperty(t *testing.T) {
	expected := float64(1337)
	var properties properties = properties{ThresholdProperty: "invalid"}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidBoolProperty(t *testing.T) {
	expected := float64(1337)
	var properties properties = properties{ThresholdProperty: true}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}
