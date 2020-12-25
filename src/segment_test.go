package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	cwd = "Projects/oh-my-posh3"
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
				"prefix": " \uE5FF ",
				"style": "folder",
				"ignore_folders": [
					"/super/secret/project"
				]
			}
		}
		`
	segment := &Segment{}
	err := json.Unmarshal([]byte(segmentJSON), segment)
	assert.NoError(t, err)
	cwd := "/super/secret/project"
	got := segment.shouldIgnoreFolder(cwd)
	assert.True(t, got)
}

func TestShouldIgnoreFolderRegex(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"Projects[\\/].*"},
		},
	}
	got := segment.shouldIgnoreFolder(cwd)
	assert.True(t, got)
}

func TestShouldIgnoreFolderRegexNonEscaped(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"Projects/.*"},
		},
	}
	got := segment.shouldIgnoreFolder(cwd)
	assert.True(t, got)
}

func TestShouldIgnoreFolderRegexInverted(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"(?!Projects[\\/]).*"},
		},
	}
	got := segment.shouldIgnoreFolder(cwd)
	assert.False(t, got)
}

func TestShouldIgnoreFolderRegexInvertedNonEscaped(t *testing.T) {
	segment := &Segment{
		Properties: map[Property]interface{}{
			IgnoreFolders: []string{"(?!Projects/).*"},
		},
	}
	got := segment.shouldIgnoreFolder(cwd)
	assert.False(t, got)
}
