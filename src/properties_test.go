package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	expected      = "expected"
	expectedColor = "#768954"
)

func TestGetString(t *testing.T) {
	var properties properties = map[Property]interface{}{TextProperty: expected}
	value := properties.getString(TextProperty, "err")
	assert.Equal(t, expected, value)
}

func TestGetStringNoEntry(t *testing.T) {
	var properties properties = map[Property]interface{}{}
	value := properties.getString(TextProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetStringNoTextEntry(t *testing.T) {
	var properties properties = map[Property]interface{}{TextProperty: true}
	value := properties.getString(TextProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetHexColor(t *testing.T) {
	expected := expectedColor
	var properties properties = map[Property]interface{}{UserColor: expected}
	value := properties.getColor(UserColor, "#789123")
	assert.Equal(t, expected, value)
}

func TestGetColor(t *testing.T) {
	expected := "yellow"
	var properties properties = map[Property]interface{}{UserColor: expected}
	value := properties.getColor(UserColor, "#789123")
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithInvalidColorCode(t *testing.T) {
	expected := expectedColor
	var properties properties = map[Property]interface{}{UserColor: "invalid"}
	value := properties.getColor(UserColor, expected)
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithUnavailableProperty(t *testing.T) {
	expected := expectedColor
	var properties properties = map[Property]interface{}{}
	value := properties.getColor(UserColor, expected)
	assert.Equal(t, expected, value)
}

func TestGetPaletteColor(t *testing.T) {
	expected := "p:red"
	var properties properties = map[Property]interface{}{Background: expected}
	value := properties.getColor(Background, "white")
	assert.Equal(t, expected, value)
}

func TestGetBool(t *testing.T) {
	expected := true
	var properties properties = map[Property]interface{}{DisplayHost: expected}
	value := properties.getBool(DisplayHost, false)
	assert.True(t, value)
}

func TestGetBoolPropertyNotInMap(t *testing.T) {
	var properties properties = map[Property]interface{}{}
	value := properties.getBool(DisplayHost, false)
	assert.False(t, value)
}

func TestGetBoolInvalidProperty(t *testing.T) {
	var properties properties = map[Property]interface{}{DisplayHost: "borked"}
	value := properties.getBool(DisplayHost, false)
	assert.False(t, value)
}

func TestGetFloat64(t *testing.T) {
	expected := float64(1337)
	var properties properties = map[Property]interface{}{"myfloat": expected}
	value := properties.getFloat64("myfloat", 9001)
	assert.Equal(t, expected, value)
}

func TestGetFloat64PropertyNotInMap(t *testing.T) {
	expected := float64(1337)
	var properties properties = map[Property]interface{}{}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidStringProperty(t *testing.T) {
	expected := float64(1337)
	var properties properties = map[Property]interface{}{ThresholdProperty: "invalid"}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidBoolProperty(t *testing.T) {
	expected := float64(1337)
	var properties properties = map[Property]interface{}{ThresholdProperty: true}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}
