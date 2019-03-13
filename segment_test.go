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
	sc.mapSegmentWithWriter(env)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(MockedEnvironment)
	sc.mapSegmentWithWriter(env)
	assert.Nil(t, sc.writer)
}
