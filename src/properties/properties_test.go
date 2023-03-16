package properties

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
	var properties = Map{Foo: expected}
	value := properties.GetString(Foo, "err")
	assert.Equal(t, expected, value)
}

func TestGetStringNoEntry(t *testing.T) {
	var properties = Map{}
	value := properties.GetString(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetStringNoTextEntry(t *testing.T) {
	var properties = Map{Foo: true}
	value := properties.GetString(Foo, expected)
	assert.Equal(t, "true", value)
}

func TestGetHexColor(t *testing.T) {
	expected := expectedColor
	var properties = Map{Foo: expected}
	value := properties.GetColor(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestGetColor(t *testing.T) {
	expected := "yellow"
	var properties = Map{Foo: expected}
	value := properties.GetColor(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithInvalidColorCode(t *testing.T) {
	expected := expectedColor
	var properties = Map{Foo: "invalid"}
	value := properties.GetColor(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithUnavailableProperty(t *testing.T) {
	expected := expectedColor
	var properties = Map{}
	value := properties.GetColor(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetPaletteColor(t *testing.T) {
	expected := "p:red"
	var properties = Map{Foo: expected}
	value := properties.GetColor(Foo, "white")
	assert.Equal(t, expected, value)
}

func TestGetBool(t *testing.T) {
	expected := true
	var properties = Map{Foo: expected}
	value := properties.GetBool(Foo, false)
	assert.True(t, value)
}

func TestGetBoolPropertyNotInMap(t *testing.T) {
	var properties = Map{}
	value := properties.GetBool(Foo, false)
	assert.False(t, value)
}

func TestGetBoolInvalidProperty(t *testing.T) {
	var properties = Map{Foo: "borked"}
	value := properties.GetBool(Foo, false)
	assert.False(t, value)
}

func TestGetFloat64(t *testing.T) {
	cases := []struct {
		Case     string
		Expected float64
		Input    interface{}
	}{
		{Case: "int", Expected: 1337, Input: 1337},
		{Case: "float64", Expected: 1337, Input: float64(1337)},
		{Case: "uint64", Expected: 1337, Input: uint64(1337)},
		{Case: "int64", Expected: 1337, Input: int64(1337)},
		{Case: "string", Expected: 9001, Input: "invalid"},
		{Case: "bool", Expected: 9001, Input: true},
	}
	for _, tc := range cases {
		properties := Map{Foo: tc.Input}
		value := properties.GetFloat64(Foo, 9001)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestGetFloat64PropertyNotInMap(t *testing.T) {
	expected := float64(1337)
	var properties = Map{}
	value := properties.GetFloat64(Foo, expected)
	assert.Equal(t, expected, value)
}
