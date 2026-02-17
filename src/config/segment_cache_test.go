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

func newCachedTextSegment(env *mock.Environment, alias string, strategy Strategy) *Segment {
	segment := &Segment{
		Type: TEXT,
		Cache: &Cache{
			Strategy: strategy,
			Duration: cache.Duration("10m"),
		},
		Alias: alias,
		env:   env,
	}
	segment.name = alias

	writer := &segments.Text{}
	writer.Init(nil, nil)
	segment.writer = writer

	return segment
}

func TestSegmentCache(t *testing.T) {
	previousTemplateCache := template.Cache
	template.Cache = &cache.Template{
		Segments: maps.NewConcurrent[any](),
	}

	defer func() {
		template.Cache = previousTemplateCache
		cache.DeleteAll(cache.Device)
	}()

	env := new(mock.Environment)
	env.On("Pwd").Return(t.TempDir())

	t.Run("round trip preserves the initialized writer", func(t *testing.T) {
		segment := newCachedTextSegment(env, "my_text_segment", Folder)
		segment.writer.SetText("Hello, Cache!")

		segment.setCache()

		newSegment := newCachedTextSegment(env, "my_text_segment", Folder)
		initializedWriter := newSegment.writer

		restored := newSegment.restoreCache()

		assert.True(t, restored, "cache should be restored")
		// Restoring must overlay the snapshot onto the writer initialized by
		// MapSegmentWithWriter, not replace it: a replacement would drop the
		// writer's runtime state (env, options) and panic on first use.
		assert.Same(t, initializedWriter, newSegment.writer, "restore should reuse the initialized writer")
		assert.Equal(t, "Hello, Cache!", newSegment.writer.Text(), "restored text should match")
	})

	t.Run("cached snapshot is immutable", func(t *testing.T) {
		segment := newCachedTextSegment(env, "immutable_segment", Folder)
		segment.writer.SetText("original")

		segment.setCache()

		// Mutating the writer after caching must not alter the cached snapshot.
		segment.writer.SetText("mutated")

		newSegment := newCachedTextSegment(env, "immutable_segment", Folder)

		assert.True(t, newSegment.restoreCache(), "cache should be restored")
		assert.Equal(t, "original", newSegment.writer.Text(), "cache should hold the state at cache time")
	})

	t.Run("legacy JSON entry is removed", func(t *testing.T) {
		segment := newCachedTextSegment(env, "legacy_segment", Device)

		key, store := segment.cacheKeyAndStore()
		cache.Set(store, key, "legacy_json_string", cache.Duration("10m"))

		assert.False(t, segment.restoreCache(), "legacy cache should not be restored")

		_, found := cache.Get[string](store, key)
		assert.False(t, found, "legacy key should be removed")
	})

	t.Run("unexpected entry type is removed", func(t *testing.T) {
		segment := newCachedTextSegment(env, "unexpected_segment", Device)

		key, store := segment.cacheKeyAndStore()
		cache.Set(store, key, 42, cache.Duration("10m"))

		assert.False(t, segment.restoreCache(), "unexpected cache type should not be restored")

		_, found := cache.Get[int](store, key)
		assert.False(t, found, "unexpected key should be removed")
	})
}
