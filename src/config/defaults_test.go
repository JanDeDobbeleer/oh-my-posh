package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplyDefaults_String(t *testing.T) {
	type testStruct struct {
		Field1 string `default:"hello"`
		Field2 string `default:"world"`
		Field3 string // no default
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.Equal(t, "hello", s.Field1)
	assert.Equal(t, "world", s.Field2)
	assert.Equal(t, "", s.Field3)
}

func TestApplyDefaults_Bool(t *testing.T) {
	type testStruct struct {
		Field1 bool `default:"true"`
		Field2 bool `default:"false"`
		Field3 bool // no default
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.True(t, s.Field1)
	assert.False(t, s.Field2)
	assert.False(t, s.Field3)
}

func TestApplyDefaults_Int(t *testing.T) {
	type testStruct struct {
		Field1 int   `default:"42"`
		Field2 int64 `default:"9999"`
		Field3 int   // no default
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.Equal(t, 42, s.Field1)
	assert.Equal(t, int64(9999), s.Field2)
	assert.Equal(t, 0, s.Field3)
}

func TestApplyDefaults_Float(t *testing.T) {
	type testStruct struct {
		Field1 float64 `default:"3.14"`
		Field2 float32 `default:"2.71"`
		Field3 float64 // no default
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.Equal(t, 3.14, s.Field1)
	// Use small delta for float32 comparison due to precision limitations
	assert.InDelta(t, 2.71, s.Field2, 0.0001)
	assert.Equal(t, 0.0, s.Field3)
}

func TestApplyDefaults_Duration(t *testing.T) {
	type testStruct struct {
		Field1 time.Duration `default:"5s"`
		Field2 time.Duration `default:"100ms"`
		Field3 time.Duration // no default
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, s.Field1)
	assert.Equal(t, 100*time.Millisecond, s.Field2)
	assert.Equal(t, time.Duration(0), s.Field3)
}

func TestApplyDefaults_Slice(t *testing.T) {
	type testStruct struct {
		Field1 []string `default:"[]"`
		Field2 []string `default:"[\"a\", \"b\", \"c\"]"`
		Field3 []string `default:"x,y,z"`
		Field4 []int    `default:"[1, 2, 3]"`
		Field5 []string // no default
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, s.Field1)
	assert.Equal(t, []string{"a", "b", "c"}, s.Field2)
	assert.Equal(t, []string{"x", "y", "z"}, s.Field3)
	assert.Equal(t, []int{1, 2, 3}, s.Field4)
	assert.Nil(t, s.Field5)
}

func TestApplyDefaults_Map(t *testing.T) {
	type testStruct struct {
		Field1 map[string]string `default:"{}"`
		Field2 map[string]string `default:"{\"key1\": \"val1\", \"key2\": \"val2\"}"`
		Field3 map[string]int    `default:"{\"a\": 1, \"b\": 2}"`
		Field4 map[string]string // no default
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{}, s.Field1)
	assert.Equal(t, map[string]string{"key1": "val1", "key2": "val2"}, s.Field2)
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, s.Field3)
	assert.Nil(t, s.Field4)
}

func TestApplyDefaults_NonZeroSkipped(t *testing.T) {
	type testStruct struct {
		Field1 string `default:"default"`
		Field2 int    `default:"42"`
		Field3 bool   `default:"false"`
	}

	s := &testStruct{
		Field1: "existing",
		Field2: 99,
		Field3: true, // true is non-zero
	}
	err := ApplyDefaults(s)

	assert.NoError(t, err)
	assert.Equal(t, "existing", s.Field1) // not overwritten
	assert.Equal(t, 99, s.Field2)         // not overwritten
	assert.True(t, s.Field3)              // true is non-zero, not overwritten
}

func TestApplyDefaults_EmbeddedStruct(t *testing.T) {
	type Base struct {
		BaseField string `default:"base_value"`
	}

	type Derived struct {
		Base
		DerivedField string `default:"derived_value"`
	}

	d := &Derived{}
	err := ApplyDefaults(d)

	assert.NoError(t, err)
	assert.Equal(t, "base_value", d.BaseField)
	assert.Equal(t, "derived_value", d.DerivedField)
}

func TestApplyDefaults_ErrorNonPointer(t *testing.T) {
	type testStruct struct {
		Field string `default:"value"`
	}

	s := testStruct{}
	err := ApplyDefaults(s)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires a pointer")
}

func TestApplyDefaults_ErrorNonStruct(t *testing.T) {
	s := "string"
	err := ApplyDefaults(&s)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires a pointer to struct")
}

func TestApplyDefaults_InvalidBool(t *testing.T) {
	type testStruct struct {
		Field bool `default:"not_a_bool"`
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid bool default")
}

func TestApplyDefaults_InvalidInt(t *testing.T) {
	type testStruct struct {
		Field int `default:"not_an_int"`
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid int default")
}

func TestApplyDefaults_InvalidFloat(t *testing.T) {
	type testStruct struct {
		Field float64 `default:"not_a_float"`
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid float default")
}

func TestApplyDefaults_InvalidDuration(t *testing.T) {
	type testStruct struct {
		Field time.Duration `default:"not_a_duration"`
	}

	s := &testStruct{}
	err := ApplyDefaults(s)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration default")
}
