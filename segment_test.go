package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSegmentWriterCanMap(t *testing.T) {
	sc := &Segment{
		Type: Session,
	}
	env := new(MockedEnvironment)
	props, err := sc.mapSegmentWithWriter(env)
	assert.NotNil(t, props)
	assert.NoError(t, err)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(MockedEnvironment)
	props, err := sc.mapSegmentWithWriter(env)
	assert.Nil(t, props)
	assert.Error(t, err)
}
