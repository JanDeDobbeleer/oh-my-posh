package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	expected := "expected"
	values := map[Property]interface{}{TextProperty: expected}
	properties := properties{
		values: values,
	}
	value := properties.getString(TextProperty, "err")
	assert.Equal(t, value, expected)
}

func TestGetStringNoEntry(t *testing.T) {
	expected := "expected"
	values := map[Property]interface{}{}
	properties := properties{
		values: values,
	}
	value := properties.getString(TextProperty, expected)
	assert.Equal(t, value, expected)
}

func TestGetStringNoTextEntry(t *testing.T) {
	expected := "expected"
	values := map[Property]interface{}{TextProperty: true}
	properties := properties{
		values: values,
	}
	value := properties.getString(TextProperty, expected)
	assert.Equal(t, value, expected)
}

func TestGetColor(t *testing.T) {
	expected := "#123456"
	values := map[Property]interface{}{UserColor: expected}
	properties := properties{
		values: values,
	}
	value := properties.getColor(UserColor, "#789123")
	assert.Equal(t, value, expected)
}

func TestDefaultColorWithInvalidColorCode(t *testing.T) {
	expected := "#123456"
	values := map[Property]interface{}{UserColor: "invalid"}
	properties := properties{
		values: values,
	}
	value := properties.getColor(UserColor, expected)
	assert.Equal(t, value, expected)
}

func TestDefaultColorWithUnavailableProperty(t *testing.T) {
	expected := "#123456"
	values := map[Property]interface{}{}
	properties := properties{
		values: values,
	}
	value := properties.getColor(UserColor, expected)
	assert.Equal(t, value, expected)
}

func TestGetBool(t *testing.T) {
	expected := true
	values := map[Property]interface{}{DisplayComputer: expected}
	properties := properties{
		values: values,
	}
	value := properties.getBool(DisplayComputer, false)
	assert.True(t, value)
}

func TestGetBoolPropertyNotInMap(t *testing.T) {
	values := map[Property]interface{}{}
	properties := properties{
		values: values,
	}
	value := properties.getBool(DisplayComputer, false)
	assert.False(t, value)
}

func TestGetBoolInvalidProperty(t *testing.T) {
	values := map[Property]interface{}{DisplayComputer: "borked"}
	properties := properties{
		values: values,
	}
	value := properties.getBool(DisplayComputer, false)
	assert.False(t, value)
}
