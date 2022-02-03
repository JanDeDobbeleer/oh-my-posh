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
	var properties Map = Map{Foo: expected}
	value := properties.GetString(Foo, "err")
	assert.Equal(t, expected, value)
}

func TestGetStringNoEntry(t *testing.T) {
	var properties Map = Map{}
	value := properties.GetString(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetStringNoTextEntry(t *testing.T) {
	var properties Map = Map{Foo: true}
	value := properties.GetString(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetHexColor(t *testing.T) {
	expected := expectedColor
	var properties Map = Map{Foo: expected}
	value := properties.GetColor(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestGetColor(t *testing.T) {
	expected := "yellow"
	var properties Map = Map{Foo: expected}
	value := properties.GetColor(Foo, "#789123")
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithInvalidColorCode(t *testing.T) {
	expected := expectedColor
	var properties Map = Map{Foo: "invalid"}
	value := properties.GetColor(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestDefaultColorWithUnavailableProperty(t *testing.T) {
	expected := expectedColor
	var properties Map = Map{}
	value := properties.GetColor(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetPaletteColor(t *testing.T) {
	expected := "p:red"
	var properties Map = Map{Foo: expected}
	value := properties.GetColor(Foo, "white")
	assert.Equal(t, expected, value)
}

func TestGetBool(t *testing.T) {
	expected := true
	var properties Map = Map{Foo: expected}
	value := properties.GetBool(Foo, false)
	assert.True(t, value)
}

func TestGetBoolPropertyNotInMap(t *testing.T) {
	var properties Map = Map{}
	value := properties.GetBool(Foo, false)
	assert.False(t, value)
}

func TestGetBoolInvalidProperty(t *testing.T) {
	var properties Map = Map{Foo: "borked"}
	value := properties.GetBool(Foo, false)
	assert.False(t, value)
}

func TestGetFloat64(t *testing.T) {
	expected := float64(1337)
	var properties Map = Map{Foo: expected}
	value := properties.GetFloat64(Foo, 9001)
	assert.Equal(t, expected, value)
}

func TestGetFloat64PropertyNotInMap(t *testing.T) {
	expected := float64(1337)
	var properties Map = Map{}
	value := properties.GetFloat64(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidStringProperty(t *testing.T) {
	expected := float64(1337)
	var properties Map = Map{Foo: "invalid"}
	value := properties.GetFloat64(Foo, expected)
	assert.Equal(t, expected, value)
}

func TestGetFloat64InvalidBoolProperty(t *testing.T) {
	expected := float64(1337)
	var properties Map = Map{Foo: true}
	value := properties.GetFloat64(Foo, expected)
	assert.Equal(t, expected, value)
}
