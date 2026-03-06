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
	defer cache.DeleteAll(cache.Device)

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

	// Test legacy cache (string value)
	legacySegment := &Segment{
		Type: TEXT,
		Cache: &Cache{
			Strategy: Folder,
			Duration: cache.Duration("10m"),
		},
		Alias: "legacy_segment",
		env:   env,
	}
	legacySegment.name = "legacy_segment"

	legacyWriter := &segments.Text{}
	legacyWriter.Init(nil, nil)
	legacySegment.writer = legacyWriter

	legacyKey, legacyStore := legacySegment.cacheKeyAndStore()
	cache.Set(legacyStore, legacyKey, "legacy_json_string", cache.Duration("10m"))

	restoredLegacy := legacySegment.restoreCache()
	assert.False(t, restoredLegacy)

	_, found := cache.Get[string](legacyStore, legacyKey)
	assert.False(t, found)

	// Test snapshot isolation (immutability)
	isoSegment := &Segment{
		Type: TEXT,
		Cache: &Cache{
			Strategy: Folder,
			Duration: cache.Duration("10m"),
		},
		Alias: "iso_segment",
		env:   env,
	}
	isoWriter := &segments.Text{}
	isoWriter.Init(nil, nil)
	isoWriter.SetText("Original")
	isoSegment.writer = isoWriter
	isoSegment.name = "iso_segment"

	isoSegment.setCache()

	// Modify the local writer after caching
	isoWriter.SetText("Modified")

	// Restore into a new segment
	newIsoSegment := &Segment{
		Type: TEXT,
		Cache: &Cache{
			Strategy: Folder,
			Duration: cache.Duration("10m"),
		},
		Alias: "iso_segment",
		env:   env,
	}
	newIsoSegment.name = "iso_segment"
	newIsoSegment.writer = &segments.Text{}
	newIsoSegment.writer.Init(nil, nil)

	newIsoSegment.restoreCache()

	// The restored value should still be "Original"
	assert.Equal(t, "Original", newIsoSegment.writer.Text(), "Restored value should be a copy (Original), not the modified local instance")
}
