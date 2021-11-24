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
	values := map[Property]interface{}{TextProperty: expected}
	properties := properties{
		values: values,
	}
	value := properties.getString(TextProperty, "err")
	assert.Equal(t, expected, value)
}

func TestGetStringNoEntry(t *testing.T) {
	values := map[Property]interface{}{}
	properties := properties{
		values: values,
	}
	value := properties.getString(TextProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetStringNoTextEntry(t *testing.T) {
	values := map[Property]interface{}{TextProperty: true}
	properties := properties{
		values: values,
	}
	value := properties.getString(TextProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetHexColor(t *testing.T) {
	expected := expectedColor
	values := map[Property]interface{}{UserColor: expected}
	properties := properties{
		values: values,
	}
	value := properties.getColor(UserColor, "#789123")
	assert.Equal(t, expected, value)
}

func TestGetColor(t *testing.T) {
	expected := "yellow"
	values := map[Property]interface{}{UserColor: expected}
	properties := properties{
		values: values,
	}
	value := properties.getColor(UserColor, "#789123")
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithInvalidColorCode(t *testing.T) {
	expected := expectedColor
	values := map[Property]interface{}{UserColor: "invalid"}
	properties := properties{
		values: values,
	}
	value := properties.getColor(UserColor, expected)
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithUnavailableProperty(t *testing.T) {
	expected := expectedColor
	values := map[Property]interface{}{}
	properties := properties{
		values: values,
	}
	value := properties.getColor(UserColor, expected)
	assert.Equal(t, expected, value)
}

func TestGetPaletteColor(t *testing.T) {
	expected := "p:red"
	values := map[Property]interface{}{Background: expected}
	properties := properties{
		values: values,
	}
	value := properties.getColor(Background, "white")
	assert.Equal(t, expected, value)
}

func TestGetBool(t *testing.T) {
	expected := true
	values := map[Property]interface{}{DisplayHost: expected}
	properties := properties{
		values: values,
	}
	value := properties.getBool(DisplayHost, false)
	assert.True(t, value)
}

func TestGetBoolPropertyNotInMap(t *testing.T) {
	values := map[Property]interface{}{}
	properties := properties{
		values: values,
	}
	value := properties.getBool(DisplayHost, false)
	assert.False(t, value)
}

func TestGetBoolInvalidProperty(t *testing.T) {
	values := map[Property]interface{}{DisplayHost: "borked"}
	properties := properties{
		values: values,
	}
	value := properties.getBool(DisplayHost, false)
	assert.False(t, value)
}

func TestGetFloat64(t *testing.T) {
	expected := float64(1337)
	values := map[Property]interface{}{"myfloat": expected}
	properties := properties{
		values: values,
	}
	value := properties.getFloat64("myfloat", 9001)
	assert.Equal(t, expected, value)
}

func TestGetFloat64PropertyNotInMap(t *testing.T) {
	expected := float64(1337)
	values := map[Property]interface{}{}
	properties := properties{
		values: values,
	}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidStringProperty(t *testing.T) {
	expected := float64(1337)
	values := map[Property]interface{}{ThresholdProperty: "invalid"}
	properties := properties{
		values: values,
	}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidBoolProperty(t *testing.T) {
	expected := float64(1337)
	values := map[Property]interface{}{ThresholdProperty: true}
	properties := properties{
		values: values,
	}
	value := properties.getFloat64(ThresholdProperty, expected)
	assert.Equal(t, expected, value)
}
