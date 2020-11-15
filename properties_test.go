package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	expected = "expected"
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
