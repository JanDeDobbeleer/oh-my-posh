package config

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/stretchr/testify/assert"
)

func TestSegmentCache(t *testing.T) {
	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}

	env := new(mock.Environment)
	env.On("Pwd").Return("/tmp")

	segment := &Segment{
		Type: TEXT,
		Cache: &Cache{
			Strategy: Folder,
			Duration: cache.Duration("10m"),
		},
		Alias: "my_text_segment",
		env:   env,
	}

	textWriter := &segments.Text{}
	textWriter.Init(nil, nil)
	textWriter.SetText("Hello, Cache!")
	segment.writer = textWriter
	segment.name = "my_text_segment"

	segment.setCache()

	newSegment := &Segment{
		Type: TEXT,
		Cache: &Cache{
			Strategy: Folder,
			Duration: cache.Duration("10m"),
		},
		Alias: "my_text_segment",
		env:   env,
	}
	newSegment.name = "my_text_segment"

	newTextWriter := &segments.Text{}
	newTextWriter.Init(nil, nil)
	newSegment.writer = newTextWriter

	restored := newSegment.restoreCache()

	assert.True(t, restored, "Cache should be restored")
	assert.NotNil(t, newSegment.writer, "Writer should be restored")
	assert.IsType(t, &segments.Text{}, newSegment.writer, "Writer should be of type *segments.Text")
	if newSegment.writer != nil {
		assert.Equal(t, "Hello, Cache!", newSegment.writer.Text(), "Restored text should match")
	}
}
