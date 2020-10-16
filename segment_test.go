package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapSegmentWriterCanMap(t *testing.T) {
	sc := &Segment{
		Type: Session,
	}
	env := new(MockedEnvironment)
	err := sc.mapSegmentWithWriter(env)
	assert.NotNil(t, sc.props)
	assert.NoError(t, err)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(MockedEnvironment)
	err := sc.mapSegmentWithWriter(env)
	assert.Nil(t, sc.props)
	assert.Error(t, err)
}

func TestParseTestSettings(t *testing.T) {
	segmentJSON :=
		`
		{
			"type": "path",
			"style": "powerline",
			"powerline_symbol": "\uE0B0",
			"foreground": "#ffffff",
			"background": "#61AFEF",
			"properties": {
				"prefix": "  ",
				"style": "folder",
				"ignore_folders": [
					"go-my-psh"
				]
			}
		}
		`
	segment := &Segment{}
	err := json.Unmarshal([]byte(segmentJSON), segment)
	assert.NoError(t, err)
	expected := "go-my-psh"
	got := segment.hasValue(IgnoreFolders, expected)
	assert.True(t, got)
}
